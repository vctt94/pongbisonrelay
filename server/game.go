package server

import (
	"context"
	"fmt"

	"github.com/companyzero/bisonrelay/zkidentity"

	"github.com/vctt94/pongbisonrelay/ponggame"
	"github.com/vctt94/pongbisonrelay/pongrpc/grpc/pong"
)

// --- game handlers ---
func (s *Server) SendInput(ctx context.Context, req *pong.PlayerInput) (*pong.GameUpdate, error) {
	var clientID zkidentity.ShortID
	clientID.FromString(req.PlayerId)
	return s.gameManager.HandlePlayerInput(clientID, req)
}

func (s *Server) StartNtfnStream(req *pong.StartNtfnStreamRequest, stream pong.PongGame_StartNtfnStreamServer) error {
	ctx, cancel := context.WithCancel(stream.Context())
	defer cancel()

	var clientID zkidentity.ShortID
	clientID.FromString(req.ClientId)
	// Ensure we delete using the correct key type
	defer s.activeNtfnStreams.Delete(clientID)
	s.log.Debugf("StartNtfnStream called by client %s", clientID)

	// Add to active streams
	s.activeNtfnStreams.Store(clientID, cancel)

	// Create player session
	player := s.gameManager.PlayerSessions.CreateSession(clientID)
	player.NotifierStream = stream

	// Track connected user so broadcast notifications (e.g. ON_WR_CREATED)
	// are delivered to all clients with active notifier streams.
	s.usersMu.Lock()
	s.users[clientID] = player
	s.usersMu.Unlock()

	// Bind any existing escrow sessions for this owner to this player so
	// watcher notifications reach the active notifier stream.
	s.escrowsMu.RLock()
	for _, es := range s.escrows {
		if es != nil && es.ownerUID == clientID {
			// Replace player binding
			es.mu.Lock()
			es.player = player
			u := es.latest
			es.mu.Unlock()
			if u.OK && u.UTXOCount > 0 {
				// Send only structured bet updates; clients derive UX from confs.
				_ = s.notify(player, &pong.NtfnStreamResponse{NotificationType: pong.NotificationType_BET_AMOUNT_UPDATE, PlayerId: es.ownerUID.String(), BetAmt: int64(es.betAtoms), Confs: u.Confs})
			}
		}
	}
	s.escrowsMu.RUnlock()

	// Escrow-first: remove legacy tips-based bet sync
	// Wait for disconnection
	<-ctx.Done()
	s.log.Debugf("Notifier stream ended for client %s", clientID)
	return ctx.Err()
}

func (s *Server) StartGameStream(req *pong.StartGameStreamRequest, stream pong.PongGame_StartGameStreamServer) error {
	ctx, cancel := context.WithCancel(stream.Context())
	defer cancel()

	var clientID zkidentity.ShortID
	clientID.FromString(req.ClientId)
	defer s.activeGameStreams.Delete(clientID)

	// Store the cancel function
	s.activeGameStreams.Store(clientID, cancel)

	s.log.Debugf("Client %s called StartGameStream", req.ClientId)

	player := s.gameManager.PlayerSessions.GetPlayer(clientID)
	if player == nil {
		return fmt.Errorf("player not found for client ID %s", clientID)
	}
	if player.NotifierStream == nil {
		return fmt.Errorf("player notifier nil %s", clientID)
	}

	// Require being in a waiting room
	if player.WR == nil {
		return fmt.Errorf("not in a waiting room")
	}
	// if not F2P, require escrow bound to waiting room
	if !s.isF2P {
		var escrowID string
		s.roomEscrowsMu.RLock()
		if m, ok := s.roomEscrows[clientID]; ok {
			escrowID = m[player.WR.ID]
		}
		var es *escrowSession
		if escrowID != "" {
			s.escrowsMu.RLock()
			es = s.escrows[escrowID]
			s.escrowsMu.RUnlock()
		}
		s.roomEscrowsMu.RUnlock()
		if es == nil {
			return fmt.Errorf("no escrow bound to waiting room for player")
		}
	}
	player.GameStream = stream
	player.Ready = true

	// Notify all players in the waiting room that this player is ready
	if player.WR != nil {
		// Marshal the waiting room state to include in notifications
		pwr := player.WR.Marshal()
		for _, p := range player.WR.Players {
			_ = s.notify(p, &pong.NtfnStreamResponse{
				NotificationType: pong.NotificationType_ON_PLAYER_READY,
				Message:          fmt.Sprintf("Player %s is ready", player.Nick),
				PlayerId:         player.ID.String(),
				Wr:               pwr,
				Ready:            true,
			})
		}
	}

	// Wait for context to end and handle disconnection
	<-ctx.Done()
	s.log.Debugf("Client %s disconnected from game stream", clientID)
	return nil
}

// UnreadyGameStream handles a request from a client who wants to signal they are no longer ready
func (s *Server) UnreadyGameStream(ctx context.Context, req *pong.UnreadyGameStreamRequest) (*pong.UnreadyGameStreamResponse, error) {
	var clientID zkidentity.ShortID
	clientID.FromString(req.ClientId)

	s.log.Debugf("Client %s called UnreadyGameStream", req.ClientId)

	// Find the player
	player := s.gameManager.PlayerSessions.GetPlayer(clientID)
	if player == nil {
		return nil, fmt.Errorf("player not found: %s", req.ClientId)
	}

	// Check if the player is in a waiting room
	if player.WR == nil {
		return nil, fmt.Errorf("player not in a waiting room")
	}
	player.Ready = false

	// First get the cancel function and call it before deleting
	if cancel, ok := s.activeGameStreams.Load(clientID); ok {
		if cancelFn, isCancel := cancel.(context.CancelFunc); isCancel {
			cancelFn()
		}
	}

	// Then delete the entry
	s.activeGameStreams.Delete(clientID)
	player.GameStream = nil

	// Notify other players in the waiting room
	pwr := player.WR.Marshal()

	for _, p := range player.WR.Players {
		_ = s.notify(p, &pong.NtfnStreamResponse{
			NotificationType: pong.NotificationType_ON_PLAYER_READY,
			Message:          fmt.Sprintf("Player %s is not ready", player.Nick),
			PlayerId:         player.ID.String(),
			RoomId:           player.WR.ID,
			Wr:               pwr,
			Ready:            false,
		})
	}

	return &pong.UnreadyGameStreamResponse{}, nil
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
			_ = s.notify(p, &pong.NtfnStreamResponse{
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
