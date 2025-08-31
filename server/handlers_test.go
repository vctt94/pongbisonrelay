package server

import (
	"encoding/hex"
	"testing"

	"github.com/decred/dcrd/chaincfg/v3"
	"github.com/decred/dcrd/txscript/v4"
	"github.com/decred/dcrd/txscript/v4/stdaddr"
)

// Note: this uses dummy XA/XB and AC values; the exact values are irrelevant
// as long as buildEscrowRedeemScript produces a deterministic redeem.
func TestPkScriptFromRedeemConsistency(t *testing.T) {
	xaHex := "02aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	xbHex := "02bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	acHex := "03cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"
	xa, _ := hex.DecodeString(xaHex)
	xb, _ := hex.DecodeString(xbHex)
	ac, _ := hex.DecodeString(acHex)
	state := &refMatchState{csv: 10, xa: xa, xb: xb}
	redeem, err := buildEscrowRedeemScript(state, ac)
	if err != nil {
		t.Fatalf("buildEscrowRedeemScript error: %v", err)
	}
	if err := setDepositFromRedeemLocked(state, hex.EncodeToString(redeem)); err != nil {
		t.Fatalf("setDepositFromRedeemLocked error: %v", err)
	}

	// Address from redeem
	h := stdaddr.Hash160(redeem)
	a1, err := stdaddr.NewAddressScriptHashV0(h, chaincfg.TestNet3Params())
	if err != nil {
		t.Fatalf("addr from redeem error: %v", err)
	}

	// Address from pkScript
	pkb, err := hex.DecodeString(state.depositPkScriptHex)
	if err != nil {
		t.Fatalf("decode pkScriptHex: %v", err)
	}
	if len(pkb) != 23 || pkb[0] != txscript.OP_HASH160 || pkb[1] != 0x14 || pkb[22] != txscript.OP_EQUAL {
		t.Fatalf("bad pkScript form: %x", pkb)
	}
	a2, err := stdaddr.NewAddressScriptHashV0(pkb[2:22], chaincfg.TestNet3Params())
	if err != nil {
		t.Fatalf("addr from pk error: %v", err)
	}

	if a1.String() != a2.String() {
		t.Fatalf("addr mismatch: redeem=%s pk=%s", a1.String(), a2.String())
	}
}

func TestAdaptorDerivationDeterminismAndEvenY(t *testing.T) {
	matchID := "m123"
	inputID := "txid:vout"
	branch := int32(1)
	sighash := "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789"
	serverPriv := "11ee22dd33cc44bb55aa66ff77ee88dd99cc00bbaa11223344556677889900aa"
	g1, T1 := deriveAdaptorGamma(matchID, inputID, branch, sighash, serverPriv)
	g2, T2 := deriveAdaptorGamma(matchID, inputID, branch, sighash, serverPriv)
	if g1 != g2 || T1 != T2 {
		t.Fatalf("adaptor not deterministic: (%s,%s) != (%s,%s)", g1, T1, g2, T2)
	}
	Tb, err := hex.DecodeString(T1)
	if err != nil || len(Tb) != 33 {
		t.Fatalf("bad T encoding: %v len=%d", err, len(Tb))
	}
	if Tb[0] != 0x02 { // even-Y
		t.Fatalf("T not even-Y: prefix=%x", Tb[0])
	}
}

// Ensure redeem script depends only on depositor pubkey and CSV, not XA/XB.
func TestRedeemScriptNoOpponentKey(t *testing.T) {
	// Same depositor key and CSV
	acHex := "03cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"
	ac, _ := hex.DecodeString(acHex)

	xa1Hex := "02aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	xb1Hex := "02bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	xa2Hex := "020101010101010101010101010101010101010101010101010101010101010101"
	xb2Hex := "020202020202020202020202020202020202020202020202020202020202020202"
	xa1, _ := hex.DecodeString(xa1Hex)
	xb1, _ := hex.DecodeString(xb1Hex)
	xa2, _ := hex.DecodeString(xa2Hex)
	xb2, _ := hex.DecodeString(xb2Hex)

	s1 := &refMatchState{csv: 10, xa: xa1, xb: xb1}
	s2 := &refMatchState{csv: 10, xa: xa2, xb: xb2}

	r1, err := buildEscrowRedeemScript(s1, ac)
	if err != nil {
		t.Fatalf("buildEscrowRedeemScript s1: %v", err)
	}
	r2, err := buildEscrowRedeemScript(s2, ac)
	if err != nil {
		t.Fatalf("buildEscrowRedeemScript s2: %v", err)
	}
	if hex.EncodeToString(r1) != hex.EncodeToString(r2) {
		t.Fatalf("redeem differs with different XA/XB; got %s vs %s", hex.EncodeToString(r1), hex.EncodeToString(r2))
	}
}
