package server

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	crand "crypto/rand"

	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/decred/dcrd/chaincfg/chainhash"
	"github.com/decred/dcrd/chaincfg/v3"
	"github.com/decred/dcrd/crypto/blake256"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/decred/dcrd/txscript/v4"
	"github.com/decred/dcrd/txscript/v4/stdaddr"
	"github.com/decred/dcrd/wire"
	pongbisonrelay "github.com/vctt94/pong-bisonrelay"
	"github.com/vctt94/pong-bisonrelay/ponggame"
	"github.com/vctt94/pong-bisonrelay/pongrpc/grpc/pong"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// SettleInput is the only thing the presign needs from funding.
// It binds identity (escrowID, ownerUID) + the exact outpoint and script.
type SettleInput struct {
	EscrowID        string           // for storage of presigs
	OwnerUID        string           // for deterministic (a,b) ordering
	InputID         string           // "txid:vout" (must be the BOUND outpoint)
	UTXO            *pong.EscrowUTXO // must include Txid, Vout, AmountAtoms
	RedeemScriptHex string           // exact redeem used to spend UTXO
	PayoutPubkey    []byte           // 33-byte compressed pubkey for payout path
}

// twoBranchDrafts holds the fully serialized drafts for A-wins and B-wins,
// and their per-input presign data (m and T) for each input using its own redeem.
// Branch mapping: 0 = pay to A, 1 = pay to B.
type twoBranchDrafts struct {
	DraftHexA string
	InputsA   []*pong.NeedPreSigs_PerInput
	DraftHexB string
	InputsB   []*pong.NeedPreSigs_PerInput
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
			if es.pkScriptHex != "" {
				if u.UTXOCount > 0 && len(u.UTXOs) > 0 {
					// Cache a shallow copy of current unspent UTXOs.
					es.lastUTXOs = append(es.lastUTXOs[:0], u.UTXOs...)
				} else {
					// No unspent outputs remain; clear cache to prevent stale bindings.
					es.lastUTXOs = es.lastUTXOs[:0]
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

func (s *Server) OpenEscrow(ctx context.Context, req *pong.OpenEscrowRequest) (*pong.OpenEscrowResponse, error) {
	if req.OwnerUid == "" {
		return nil, status.Errorf(codes.InvalidArgument, "missing owner uid")
	}
	if req.BetAtoms == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "missing bet atoms")
	}
	if req.CsvBlocks == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "missing csv blocks")
	}
	if len(req.CompPubkey) != 33 {
		return nil, status.Errorf(codes.InvalidArgument, "invalid compressed pubkey")
	}
	if len(req.PayoutPubkey) != 33 {
		return nil, status.Errorf(codes.InvalidArgument, "invalid payout address")
	}

	// Canonicalize compressed pubkey.
	pub, err := secp256k1.ParsePubKey(req.CompPubkey)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "bad compressed pubkey: %v", err)
	}
	comp := pub.SerializeCompressed()
	// Canonicalize payout compressed pubkey.
	pp, err := secp256k1.ParsePubKey(req.PayoutPubkey)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "bad payout compressed pubkey: %v", err)
	}
	payoutPubkey := pp.SerializeCompressed()

	s.Lock()
	defer s.Unlock()

	if s.escrows == nil {
		s.escrows = make(map[string]*escrowSession)
	}

	// Build depositor-only redeem script (no opponent key).
	redeem, err := pongbisonrelay.BuildPerDepositorRedeemScript(comp, req.CsvBlocks)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "build redeem: %v", err)
	}

	params := chaincfg.TestNet3Params() // TODO: use server-configured params

	pkScriptHex, addr, err := pkScriptAndAddrFromRedeem(redeem, params)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "derive address: %v", err)
	}

	// Convert owner UID string to zkidentity.ShortID before looking up player.
	var ownerID zkidentity.ShortID
	if err := ownerID.FromString(req.OwnerUid); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid owner uid: %v", err)
	}
	player := s.gameManager.PlayerSessions.GetPlayer(ownerID)

	// Secure random escrow ID.
	var rnd [8]byte
	if _, err := crand.Read(rnd[:]); err != nil {
		return nil, status.Errorf(codes.Internal, "rand: %v", err)
	}
	eid := "e" + hex.EncodeToString(rnd[:])

	es := &escrowSession{
		escrowID:        eid,
		ownerUID:        req.OwnerUid,
		compPubkey:      append([]byte(nil), comp...),
		payoutPubkey:    append([]byte(nil), payoutPubkey...),
		betAtoms:        req.BetAtoms,
		csvBlocks:       req.CsvBlocks,
		redeemScriptHex: hex.EncodeToString(redeem),
		pkScriptHex:     pkScriptHex,
		player:          player,
	}
	s.escrows[eid] = es

	// Subscribe right here; trackEscrow updates es.latest and es.lastUTXOs.
	if s.watcher != nil {
		ch, unsub := s.watcher.Subscribe(es.pkScriptHex)
		es.unsubW = unsub
		// Run trackEscrow on a long-lived context independent of this RPC's ctx.
		trackCtx, cancel := context.WithCancel(context.Background())
		es.cancelTrack = cancel
		go s.trackEscrow(trackCtx, es, ch)
	}

	return &pong.OpenEscrowResponse{
		EscrowId:       eid,
		DepositAddress: addr,
		PkScriptHex:    pkScriptHex,
	}, nil
}

