package server

import (
	"encoding/hex"
	"fmt"

	"github.com/decred/dcrd/crypto/blake256"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/decred/dcrd/txscript/v4"
	"github.com/vctt94/pong-bisonrelay/ponggame"
	"github.com/vctt94/pong-bisonrelay/pongrpc/grpc/pong"
)

// Unit conversion constants and helpers between matoms, atoms, and DCR.
const (
	AtomsPerDCR = int64(100_000_000) // 1e8 atoms per DCR
)

func atomsToDCR(a int64) float64 { return float64(a) / float64(AtomsPerDCR) }

func buildPerDepositorRedeemScript(comp33 []byte, csvBlocks uint32) ([]byte, error) {
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

// escrowForRoomPlayer returns the escrow session bound to (wrID, ownerUID)
// via the roomEscrows mapping. It does not attempt to pick a "newest" escrow.
func (s *Server) escrowForRoomPlayer(wrID, ownerUID string) *escrowSession {
	s.RLock()
	defer s.RUnlock()
	m := s.roomEscrows[wrID]
	if m == nil {
		return nil
	}
	escrowID := m[ownerUID]
	if escrowID == "" {
		return nil
	}
	return s.escrows[escrowID]
}

// ensureBoundFunding either binds the canonical funding input if not yet bound
// (requiring exactly one UTXO), or verifies the previously-bound input still
// exists and that no extra deposits were made.
func (s *Server) ensureBoundFunding(es *escrowSession) error {
	es.mu.Lock()
	defer es.mu.Unlock()

	// Must have a watcher snapshot that says "funded".
	if !es.latest.OK || es.latest.UTXOCount == 0 {
		return fmt.Errorf("escrow not yet funded")
	}

	// Normalize current UTXO list.
	var current []*pong.EscrowUTXO
	for _, u := range es.lastUTXOs {
		if u != nil && u.Txid != "" {
			current = append(current, u)
		}
	}
	// If we haven't cached details yet, we can't bind.
	if len(current) == 0 {
		return fmt.Errorf("escrow UTXO details not available yet; wait for index")
	}

	if es.boundInputID == "" {
		// Not yet bound: require exactly one deposit to avoid ambiguity.
		if len(current) != 1 || es.latest.UTXOCount != 1 {
			return fmt.Errorf("multiple deposits detected (%d); escrow requires exactly one funding UTXO", len(current))
		}
		u := current[0]
		boundID := fmt.Sprintf("%s:%d", u.Txid, u.Vout)
		es.boundInputID = boundID
		es.boundInput = u
		// Notify outside of the lock to avoid deadlocks.
		p := es.player
		go func(pid string, pl *ponggame.Player) {
			if pl != nil {
				s.notify(pl, fmt.Sprintf("Escrow bound to %s. You can proceed.", boundID))
			}
		}(boundID, p)
		return nil
	}

	// Already bound: verify the exact input is still present.
	var found *pong.EscrowUTXO
	for _, u := range current {
		if fmt.Sprintf("%s:%d", u.Txid, u.Vout) == es.boundInputID {
			found = u
			break
		}
	}
	if found == nil {
		// The canonical input disappeared (reorg/spent?) -> invalidate.
		return fmt.Errorf("bound funding UTXO %s not present", es.boundInputID)
	}
	es.boundInput = found

	// Enforce no extra deposits beyond the bound one.
	if len(current) != 1 || es.latest.UTXOCount != 1 {
		return fmt.Errorf("unexpected additional deposits (%d); only the bound %s is allowed", len(current), es.boundInputID)
	}

	return nil
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
