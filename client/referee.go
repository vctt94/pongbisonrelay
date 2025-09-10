package client

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"sort"

	"github.com/decred/dcrd/crypto/blake256"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/decred/dcrd/txscript/v4"
	"github.com/decred/dcrd/wire"
	"github.com/vctt94/pong-bisonrelay/pongrpc/grpc/pong"
)

// computePreSig derives k via RFC6979 and enforces even-Y on R' = k*G + T.
// Returns r_x, s', R' (all hex).
func computePreSig(xPrivHex, mHex, TCompHex string) (rXHex string, sPrimeHex string, rCompHex string, err error) {
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

	// Domain separation for nonce derivation:
	// extra = BLAKE256(tag32 || Tcompressed [|| branch/inputID...])
	extra := blake256.Sum256(append(schnorrV0ExtraTag[:], Tb...))
	var version []byte // optional 16-byte version tag; nil is fine

	// Deterministic retry loop: iterate RFC6979 stream until all constraints pass.
	for iter := uint32(0); ; iter++ {
		k := secp256k1.NonceRFC6979(xb, mb, extra[:], version, iter)
		if k == nil || k.IsZero() {
			continue // ultra-rare, but stay deterministic
		}
		kbArr := k.Bytes()

		// R = k·G
		R := secp256k1.PrivKeyFromBytes(kbArr[:]).PubKey()

		// R' = R + T  (retry on infinity)
		Rp, err := addPoints(R, Tpub)
		if err != nil { // point at infinity
			continue
		}
		cp := Rp.SerializeCompressed()
		if len(cp) != 33 {
			continue
		}
		// Enforce even-Y on R' (required by EC-Schnorr-DCRv0 verification).
		if cp[0] != 0x02 {
			continue
		}
		// e = BLAKE256(r_x || m). Retry if e >= n (spec says to try next nonce).
		// https://github.com/decred/dcrd/blob/master/dcrec/secp256k1/schnorr/README.md?plain=1#L257
		var r32 [32]byte
		copy(r32[:], cp[1:33])
		h := blake256.Sum256(append(r32[:], mb...))

		var e secp256k1.ModNScalar
		if overflow := e.SetByteSlice(h[:]); overflow {
			continue // e >= n -> try next deterministic nonce
		}

		// s' = k - e*x  (mod n)
		var ex, sLine, negex secp256k1.ModNScalar
		ex.Set(&e)
		ex.Mul(&x)
		negex.NegateVal(&ex)
		sLine.Set(k) // k is already a *ModNScalar from NonceRFC6979
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

// BuildVerifyOk validates draft and inputs and constructs PreSig list and ack digest.
func BuildVerifyOk(xPrivHex string, req *pong.NeedPreSigs) (*pong.VerifyOk, error) {
	txb, err := hex.DecodeString(req.DraftTxHex)
	if err != nil {
		return nil, err
	}
	var tx wire.MsgTx
	if err := tx.Deserialize(bytes.NewReader(txb)); err != nil {
		return nil, err
	}
	inputs := make([]*pong.NeedPreSigs_PerInput, len(req.Inputs))
	copy(inputs, req.Inputs)
	sort.Slice(inputs, func(i, j int) bool { return inputs[i].InputId < inputs[j].InputId })
	var concat []byte
	concat = append(concat, []byte(req.DraftTxHex)...)
	for _, in := range inputs {
		concat = append(concat, []byte(in.InputId)...)
		concat = append(concat, []byte(in.MHex)...)
		concat = append(concat, in.TCompressed...)
	}
	ack := blake256.Sum256(concat)
	pres := make([]*pong.PreSig, 0, len(inputs))
	for _, in := range inputs {
		redeem, err := hex.DecodeString(in.RedeemScriptHex)
		if err != nil {
			return nil, fmt.Errorf("bad redeem: %w", err)
		}
		idx, err := findInputIndex(&tx, in.InputId)
		if err != nil {
			return nil, err
		}
		mBytes, err := txscript.CalcSignatureHash(redeem, txscript.SigHashAll, &tx, idx, nil)
		if err != nil || len(mBytes) != 32 {
			return nil, fmt.Errorf("sighash")
		}
		mHex := hex.EncodeToString(mBytes)
		if mHex != in.MHex {
			return nil, fmt.Errorf("m mismatch")
		}
		rComp, sLine, err := DeriveAdaptorPreSig(xPrivHex, mHex, hex.EncodeToString(in.TCompressed))
		if err != nil {
			return nil, err
		}
		pres = append(pres, &pong.PreSig{InputId: in.InputId, RLineCompressed: rComp, SLine32: sLine})
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
	return pc.rc.GetFinalizeBundle(ctx, &pong.GetFinalizeBundleRequest{MatchId: matchID, WinnerUid: pc.ID})
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
	return pc.RefOpenEscrow(pc.ID, pubBytes, payoutPubkey, betAtoms, csvBlocks)
}

// StartSettlementHandshake performs HELLO -> REQ -> VERIFY_OK -> SERVER_OK.
func (pc *PongClient) StartSettlementHandshake(ctx context.Context, matchID string) error {
	priv, _, err := pc.RequireSettlementSessionKey()
	if err != nil {
		return err
	}
	stream, err := pc.RefStartSettlementStream(ctx)
	if err != nil {
		return err
	}
	pubHex := pc.settlePubHex
	pubBytes, _ := hex.DecodeString(pubHex)
	_ = stream.Send(&pong.ClientMsg{
		MatchId: matchID,
		Kind: &pong.ClientMsg_Hello{
			Hello: &pong.Hello{
				MatchId:       matchID,
				CompPubkey:    pubBytes,
				ClientVersion: "poc",
			}}})
	for {
		in, err := stream.Recv()
		if err != nil {
			_ = stream.CloseSend()
			return err
		}
		if req := in.GetReq(); req != nil {
			verify, err := BuildVerifyOk(priv, req)
			if err != nil {
				_ = stream.CloseSend()
				return err
			}
			_ = stream.Send(&pong.ClientMsg{
				MatchId: matchID,
				Kind: &pong.ClientMsg_VerifyOk{
					VerifyOk: verify,
				}})
			continue
		}
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
	// TODO: Implement this
	return nil, nil
}
