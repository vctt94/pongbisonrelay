package server

import (
	"bytes"
	"context"
	crand "crypto/rand"
	"encoding/hex"
	"fmt"
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
		// Look up the exact PreSignCtx created during presign and reuse its
		// DraftHex and MHex to derive gamma. Do NOT rebuild drafts here.
		s.RLock()
		var es *escrowSession
		for _, e := range s.escrows {
			if e != nil && e.ownerUID == winnerID {
				es = e
				break
			}
		}
		s.RUnlock()
		if es != nil && es.preSign != nil {
			for _, ctxp := range es.preSign {
				// bind gamma to the same input_id + m used during presign
				const pocServerPrivHex = "11ee22dd33cc44bb55aa66ff77ee88dd99cc00bbaa11223344556677889900aa"
				gammaHex, _ := deriveAdaptorGamma("", ctxp.InputID, 0, ctxp.MHex, pocServerPrivHex)
				gb, err := hex.DecodeString(gammaHex)
				if err != nil || len(gb) != 32 {
					continue
				}
				w := s.gameManager.PlayerSessions.GetPlayer(*winner)
				if w != nil && w.NotifierStream != nil {
					w.NotifierStream.Send(&pong.NtfnStreamResponse{
						NotificationType: pong.NotificationType_MESSAGE,
						Message:          fmt.Sprintf("Gamma received; finalizing… (%x)\nResult draft tx hex: %s", gb, ctxp.DraftHex),
					})
				}
				break
			}
		}
	}
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

	// Idempotency key: owner|comp|bet|csv
	key := fmt.Sprintf("%s|%x|%d|%d", req.OwnerUid, comp, req.BetAtoms, req.CsvBlocks)

	s.Lock()
	defer s.Unlock()

	if s.escrows == nil {
		s.escrows = make(map[string]*escrowSession)
	}
	if s.escrowByKey == nil {
		s.escrowByKey = make(map[string]string)
	}
	if id, ok := s.escrowByKey[key]; ok {
		if es := s.escrows[id]; es != nil {
			return &pong.OpenEscrowResponse{
				EscrowId:       es.escrowID,
				DepositAddress: es.depositAddress,
				PkScriptHex:    es.pkScriptHex,
			}, nil
		}
	}

	// Build depositor-only redeem script (no opponent key).
	redeem, err := buildPerDepositorRedeemScript(comp, req.CsvBlocks)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "build redeem: %v", err)
	}

	// TODO: prefer a server field like s.netParams stdaddr.AddressParams.
	// For now, use testnet3 params (or mainnet/simnet as appropriate).
	params := chaincfg.TestNet3Params() // <- replace with your server’s configured params

	pkScriptHex, addr, err := pkScriptAndAddrFromRedeem(redeem, params)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "derive address: %v", err)
	}

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
		updatedAt:       now,
	}
	s.escrows[eid] = es
	s.escrowByKey[key] = eid

	return &pong.OpenEscrowResponse{
		EscrowId:       eid,
		DepositAddress: addr,
		PkScriptHex:    pkScriptHex,
	}, nil
}

