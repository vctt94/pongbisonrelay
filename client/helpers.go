package client

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/decred/dcrd/chaincfg/chainhash"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/decred/dcrd/wire"
)

// https://github.com/decred/dcrd/blob/master/dcrec/secp256k1/schnorr/README.md?plain=1#L249
var schnorrV0ExtraTag = func() [32]byte {
	// EC-Schnorr-DCRv0 nonce domain-sep tag from the spec.
	const tagHex = "0b75f97b60e8a5762876c004829ee9b926fa6f0d2eeaec3a4fd1446a768331cb"
	b, _ := hex.DecodeString(tagHex)
	var out [32]byte
	copy(out[:], b)
	return out
}()

func findInputIndex(tx *wire.MsgTx, inputID string) (int, error) {
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
