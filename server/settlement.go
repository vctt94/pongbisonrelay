package server

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"sort"

	"github.com/decred/dcrd/chaincfg/chainhash"
	"github.com/decred/dcrd/chaincfg/v3"
	"github.com/decred/dcrd/crypto/blake256"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/decred/dcrd/txscript/v4"
	"github.com/decred/dcrd/txscript/v4/stdaddr"
	"github.com/decred/dcrd/txscript/v4/stdscript"
	"github.com/decred/dcrd/wire"
	"github.com/vctt94/pong-bisonrelay/ponggame"
	"github.com/vctt94/pong-bisonrelay/pongrpc/grpc/pong"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

// buildP2PKAltScript builds a payout script that pays to a single compressed
// pubkey using OP_CHECKSIGALT with the Schnorr-secp256k1 signature type (2).
// Script: <Ac> 2 OP_CHECKSIGALT
func buildP2PKAltScript(comp33 []byte) ([]byte, error) {
	if len(comp33) != 33 {
		return nil, fmt.Errorf("need 33-byte compressed pubkey")
	}
	b := txscript.NewScriptBuilder()
	b.AddData(comp33).
		AddInt64(2).
		AddOp(txscript.OP_CHECKSIGALT)
	return b.Script()
}

// pkScriptAndAddrFromRedeem takes a raw redeem script and returns the P2SH pkScript (hex)
// and its human-readable address for the given Decred network params.
// Build P2SH pkScript+address from a redeem script.
// NOTE: stdaddr wants (scriptVersion, redeem, params), then use addr.PaymentScript().
func pkScriptAndAddrFromRedeem(redeem []byte, params stdaddr.AddressParams) (string, string, error) {
	a, err := stdaddr.NewAddressScriptHash(0, redeem, params)
	if err != nil {
		return "", "", err
	}
	_, pk := a.PaymentScript()
	return hex.EncodeToString(pk), a.String(), nil
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

// verifyAndStorePresig verifies s'·G + e·X + T == R' and persists the context.
func (s *Server) verifyAndStorePresig(X *secp256k1.PublicKey, ps *pong.PreSig, in *pong.NeedPreSigs_PerInput, draftHex string, winner *escrowSession, branch int32) error {
	if len(ps.RprimeCompressed) != 33 || len(ps.Sprime32) != 32 {
		return status.Error(codes.InvalidArgument, "bad presig sizes")
	}
	if ps.RprimeCompressed[0] != 0x02 {
		return status.Error(codes.InvalidArgument, "R' must be even-Y (0x02)")
	}
	Rp, err := secp256k1.ParsePubKey(ps.RprimeCompressed)
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "parse R': %v", err)
	}
	mh, err := hex.DecodeString(in.MHex)
	if err != nil || len(mh) != 32 {
		return status.Error(codes.InvalidArgument, "bad m_hex")
	}
	var r32 [32]byte
	copy(r32[:], ps.RprimeCompressed[1:33])
	ch := blake256.Sum256(append(r32[:], mh...))
	var e secp256k1.ModNScalar
	if overflow := e.SetByteSlice(ch[:]); overflow {
		return status.Error(codes.InvalidArgument, "e overflow")
	}
	var sp secp256k1.ModNScalar
	if overflow := sp.SetByteSlice(ps.Sprime32); overflow {
		return status.Error(codes.InvalidArgument, "s' overflow")
	}
	// Compute k·G = s'·G + e·X
	spb := sp.Bytes()
	spk := secp256k1.PrivKeyFromBytes(spb[:]).PubKey()
	var Xj secp256k1.JacobianPoint
	X.AsJacobian(&Xj)
	var out secp256k1.JacobianPoint
	secp256k1.ScalarMultNonConst(&e, &Xj, &out)
	out.ToAffine()
	Ex := secp256k1.NewPublicKey(&out.X, &out.Y)
	lhs1, err := addPoints(spk, Ex)
	if err != nil {
		return status.Error(codes.InvalidArgument, "lhs infinity")
	}
	// Compute R' - T
	T, err := secp256k1.ParsePubKey(in.TCompressed)
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "parse T: %v", err)
	}
	var tj secp256k1.JacobianPoint
	T.AsJacobian(&tj)
	tj.Y.Negate(1)
	var rj secp256k1.JacobianPoint
	Rp.AsJacobian(&rj)
	var diff secp256k1.JacobianPoint
	secp256k1.AddNonConst(&rj, &tj, &diff)
	if diff.Z.IsZero() {
		return status.Error(codes.InvalidArgument, "R'-T infinity")
	}
	diff.ToAffine()
	RminusT := secp256k1.NewPublicKey(&diff.X, &diff.Y)
	if !bytes.Equal(lhs1.SerializeCompressed(), RminusT.SerializeCompressed()) {
		return status.Error(codes.InvalidArgument, "adaptor relation failed")
	}
	// Store context under the client's escrow session for later finalization
	s.Lock()
	if winner.preSign == nil {
		winner.preSign = make(map[string]*PreSignCtx)
	}
	winner.preSign[in.InputId] = &PreSignCtx{
		InputID:          in.InputId,
		RedeemScriptHex:  in.RedeemScriptHex,
		DraftHex:         draftHex,
		MHex:             in.MHex,
		RPrimeCompressed: append([]byte(nil), ps.RprimeCompressed...),
		SPrime32:         append([]byte(nil), ps.Sprime32...),
		TCompressed:      append([]byte(nil), in.TCompressed...),
		WinnerUID:        winner.ownerUID,
		Branch:           branch,
	}
	s.Unlock()
	return nil
}