func (s *Server) WaitFunding(req *pong.WaitFundingRequest, stream pong.PongReferee_WaitFundingServer) error {
	ctx := stream.Context()
	if req.EscrowId == "" {
		return status.Error(codes.InvalidArgument, "missing escrow_id")
	}

	// Register this pkScript with the watcher only once
	var registered sync.Once
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case <-ticker.C:
			s.RLock()
			es := s.escrows[req.EscrowId]
			s.RUnlock()
			if es == nil {
				return status.Error(codes.NotFound, "unknown escrow id")
			}
			if s.watcher == nil {
				return status.Error(codes.FailedPrecondition, "chain watcher not initialized")
			}

			registered.Do(func() {
				s.watcher.registerDeposit(es.pkScriptHex, es.redeemScriptHex, req.EscrowId)
				s.log.Infof("watcher: registered escrow_id=%s pkScript=%s", req.EscrowId, es.pkScriptHex)
			})

			utxos, confs, ok := s.watcher.queryDeposit(es.pkScriptHex)

			funded := ok && len(utxos) > 0
			confirmed := funded && confs >= 1

			resp := &pong.WaitFundingResponse{
				Confs: confs,
			}

			if funded {
				u := utxos[0] // safe now
				resp.Value = u.Value
				resp.Utxo = &pong.EscrowUTXO{
					Txid:            u.Txid,
					Vout:            u.Vout,
					Value:           u.Value,
					RedeemScriptHex: es.redeemScriptHex,
					PkScriptHex:     es.pkScriptHex,
					Owner:           es.ownerUID,
				}
			}

			if err := stream.Send(resp); err != nil {
				return status.Errorf(codes.Unavailable, "stream send failed: %v", err)
			}

			if confirmed {
				s.Lock()
				es.confs = confs
				es.updatedAt = time.Now()
				s.Unlock()
				return nil
			}
		}
	}
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

	// Update ready state idempotently and snapshot for notifications without holding the lock.
	var (
		already     bool
		allReady    bool
		playersSnap []*ponggame.Player
		nick        = player.Nick
	)

	game.Lock()
	if game.PlayersReady == nil {
		game.PlayersReady = make(map[string]bool)
	}
	key := clientID.String()
	if game.PlayersReady[key] {
		already = true
	} else {
		game.PlayersReady[key] = true
	}

	// Count readiness.
	readyCount := 0
	total := len(game.Players)
	for _, p := range game.Players {
		if game.PlayersReady[p.ID.String()] {
			readyCount++
		}
	}
	allReady = (total > 0 && readyCount == total)

	// Snapshot players slice for notifying outside the lock.
	playersSnap = append(playersSnap, game.Players...)
	game.Unlock()

	// Notify: this player is ready (only once).
	if !already {
		s.broadcastReady(playersSnap, req.ClientId, game.Id, fmt.Sprintf("Player %s is ready to start the game", nick))
	}

	// // If everyone is ready, signal next phase (start game, fetch drafts/adaptors, etc.).
	if allReady {
		// s.broadcastAllReady(playersSnap, game.Id)
		// go s.onAllPlayersReady(game) // hook to kick the next phase safely out-of-band
	}

	return &pong.SignalReadyToPlayResponse{
		Success: true,
		Message: "Ready signal received",
	}, nil
}

func (s *Server) broadcastReady(players []*ponggame.Player, playerID, gameID, msg string) {
	for _, p := range players {
		if p.NotifierStream == nil {
			continue
		}
		_ = p.NotifierStream.Send(&pong.NtfnStreamResponse{
			NotificationType: pong.NotificationType_ON_PLAYER_READY,
			Message:          msg,
			PlayerId:         playerID,
			GameId:           gameID,
			Ready:            true,
		})
	}
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
	Ac := X.SerializeCompressed()
	_ = Ac

	return s.settlementStreamTwoInputsAuto(stream, X, hel.CompPubkey)
}

