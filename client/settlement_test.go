package client

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/decred/dcrd/dcrec/secp256k1/v4/schnorr"
	"github.com/decred/dcrd/txscript/v4"
	"github.com/decred/dcrd/wire"
)

// buildWinnerRedeem builds the minimal winner-branch snippet used in this repo:
// <P_c> OP_SWAP OP_2 OP_CHECKSIGALTVERIFY OP_TRUE
func buildWinnerRedeem(pub33 []byte) []byte {
	b := txscript.NewScriptBuilder()
	b.AddData(pub33)
	b.AddOp(txscript.OP_SWAP)
	b.AddOp(txscript.OP_2)
	b.AddOp(txscript.OP_CHECKSIGALTVERIFY)
	b.AddOp(txscript.OP_TRUE)
	s, _ := b.Script()
	return s
}

// deriveEvenYGamma derives a random gamma and returns (gamma32, Tcompressed-evenY).
func deriveEvenYGamma() ([32]byte, []byte) {
	// Use a random private key as gamma source.
	priv, _ := secp256k1.GeneratePrivateKey()
	var g32 [32]byte
	copy(g32[:], priv.Serialize())
	T := priv.PubKey()
	comp := T.SerializeCompressed()
	if comp[0] == 0x03 { // odd Y → negate gamma
		var sc secp256k1.ModNScalar
		sc.SetByteSlice(g32[:])
		var neg secp256k1.ModNScalar
		neg.NegateVal(&sc)
		g32 = [32]byte{}
		nb := neg.Bytes()
		copy(g32[:], nb[:])
		priv = secp256k1.PrivKeyFromBytes(nb[:])
		T = priv.PubKey()
		comp = T.SerializeCompressed()
	}
	return g32, comp
}

func TestFinalizeSchnorrSuccess(t *testing.T) {
	// Key pair for the depositor
	xPriv, _ := secp256k1.GeneratePrivateKey()
	X := xPriv.PubKey().SerializeCompressed()
	redeem := buildWinnerRedeem(X)

	// Draft tx spending arbitrary prevout with 1 output (values irrelevant for m here)
	var prev wire.OutPoint
	tx := wire.NewMsgTx()
	tx.AddTxIn(&wire.TxIn{PreviousOutPoint: prev})
	tx.AddTxOut(&wire.TxOut{Value: 1000, PkScript: []byte{}})

	// Compute node-style m for input 0 with redeem as subscript
	mNode, err := txscript.CalcSignatureHash(redeem, txscript.SigHashAll, tx, 0, nil)
	if err != nil || len(mNode) != 32 {
		t.Fatalf("CalcSignatureHash failed: %v", err)
	}
	mHex := hex.EncodeToString(mNode)

	// Adaptor secret and point (even-Y)
	gamma32, Tcomp := deriveEvenYGamma()
	THex := hex.EncodeToString(Tcomp)

	// Compute pre-signature (R', s') bound to m and T
	rComp, sPrime, err := ComputePreSigMinusFull(hex.EncodeToString(xPriv.Serialize()), mHex, THex)
	if err != nil {
		t.Fatalf("ComputePreSigMinusFull: %v", err)
	}
	if rComp[0] != 0x02 {
		t.Fatalf("R' not even-Y")
	}

	// Finalize: s = s' + gamma
	var sp, gg secp256k1.ModNScalar
	if overflow := sp.SetByteSlice(sPrime); overflow {
		t.Fatalf("s' overflow")
	}
	if overflow := gg.SetByteSlice(gamma32[:]); overflow {
		t.Fatalf("gamma overflow")
	}
	sp.Add(&gg)
	spb := sp.Bytes()
	sig64 := append(append([]byte{}, rComp[1:33]...), spb[:]...)

	// Verify with schnorr over mNode and X
	sobj, err := schnorr.ParseSignature(sig64)
	if err != nil {
		t.Fatalf("parse sig: %v", err)
	}
	Q, _ := schnorr.ParsePubKey(X)
	if !sobj.Verify(mNode, Q) {
		t.Fatalf("local verify failed")
	}
}

func TestFinalizeSchnorrFailsOnMutation(t *testing.T) {
	xPriv, _ := secp256k1.GeneratePrivateKey()
	X := xPriv.PubKey().SerializeCompressed()
	redeem := buildWinnerRedeem(X)

	var prev wire.OutPoint
	tx := wire.NewMsgTx()
	tx.AddTxIn(&wire.TxIn{PreviousOutPoint: prev})
	tx.AddTxOut(&wire.TxOut{Value: 1000, PkScript: []byte{}})

	mNode, err := txscript.CalcSignatureHash(redeem, txscript.SigHashAll, tx, 0, nil)
	if err != nil || len(mNode) != 32 {
		t.Fatalf("CalcSignatureHash failed: %v", err)
	}
	mHex := hex.EncodeToString(mNode)

	gamma32, Tcomp := deriveEvenYGamma()
	THex := hex.EncodeToString(Tcomp)
	rComp, sPrime, err := ComputePreSigMinusFull(hex.EncodeToString(xPriv.Serialize()), mHex, THex)
	if err != nil {
		t.Fatalf("ComputePreSigMinusFull: %v", err)
	}

	var sp, gg secp256k1.ModNScalar
	if overflow := sp.SetByteSlice(sPrime); overflow {
		t.Fatalf("s' overflow")
	}
	if overflow := gg.SetByteSlice(gamma32[:]); overflow {
		t.Fatalf("gamma overflow")
	}
	sp.Add(&gg)
	spb := sp.Bytes()
	sig64 := append(append([]byte{}, rComp[1:33]...), spb[:]...)
	sobj, _ := schnorr.ParseSignature(sig64)
	Q, _ := schnorr.ParsePubKey(X)

	// Mutate: flip one byte of the draft by adding a new output → changes m
	tx2 := wire.NewMsgTx()
	// copy inputs
	for _, ti := range tx.TxIn {
		tx2.AddTxIn(&wire.TxIn{PreviousOutPoint: ti.PreviousOutPoint})
	}
	// copy outputs
	for _, to := range tx.TxOut {
		tx2.AddTxOut(&wire.TxOut{Value: to.Value, PkScript: append([]byte{}, to.PkScript...)})
	}
	// add one more output to mutate
	tx2.AddTxOut(&wire.TxOut{Value: 1, PkScript: []byte{}})
	m2, err := txscript.CalcSignatureHash(redeem, txscript.SigHashAll, tx2, 0, nil)
	if err != nil || len(m2) != 32 {
		t.Fatalf("CalcSignatureHash2 failed: %v", err)
	}
	if bytes.Equal(m2, mNode) {
		t.Fatalf("expected different m after mutation")
	}
	if sobj.Verify(m2, Q) {
		t.Fatalf("verify unexpectedly succeeded with mutated draft")
	}

	// Mutate redeem: flip one byte
	redeem2 := append([]byte{}, redeem...)
	redeem2[len(redeem2)-1] ^= 0x01
	m3, err := txscript.CalcSignatureHash(redeem2, txscript.SigHashAll, tx, 0, nil)
	if err != nil || len(m3) != 32 {
		t.Fatalf("CalcSignatureHash3 failed: %v", err)
	}
	if bytes.Equal(m3, mNode) {
		t.Fatalf("expected different m after redeem mutation")
	}
	if sobj.Verify(m3, Q) {
		t.Fatalf("verify unexpectedly succeeded with mutated redeem")
	}
}