func (s *Server) SettlementStream(stream pong.PongReferee_SettlementStreamServer) error {
	// Step 1: receive HELLO
	first, err := stream.Recv()
	if err != nil || first.GetHello() == nil {
		return status.Errorf(codes.InvalidArgument, "expected HELLO: %v", err)
	}
	hel := first.GetHello()
	if len(hel.CompPubkey) != 33 {
		return status.Error(codes.InvalidArgument, "bad comp pubkey")
	}
	// Canonicalize client pubkey
	X, err := secp256k1.ParsePubKey(hel.CompPubkey)
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "parse pubkey: %v", err)
	}
	_ = X.SerializeCompressed()

	// New: room-scoped settlement. Expect match_id as "<wrID>|<hostUID>".
	matchID := first.GetMatchId()
	if matchID == "" {
		matchID = hel.GetMatchId()
	}
	if matchID == "" {
		return status.Error(codes.InvalidArgument, "missing match_id in ClientMsg/Hello")
	}
	var wrID, hostUID string
	parts := strings.SplitN(matchID, "|", 2)
	wrID = strings.TrimSpace(parts[0])
	if wrID == "" {
		return status.Error(codes.InvalidArgument, "invalid match_id: missing room id")
	}

	wr := s.gameManager.GetWaitingRoom(wrID)
	if wr == nil {
		return status.Errorf(codes.FailedPrecondition, "waiting room %s not found", wrID)
	}
	if len(parts) == 2 {
		hostUID = strings.TrimSpace(parts[1])
	}
	if hostUID == "" && wr.HostID != nil {
		hostUID = wr.HostID.String()
	}
	if hostUID == "" {
		return status.Error(codes.FailedPrecondition, "could not determine host uid for match")
	}

	// Determine caller by matching HELLO comp pubkey to an escrow bound in this room.
	xBytes := X.SerializeCompressed()
	var callerUID string
	s.RLock()
	if m := s.roomEscrows[wrID]; m != nil {
		for uid, eid := range m {
			if es := s.escrows[eid]; es != nil && len(es.compPubkey) == 33 {
				if bytes.Equal(es.compPubkey, xBytes) {
					callerUID = uid
					break
				}
			}
		}
	}
	s.RUnlock()
	if callerUID == "" {
		return status.Error(codes.FailedPrecondition, "caller escrow not found in room")
	}

	var host zkidentity.ShortID
	if err := host.FromString(hostUID); err != nil {
		return status.Error(codes.InvalidArgument, "bad host uid in match_id")
	}

	s.log.Debugf("SettlementStream: match_id=%q wr=%s host=%s caller=%s", matchID, wrID, hostUID, callerUID)

	var caller zkidentity.ShortID
	if err := caller.FromString(callerUID); err != nil {
		return status.Error(codes.InvalidArgument, "bad caller uid")
	}
	return s.settlementStreamForRoom(stream, X, wr, caller, host)
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

