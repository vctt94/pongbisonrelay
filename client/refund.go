package client

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/decred/dcrd/chaincfg/chainhash"
	"github.com/decred/dcrd/chaincfg/v3"
	"github.com/decred/dcrd/crypto/blake256"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/decred/dcrd/dcrutil/v4"
	"github.com/decred/dcrd/txscript/v4"
	"github.com/decred/dcrd/txscript/v4/stdaddr"
	"github.com/decred/dcrd/wire"
	pongbisonrelay "github.com/vctt94/pongbisonrelay"
)

const (
	DefaultBetAtoms = 100000000
)

// signSchnorrV0 deterministically signs the provided 32-byte message digest using
// EC-Schnorr-DCRv0 with RFC6979 and enforcing an even-Y nonce point. Returns
// the 65-byte signature: r_x (32B) || s (32B) || 0x01 (SigHashAll).
func signSchnorrV0(privHex, mHex string) ([]byte, error) {
	xb, err := hex.DecodeString(strings.TrimSpace(privHex))
	if err != nil {
		return nil, fmt.Errorf("bad privkey hex: %w", err)
	}
	if len(xb) != 32 {
		return nil, fmt.Errorf("privkey must be 32 bytes")
	}
	mb, err := hex.DecodeString(strings.TrimSpace(mHex))
	if err != nil || len(mb) != 32 {
		return nil, fmt.Errorf("bad m (need 32B hex)")
	}

	var x secp256k1.ModNScalar
	if overflow := x.SetByteSlice(xb); overflow || x.IsZero() {
		return nil, fmt.Errorf("invalid private key scalar")
	}

	extra := blake256.Sum256(pongbisonrelay.SchnorrV0ExtraTag[:])
	var version []byte
	for iter := uint32(0); ; iter++ {
		k := secp256k1.NonceRFC6979(xb, mb, extra[:], version, iter)
		if k == nil || k.IsZero() {
			continue
		}
		kb := k.Bytes()
		R := secp256k1.PrivKeyFromBytes(kb[:]).PubKey()
		cp := R.SerializeCompressed()
		if len(cp) != 33 || cp[0] != 0x02 { // enforce even-Y
			continue
		}
		var r32 [32]byte
		copy(r32[:], cp[1:33])
		h := blake256.Sum256(append(r32[:], mb...))
		var e secp256k1.ModNScalar
		if overflow := e.SetByteSlice(h[:]); overflow { // e >= n -> retry
			continue
		}
		var ex, negex, s secp256k1.ModNScalar
		ex.Set(&e)
		ex.Mul(&x)
		negex.NegateVal(&ex)
		s.Set(k)
		s.Add(&negex)

		sBytes := s.Bytes()
		sig := make([]byte, 0, 65)
		sig = append(sig, r32[:]...)
		sig = append(sig, sBytes[:]...)
		sig = append(sig, byte(txscript.SigHashAll))
		return sig, nil
	}
}

// BuildCSVRefundTx constructs a raw transaction hex that spends a single escrow
// deposit UTXO via the CSV timeout path back to a destination address.
//
// Parameters:
//   - privHex:      32-byte hex of the depositor's private key (controls A_c)
//   - utxoTxid:     funding txid
//   - utxoVout:     funding vout index
//   - utxoValue:    amount in atoms
//   - redeemHex:    redeem script hex for the deposit P2SH
//   - destAddr:     Decred address to receive refund
//   - feeAtoms:     fee to subtract in atoms (e.g., 20000)
//   - csvBlocks:    relative timelock blocks to satisfy (sequence >= csvBlocks)
//
// It returns the serialized transaction hex. The caller is responsible for
// broadcasting it once CSV has matured on-chain.
func BuildCSVRefundTx(privHex, utxoTxid string, utxoVout uint32, utxoValue uint64, redeemHex, destAddr string, feeAtoms uint64, csvBlocks uint32) (string, error) {
	// Decode destination address across known networks and get payment script.
	paramsList := []*chaincfg.Params{
		chaincfg.TestNet3Params(),
		chaincfg.MainNetParams(),
		chaincfg.SimNetParams(),
		chaincfg.RegNetParams(),
	}
	var pkScript []byte
	var lastErr error
	for _, p := range paramsList {
		addr, err := stdaddr.DecodeAddress(strings.TrimSpace(destAddr), p)
		if err != nil {
			lastErr = err
			continue
		}
		_, pk := addr.PaymentScript()
		pkScript = pk
		lastErr = nil
		break
	}
	if pkScript == nil {
		return "", fmt.Errorf("bad dest address: %v", lastErr)
	}

	if feeAtoms == 0 {
		feeAtoms = 20000
	}
	payout := int64(utxoValue) - int64(feeAtoms)
	if payout <= 0 {
		return "", fmt.Errorf("fee %d exceeds input %d", feeAtoms, utxoValue)
	}

	// Build tx: version 1, one input, one output.
	var h chainhash.Hash
	if err := chainhash.Decode(&h, strings.TrimSpace(utxoTxid)); err != nil {
		return "", fmt.Errorf("bad txid: %w", err)
	}
	tx := wire.NewMsgTx()
	tx.Version = 1
	tx.AddTxIn(&wire.TxIn{PreviousOutPoint: wire.OutPoint{Hash: h, Index: utxoVout}, ValueIn: int64(utxoValue)})
	// Set sequence to satisfy CSV.
	tx.TxIn[0].Sequence = csvBlocks
	tx.AddTxOut(&wire.TxOut{Value: payout, PkScript: pkScript})

	// Compute sighash against redeem script.
	redeem, err := hex.DecodeString(strings.TrimSpace(redeemHex))
	if err != nil {
		return "", fmt.Errorf("bad redeem hex: %w", err)
	}
	mBytes, err := txscript.CalcSignatureHash(redeem, txscript.SigHashAll, tx, 0, nil)
	if err != nil || len(mBytes) != 32 {
		return "", fmt.Errorf("sighash failed: %v", err)
	}
	mHex := hex.EncodeToString(mBytes)

	// Sign Schnorr v0 and build CSV branch scriptSig.
	sig65, err := signSchnorrV0(privHex, mHex)
	if err != nil {
		return "", err
	}
	sb := txscript.NewScriptBuilder()
	sb.AddData(sig65)
	sb.AddOp(txscript.OP_0) // select CSV branch (ELSE)
	sb.AddData(redeem)
	sigScript, err := sb.Script()
	if err != nil {
		return "", fmt.Errorf("build scriptSig: %w", err)
	}
	tx.TxIn[0].SignatureScript = sigScript

	// Optional local VM verification (P2SH wrapper):
	sh := dcrutil.Hash160(redeem)
	p2sh, err := txscript.NewScriptBuilder().AddOp(txscript.OP_HASH160).AddData(sh).AddOp(txscript.OP_EQUAL).Script()
	if err != nil {
		return "", fmt.Errorf("build pkScript: %w", err)
	}
	vm, err := txscript.NewEngine(p2sh, tx, 0, 0, 0, nil)
	if err != nil {
		return "", fmt.Errorf("engine init: %w", err)
	}
	if err := vm.Execute(); err != nil {
		return "", fmt.Errorf("local VM verify failed: %w", err)
	}

	var out bytes.Buffer
	_ = tx.Serialize(&out)
	return hex.EncodeToString(out.Bytes()), nil
}
