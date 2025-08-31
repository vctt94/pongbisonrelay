package client

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/decred/dcrd/chaincfg/chainhash"
	"github.com/decred/dcrd/crypto/blake256"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/decred/dcrd/dcrec/secp256k1/v4/schnorr"
	"github.com/decred/dcrd/wire"
)

const (
	DefaultBetAtoms = 100000000
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

// ComputePreSigMinusFull computes a minus-variant adaptor pre-signature ensuring a
// normalized (even-Y) R' and returns the full compressed R' and s' bytes.
// It uses RFC6979 deterministic nonces and retries until the derived R' has
// even Y as required by EC-Schnorr-DCRv0's verification equation.
func ComputePreSigMinusFull(xPrivHex, mHex, TCompHex string) (rPrimeCompressed []byte, sPrime32 []byte, err error) {
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
