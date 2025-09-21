package pongbisonrelay

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/decred/dcrd/chaincfg/chainhash"
	"github.com/decred/dcrd/chaincfg/v3"
	"github.com/decred/dcrd/crypto/blake256"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/decred/dcrd/dcrutil/v4"
	"github.com/decred/dcrd/txscript/v4"
	"github.com/decred/dcrd/txscript/v4/stdaddr"
	"github.com/decred/dcrd/wire"
	"github.com/vctt94/pongbisonrelay/pongrpc/grpc/pong"
)

func BuildPerDepositorRedeemScript(comp33 []byte, csvBlocks uint32) ([]byte, error) {
	if len(comp33) != 33 {
		return nil, fmt.Errorf("need 33-byte compressed pubkey")
	}
	b := txscript.NewScriptBuilder()

	// Winner branch:
	//   initial stack after P2SH/IF: [sig]
	//   push <pub>, 2 -> [sig, pub, 2]
	//   OP_CHECKSIGALTVERIFY pops (2, pub, sig) == (hashtype, pubkey, signature)
	b.AddOp(txscript.OP_IF).
		AddData(comp33).
		AddInt64(2). // schnorr-secp256k1
		AddOp(txscript.OP_CHECKSIGALTVERIFY).
		AddOp(txscript.OP_TRUE).

		// Timeout branch:
		AddOp(txscript.OP_ELSE).
		AddInt64(int64(csvBlocks)).
		AddOp(txscript.OP_CHECKSEQUENCEVERIFY).
		AddOp(txscript.OP_DROP).
		AddData(comp33).
		AddInt64(2). // schnorr-secp256k1
		AddOp(txscript.OP_CHECKSIGALTVERIFY).
		AddOp(txscript.OP_TRUE).
		AddOp(txscript.OP_ENDIF)

	return b.Script()
}

// deriveAdaptorGamma deterministically derives a per-input adaptor secret bound to
// (matchID, inputID, branch, sighash) and the server's private key. It returns a
// 32-byte gamma scalar (hex) and the corresponding even-Y compressed point T.
func DeriveAdaptorGamma(matchID, inputID string, branch int32, sighashHex string, serverPrivHex string) (gammaHex string, TCompHex string) {
	// Hash context with domain separation tags; reduce mod n.
	h := blake256.New()
	h.Write([]byte("Adaptor/Derive/v1"))
	h.Write([]byte(matchID))
	h.Write([]byte{'|'})
	h.Write([]byte(inputID))
	h.Write([]byte{'|'})
	// branch is small; include as 1 byte
	h.Write([]byte{byte(branch)})
	h.Write([]byte{'|'})
	if b, err := hex.DecodeString(sighashHex); err == nil {
		h.Write(b)
	} else {
		h.Write([]byte(sighashHex))
	}
	h.Write([]byte{'|'})
	if sp, err := hex.DecodeString(serverPrivHex); err == nil {
		h.Write(sp)
	} else {
		h.Write([]byte(serverPrivHex))
	}
	sum := h.Sum(nil)

	var sc secp256k1.ModNScalar
	sc.SetByteSlice(sum)
	// Ensure non-zero
	if sc.IsZero() {
		// Add 1 to avoid zero
		var one secp256k1.ModNScalar
		one.SetInt(1)
		sc.Add(&one)
	}
	// Compute T = gamma*G and normalize to even-Y by possibly negating gamma.
	g := sc.Bytes()
	priv := secp256k1.PrivKeyFromBytes(g[:])
	T := priv.PubKey()
	comp := T.SerializeCompressed()
	// If odd Y prefix (0x03), negate gamma => n - gamma, which negates T as well.
	if len(comp) == 33 && comp[0] == 0x03 {
		// gamma = -gamma mod n
		var neg secp256k1.ModNScalar
		neg.NegateVal(&sc)
		sc = neg
		g = sc.Bytes()
		priv = secp256k1.PrivKeyFromBytes(g[:])
		T = priv.PubKey()
		comp = T.SerializeCompressed()
	}
	g = sc.Bytes()
	gammaHex = hex.EncodeToString(g[:])
	TCompHex = hex.EncodeToString(comp)
	return
}

// https://github.com/decred/dcrd/blob/master/dcrec/secp256k1/schnorr/README.md?plain=1#L249
var SchnorrV0ExtraTag = func() [32]byte {
	// EC-Schnorr-DCRv0 nonce domain-sep tag from the spec.
	const tagHex = "0b75f97b60e8a5762876c004829ee9b926fa6f0d2eeaec3a4fd1446a768331cb"
	b, _ := hex.DecodeString(tagHex)
	var out [32]byte
	copy(out[:], b)
	return out
}()