// settlementStreamForRoom looks up each player's room-bound escrow, ensures bound funding,
// builds SettleInput contexts, and calls settlementStreamTwoInputsCore.
func (s *Server) settlementStreamForRoom(
	stream pong.PongReferee_SettlementStreamServer,
	X *secp256k1.PublicKey,
	wr *ponggame.WaitingRoom,
	caller zkidentity.ShortID,
	host zkidentity.ShortID,
) error {
	if wr == nil {
		return fmt.Errorf("waiting room is nil")
	}

	// Snapshot players without requiring Ready() here.
	wr.Lock()
	playersSnap := append([]*ponggame.Player(nil), wr.Players...)
	wr.Unlock()

	if len(playersSnap) != 2 {
		return fmt.Errorf("need exactly 2 players in room (have %d)", len(playersSnap))
	}

	// Identify my/opp by caller id.
	var myPlayer, oppPlayer *ponggame.Player
	if playersSnap[0].ID.String() == caller.String() {
		myPlayer, oppPlayer = playersSnap[0], playersSnap[1]
	} else if playersSnap[1].ID.String() == caller.String() {
		myPlayer, oppPlayer = playersSnap[1], playersSnap[0]
	} else {
		return fmt.Errorf("caller %s not in waiting room %s", caller, wr.ID)
	}

	// Resolve room-bound escrows.
	myEscrow := s.escrowForRoomPlayer(wr.ID, myPlayer.ID.String())
	oppEscrow := s.escrowForRoomPlayer(wr.ID, oppPlayer.ID.String())
	if myEscrow == nil || oppEscrow == nil {
		return fmt.Errorf("room-bound escrows missing (my=%v opp=%v)", myEscrow != nil, oppEscrow != nil)
	}

	// Bound funding only (identity over count).
	if err := s.ensureBoundFunding(myEscrow); err != nil {
		return fmt.Errorf("my funding not bound: %w", err)
	}
	if err := s.ensureBoundFunding(oppEscrow); err != nil {
		return fmt.Errorf("opponent funding not bound: %w", err)
	}

	myIn, err := s.makeSettleInputFromEscrow(myEscrow)
	if err != nil {
		return err
	}
	oppIn, err := s.makeSettleInputFromEscrow(oppEscrow)
	if err != nil {
		return err
	}

	// Anchor a/b ordering to host; pass host uid string
	return s.settlementStreamTwoInputs(stream, X, myEscrow, oppEscrow, myIn.UTXO, oppIn.UTXO, host.String())
}

// buildP2PKScript builds a standard ECDSA P2PK payout script that pays to a
// single compressed pubkey using OP_CHECKSIG.
// Script: <Ac> OP_CHECKSIG
func buildP2PKScript(comp33 []byte) ([]byte, error) {
	if len(comp33) != 33 {
		return nil, fmt.Errorf("need 33-byte compressed pubkey")
	}
	b := txscript.NewScriptBuilder()
	b.AddData(comp33).
		AddOp(txscript.OP_CHECKSIG)
	return b.Script()
}

// buildTwoInputDrafts builds two drafts spending one UTXO from each escrow and
// paying the sum minus fee to winner's P2PK address. Inputs are deterministically
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
		pkScript, err := buildP2PKScript(payKey)
		if err != nil {
			return "", nil, err
		}
		tx.AddTxOut(&wire.TxOut{Value: payout, PkScript: pkScript})

		// Derive a single adaptor point T per branch and reuse across inputs so a single
		// gamma reveal finalizes all inputs of the winning branch.
		branch := int32(0)
		if payTo == b {
			branch = 1
		}
		// Derive branch-wide adaptor point using server's configured secret.
		serverSecret := s.adaptorSecret
		if strings.TrimSpace(serverSecret) == "" {
			return "", nil, status.Error(codes.FailedPrecondition, "server adaptor secret not configured")
		}
		// Bind to a stable tag for the branch.
		_, TCompHexBranch := pongbisonrelay.DeriveAdaptorGamma("", fmt.Sprintf("branch-%d", branch), branch, fmt.Sprintf("branch-%d", branch), serverSecret)
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

