package server

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/decred/dcrd/chaincfg/v3"
	"github.com/decred/dcrd/crypto/blake256"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/decred/dcrd/txscript/v4"
	"github.com/decred/dcrd/txscript/v4/stdaddr"
	"github.com/decred/dcrd/txscript/v4/stdscript"
)

// Unit conversion constants and helpers between matoms, atoms, and DCR.
const (
	MatomsPerAtom = int64(1_000)                // 1000 matoms = 1 atom
	AtomsPerDCR   = int64(100_000_000)          // 1e8 atoms per DCR
	MatomsPerDCR  = MatomsPerAtom * AtomsPerDCR // 1e11 matoms per DCR
)

func matomsToAtoms(m int64) int64 { return m / MatomsPerAtom }
func atomsToMatoms(a int64) int64 { return a * MatomsPerAtom }
func matomsToDCR(m int64) float64 { return float64(m) / float64(MatomsPerDCR) }
func atomsToDCR(a int64) float64  { return float64(a) / float64(AtomsPerDCR) }

// setDepositFromRedeemLocked normalizes redeem hex, derives the corresponding
// standard P2SH (v0) pkScript, lowercases both, and stores in the match state.
// Caller must hold the appropriate lock protecting the state.
func setDepositFromRedeemLocked(state *refMatchState, redeemHex string) error {
	norm := strings.ToLower(redeemHex)
	rb, err := hex.DecodeString(norm)
	if err != nil {
		return err
	}
	h := stdaddr.Hash160(rb)
	pkb, err := txscript.NewScriptBuilder().
		AddOp(txscript.OP_HASH160).
		AddData(h[:]).
		AddOp(txscript.OP_EQUAL).
		Script()
	if err != nil {
		return err
	}
	state.DepositRedeemScriptHex = strings.ToLower(hex.EncodeToString(rb))
	state.DepositPkScriptHex = strings.ToLower(hex.EncodeToString(pkb))
	return nil
}

// addrFromPkScript returns the canonical address encoded by a v0 P2SH pkScript.
func addrFromPkScript(pkHex string, params *chaincfg.Params) (string, error) {
	pkb, err := hex.DecodeString(pkHex)
	if err != nil {
		return "", fmt.Errorf("decode pkScript hex: %w", err)
	}
	// Script version 0 is correct for standard P2SH.
	st := stdscript.DetermineScriptType(0, pkb)
	if st != stdscript.STScriptHash {
		return "", fmt.Errorf("pkScript is not P2SH v0 (got %v)", st)
	}
	_, addrs := stdscript.ExtractAddrs(0, pkb, params)
	if len(addrs) == 0 {
		return "", fmt.Errorf("no address extracted from pkScript")
	}
	return addrs[0].String(), nil
}

// deriveAdaptorGamma deterministically derives a per-input adaptor secret bound to
// (matchID, inputID, branch, sighash) and the server's private key. It returns a
// 32-byte gamma scalar (hex) and the corresponding even-Y compressed point T.
func deriveAdaptorGamma(matchID, inputID string, branch int32, sighashHex string, serverPrivHex string) (gammaHex string, TCompHex string) {
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
