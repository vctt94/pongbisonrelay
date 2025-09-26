package client

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"sort"
	"time"

	"github.com/decred/dcrd/crypto/blake256"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/decred/dcrd/txscript/v4"
	"github.com/decred/dcrd/wire"
	pongbisonrelay "github.com/vctt94/pongbisonrelay"
	"github.com/vctt94/pongbisonrelay/pongrpc/grpc/pong"
)

// computePreSig derives adaptor pre-signature for (x, m, T).
// Math (DCRv0):
//
//	R  = k·G
//	R' = R + T, require even-Y
//	e  = BLAKE256(r_x || m) mod n
//	s' = k - e·x  (mod n)
//
// Returns hex: (r_x, s', R').
//
// Notes:
//   - k via RFC6979 with domain-sep “extra” binding to T (and optional branch).
//   - Deterministic retry until: k≠0, R'≠∞, even-Y(R'), e<n.
func computePreSig(xPrivHex, mHex, TCompHex string) (rXHex string, sPrimeHex string, rCompHex string, err error) {
	// Inputs: x (scalar), m (32B sighash), T (33B adaptor point).
	xb, err := hex.DecodeString(xPrivHex)
	if err != nil {
		return "", "", "", err
	}
	mb, err := hex.DecodeString(mHex)
	if err != nil || len(mb) != 32 {
		return "", "", "", fmt.Errorf("bad m")
	}
	Tb, err := hex.DecodeString(TCompHex)
	if err != nil {
		return "", "", "", err
	}

	// Validate x and parse T.
	var x secp256k1.ModNScalar
	if overflow := x.SetByteSlice(xb); overflow || x.IsZero() {
		return "", "", "", fmt.Errorf("bad x scalar")
	}
	Tpub, err := secp256k1.ParsePubKey(Tb)
	if err != nil {
		return "", "", "", err
	}

	// RFC6979 domain separation: bind nonce to T (and version if used).
	// extra = BLAKE256(tag32 || T_compressed [|| branch/inputID...])
	extra := blake256.Sum256(append(pongbisonrelay.SchnorrV0ExtraTag[:], Tb...))
	var version []byte // optional 16B tag; nil = unused

	// Deterministic retry loop: enforce all constraints.
	for iter := uint32(0); ; iter++ {
		k := secp256k1.NonceRFC6979(xb, mb, extra[:], version, iter)
		if k == nil || k.IsZero() {
			continue
		}
		kbArr := k.Bytes()

		// R = kG
		R := secp256k1.PrivKeyFromBytes(kbArr[:]).PubKey()

		// R' = R + T ; skip if infinity.
		Rp, err := pongbisonrelay.AddPoints(R, Tpub)
		if err != nil { // infinity
			continue
		}
		cp := Rp.SerializeCompressed()
		if len(cp) != 33 {
			continue
		}

		// Even-Y(R') (DCRv0 normalization): compressed prefix must be 0x02.
		if cp[0] != 0x02 {
			continue
		}

		// e = H(r_x || m). Retry if e ≥ n (per spec).
		// https://github.com/decred/dcrd/blob/master/dcrec/secp256k1/schnorr/README.md?plain=1#L257
		var r32 [32]byte
		copy(r32[:], cp[1:33])
		h := blake256.Sum256(append(r32[:], mb...))
		var e secp256k1.ModNScalar
		if overflow := e.SetByteSlice(h[:]); overflow {
			continue
		}

		// s' = k - e·x (mod n)
		var ex, sLine, negex secp256k1.ModNScalar
		ex.Set(&e)
		ex.Mul(&x)
		negex.NegateVal(&ex)
		sLine.Set(k) // k is already ModNScalar
		sLine.Add(&negex)

		spb := sLine.Bytes()
		return hex.EncodeToString(r32[:]), hex.EncodeToString(spb[:]), hex.EncodeToString(cp), nil
	}
}

// DeriveAdaptorPreSig computes a minus-variant adaptor pre-signature ensuring a
// normalized (even-Y) R' and returns (R' compressed, s'). Uses RFC6979 nonces
// and retries until the derived R' has even Y per EC-Schnorr-DCRv0.
func DeriveAdaptorPreSig(xPrivHex, mHex, TCompHex string) (rPrimeCompressed []byte, sPrime32 []byte, err error) {
	rXHex, sPrimeHex, rCompHex, err := computePreSig(xPrivHex, mHex, TCompHex)
	_ = rXHex // x-coordinate is intentionally unused by callers of the minimal handshake
	if err != nil {
		return nil, nil, err
	}
	rb, err := hex.DecodeString(rCompHex)
	if err != nil {
		return nil, nil, err
	}
	sb, err := hex.DecodeString(sPrimeHex)
	if err != nil {
		return nil, nil, err
	}
	return rb, sb, nil
}