// verifyAndStorePresig verifies s'·G + e·X + T == R' and persists the context.
func (s *Server) verifyAndStorePresig(X *secp256k1.PublicKey, ps *pong.PreSig, in *pong.NeedPreSigs_PerInput, draftHex string, winner *escrowSession, branch int32) error {
	if len(ps.RLineCompressed) != 33 || len(ps.SLine32) != 32 {
		return status.Error(codes.InvalidArgument, "bad presig sizes")
	}
	if ps.RLineCompressed[0] != 0x02 {
		return status.Error(codes.InvalidArgument, "R' must be even-Y (0x02)")
	}
	Rp, err := secp256k1.ParsePubKey(ps.RLineCompressed)
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "parse R': %v", err)
	}
	mh, err := hex.DecodeString(in.MHex)
	if err != nil || len(mh) != 32 {
		return status.Error(codes.InvalidArgument, "bad m_hex")
	}
	var r32 [32]byte
	copy(r32[:], ps.RLineCompressed[1:33])
	ch := blake256.Sum256(append(r32[:], mh...))
	var e secp256k1.ModNScalar
	if overflow := e.SetByteSlice(ch[:]); overflow {
		return status.Error(codes.InvalidArgument, "e overflow")
	}
	var sp secp256k1.ModNScalar
	if overflow := sp.SetByteSlice(ps.SLine32); overflow {
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
	lhs1, err := pongbisonrelay.AddPoints(spk, Ex)
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
		InputID:         in.InputId,
		RedeemScriptHex: in.RedeemScriptHex,
		DraftHex:        draftHex,
		MHex:            in.MHex,
		RLineCompressed: append([]byte(nil), ps.RLineCompressed...),
		SLine32:         append([]byte(nil), ps.SLine32...),
		TCompressed:     append([]byte(nil), in.TCompressed...),
		WinnerUID:       winner.ownerUID,
		Branch:          branch,
	}
	s.Unlock()
	return nil
}

