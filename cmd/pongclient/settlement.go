package main

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/vctt94/pong-bisonrelay/client"

	"github.com/decred/dcrd/chaincfg/chainhash"
	"github.com/decred/dcrd/crypto/blake256"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/decred/dcrd/dcrec/secp256k1/v4/schnorr"
	"github.com/decred/dcrd/txscript/v4"
	"github.com/decred/dcrd/wire"
	"github.com/vctt94/pong-bisonrelay/pongrpc/grpc/pong"
)

// finalizeAndVerify returns the canonical 64-byte EC-Schnorr-DCRv0 signature.
func finalizeAndVerify(m32, rX32, sPrime32, t32, pubCompressed []byte) ([]byte, error) {
	if len(m32) != 32 || len(rX32) != 32 || len(sPrime32) != 32 || len(t32) != 32 {
		return nil, errors.New("bad input sizes")
	}

	// s = s' + t  (mod n)
	var sp, tt, s secp256k1.ModNScalar
	if overflow := sp.SetByteSlice(sPrime32); overflow {
		return nil, errors.New("s' overflow")
	}
	if overflow := tt.SetByteSlice(t32); overflow {
		return nil, errors.New("t overflow")
	}
	s.Set(&sp)
	s.Add(&tt)
	sb := s.Bytes()

	// Build 64-byte sig = r || s and parse via schnorr.
	sigBytes := append(append([]byte{}, rX32...), sb[:]...)
	sig, err := schnorr.ParseSignature(sigBytes)
	if err != nil {
		return nil, err
	}

	// Parse pubkey and verify per EC-Schnorr-DCRv0 (includes even-Y check on computed R).
	pub, err := schnorr.ParsePubKey(pubCompressed)
	if err != nil {
		return nil, err
	}
	if !sig.Verify(m32, pub) {
		return nil, errors.New("bad signature")
	}
	return sig.Serialize(), nil
}

func findInputIndex(tx *wire.MsgTx, inputID string) (int, error) {
	var txid string
	var vout uint32
	if _, err := fmt.Sscanf(inputID, "%s:%d", &txid, &vout); err != nil {
		return -1, fmt.Errorf("bad input_id %q: %w", inputID, err)
	}
	var h chainhash.Hash
	if err := chainhash.Decode(&h, txid); err != nil {
		return -1, fmt.Errorf("bad txid %q: %w", txid, err)
	}
	for i, ti := range tx.TxIn {
		if ti.PreviousOutPoint.Hash == h && ti.PreviousOutPoint.Index == vout {
			return i, nil
		}
	}
	return -1, fmt.Errorf("input %s not found in draft", inputID)
}

// addPoints returns R+S as a *secp256k1.PublicKey using Jacobian add and affine conversion.
func addPoints(R, S *secp256k1.PublicKey) (*secp256k1.PublicKey, error) {
	var rj, sj, sum secp256k1.JacobianPoint
	R.AsJacobian(&rj)
	S.AsJacobian(&sj)

	// sum = rj + sj (Jacobian)
	secp256k1.AddNonConst(&rj, &sj, &sum)

	// Infinity if Z == 0 in Jacobian coords.
	if sum.Z.IsZero() {
		return nil, fmt.Errorf("R' is point at infinity")
	}

	// Convert in place to affine, then build a PublicKey.
	sum.ToAffine()

	var ax, ay secp256k1.FieldVal
	ax.Set(&sum.X)
	ay.Set(&sum.Y)

	return secp256k1.NewPublicKey(&ax, &ay), nil
}

// https://github.com/decred/dcrd/blob/master/dcrec/secp256k1/schnorr/README.md?plain=1#L249
var schnorrV0ExtraTag = func() [32]byte {
	// EC-Schnorr-DCRv0 nonce domain-sep tag from the spec.
	const tagHex = "0b75f97b60e8a5762876c004829ee9b926fa6f0d2eeaec3a4fd1446a768331cb"
	b, _ := hex.DecodeString(tagHex)
	var out [32]byte
	copy(out[:], b)
	return out
}()

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
		var ex, sprime, negex secp256k1.ModNScalar
		ex.Set(&e)
		ex.Mul(&x)
		negex.NegateVal(&ex)
		sprime.Set(k) // k is already a *ModNScalar from NonceRFC6979
		sprime.Add(&negex)

		spb := sprime.Bytes()
		return hex.EncodeToString(r32[:]), hex.EncodeToString(spb[:]), hex.EncodeToString(cp), nil
	}
}

// handleSubmitPreSig triggers the streaming settlement flow (Phase 1 UX).
func (m *appstate) handleSubmitPreSig() tea.Cmd {
	return func() tea.Msg {
		m.startSettlementStream()
		return client.UpdatedMsg{}
	}
}