func FindInputIndex(tx *wire.MsgTx, inputID string) (int, error) {
	parts := strings.Split(inputID, ":")
	if len(parts) != 2 {
		return -1, fmt.Errorf("bad input_id %q: want txid:vout", inputID)
	}
	// fmt.Println("parts", parts)
	txid := parts[0]
	voutU64, err := strconv.ParseUint(parts[1], 10, 32)
	if err != nil {
		return -1, fmt.Errorf("bad input_id %q: %w", inputID, err)
	}
	var h chainhash.Hash
	if err := chainhash.Decode(&h, txid); err != nil {
		return -1, fmt.Errorf("bad txid %q: %w", txid, err)
	}
	vout := uint32(voutU64)

	matchCount := 0
	matchIdx := -1
	for i, ti := range tx.TxIn {
		if ti.PreviousOutPoint.Hash == h && ti.PreviousOutPoint.Index == vout {
			matchCount++
			matchIdx = i
		}
	}
	if matchCount == 0 {
		return -1, fmt.Errorf("input %s not found in draft", inputID)
	}
	if matchCount > 1 {
		return -1, fmt.Errorf("input %s matches %d inputs in draft (ambiguous)", inputID, matchCount)
	}
	return matchIdx, nil
}

// addPoints returns R+S as a *secp256k1.PublicKey using Jacobian add and affine conversion.
func AddPoints(R, S *secp256k1.PublicKey) (*secp256k1.PublicKey, error) {
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
		if len(ps.RLineCompressed) != 33 || len(ps.SLine32) != 32 {
			return "", fmt.Errorf("bad presig sizes for %s", id)
		}
		if ps.RLineCompressed[0] != 0x02 {
			return "", fmt.Errorf("R' must have even Y (0x02) for %s, got 0x%02x", id, ps.RLineCompressed[0])
		}
		var sPrime secp256k1.ModNScalar
		if overflow := sPrime.SetByteSlice(ps.SLine32); overflow {
			return "", fmt.Errorf("s' overflow for %s", id)
		}
		sPrime.Add(&gamma)
		sBytes := sPrime.Bytes()
		rX := ps.RLineCompressed[1:33]
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

// PayoutPubkeyFromConfHex parses a 33B/65B pubkey hex or a Decred P2PK ("pubkeyaddr")
// for any known net and returns a 33-byte compressed secp256k1 pubkey.
//
// Examples it will accept:
//   - "034bbf63cf03ecd1...f2e3"              (33B hex)
//   - "04...."                               (65B hex -> compressed)
//   - "TkQ3wZu4yVeS1cfxkXJjmYvHd2FEMfSaE5yW1jhrkKFkiwYVX7bGK"  (testnet P2PK)
//
// It will *not* accept a P2PKH like "TsoY9N6..." (can’t derive pubkey from hash).
func PayoutPubkeyFromConfHex(addrOrHex string) ([]byte, error) {
	s := strings.TrimSpace(addrOrHex)
	if s == "" {
		return nil, fmt.Errorf("empty -address")
	}

	// 1) Try hex.
	if b, err := hex.DecodeString(s); err == nil {
		switch len(b) {
		case 33:
			return b, nil
		case 65:
			pk, err := secp256k1.ParsePubKey(b)
			if err != nil {
				return nil, fmt.Errorf("invalid 65-byte pubkey: %w", err)
			}
			return pk.SerializeCompressed(), nil
		default:
			return nil, fmt.Errorf("pubkey hex must be 33 or 65 bytes, got %d", len(b))
		}
	}

	// 2) Try as Decred address (prefer testnet first).
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

		// Most pubkey address types in stdaddr expose SerializedPubKey().
		type hasSerializedPubKey interface{ SerializedPubKey() []byte }
		if pkAddr, ok := addr.(hasSerializedPubKey); ok {
			b := pkAddr.SerializedPubKey()
			switch len(b) {
			case 33:
				return b, nil
			case 65:
				pk, err := secp256k1.ParsePubKey(b)
				if err != nil {
					return nil, fmt.Errorf("invalid pubkey inside address: %w", err)
				}
				return pk.SerializeCompressed(), nil
			default:
				return nil, fmt.Errorf("pubkey in address has unexpected length %d", len(b))
			}
		}

		// If we got here, it decoded as some *other* address type (e.g., P2PKH).
		return nil, fmt.Errorf("address is not P2PK (pubkey). Pass a pubkey address (the 'pubkeyaddr' value) or a 33-byte pubkey hex")
	}
	return nil, fmt.Errorf("decode failed as hex or address: %v", lastErr)
}