// settlementStreamTwoInputs performs the two-branch, two-input presign handshake
// with a single client. For each branch (A wins, B wins), it sends a presign
// request containing only the client's own input. The opponent performs the
// same handshake on their stream. The server persists both branches' presigs
// and drafts for later finalization delivery to the winner.
func (s *Server) settlementStreamTwoInputs(
	stream pong.PongReferee_SettlementStreamServer,
	X *secp256k1.PublicKey,
	myEscrow, oppEscrow *escrowSession,
	myUTXO, oppUTXO *pong.EscrowUTXO,
	hostUID string,
) error {
	// 0) Enforce identity: the UTXOs passed in MUST be the bound inputs.
	if err := s.ensureBoundFunding(myEscrow); err != nil {
		return status.Errorf(codes.FailedPrecondition, "my funding not bound: %v", err)
	}
	if err := s.ensureBoundFunding(oppEscrow); err != nil {
		return status.Errorf(codes.FailedPrecondition, "opponent funding not bound: %v", err)
	}
	myBound := ""
	oppBound := ""
	myEscrow.mu.RLock()
	myBound = myEscrow.boundInputID
	myEscrow.mu.RUnlock()
	oppEscrow.mu.RLock()
	oppBound = oppEscrow.boundInputID
	oppEscrow.mu.RUnlock()

	myWant := fmt.Sprintf("%s:%d", myUTXO.Txid, myUTXO.Vout)
	oppWant := fmt.Sprintf("%s:%d", oppUTXO.Txid, oppUTXO.Vout)
	if myBound != myWant {
		return status.Errorf(codes.InvalidArgument, "my input mismatch: bound=%s != passed=%s", myBound, myWant)
	}
	if oppBound != oppWant {
		return status.Errorf(codes.InvalidArgument, "opp input mismatch: bound=%s != passed=%s", oppBound, oppWant)
	}

	// 1) Anchor (a,b) ordering to host: a = host, b = other.
	var aEscrow, bEscrow *escrowSession
	var aUTXO, bUTXO *pong.EscrowUTXO
	if myEscrow.ownerUID == hostUID {
		aEscrow, aUTXO = myEscrow, myUTXO
		bEscrow, bUTXO = oppEscrow, oppUTXO
	} else if oppEscrow.ownerUID == hostUID {
		aEscrow, aUTXO = oppEscrow, oppUTXO
		bEscrow, bUTXO = myEscrow, myUTXO
	} else {
		return status.Error(codes.FailedPrecondition, "host escrow not bound in this room")
	}

	// 2) Build drafts (your existing builder).
	drafts, err := s.buildTwoInputDrafts(aEscrow, aUTXO, bEscrow, bUTXO)
	if err != nil {
		return status.Errorf(codes.Internal, "build two-input drafts: %v", err)
	}

	// 3) Sanity log + assert each branch actually contains the exact input IDs.
	aID := fmt.Sprintf("%s:%d", aUTXO.Txid, aUTXO.Vout)
	bID := fmt.Sprintf("%s:%d", bUTXO.Txid, bUTXO.Vout)
	s.log.Debugf("settle: a.Owner=%s a.InputID=%s  b.Owner=%s b.InputID=%s  (branch 0 pays a)",
		aEscrow.ownerUID, aID, bEscrow.ownerUID, bID)

	ensureHas := func(who, want string, all []*pong.NeedPreSigs_PerInput) error {
		for _, in := range all {
			if in.InputId == want {
				if len(in.MHex) != 64 { // 32-byte digest in hex
					return fmt.Errorf("%s: mHex wrong size (%d chars)", who, len(in.MHex))
				}
				if len(in.TCompressed) != 33 {
					return fmt.Errorf("%s: TCompressed not 33 bytes", who)
				}
				return nil
			}
		}
		return fmt.Errorf("%s: input %s not present in draft", who, want)
	}
	// Each branch should include per-input data for BOTH inputs.
	if err := ensureHas("A-inputs(a)", aID, drafts.InputsA); err != nil {
		return status.Errorf(codes.Internal, "draft sanity: %v", err)
	}
	if err := ensureHas("A-inputs(b)", bID, drafts.InputsA); err != nil {
		return status.Errorf(codes.Internal, "draft sanity: %v", err)
	}
	if err := ensureHas("B-inputs(a)", aID, drafts.InputsB); err != nil {
		return status.Errorf(codes.Internal, "draft sanity: %v", err)
	}
	if err := ensureHas("B-inputs(b)", bID, drafts.InputsB); err != nil {
		return status.Errorf(codes.Internal, "draft sanity: %v", err)
	}

	// 4) Helper to pick THIS client's per-input by exact InputID.
	pickMine := func(all []*pong.NeedPreSigs_PerInput) (*pong.NeedPreSigs_PerInput, error) {
		want := fmt.Sprintf("%s:%d", myUTXO.Txid, myUTXO.Vout)
		for _, in := range all {
			if in.InputId == want {
				return in, nil
			}
		}
		return nil, fmt.Errorf("client input %s not found in draft inputs", want)
	}

	// 5) Compute which branch corresponds to the caller winning.
	// Branch 0 = host wins (a), Branch 1 = non-host wins (b).
	clientWinsBranch := int32(0)
	if aEscrow != myEscrow {
		clientWinsBranch = 1
	}

	// 6) WIN branch handshake (unchanged, but now we’re sure InputId matches).
	var inWin *pong.NeedPreSigs_PerInput
	if clientWinsBranch == 0 {
		inWin, err = pickMine(drafts.InputsA)
	} else {
		inWin, err = pickMine(drafts.InputsB)
	}
	if err != nil {
		return status.Errorf(codes.Internal, "pick input (win): %v", err)
	}
	winDraftHex := drafts.DraftHexA
	if clientWinsBranch == 1 {
		winDraftHex = drafts.DraftHexB
	}
	if err := stream.Send(&pong.ServerMsg{
		Kind: &pong.ServerMsg_Req{
			Req: &pong.NeedPreSigs{DraftTxHex: winDraftHex, Inputs: []*pong.NeedPreSigs_PerInput{inWin}},
		},
	}); err != nil {
		return status.Errorf(codes.Unavailable, "send req (win): %v", err)
	}
	msgWin, err := stream.Recv()
	if err != nil || msgWin.GetVerifyOk() == nil {
		return status.Errorf(codes.InvalidArgument, "expected VERIFY_OK (win): %v", err)
	}
	verifyWin := msgWin.GetVerifyOk()
	var concatWin []byte
	concatWin = append(concatWin, []byte(winDraftHex)...)
	concatWin = append(concatWin, []byte(inWin.InputId)...)
	concatWin = append(concatWin, []byte(inWin.MHex)...)
	concatWin = append(concatWin, inWin.TCompressed...)
	expectedWin := blake256.Sum256(concatWin)
	if len(verifyWin.AckDigest) != 32 || !bytes.Equal(verifyWin.AckDigest, expectedWin[:]) {
		s.log.Warnf("ack mismatch (win): want=%x got=%x input=%s m=%s", expectedWin[:], verifyWin.AckDigest, inWin.InputId, inWin.MHex)
		return status.Error(codes.InvalidArgument, "ack digest mismatch (win)")
	}
	if len(verifyWin.Presigs) != 1 || verifyWin.Presigs[0].InputId != inWin.InputId {
		return status.Error(codes.InvalidArgument, "presig input mismatch (win)")
	}
	if err := s.verifyAndStorePresig(X, verifyWin.Presigs[0], inWin, winDraftHex, myEscrow, clientWinsBranch); err != nil {
		return err
	}

	// 7) LOSE branch handshake (opponent wins).
	oppWinsBranch := int32(1 - clientWinsBranch)
	var inLose *pong.NeedPreSigs_PerInput
	loseDraftHex := drafts.DraftHexB
	if oppWinsBranch == 0 {
		inLose, err = pickMine(drafts.InputsA)
		loseDraftHex = drafts.DraftHexA
	} else {
		inLose, err = pickMine(drafts.InputsB)
	}
	if err != nil {
		return status.Errorf(codes.Internal, "pick input (lose): %v", err)
	}
	if err := stream.Send(&pong.ServerMsg{
		Kind: &pong.ServerMsg_Req{
			Req: &pong.NeedPreSigs{DraftTxHex: loseDraftHex, Inputs: []*pong.NeedPreSigs_PerInput{inLose}},
		},
	}); err != nil {
		return status.Errorf(codes.Unavailable, "send req (lose): %v", err)
	}
	msgLose, err := stream.Recv()
	if err != nil || msgLose.GetVerifyOk() == nil {
		return status.Errorf(codes.InvalidArgument, "expected VERIFY_OK (lose): %v", err)
	}
	verifyLose := msgLose.GetVerifyOk()
	var concatLose []byte
	concatLose = append(concatLose, []byte(loseDraftHex)...)
	concatLose = append(concatLose, []byte(inLose.InputId)...)
	concatLose = append(concatLose, []byte(inLose.MHex)...)
	concatLose = append(concatLose, inLose.TCompressed...)
	expectedLose := blake256.Sum256(concatLose)
	if len(verifyLose.AckDigest) != 32 || !bytes.Equal(verifyLose.AckDigest, expectedLose[:]) {
		s.log.Warnf("ack mismatch (lose): want=%x got=%x input=%s m=%s", expectedLose[:], verifyLose.AckDigest, inLose.InputId, inLose.MHex)
		return status.Error(codes.InvalidArgument, "ack digest mismatch (lose)")
	}
	if len(verifyLose.Presigs) != 1 || verifyLose.Presigs[0].InputId != inLose.InputId {
		return status.Error(codes.InvalidArgument, "presig input mismatch (lose)")
	}
	if err := s.verifyAndStorePresig(X, verifyLose.Presigs[0], inLose, loseDraftHex, oppEscrow, oppWinsBranch); err != nil {
		return err
	}

	// 8) Final ACK
	if err := stream.Send(&pong.ServerMsg{
		Kind: &pong.ServerMsg_Ok{Ok: &pong.ServerOk{AckDigest: expectedLose[:]}},
	}); err != nil {
		return status.Errorf(codes.Unavailable, "send ok: %v", err)
	}
	return nil
}