// startSettlementStream starts listening to the server-driven settlement stream (Phase 1).
func (m *appstate) startSettlementStream() {
	// Ensure key and minimal allocation if needed
	if m.genPrivHex == "" {
		p, _ := secp256k1.GeneratePrivateKey()
		m.genPrivHex = hex.EncodeToString(p.Serialize())
		m.genPubHex = hex.EncodeToString(p.PubKey().SerializeCompressed())
		m.settle.aCompHex = m.genPubHex
	}
	if m.settle.matchID == "" {
		if m.settle.betAtoms == 0 {
			m.settle.betAtoms = 100000000
		}
		if m.settle.csvBlocks == 0 {
			m.settle.csvBlocks = 64
		}
		if res, err := m.pc.RefAllocateEscrow(m.pc.ID, m.settle.aCompHex, m.settle.betAtoms, m.settle.csvBlocks); err == nil {
			m.settle.matchID = res.MatchId
			m.settle.xA = res.XA
			m.settle.xB = res.XB
			// Defer reflecting bet amount until we see deposit in mempool/blocks
		} else {
			m.notification = err.Error()
			m.msgCh <- client.UpdatedMsg{}
			return
		}
	}
	// Start WaitFunding stream to reflect bet when funding appears
	if m.fundingCancel != nil {
		m.fundingCancel()
		m.fundingCancel = nil
	}
	fctx, cancel := context.WithCancel(context.Background())
	m.fundingCancel = cancel
	go func(matchID string) {
		stream, err := m.pc.RefWaitFunding(fctx, matchID)
		if err != nil {
			m.notification = fmt.Sprintf("wait funding error: %v", err)
			m.msgCh <- client.UpdatedMsg{}
			return
		}
		for {
			resp, err := stream.Recv()
			if err != nil {
				m.notification = fmt.Sprintf("funding stream closed: %v", err)
				m.msgCh <- client.UpdatedMsg{}
				return
			}
			if (resp.UtxoA != nil || resp.UtxoB != nil) && m.settle.betAtoms > 0 {
				m.betAmount = float64(m.settle.betAtoms) / 1e8
				m.msgCh <- client.UpdatedMsg{}
			}
			if resp.Confirmed {
				return
			}
		}
	}(m.settle.matchID)
	ctx := m.ctx
	stream, err := m.pc.RefStartSettlementStream(ctx)
	if err != nil {
		m.notification = fmt.Sprintf("stream error: %v", err)
		m.msgCh <- client.UpdatedMsg{}
		return
	}
	pubBytes, _ := hex.DecodeString(m.settle.aCompHex)
	_ = stream.Send(&pong.ClientMsg{MatchId: m.settle.matchID, Kind: &pong.ClientMsg_Hello{Hello: &pong.Hello{CompPubkey: pubBytes, ClientVersion: "poc"}}})
	go func() {
		for {
			in, err := stream.Recv()
			if err != nil {
				m.notification = fmt.Sprintf("settlement stream closed: %v", err)
				m.msgCh <- client.UpdatedMsg{}
				return
			}
			if role := in.GetRole(); role != nil {
				m.notification = fmt.Sprintf("Assigned role received. required=%d", role.RequiredAtoms)
				m.msgCh <- client.UpdatedMsg{}
				continue
			}
			if req := in.GetReq(); req != nil {
				build := func(br pong.Branch) ([]*pong.PreSigBatch_Sig, error) {
					draftHex := req.DraftAwinsTxHex
					if br == pong.Branch_BRANCH_B {
						draftHex = req.DraftBwinsTxHex
					}
					txb, err := hex.DecodeString(draftHex)
					if err != nil {
						return nil, err
					}
					var tx wire.MsgTx
					if err := tx.Deserialize(bytes.NewReader(txb)); err != nil {
						return nil, err
					}
					out := make([]*pong.PreSigBatch_Sig, 0, len(req.Inputs))
					for _, in := range req.Inputs {
						redeem, _ := hex.DecodeString(in.RedeemScriptHex)
						idx, err := findInputIndex(&tx, in.InputId)
						if err != nil {
							return nil, err
						}
						mBytes, err := txscript.CalcSignatureHash(redeem, txscript.SigHashAll, &tx, idx, nil)
						if err != nil || len(mBytes) != 32 {
							return nil, fmt.Errorf("sighash")
						}
						mHex := hex.EncodeToString(mBytes)
						var tComp []byte
						if br == pong.Branch_BRANCH_A {
							if mHex != in.MAwinsHex {
								return nil, fmt.Errorf("m mismatch A")
							}
							tComp = in.TAwinsCompressed
						} else {
							if mHex != in.MBwinsHex {
								return nil, fmt.Errorf("m mismatch B")
							}
							tComp = in.TBwinsCompressed
						}
						rX, sPrime, _, err := computePreSig(m.genPrivHex, mHex, hex.EncodeToString(tComp))
						if err != nil {
							return nil, err
						}
						rb, _ := hex.DecodeString(rX)
						sb, _ := hex.DecodeString(sPrime)
						out = append(out, &pong.PreSigBatch_Sig{InputId: in.InputId, Rprime32: rb, Sprime32: sb})
					}
					return out, nil
				}
				for _, br := range req.BranchesToPresign {
					presigs, err := build(br)
					if err != nil {
						m.notification = err.Error()
						m.msgCh <- client.UpdatedMsg{}
						return
					}
					_ = stream.Send(&pong.ClientMsg{MatchId: m.settle.matchID, Kind: &pong.ClientMsg_Presigs{Presigs: &pong.PreSigBatch{Branch: br, Presigs: presigs}}})
				}
				m.notification = "Pre-sig batches sent via stream."
				m.msgCh <- client.UpdatedMsg{}
				continue
			}
			if rev := in.GetReveal(); rev != nil {
				_ = rev
				m.notification = "Gamma revealed (winner can finalize)."
				m.msgCh <- client.UpdatedMsg{}
				continue
			}
		}
	}()
}
