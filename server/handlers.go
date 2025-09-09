package server

import (
	"bytes"
	"context"
	crand "crypto/rand"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/companyzero/bisonrelay/clientrpc/types"
	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/decred/dcrd/chaincfg/v3"
	"github.com/decred/dcrd/crypto/blake256"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"

	"github.com/vctt94/pong-bisonrelay/ponggame"
	"github.com/vctt94/pong-bisonrelay/pongrpc/grpc/pong"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// --- game handlers ---

func (s *Server) handleGameLifecycle(ctx context.Context, players []*ponggame.Player, tips []*types.ReceivedTip) {
	game, err := s.gameManager.StartGame(ctx, players)
	if err != nil {
		s.log.Errorf("Failed to start game: %v", err)
		return
	}

	defer func() {
		// reset player status
		for _, player := range game.Players {
			player.ResetPlayer()
		}
		// remove game from gameManager after it ended
		delete(s.gameManager.Games, game.Id)
		s.log.Debugf("Game %s cleaned up", game.Id)
	}()

	game.Run()

	var wg sync.WaitGroup
	for _, player := range players {
		wg.Add(1)
		go func(player *ponggame.Player) {
			defer wg.Done()
			if player.NotifierStream != nil {
				err := player.NotifierStream.Send(&pong.NtfnStreamResponse{
					NotificationType: pong.NotificationType_GAME_START,
					Message:          "Game started with ID: " + game.Id,
					Started:          true,
					GameId:           game.Id,
				})
				if err != nil {
					s.log.Warnf("Failed to notify player %s: %v", player.ID, err)
				}
			}
			s.sendGameUpdates(ctx, player, game)
		}(player)
	}

	wg.Wait() // Wait for both players' streams to finish

	s.handleGameEnd(ctx, game, players, tips)
}

func (s *Server) handleGameEnd(ctx context.Context, game *ponggame.GameInstance, players []*ponggame.Player, tips []*types.ReceivedTip) {
	winner := game.Winner
	var winnerID string
	if winner != nil {
		winnerID = winner.String()
		s.log.Infof("Game ended. Winner: %s", winnerID)
	} else {
		s.log.Infof("Game ended in a draw.")
	}

	// Notify players of game outcome (no transfers)
	for _, player := range players {
		message := "Game ended in a draw."
		if player.ID == winner {
			message = "Congratulations, you won!"
		} else if winner != nil {
			message = "Sorry, you lost."
		}
		player.NotifierStream.Send(&pong.NtfnStreamResponse{
			NotificationType: pong.NotificationType_GAME_END,
			Message:          message,
			GameId:           game.Id,
		})
		// delete player from gameManager PlayerGameMap
		delete(s.gameManager.PlayerGameMap, *player.ID)
	}

	// POC: deliver gamma to winner via NtfnStream for finalization
	if winner != nil {
		// Determine branch index anchored to room host: branch 0 pays host (a), 1 pays non-host (b).
		winnerBranch := int32(0)
		var hostUID string
		if len(players) >= 2 {
			if players[0].WR != nil && players[0].WR.HostID != nil {
				hostUID = players[0].WR.HostID.String()
			} else if players[1].WR != nil && players[1].WR.HostID != nil {
				hostUID = players[1].WR.HostID.String()
			}
			if hostUID != "" && winnerID != hostUID {
				winnerBranch = 1
			}
		}
		s.log.Debugf("finalize: computed winnerBranch=%d host=%s winner=%s", winnerBranch, hostUID, winnerID)

		// Look up the winner's escrow session bound to this room and find the presign context
		// for the computed winner branch.
		var wrID string
		for _, p := range players {
			if p != nil && p.WR != nil {
				wrID = p.WR.ID
				break
			}
		}
		s.RLock()
		var es *escrowSession
		if wrID != "" && s.roomEscrows != nil {
			if m := s.roomEscrows[wrID]; m != nil {
				if eid := m[winnerID]; eid != "" {
					es = s.escrows[eid]
				}
			}
		}
		s.RUnlock()
		if es == nil {
			s.log.Errorf("finalize: no room-bound escrow session found for winner %s in wr %s", winnerID, wrID)
			return
		}
		if es == nil {
			s.log.Warnf("finalize: no escrow session found for winner %s", winnerID)
			return
		}
		if es.preSign == nil {
			s.log.Warnf("finalize: no presign contexts stored for winner %s", winnerID)
			return
		}
		var chosen *PreSignCtx
		var branches []int32
		for _, ctxp := range es.preSign {
			branches = append(branches, ctxp.Branch)
			if ctxp.Branch == winnerBranch {
				chosen = ctxp
			}
		}
		if chosen == nil {
			s.log.Warnf("finalize: no presign context for branch %d; have branches=%v", winnerBranch, branches)
			return
		}
		// Derive gamma using the same branch-bound domain separation used during presign.
		const pocServerPrivHex = "11ee22dd33cc44bb55aa66ff77ee88dd99cc00bbaa11223344556677889900aa"
		branchTag := fmt.Sprintf("branch-%d", chosen.Branch)
		gammaHex, _ := deriveAdaptorGamma("", branchTag, chosen.Branch, branchTag, pocServerPrivHex)
		if gb, err := hex.DecodeString(gammaHex); err == nil && len(gb) == 32 {
			if w := s.gameManager.PlayerSessions.GetPlayer(*winner); w != nil && w.NotifierStream != nil {
				_ = w.NotifierStream.Send(&pong.NtfnStreamResponse{
					NotificationType: pong.NotificationType_MESSAGE,
					Message:          fmt.Sprintf("Gamma received; finalizing… (%x)\nResult draft tx hex: %s", gb, chosen.DraftHex),
				})
			}
		} else if err != nil {
			s.log.Warnf("failed to decode gamma hex: %v", err)
		}
	}
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

	// Derive gamma for that branch (same domain sep. as presign)
	const pocServerPrivHex = "11ee22dd33cc44bb55aa66ff77ee88dd99cc00bbaa11223344556677889900aa"
	branchTag := fmt.Sprintf("branch-%d", branch)
	gammaHex, _ := deriveAdaptorGamma("", branchTag, branch, branchTag, pocServerPrivHex)
	gb, err := hex.DecodeString(gammaHex)
	if err != nil || len(gb) != 32 {
		return nil, status.Error(codes.Internal, "failed to derive gamma")
	}

	resp := &pong.GetFinalizeBundleResponse{DraftTxHex: draftHex, Gamma32: gb}
	for _, p := range list {
		resp.Inputs = append(resp.Inputs, &pong.FinalizeInput{
			InputId:          p.id,
			RedeemScriptHex:  p.ctx.RedeemScriptHex,
			RprimeCompressed: append([]byte(nil), p.ctx.RPrimeCompressed...),
			Sprime32:         append([]byte(nil), p.ctx.SPrime32...),
		})
	}
	return resp, nil
}

// --- waiting room handlers ---

func (s *Server) handleWaitingRoomRemoved(wr *pong.WaitingRoom) {
	s.log.Infof("Waiting room %s removed", wr.Id)

	// Notify all users about the waiting room removal
	for _, user := range s.users {
		user.NotifierStream.Send(&pong.NtfnStreamResponse{
			NotificationType: pong.NotificationType_ON_WR_REMOVED,
			Message:          fmt.Sprintf("Waiting room %s was removed", wr.Id),
			RoomId:           wr.Id,
		})
	}
}
func (s *Server) RefOpenEscrow(ctx context.Context, req *pong.OpenEscrowRequest) (*pong.OpenEscrowResponse, error) {
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
	redeem, err := buildPerDepositorRedeemScript(comp, req.CsvBlocks)
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

	now := time.Now()
	es := &escrowSession{
		escrowID:        eid,
		ownerUID:        req.OwnerUid,
		compPubkey:      append([]byte(nil), comp...),
		payoutPubkey:    append([]byte(nil), payoutPubkey...),
		betAtoms:        req.BetAtoms,
		csvBlocks:       req.CsvBlocks,
		redeemScriptHex: hex.EncodeToString(redeem),
		pkScriptHex:     pkScriptHex,
		depositAddress:  addr,
		createdAt:       now,
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

// SignalReadyToPlay marks the caller ready and, if everyone is ready, triggers next phase.
func (s *Server) SignalReadyToPlay(ctx context.Context, req *pong.SignalReadyToPlayRequest) (*pong.SignalReadyToPlayResponse, error) {
	var clientID zkidentity.ShortID
	if err := clientID.FromString(req.ClientId); err != nil {
		return nil, fmt.Errorf("invalid client id: %w", err)
	}

	s.log.Debugf("Client %s signaling ready for game %s", req.ClientId, req.GameId)

	player := s.gameManager.PlayerSessions.GetPlayer(clientID)
	if player == nil {
		return nil, fmt.Errorf("player not found for client %s", clientID)
	}

	game := s.gameManager.GetPlayerGame(clientID)
	if game == nil || (req.GameId != "" && game.Id != req.GameId) {
		return nil, fmt.Errorf("game instance not found for client %s", clientID)
	}

	// Mark this client ready (idempotent) and check if everyone is ready.
	game.Lock()
	if game.PlayersReady == nil {
		game.PlayersReady = make(map[string]bool)
	}
	key := clientID.String()
	game.PlayersReady[key] = true
	// Compute readiness based on number of players
	total := len(game.Players)
	allReady := (total > 0 && len(game.PlayersReady) == total)
	playersSnap := append([]*ponggame.Player(nil), game.Players...)
	game.Unlock()

	if allReady {
		for _, p := range playersSnap {
			if p.NotifierStream == nil {
				continue
			}
			_ = p.NotifierStream.Send(&pong.NtfnStreamResponse{
				NotificationType: pong.NotificationType_ON_PLAYER_READY,
				Message:          "all players ready to start the game",
				PlayerId:         player.ID.String(),
				GameId:           game.Id,
				Ready:            true,
			})
		}
	}

	return &pong.SignalReadyToPlayResponse{
		Success: true,
		Message: "Ready signal received",
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