// settlementStreamTwoInputs performs the two-branch, two-input presign handshake
// with a single client. For each branch (A wins, B wins), it sends a presign
// request containing only the client's own input. The opponent performs the
// same handshake on their stream. The server persists both branches' presigs
// and drafts for later finalization delivery to the winner.
func (s *Server) settlementStreamTwoInputs(stream pong.PongReferee_SettlementStreamServer, X *secp256k1.PublicKey, myEscrow, oppEscrow *escrowSession, myUTXO, oppUTXO *pong.EscrowUTXO) error {
	// Build two-branch drafts for both inputs.
	drafts, err := s.buildTwoInputDrafts(myEscrow, myUTXO, oppEscrow, oppUTXO)
	if err != nil {
		return status.Errorf(codes.Internal, "build two-input drafts: %v", err)
	}

	// Helper to pick this client's PerInput from a full inputs slice.
	pickMine := func(all []*pong.NeedPreSigs_PerInput) (*pong.NeedPreSigs_PerInput, error) {
		want := fmt.Sprintf("%s:%d", myUTXO.Txid, myUTXO.Vout)
		for _, in := range all {
			if in.InputId == want {
				return in, nil
			}
		}
		return nil, fmt.Errorf("client input %s not found in draft inputs", want)
	}

	// Branch 0: A-wins (myEscrow as winner)
	inA, err := pickMine(drafts.InputsA)
	if err != nil {
		return status.Errorf(codes.Internal, "pick input A: %v", err)
	}
	if err := stream.Send(&pong.ServerMsg{Kind: &pong.ServerMsg_Req{Req: &pong.NeedPreSigs{DraftTxHex: drafts.DraftHexA, Inputs: []*pong.NeedPreSigs_PerInput{inA}}}}); err != nil {
		return status.Errorf(codes.Unavailable, "send req A: %v", err)
	}
	msgA, err := stream.Recv()
	if err != nil || msgA.GetVerifyOk() == nil {
		return status.Errorf(codes.InvalidArgument, "expected VERIFY_OK (A): %v", err)
	}
	verifyA := msgA.GetVerifyOk()
	// Ack digest over draft + canonical(single) input
	var concatA []byte
	concatA = append(concatA, []byte(drafts.DraftHexA)...)
	concatA = append(concatA, []byte(inA.InputId)...)
	concatA = append(concatA, []byte(inA.MHex)...)
	concatA = append(concatA, inA.TCompressed...)
	expectedA := blake256.Sum256(concatA)
	if len(verifyA.AckDigest) != 32 || !bytes.Equal(verifyA.AckDigest, expectedA[:]) {
		return status.Error(codes.InvalidArgument, "ack digest mismatch (A)")
	}
	if len(verifyA.Presigs) != 1 || verifyA.Presigs[0].InputId != inA.InputId {
		return status.Error(codes.InvalidArgument, "presig input mismatch (A)")
	}
	if err := s.verifyAndStorePresig(X, verifyA.Presigs[0], inA, drafts.DraftHexA, myEscrow, 0); err != nil {
		return err
	}

	// Branch 1: B-wins (oppEscrow as winner)
	inB, err := pickMine(drafts.InputsB)
	if err != nil {
		return status.Errorf(codes.Internal, "pick input B: %v", err)
	}
	if err := stream.Send(&pong.ServerMsg{Kind: &pong.ServerMsg_Req{Req: &pong.NeedPreSigs{DraftTxHex: drafts.DraftHexB, Inputs: []*pong.NeedPreSigs_PerInput{inB}}}}); err != nil {
		return status.Errorf(codes.Unavailable, "send req B: %v", err)
	}
	msgB, err := stream.Recv()
	if err != nil || msgB.GetVerifyOk() == nil {
		return status.Errorf(codes.InvalidArgument, "expected VERIFY_OK (B): %v", err)
	}
	verifyB := msgB.GetVerifyOk()
	var concatB []byte
	concatB = append(concatB, []byte(drafts.DraftHexB)...)
	concatB = append(concatB, []byte(inB.InputId)...)
	concatB = append(concatB, []byte(inB.MHex)...)
	concatB = append(concatB, inB.TCompressed...)
	expectedB := blake256.Sum256(concatB)
	if len(verifyB.AckDigest) != 32 || !bytes.Equal(verifyB.AckDigest, expectedB[:]) {
		return status.Error(codes.InvalidArgument, "ack digest mismatch (B)")
	}
	if len(verifyB.Presigs) != 1 || verifyB.Presigs[0].InputId != inB.InputId {
		return status.Error(codes.InvalidArgument, "presig input mismatch (B)")
	}
	if err := s.verifyAndStorePresig(X, verifyB.Presigs[0], inB, drafts.DraftHexB, oppEscrow, 1); err != nil {
		return err
	}

	// Ack overall success
	if err := stream.Send(&pong.ServerMsg{Kind: &pong.ServerMsg_Ok{Ok: &pong.ServerOk{AckDigest: expectedB[:]}}}); err != nil {
		return status.Errorf(codes.Unavailable, "send ok: %v", err)
	}
	return nil
}

// settlementStreamTwoInputsAuto locates the caller's escrow and an opponent's
// funded escrow, fetches one UTXO from each, and runs the two-branch handshake.
func (s *Server) settlementStreamTwoInputsAuto(stream pong.PongReferee_SettlementStreamServer, X *secp256k1.PublicKey, compPubkey []byte) error {
	// Locate my escrow by compPubkey and ensure it has a UTXO.
	s.RLock()
	var myEscrow *escrowSession
	for _, cand := range s.escrows {
		if cand != nil && bytes.Equal(cand.compPubkey, compPubkey) {
			myEscrow = cand
			break
		}
	}
	s.RUnlock()
	if myEscrow == nil {
		return status.Error(codes.FailedPrecondition, "no funded escrow for this pubkey")
	}
	aus, _, _ := s.watcher.queryDeposit(myEscrow.pkScriptHex)
	if len(aus) == 0 {
		return status.Error(codes.FailedPrecondition, "watcher lost utxo state")
	}

	// Find any opponent escrow with a UTXO.
	s.RLock()
	var oppEscrow *escrowSession
	for _, cand := range s.escrows {
		if cand == nil || cand == myEscrow {
			continue
		}
		if s.watcher != nil && cand.pkScriptHex != "" {
			if bus, _, ok := s.watcher.queryDeposit(cand.pkScriptHex); ok && len(bus) > 0 {
				oppEscrow = cand
				break
			}
		}
	}
	s.RUnlock()
	if oppEscrow == nil {
		return status.Error(codes.FailedPrecondition, "opponent escrow not yet funded")
	}
	bus, _, _ := s.watcher.queryDeposit(oppEscrow.pkScriptHex)
	if len(bus) == 0 {
		return status.Error(codes.FailedPrecondition, "opponent watcher lost utxo state")
	}
	return s.settlementStreamTwoInputs(stream, X, myEscrow, oppEscrow, aus[0], bus[0])
}
