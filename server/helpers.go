package server

import (
	"encoding/hex"
	"fmt"

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

// buildPerDepositorRedeemScript builds a per-depositor redeem script that does NOT
// depend on the opponent's key. Winner path accepts a valid Schnorr (alt) sig
// under depositor's compressed pubkey; timeout path lets depositor recover after CSV.
//
// Spend patterns (scriptSig / witness items pushed by spender, last item is top):
//   - Winner path:  <sig_final> 1
//   - Timeout path: <sig_owner> <csvBlocks> 0
//
// Script (pseudocode):
//
//	OP_IF
//	  <P_c> OP_CHECKSIGALTVERIFY
//	  OP_TRUE
//	OP_ELSE
//	  <csv> OP_CHECKSEQUENCEVERIFY OP_DROP
//	  <P_c> OP_CHECKSIGALTVERIFY
//	  OP_TRUE
//	OP_ENDIF
func buildPerDepositorRedeemScript(comp33 []byte, csvBlocks uint32) ([]byte, error) {
	if len(comp33) != 33 {
		return nil, fmt.Errorf("need 33-byte compressed pubkey")
	}
	b := txscript.NewScriptBuilder()

	b.AddOp(txscript.OP_IF).
		AddData(comp33).
		AddOp(txscript.OP_CHECKSIGALTVERIFY).
		AddOp(txscript.OP_TRUE).
		AddOp(txscript.OP_ELSE).
		AddInt64(int64(csvBlocks)).
		AddOp(txscript.OP_CHECKSEQUENCEVERIFY).
		AddOp(txscript.OP_DROP).
		AddData(comp33).
		AddOp(txscript.OP_CHECKSIGALTVERIFY).
		AddOp(txscript.OP_TRUE).
		AddOp(txscript.OP_ENDIF)

	scr, err := b.Script()
	if err != nil {
		return nil, err
	}
	return scr, nil
}

// pkScriptAndAddrFromRedeem takes a raw redeem script and returns the P2SH pkScript (hex)
// and its human-readable address for the given Decred network params.
// Build P2SH pkScript+address from a redeem script.
// NOTE: stdaddr wants (scriptVersion, redeem, params), then use addr.PaymentScript().
func pkScriptAndAddrFromRedeem(redeem []byte, params stdaddr.AddressParams) (pkScriptHex, addr string, err error) {
	if len(redeem) == 0 {
		return "", "", fmt.Errorf("empty redeem script")
	}
	// v0 script version for Decred standard scripts.
	a, err := stdaddr.NewAddressScriptHash(0, redeem, params)
	if err != nil {
		return "", "", fmt.Errorf("NewAddressScriptHash: %w", err)
	}
	sv, pkScript := a.PaymentScript() // returns (version, script)
	if sv != 0 {
		return "", "", fmt.Errorf("unexpected script version %d", sv)
	}
	return hex.EncodeToString(pkScript), a.String(), nil
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