// buildTwoInputDrafts builds two drafts spending one UTXO from each escrow and
// paying the sum minus fee to winner's P2PK-alt address. Inputs are deterministically
// ordered by (txid asc, vout asc). Branch 0 pays to a, branch 1 pays to b.
func (s *Server) buildTwoInputDrafts(a *escrowSession, au *pong.EscrowUTXO, b *escrowSession, bu *pong.EscrowUTXO) (*twoBranchDrafts, error) {
	// Deterministic order
	type inCtx struct {
		utxo *pong.EscrowUTXO
		es   *escrowSession
	}
	ins := []inCtx{{utxo: au, es: a}, {utxo: bu, es: b}}
	sort.Slice(ins, func(i, j int) bool {
		if ins[i].utxo.Txid == ins[j].utxo.Txid {
			return ins[i].utxo.Vout < ins[j].utxo.Vout
		}
		return ins[i].utxo.Txid < ins[j].utxo.Txid
	})

	build := func(payTo *escrowSession) (string, []*pong.NeedPreSigs_PerInput, error) {
		tx := wire.NewMsgTx()
		tx.Version = 1
		total := int64(0)
		for _, ic := range ins {
			var h chainhash.Hash
			if err := chainhash.Decode(&h, ic.utxo.Txid); err != nil {
				return "", nil, fmt.Errorf("bad utxo txid: %v", err)
			}
			tx.AddTxIn(&wire.TxIn{PreviousOutPoint: wire.OutPoint{Hash: h, Index: ic.utxo.Vout}, ValueIn: int64(ic.utxo.Value)})
			total += int64(ic.utxo.Value)
		}
		feeAtoms := s.pocFeeAtoms
		if feeAtoms == 0 {
			feeAtoms = 20000
		}
		payout := total - int64(feeAtoms)
		if payout <= 0 {
			return "", nil, fmt.Errorf("sum too small to pay fee")
		}
		payKey := payTo.payoutPubkey
		if len(payKey) != 33 {
			payKey = payTo.compPubkey
		}
		pkAlt, err := buildP2PKAltScript(payKey)
		if err != nil {
			return "", nil, err
		}
		tx.AddTxOut(&wire.TxOut{Value: payout, PkScript: pkAlt})

		// Derive a single adaptor point T per branch and reuse across inputs so a single
		// gamma reveal finalizes all inputs of the winning branch.
		branch := int32(0)
		if payTo == b {
			branch = 1
		}
		const pocServerPrivHex = "11ee22dd33cc44bb55aa66ff77ee88dd99cc00bbaa11223344556677889900aa"
		_, TCompHexBranch := deriveAdaptorGamma("", fmt.Sprintf("branch-%d", branch), branch, fmt.Sprintf("branch-%d", branch), pocServerPrivHex)
		TcompBranch, _ := hex.DecodeString(TCompHexBranch)

		inputs := make([]*pong.NeedPreSigs_PerInput, 0, len(ins))
		for idx, ic := range ins {
			redeemB, _ := hex.DecodeString(ic.es.redeemScriptHex)
			mBytes, err := txscript.CalcSignatureHash(redeemB, txscript.SigHashAll, tx, idx, nil)
			if err != nil || len(mBytes) != 32 {
				return "", nil, fmt.Errorf("sighash for %s:%d failed", ic.utxo.Txid, ic.utxo.Vout)
			}
			inputID := fmt.Sprintf("%s:%d", ic.utxo.Txid, ic.utxo.Vout)
			inputs = append(inputs, &pong.NeedPreSigs_PerInput{
				InputId:         inputID,
				RedeemScriptHex: ic.es.redeemScriptHex,
				MHex:            hex.EncodeToString(mBytes),
				TCompressed:     TcompBranch,
			})
		}
		var buf bytes.Buffer
		_ = tx.Serialize(&buf)
		return hex.EncodeToString(buf.Bytes()), inputs, nil
	}

	dhA, inA, err := build(a)
	if err != nil {
		return nil, err
	}
	dhB, inB, err := build(b)
	if err != nil {
		return nil, err
	}
	return &twoBranchDrafts{DraftHexA: dhA, InputsA: inA, DraftHexB: dhB, InputsB: inB}, nil
}

// --- Helpers ---------------------------------------------------------------