// GetFinalizeBundle returns the winning draft, gamma and both presigs so the winner can finalize.
func (s *Server) GetFinalizeBundle(ctx context.Context, req *pong.GetFinalizeBundleRequest) (*pong.GetFinalizeBundleResponse, error) {
	if req == nil || strings.TrimSpace(req.MatchId) == "" || strings.TrimSpace(req.WinnerUid) == "" {
		return nil, status.Error(codes.InvalidArgument, "missing match_id or winner_uid")
	}

	parts := strings.SplitN(req.MatchId, "|", 2)
	wrID := strings.TrimSpace(parts[0])
	if wrID == "" {
		return nil, status.Error(codes.InvalidArgument, "bad match_id")
	}

	s.RLock()
	var es *escrowSession
	if m := s.roomEscrows[wrID]; m != nil {
		if eid := m[req.WinnerUid]; eid != "" {
			es = s.escrows[eid]
		}
	}
	s.RUnlock()
	if es == nil {
		return nil, status.Errorf(codes.NotFound, "no escrow bound for winner in room %s", wrID)
	}
	if es.preSign == nil || len(es.preSign) == 0 {
		return nil, status.Error(codes.FailedPrecondition, "no presign contexts stored for winner")
	}

	// Determine winner branch from stored contexts (they share branch per win/lose per player)
	var branch int32 = -1
	for _, ctxp := range es.preSign {
		branch = ctxp.Branch
		break
	}
	if branch < 0 {
		return nil, status.Error(codes.Internal, "failed to determine branch")
	}

	// Collect presigs for the chosen branch
	var draftHex string
	type kv struct {
		id  string
		ctx *PreSignCtx
	}
	var list []kv
	for id, ctxp := range es.preSign {
		if ctxp.Branch != branch {
			continue
		}
		list = append(list, kv{id: id, ctx: ctxp})
		if draftHex == "" {
			draftHex = ctxp.DraftHex
		}
	}
	if len(list) == 0 || draftHex == "" {
		return nil, status.Error(codes.FailedPrecondition, "no presigs for winner branch")
	}
	sort.Slice(list, func(i, j int) bool { return list[i].id < list[j].id })

	// Derive gamma for that branch using configured server secret.
	serverSecret := s.adaptorSecret
	if strings.TrimSpace(serverSecret) == "" {
		return nil, status.Error(codes.FailedPrecondition, "server adaptor secret not configured")
	}
	branchTag := fmt.Sprintf("branch-%d", branch)
	gammaHex, _ := pongbisonrelay.DeriveAdaptorGamma("", branchTag, branch, branchTag, serverSecret)
	gb, err := hex.DecodeString(gammaHex)
	if err != nil || len(gb) != 32 {
		return nil, status.Error(codes.Internal, "failed to derive gamma")
	}

	resp := &pong.GetFinalizeBundleResponse{DraftTxHex: draftHex, Gamma32: gb}
	for _, p := range list {
		resp.Inputs = append(resp.Inputs, &pong.FinalizeInput{
			InputId:         p.id,
			RedeemScriptHex: p.ctx.RedeemScriptHex,
			RLineCompressed: append([]byte(nil), p.ctx.RLineCompressed...),
			SLine32:         append([]byte(nil), p.ctx.SLine32...),
		})
	}
	return resp, nil
}
