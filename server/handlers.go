package server

import (
	"context"
	crand "crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"bytes"

	"github.com/decred/dcrd/crypto/blake256"

	"github.com/companyzero/bisonrelay/clientrpc/types"
	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/decred/dcrd/chaincfg/chainhash"
	"github.com/decred/dcrd/chaincfg/v3"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/decred/dcrd/txscript/v4"
	"github.com/decred/dcrd/wire"

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
	if req.OwnerUid == "" || req.BetAtoms == 0 || req.CsvBlocks == 0 || len(req.CompPubkey) != 33 {
		return nil, status.Errorf(codes.InvalidArgument, "invalid escrow open request")
	}
	// Canonicalize compressed pubkey.
	pub, err := secp256k1.ParsePubKey(req.CompPubkey)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "bad compressed pubkey: %v", err)
	}
	comp := pub.SerializeCompressed()

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
	ctx := stream.Context()
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

	// Find a funded escrow session matching this comp pubkey
	s.RLock()
	var es *escrowSession
	for _, cand := range s.escrows {
		if cand != nil && bytes.Equal(cand.compPubkey, hel.CompPubkey) {
			// Require a UTXO (0-conf)
			if s.watcher != nil && cand.pkScriptHex != "" {
				utxos, _, ok := s.watcher.queryDeposit(cand.pkScriptHex)
				if ok && len(utxos) > 0 {
					es = cand
					break
				}
			}
		}
	}
	s.RUnlock()
	if es == nil {
		return status.Error(codes.FailedPrecondition, "no funded escrow for this pubkey")
	}
	utxos, _, _ := s.watcher.queryDeposit(es.pkScriptHex)
	if len(utxos) == 0 {
		return status.Error(codes.FailedPrecondition, "watcher lost utxo state")
	}
	u := utxos[0]
	// Build a simple draft that spends this UTXO back to the same pkScript (POC)
	var prev chainhash.Hash
	if err := chainhash.Decode(&prev, u.Txid); err != nil {
		return status.Errorf(codes.InvalidArgument, "bad utxo txid: %v", err)
	}
	tx := wire.NewMsgTx()
	tx.Version = 1
	tx.AddTxIn(&wire.TxIn{PreviousOutPoint: wire.OutPoint{Hash: prev, Index: u.Vout}})
	pkb, _ := hex.DecodeString(es.pkScriptHex)
	tx.AddTxOut(&wire.TxOut{Value: int64(u.Value), PkScript: pkb})
	var buf bytes.Buffer
	_ = tx.Serialize(&buf)
	draftHex := hex.EncodeToString(buf.Bytes())
	// Compute m for input 0 using the redeem script
	redeem, _ := hex.DecodeString(es.redeemScriptHex)
	mBytes, err := txscript.CalcSignatureHash(redeem, txscript.SigHashAll, tx, 0, nil)
	if err != nil || len(mBytes) != 32 {
		return status.Error(codes.Internal, "failed to compute sighash")
	}
	mHex := hex.EncodeToString(mBytes)
	// Derive adaptor T (POC secret)
	const pocServerPrivHex = "11ee22dd33cc44bb55aa66ff77ee88dd99cc00bbaa11223344556677889900aa"
	gammaHex, TCompHex := deriveAdaptorGamma("", fmt.Sprintf("%s:%d", u.Txid, u.Vout), 0, mHex, pocServerPrivHex)
	_ = gammaHex // not revealed here
	Tcomp, _ := hex.DecodeString(TCompHex)

	req := &pong.NeedPreSigs{
		DraftTxHex: draftHex,
		Inputs: []*pong.NeedPreSigs_PerInput{
			{
				InputId:         fmt.Sprintf("%s:%d", u.Txid, u.Vout),
				RedeemScriptHex: es.redeemScriptHex,
				MHex:            mHex,
				TCompressed:     Tcomp,
			},
		},
	}
	if err := stream.Send(&pong.ServerMsg{Kind: &pong.ServerMsg_Req{Req: req}}); err != nil {
		return status.Errorf(codes.Unavailable, "send req: %v", err)
	}

	// Step 3: receive VERIFY_OK
	vmsg, err := stream.Recv()
	if err != nil || vmsg.GetVerifyOk() == nil {
		return status.Errorf(codes.InvalidArgument, "expected VERIFY_OK: %v", err)
	}
	verify := vmsg.GetVerifyOk()
	// Compute expected ack digest for the sent req
	var concat []byte
	concat = append(concat, []byte(req.DraftTxHex)...)
	for _, in := range req.Inputs {
		concat = append(concat, []byte(in.InputId)...)
		concat = append(concat, []byte(in.MHex)...)
		concat = append(concat, in.TCompressed...)
	}
	expected := blake256.Sum256(concat)
	if len(verify.AckDigest) != 32 || !bytes.Equal(verify.AckDigest, expected[:]) {
		return status.Error(codes.InvalidArgument, "ack digest mismatch")
	}
	// Verify each presig: s'·G + e·X + T == R'
	for _, ps := range verify.Presigs {
		if len(ps.RprimeCompressed) != 33 || len(ps.Sprime32) != 32 {
			return status.Error(codes.InvalidArgument, "bad presig sizes")
		}
		Rp, err := secp256k1.ParsePubKey(ps.RprimeCompressed)
		if err != nil {
			return status.Errorf(codes.InvalidArgument, "parse R': %v", err)
		}
		// Lookup input info
		var in *pong.NeedPreSigs_PerInput
		for _, ii := range req.Inputs {
			if ii.InputId == ps.InputId {
				in = ii
				break
			}
		}
		if in == nil {
			return status.Error(codes.InvalidArgument, "unknown input_id in presig")
		}
		// Compute e = BLAKE256(r_x || m)
		var r32 [32]byte
		copy(r32[:], ps.RprimeCompressed[1:33])
		mh, err := hex.DecodeString(in.MHex)
		if err != nil || len(mh) != 32 {
			return status.Error(codes.InvalidArgument, "bad m_hex")
		}
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
		// Check equality by bytes
		if !bytes.Equal(lhs1.SerializeCompressed(), RminusT.SerializeCompressed()) {
			return status.Error(codes.InvalidArgument, "adaptor relation failed")
		}
	}
	// All good → send SERVER_OK and return
	if err := stream.Send(&pong.ServerMsg{Kind: &pong.ServerMsg_Ok{Ok: &pong.ServerOk{AckDigest: expected[:]}}}); err != nil {
		return status.Errorf(codes.Unavailable, "send ok: %v", err)
	}
	// Do not keep the stream open
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}