// notify sends a simple text notification if the stream is live.
// (Optional: de-dupe by caching the last message per player to avoid spam.)
func (s *Server) notify(p *ponggame.Player, msg string) {
	if p != nil && p.NotifierStream != nil {
		if p.ID != nil {
			s.log.Debugf("notify: to=%s msg=%q", p.ID.String(), msg)
		} else {
			s.log.Debugf("notify: to=<nil-id> msg=%q", msg)
		}
		if err := p.NotifierStream.Send(&pong.NtfnStreamResponse{
			NotificationType: pong.NotificationType_MESSAGE,
			Message:          msg,
		}); err != nil {
			var idStr string
			if p.ID != nil {
				idStr = p.ID.String()
			} else {
				idStr = "<nil-id>"
			}
			s.log.Warnf("notify: failed to deliver to %s: %v", idStr, err)
		}
	} else {
		s.log.Debugf("notify: dropped (player or stream nil) msg=%q", msg)
	}
}

func (s *Server) trackEscrow(ctx context.Context, es *escrowSession, ch <-chan DepositUpdate) {
	for {
		select {
		case <-ctx.Done():
			return
		case u, ok := <-ch:
			if !ok {
				return
			}
			es.mu.Lock()
			prev := es.latest
			es.latest = u
			// keep a small cache if you use it elsewhere
			if u.UTXOCount > 0 && es.pkScriptHex != "" {
				// Cache a shallow copy of UTXOs for settlement handshake.
				if len(u.UTXOs) > 0 {
					es.lastUTXOs = append(es.lastUTXOs[:0], u.UTXOs...)
				}
			}
			es.mu.Unlock()

			// Optional UX nudges. Notify on first funding sighting and on confirmation.
			if es.ownerUID != "" {
				// First time we see any UTXO for this escrow.
				if prev.UTXOCount == 0 && u.UTXOCount > 0 {
					if u.Confs == 0 {
						s.log.Debugf("trackEscrow: mempool seen for owner=%s pk=%s utxos=%d", es.ownerUID, u.PkScriptHex, u.UTXOCount)
						s.notify(es.player, "Deposit seen in mempool. Waiting confirmations.")
					} else {
						s.log.Debugf("trackEscrow: confirmed on first sight for owner=%s pk=%s confs=%d", es.ownerUID, u.PkScriptHex, u.Confs)
						s.notify(es.player, "Deposit confirmed.")
					}
				} else if prev.Confs < 1 && u.Confs >= 1 && u.UTXOCount > 0 {
					s.log.Debugf("trackEscrow: transitioned to confirmed for owner=%s pk=%s confs=%d", es.ownerUID, u.PkScriptHex, u.Confs)
					// Transition to confirmed after already seeing funding.
					s.notify(es.player, "Deposit confirmed.")
				}
			}
		}
	}
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

// makeSettleInputFromEscrow builds a minimal settlement context from a bound escrow.
// It requires es.boundInputID and es.boundInput to be set (use ensureBoundFunding first).
func (s *Server) makeSettleInputFromEscrow(es *escrowSession) (*SettleInput, error) {
	if es == nil {
		return nil, fmt.Errorf("nil escrow session")
	}
	es.mu.RLock()
	defer es.mu.RUnlock()
	if es.boundInputID == "" || es.boundInput == nil {
		return nil, fmt.Errorf("escrow not bound to a specific funding input yet")
	}
	payout := es.payoutPubkey
	if len(payout) != 33 {
		payout = es.compPubkey
	}
	return &SettleInput{
		EscrowID:        es.escrowID,
		OwnerUID:        es.ownerUID,
		InputID:         es.boundInputID,
		UTXO:            es.boundInput,
		RedeemScriptHex: es.redeemScriptHex,
		PayoutPubkey:    append([]byte(nil), payout...),
	}, nil
}

// resolveFundedEscrow returns the escrow after strictly validating:
//   - escrow exists and is owned by ownerUID
//   - funding is bound to a specific outpoint (txid:vout)
//   - that exact outpoint is present and is the *only* deposit
func (s *Server) resolveFundedEscrow(ownerUID, escrowID string) (*escrowSession, error) {
	if escrowID == "" {
		return nil, fmt.Errorf("escrow ID is required")
	}

	// Exact escrow lookup & ownership check.
	s.RLock()
	es := s.escrows[escrowID]
	s.RUnlock()
	if es == nil {
		return nil, fmt.Errorf("escrow not found: %s", escrowID)
	}
	if es.ownerUID != ownerUID {
		return nil, fmt.Errorf("escrow not owned by user")
	}

	// Bind (if first time) or verify the canonical funding input.
	if err := s.ensureBoundFunding(es); err != nil {
		return nil, err
	}

	// Double-check the bound identity is set (belt-and-suspenders).
	es.mu.RLock()
	bound := es.boundInputID
	es.mu.RUnlock()
	if bound == "" {
		return nil, fmt.Errorf("escrow not bound to a specific funding input yet")
	}

	return es, nil
}
