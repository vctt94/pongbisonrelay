package client

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	"github.com/decred/dcrd/chaincfg/v3"
	"github.com/decred/dcrd/crypto/blake256"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/decred/dcrd/dcrutil/v4"
	"github.com/decred/dcrd/txscript/v4"
	"github.com/decred/dcrd/txscript/v4/stdaddr"
	"github.com/decred/dcrd/wire"

	"github.com/vctt94/pong-bisonrelay/pongrpc/grpc/pong"
)

const (
	DefaultBetAtoms = 100000000
)

// computePreSig derives k via RFC6979 and enforces even-Y on R' = k*G + T.
// Returns r_x, s', R' (all hex).
func ComputePreSig(xPrivHex, mHex, TCompHex string) (rXHex string, sPrimeHex string, rCompHex string, err error) {
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

// DeriveAdaptorPreSig computes a minus-variant adaptor pre-signature ensuring a
// normalized (even-Y) R' and returns (R' compressed, s'). Uses RFC6979 nonces
// and retries until the derived R' has even Y per EC-Schnorr-DCRv0.
func DeriveAdaptorPreSig(xPrivHex, mHex, TCompHex string) (rPrimeCompressed []byte, sPrime32 []byte, err error) {
	rXHex, sPrimeHex, rCompHex, err := ComputePreSig(xPrivHex, mHex, TCompHex)
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

// PayoutPubkeyFromConfHex parses -address flag input as 33B hex or a P2PK address (any known net).
func PayoutPubkeyFromConfHex(addrOrHex string) ([]byte, error) {
	s := strings.TrimSpace(addrOrHex)
	// Try as hex compressed pubkey first
	if b, err := hex.DecodeString(s); err == nil {
		if len(b) == 33 {
			return b, nil
		}
	}
	// Try as Decred address (prefer testnet, then mainnet/simnet/regnet)
	paramsList := []*chaincfg.Params{
		chaincfg.TestNet3Params(),
		chaincfg.MainNetParams(),
		chaincfg.SimNetParams(),
		chaincfg.RegNetParams(),
	}
	var lastErr error
	for _, p := range paramsList {
		addr, err := stdaddr.DecodeAddress(s, p)
		if err != nil {
			lastErr = err
			continue
		}
		// Check if this is a P2PK (raw pubkey) address
		type compressedPubKeyer interface{ SerializedPubKey() []byte }
		if pk, ok := addr.(compressedPubKeyer); ok {
			b := pk.SerializedPubKey()
			if len(b) == 33 {
				return b, nil
			}
			if len(b) == 65 {
				pp, err := secp256k1.ParsePubKey(b)
				if err != nil {
					return nil, fmt.Errorf("invalid pubkey inside address: %v", err)
				}
				return pp.SerializeCompressed(), nil
			}
			return nil, fmt.Errorf("pubkey in address has unexpected length %d", len(b))
		}
		return nil, fmt.Errorf("address is not P2PK; pass a P2PK address or 33-byte pubkey hex")
	}
	return nil, fmt.Errorf("decode -address failed: %v", lastErr)
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
		rComp, sPrime, err := DeriveAdaptorPreSig(xPrivHex, mHex, hex.EncodeToString(in.TCompressed))
		if err != nil {
			return nil, err
		}
		pres = append(pres, &pong.PreSig{InputId: in.InputId, RprimeCompressed: rComp, Sprime32: sPrime})
	}
	return &pong.VerifyOk{AckDigest: ack[:], Presigs: pres}, nil
}

// FinalizeWinner attaches winner-path sigs for the given draft and presigs and returns the tx hex.
func FinalizeWinner(gammaHex, draftHex string, inputsForDraft []*pong.NeedPreSigs_PerInput, presigsForDraft map[string]*pong.PreSig) (string, error) {
	if presigsForDraft == nil || len(presigsForDraft) == 0 || inputsForDraft == nil || len(inputsForDraft) == 0 {
		return "", fmt.Errorf("no presigs stored for this draft")
	}

	raw, err := hex.DecodeString(strings.TrimSpace(draftHex))
	if err != nil {
		return "", fmt.Errorf("decode draft hex: %w", err)
	}
	var tx wire.MsgTx
	if err := tx.Deserialize(bytes.NewReader(raw)); err != nil {
		return "", fmt.Errorf("deserialize draft: %w", err)
	}

	idxByID := make(map[string]int, len(tx.TxIn))
	have := make([]string, 0, len(tx.TxIn))
	for i, ti := range tx.TxIn {
		k := strings.ToLower(fmt.Sprintf("%s:%d", ti.PreviousOutPoint.Hash.String(), ti.PreviousOutPoint.Index))
		idxByID[k] = i
		have = append(have, k)
	}
	sort.Strings(have)

	redeemByID := make(map[string]string, len(inputsForDraft))
	for _, in := range inputsForDraft {
		redeemByID[strings.ToLower(in.InputId)] = in.RedeemScriptHex
	}

	gb, err := hex.DecodeString(strings.TrimSpace(gammaHex))
	if err != nil || len(gb) != 32 {
		return "", fmt.Errorf("bad gamma")
	}
	var gamma secp256k1.ModNScalar
	if overflow := gamma.SetByteSlice(gb); overflow {
		return "", fmt.Errorf("gamma overflow")
	}

	for id := range idxByID {
		if _, ok := presigsForDraft[id]; !ok {
			return "", fmt.Errorf("missing presig for input %s; cannot finalize", id)
		}
	}

	for id, ps := range presigsForDraft {
		key := strings.ToLower(strings.TrimSpace(id))
		idx, ok := idxByID[key]
		if !ok {
			return "", fmt.Errorf("input %s not found in draft. draft has: %v", id, have)
		}
		redHex, ok := redeemByID[key]
		if !ok {
			return "", fmt.Errorf("no redeem script stored for input %s", id)
		}
		redeem, err := hex.DecodeString(redHex)
		if err != nil {
			return "", fmt.Errorf("decode redeem for %s: %w", id, err)
		}
		if len(ps.RprimeCompressed) != 33 || len(ps.Sprime32) != 32 {
			return "", fmt.Errorf("bad presig sizes for %s", id)
		}
		if ps.RprimeCompressed[0] != 0x02 {
			return "", fmt.Errorf("R' must have even Y (0x02) for %s, got 0x%02x", id, ps.RprimeCompressed[0])
		}
		var sPrime secp256k1.ModNScalar
		if overflow := sPrime.SetByteSlice(ps.Sprime32); overflow {
			return "", fmt.Errorf("s' overflow for %s", id)
		}
		sPrime.Add(&gamma)
		sBytes := sPrime.Bytes()
		rX := ps.RprimeCompressed[1:33]
		sig65 := make([]byte, 0, 65)
		sig65 = append(sig65, rX...)
		sig65 = append(sig65, sBytes[:]...)
		sig65 = append(sig65, byte(txscript.SigHashAll))
		sb := txscript.NewScriptBuilder()
		sb.AddData(sig65)
		sb.AddOp(txscript.OP_1)
		sb.AddData(redeem)
		sigScript, err := sb.Script()
		if err != nil {
			return "", fmt.Errorf("build scriptSig for %s: %w", id, err)
		}
		tx.TxIn[idx].SignatureScript = sigScript

		sh := dcrutil.Hash160(redeem)
		pkScript, err := txscript.NewScriptBuilder().AddOp(txscript.OP_HASH160).AddData(sh).AddOp(txscript.OP_EQUAL).Script()
		if err != nil {
			return "", fmt.Errorf("build pkScript for %s: %w", id, err)
		}
		vm, err := txscript.NewEngine(pkScript, &tx, idx, 0, 0, nil)
		if err != nil {
			return "", fmt.Errorf("engine init for %s: %w", id, err)
		}
		if err := vm.Execute(); err != nil {
			return "", fmt.Errorf("local VM verify failed on %s: %w", id, err)
		}
	}
	var out bytes.Buffer
	_ = tx.Serialize(&out)
	return hex.EncodeToString(out.Bytes()), nil
}
