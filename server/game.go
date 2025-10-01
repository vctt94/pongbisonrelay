package server

import (
	"context"
	"fmt"
	"strconv"
	"sync/atomic"
	"time"

	"google.golang.org/grpc/metadata"

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

// runNtfnSender consumes the player's notification queue and sends over the gRPC stream.
func (s *Server) runNtfnSender(ctx context.Context, p *ponggame.Player, stream pong.PongGame_StartNtfnStreamServer) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-p.NtfnQueue():
			if !ok {
				return
			}
			if err := stream.Send(msg); err != nil {
				s.log.Warnf("ntfn send failed for %s: %v", p.ID, err)
				return
			}
		}
	}
}

// runGameSender consumes the player's game bytes queue and sends over the gRPC stream.
func (s *Server) runGameSender(ctx context.Context, p *ponggame.Player, stream pong.PongGame_StartGameStreamServer) {
	// Determine desired per-client FPS from incoming metadata (optional).
	// Supports keys: "desired-fps" or "x-desired-fps" (string number).
	desiredFPS := ponggame.DEFAULT_FPS
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if vals := md.Get("desired-fps"); len(vals) > 0 {
			if v, err := strconv.ParseFloat(vals[0], 64); err == nil && v > 0 {
				desiredFPS = v
			}
		} else if vals := md.Get("x-desired-fps"); len(vals) > 0 {
			if v, err := strconv.ParseFloat(vals[0], 64); err == nil && v > 0 {
				desiredFPS = v
			}
		}
	}
	// Clamp to sensible bounds.
	if desiredFPS < 10 {
		desiredFPS = 10
	}
	if desiredFPS > 144 {
		desiredFPS = 144
	}

	s.log.Debugf("game sender for %s using desired FPS: %.2f", p.ID, desiredFPS)
	tick := time.NewTicker(time.Duration(float64(time.Second) / desiredFPS))
	defer tick.Stop()

	// Atomic holder for latest frame; initialize with a typed nil.
	var latest atomic.Value
	latest.Store([]byte(nil))

	// Drainer goroutine: read frames as they arrive and stash the latest.
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case b, ok := <-p.GameQueue():
				if !ok {
					return
				}
				latest.Store(b)
				// Opportunistically drain additional backlog without blocking.
				for i := 0; i < 16; i++ {
					select {
					case b2, ok2 := <-p.GameQueue():
						if !ok2 {
							return
						}
						latest.Store(b2)
					default:
						i = 16
					}
				}
			}
		}
	}()

	// Sender loop: on each tick, read latest and send.
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			b, _ := latest.Load().([]byte)
			if len(b) == 0 {
				continue
			}
			if err := stream.Send(&pong.GameUpdateBytes{Data: b}); err != nil {
				s.log.Warnf("game send failed for %s: %v", p.ID, err)
				return
			}
		}
	}
}

func (s *Server) StartNtfnStream(req *pong.StartNtfnStreamRequest, stream pong.PongGame_StartNtfnStreamServer) error {
	ctx, cancel := context.WithCancel(stream.Context())
	defer cancel()

	var clientID zkidentity.ShortID
	clientID.FromString(req.ClientId)
	// Ensure we delete using the correct key type
	defer s.activeNtfnStreams.Delete(clientID)
	s.log.Debugf("StartNtfnStream called by client %s", clientID)

	// Add to active streams: cancel any previous sender and store the new one

	// Create player session and ensure queues are initialized
	player := s.gameManager.PlayerSessions.CreateSession(clientID)

	// Start dedicated ntfn sender for this player's stream.
	// Cancel any previous one for this player if present.
	if prev, ok := s.activeNtfnStreams.Load(clientID); ok {
		if prevFn, is := prev.(context.CancelFunc); is {
			prevFn()
		}
	}
	s.activeNtfnStreams.Store(clientID, cancel)
	go s.runNtfnSender(ctx, player, stream)

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
	// Wait for disconnection (sender goroutine exits on ctx.Done as well)
	<-ctx.Done()
	s.log.Debugf("Notifier stream ended for client %s", clientID)
	s.handleDisconnect(clientID)
	return ctx.Err()
}

func (s *Server) StartGameStream(req *pong.StartGameStreamRequest, stream pong.PongGame_StartGameStreamServer) error {
	ctx, cancel := context.WithCancel(stream.Context())
	defer cancel()

	var clientID zkidentity.ShortID
	clientID.FromString(req.ClientId)
	defer s.activeGameStreams.Delete(clientID)

	// Store the cancel function after canceling any previous sender

	s.log.Debugf("Client %s called StartGameStream", req.ClientId)

	player := s.gameManager.PlayerSessions.GetPlayer(clientID)
	if player == nil {
		return fmt.Errorf("player not found for client ID %s", clientID)
	}
	// Require an active notifier sender for this player.
	if _, ok := s.activeNtfnStreams.Load(clientID); !ok {
		return fmt.Errorf("player notifier not active %s", clientID)
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
	player.Lock()
	player.Ready = true
	player.Unlock()

	// Start dedicated game sender for this player's stream.
	if prev, ok := s.activeGameStreams.Load(clientID); ok {
		if prevFn, is := prev.(context.CancelFunc); is {
			prevFn()
		}
	}
	s.activeGameStreams.Store(clientID, cancel)
	go s.runGameSender(ctx, player, stream)

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
	player.Lock()
	player.Ready = false
	player.Unlock()

	// First get the cancel function and call it before deleting
	if cancel, ok := s.activeGameStreams.Load(clientID); ok {
		if cancelFn, isCancel := cancel.(context.CancelFunc); isCancel {
			cancelFn()
		}
	}
	// Then delete the entry
	s.activeGameStreams.Delete(clientID)

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