// BuildVerifyOk validates server REQ and produces adaptor pre-sigs.
//
// Math (minus-variant, DCRv0):
//
//	e = BLAKE256(r_x || m) mod n
//	s' = k - e·x
//	R' = k·G + T  (normalized to even-Y in DeriveAdaptorPreSig)
//
// Server check: s'G ?= R' - eX - T
//
// Ack binds what we sign to: draft tx, per-input (InputId, m, T).
func BuildVerifyOk(xPrivHex string, req *pong.NeedPreSigs) (*pong.VerifyOk, error) {
	// Parse draft tx.
	txb, err := hex.DecodeString(req.DraftTxHex)
	if err != nil {
		return nil, err
	}
	var tx wire.MsgTx
	if err := tx.Deserialize(bytes.NewReader(txb)); err != nil {
		return nil, err
	}

	// Canonicalize per-input order for deterministic ack/presigs.
	inputs := make([]*pong.NeedPreSigs_PerInput, len(req.Inputs))
	copy(inputs, req.Inputs)
	sort.Slice(inputs, func(i, j int) bool { return inputs[i].InputId < inputs[j].InputId })

	// Build ack = BLAKE256(draftHex || ⊕_i(InputId || m_i || T_i)).
	var concat []byte
	concat = append(concat, []byte(req.DraftTxHex)...)
	for _, in := range inputs {
		concat = append(concat, []byte(in.InputId)...)
		concat = append(concat, []byte(in.MHex)...)
		concat = append(concat, in.TCompressed...)
	}
	ack := blake256.Sum256(concat)

	// For each input: recompute m, then derive (R', s').
	pres := make([]*pong.PreSig, 0, len(inputs))
	for _, in := range inputs {
		redeem, err := hex.DecodeString(in.RedeemScriptHex)
		if err != nil {
			return nil, fmt.Errorf("bad redeem: %w", err)
		}
		// Map InputId → tx input index.
		idx, err := pongbisonrelay.FindInputIndex(&tx, in.InputId)
		if err != nil {
			return nil, err
		}
		// Recompute Decred SigHashAll(m) and compare with REQ.
		mBytes, err := txscript.CalcSignatureHash(redeem, txscript.SigHashAll, &tx, idx, nil)
		if err != nil || len(mBytes) != 32 {
			return nil, fmt.Errorf("sighash")
		}
		mHex := hex.EncodeToString(mBytes)
		if mHex != in.MHex {
			return nil, fmt.Errorf("m mismatch")
		}

		// Derive adaptor pre-sig using session x and server T:
		//   RFC6979 k, R' = kG + T (even-Y), e = H(r_x||m), s' = k - e·x.
		rComp, sLine, err := DeriveAdaptorPreSig(xPrivHex, mHex, hex.EncodeToString(in.TCompressed))
		if err != nil {
			return nil, err
		}
		pres = append(pres, &pong.PreSig{
			InputId:         in.InputId,
			RLineCompressed: rComp, // R' (33B)
			SLine32:         sLine, // s' (32B)
		})
	}

	return &pong.VerifyOk{AckDigest: ack[:], Presigs: pres}, nil
}

// Escrow-first referee helpers
func (pc *PongClient) RefOpenEscrow(ownerID string, compPub []byte, payoutPubkey []byte, betAtoms uint64, csv uint32) (*pong.OpenEscrowResponse, error) {
	ctx := context.Background()
	return pc.rc.OpenEscrow(ctx, &pong.OpenEscrowRequest{OwnerUid: ownerID, CompPubkey: compPub, PayoutPubkey: payoutPubkey, BetAtoms: betAtoms, CsvBlocks: csv})
}

// RefStartSettlementStream starts the bidirectional settlement stream.
func (pc *PongClient) RefStartSettlementStream(ctx context.Context) (pong.PongReferee_SettlementStreamClient, error) {
	return pc.rc.SettlementStream(ctx)
}

// RefGetFinalizeBundle fetches gamma and both presigs for the winning branch.
func (pc *PongClient) RefGetFinalizeBundle(matchID string) (*pong.GetFinalizeBundleResponse, error) {
	ctx := context.Background()
	return pc.rc.GetFinalizeBundle(ctx, &pong.GetFinalizeBundleRequest{MatchId: matchID, WinnerUid: pc.id})
}

