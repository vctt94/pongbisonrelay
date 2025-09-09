package server

import (
	"bytes"
	"encoding/hex"
	"testing"

	clientpkg "github.com/vctt94/pong-bisonrelay/client"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/decred/dcrd/dcrec/secp256k1/v4/schnorr"
	"github.com/decred/dcrd/txscript/v4"
	"github.com/decred/dcrd/wire"
)

func TestFinalizeSchnorrSuccess(t *testing.T) {
	// Key pair for the depositor
	xPriv, _ := secp256k1.GeneratePrivateKey()
	X := xPriv.PubKey().SerializeCompressed()
	redeem, err := buildPerDepositorRedeemScript(X, 10)
	if err != nil {
		t.Fatalf("build redeem: %v", err)
	}

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

	// Adaptor secret and point (even-Y) using server's derivation method
	const pocServerPrivHex = "11ee22dd33cc44bb55aa66ff77ee88dd99cc00bbaa11223344556677889900aa"
	gammaHex, THex := deriveAdaptorGamma("", "test", 0, "test", pocServerPrivHex)
	gb, err := hex.DecodeString(gammaHex)
	if err != nil || len(gb) != 32 {
		t.Fatalf("derive gamma: %v", err)
	}
	var gamma32 [32]byte
	copy(gamma32[:], gb)

	// Compute pre-signature (R', s') bound to m and T
	rComp, sPrime, err := clientpkg.ComputePreSigMinusFull(hex.EncodeToString(xPriv.Serialize()), mHex, THex)
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
	redeem, err := buildPerDepositorRedeemScript(X, 10)
	if err != nil {
		t.Fatalf("build redeem: %v", err)
	}

	var prev wire.OutPoint
	tx := wire.NewMsgTx()
	tx.AddTxIn(&wire.TxIn{PreviousOutPoint: prev})
	tx.AddTxOut(&wire.TxOut{Value: 1000, PkScript: []byte{}})

	mNode, err := txscript.CalcSignatureHash(redeem, txscript.SigHashAll, tx, 0, nil)
	if err != nil || len(mNode) != 32 {
		t.Fatalf("CalcSignatureHash failed: %v", err)
	}
	mHex := hex.EncodeToString(mNode)

	const pocServerPrivHex2 = "11ee22dd33cc44bb55aa66ff77ee88dd99cc00bbaa11223344556677889900aa"
	gammaHex2, THex := deriveAdaptorGamma("", "test", 0, "test", pocServerPrivHex2)
	gb2, err := hex.DecodeString(gammaHex2)
	if err != nil || len(gb2) != 32 {
		t.Fatalf("derive gamma: %v", err)
	}
	var gamma32 [32]byte
	copy(gamma32[:], gb2)
	rComp, sPrime, err := clientpkg.ComputePreSigMinusFull(hex.EncodeToString(xPriv.Serialize()), mHex, THex)
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