// OpenEscrowWithSession opens an escrow using the cached settlement session pubkey.
func (pc *PongClient) OpenEscrowWithSession(ctx context.Context, payoutPubkey []byte, betAtoms uint64, csvBlocks uint32) (*pong.OpenEscrowResponse, error) {
	pc.RLock()
	pubHex := pc.settlePubHex
	pc.RUnlock()
	if pubHex == "" {
		return nil, fmt.Errorf("no settlement session key; generate one with GenerateNewSettlementSessionKey()")
	}
	pubBytes, err := hex.DecodeString(pubHex)
	if err != nil {
		return nil, fmt.Errorf("bad session pubkey: %w", err)
	}
	return pc.RefOpenEscrow(pc.id, pubBytes, payoutPubkey, betAtoms, csvBlocks)
}

// RefStartSettlementHandshake starts a schnorr adaptor pre-sign handshake:
//
//	C→S HELLO{X}          // publish session X so server can bind/verify pre-sigs
//	S→C REQ{tx, m_i, T_i} // draft, per-input sighash m, adaptor T
//	C→S VERIFY_OK{ack, (R'_i, s'_i)_i} // pre-sigs created with x via RFC6979 nonces
//	S→C OK                // server accepted all (R', s')
func (pc *PongClient) RefStartSettlementHandshake(ctx context.Context, matchID string) error {
	// x := session private scalar used to derive all (R', s') in BuildVerifyOk.
	// We now require the session key to be generated before opening escrow.
	pc.RLock()
	priv := pc.settlePrivHex
	pubHex := pc.settlePubHex
	pc.RUnlock()
	if len(priv) == 0 || len(pubHex) == 0 {
		return fmt.Errorf("no settlement session key present; generate one before presigning")
	}

	// Open stream to referee.
	stream, err := pc.RefStartSettlementStream(ctx)
	if err != nil {
		return fmt.Errorf("settlement stream open failed: %w", err)
	}

	// HELLO — publish session pubkey X (= xG) in compressed form (33B).
	// Server will later verify s'G ?= R' - eX - T using this X.
	pubBytes, _ := hex.DecodeString(pubHex)
	sendStart := time.Now()
	if err := stream.Send(&pong.ClientMsg{
		MatchId: matchID,
		Kind: &pong.ClientMsg_Hello{
			Hello: &pong.Hello{
				MatchId:       matchID,
				CompPubkey:    pubBytes, // X
				ClientVersion: "poc",
			},
		},
	}); err != nil {
		_ = stream.CloseSend()
		return fmt.Errorf("send HELLO failed: %w", err)
	}

	// Drive REQ → VERIFY_OK → OK.
	for {
		in, err := stream.Recv()
		if err != nil {
			_ = stream.CloseSend()
			return fmt.Errorf("recv failed after %.3fs: %w", time.Since(sendStart).Seconds(), err)
		}

		// On REQ: build adaptor pre-sigs bound to the server’s challenge.
		// BuildVerifyOk does:
		//   • Parse draft tx; recompute each m_i = SigHashAll(redeem, tx, idx) and match REQ.m_i.
		//   • ack = BLAKE256(draft || ⊕_i(InputId || m_i || T_i)) to bind tx/input order and T_i.
		//   • For each input i:
		//       - RFC6979 derive k
		//       - R' = kG + T_i, enforce even-Y
		//       - e = H(r_x || m_i), s' = k - e·x
		//       - emit PreSig{RLineCompressed=R', SLine32=s'}.
		if req := in.GetReq(); req != nil {
			verify, err := BuildVerifyOk(priv, req)
			if err != nil {
				_ = stream.CloseSend()
				return fmt.Errorf("BuildVerifyOk failed: %w", err)
			}
			if err := stream.Send(&pong.ClientMsg{
				MatchId: matchID,
				Kind: &pong.ClientMsg_VerifyOk{
					VerifyOk: verify, // {ack, presigs_i=(R'_i, s'_i)}
				},
			}); err != nil {
				_ = stream.CloseSend()
				return fmt.Errorf("send VERIFY_OK failed: %w", err)
			}
			continue
		}

		// On OK: server validated s'G == R' - eX - T for all inputs → handshake complete.
		if ok := in.GetOk(); ok != nil {
			_ = stream.CloseSend()
			return nil
		}
	}
}

// GetEscrowUTXO opens the server WaitFunding stream and returns the first
// non-nil escrow UTXO seen for the given escrow ID. It uses a short timeout.
func (pc *PongClient) GetEscrowUTXO(escrowID string) (*pong.EscrowUTXO, error) {
	if escrowID == "" {
		return nil, fmt.Errorf("escrowID required")
	}
	// The notifier stream should already be running from init; do not start another here.
	if pc.notifier == nil {
		return nil, fmt.Errorf("notification stream not started")
	}
	// Server does not expose a direct unary query for the exact UTXO yet.
	// Caller should listen to the existing ntfn stream or add a dedicated RPC.
	return nil, fmt.Errorf("GetEscrowUTXO unsupported: server does not expose UTXO query")
}
