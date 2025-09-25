package server

import (
	"context"
	"fmt"

	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/vctt94/pongbisonrelay/ponggame"
	"github.com/vctt94/pongbisonrelay/pongrpc/grpc/pong"
)

func (s *Server) GetWaitingRoom(ctx context.Context, req *pong.WaitingRoomRequest) (*pong.WaitingRoomResponse, error) {
	wr := s.gameManager.GetWaitingRoom(req.RoomId)
	if wr == nil {
		return nil, fmt.Errorf("waiting room not found: %s", req.RoomId)
	}

	return &pong.WaitingRoomResponse{
		Wr: wr.Marshal(),
	}, nil
}

func (s *Server) GetWaitingRooms(ctx context.Context, req *pong.WaitingRoomsRequest) (*pong.WaitingRoomsResponse, error) {
	pongWaitingRooms := make([]*pong.WaitingRoom, len(s.gameManager.WaitingRooms))
	for i, room := range s.gameManager.WaitingRooms {
		wr := room.Marshal()
		pongWaitingRooms[i] = wr
	}

	return &pong.WaitingRoomsResponse{
		Wr: pongWaitingRooms,
	}, nil
}

// resolveFundedEscrow returns the escrow after strictly validating:
//   - escrow exists and is owned by ownerUID
//   - funding is bound to a specific outpoint (txid:vout)
//   - that exact outpoint is present and is the *only* deposit
func (s *Server) resolveFundedEscrow(ownerUID zkidentity.ShortID, escrowID string) (*escrowSession, error) {
	if escrowID == "" {
		return nil, fmt.Errorf("escrow ID is required")
	}

	// Exact escrow lookup & ownership check.
	s.escrowsMu.RLock()
	es := s.escrows[escrowID]
	s.escrowsMu.RUnlock()
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

func (s *Server) CreateWaitingRoom(ctx context.Context, req *pong.CreateWaitingRoomRequest) (*pong.CreateWaitingRoomResponse, error) {
	var hostID zkidentity.ShortID
	if err := hostID.FromString(req.HostId); err != nil {
		return nil, err
	}
	hostPlayer := s.gameManager.PlayerSessions.GetPlayer(hostID)
	if hostPlayer == nil {
		return nil, fmt.Errorf("player not found: %s", req.HostId)
	}
	// Allow F2P; otherwise enforce min bet when req.BetAmt > 0.
	if !s.isF2P && req.BetAmt > 0 && float64(req.BetAmt)/1e11 < s.minBetAmt {
		return nil, fmt.Errorf("bet needs to be higher than %.8f", s.minBetAmt)
	}
	if hostPlayer.WR != nil {
		return nil, fmt.Errorf("player %s is already in a waiting room", hostID.String())
	}

	s.log.Debugf("creating waiting room. Host ID: %s", hostID)

	// F2P mode: allow creating a waiting room without an escrow.
	if s.isF2P {
		// In F2P use the requested bet amount (may be zero for truly free games).
		wr, err := ponggame.NewWaitingRoom(hostPlayer, req.BetAmt)
		if err != nil {
			return nil, fmt.Errorf("failed to create waiting room: %v", err)
		}
		hostPlayer.WR = wr

		// Add to list of rooms.
		totalRooms := s.gameManager.AppendWaitingRoom(wr)

		s.log.Debugf("waiting room created (F2P). Total rooms: %d", totalRooms)

		// Signal creation (non-blocking).
		select {
		case s.waitingRoomCreated <- struct{}{}:
		default:
		}

		pongWR := wr.Marshal()

		return &pong.CreateWaitingRoomResponse{Wr: pongWR}, nil
	}

	// Non-F2P: require funded escrow (0-conf OK).
	es, err := s.resolveFundedEscrow(hostID, req.EscrowId)
	if err != nil {
		return nil, fmt.Errorf("require funded escrow to create room (0-conf ok): %w", err)
	}

	// Create room; bet uses host escrow's betAtoms.
	wr, err := ponggame.NewWaitingRoom(hostPlayer, int64(es.betAtoms))
	if err != nil {
		return nil, fmt.Errorf("failed to create waiting room: %v", err)
	}
	hostPlayer.WR = wr

	// Bind host escrow to room.
	s.roomEscrowsMu.Lock()
	if s.roomEscrows == nil {
		s.roomEscrows = make(map[zkidentity.ShortID]map[string]string)
	}
	if s.roomEscrows[*hostPlayer.ID] == nil {
		s.roomEscrows[*hostPlayer.ID] = make(map[string]string)
	}
	s.roomEscrows[*hostPlayer.ID][wr.ID] = es.escrowID
	s.roomEscrowsMu.Unlock()

	// Add to list of rooms.
	totalRooms := s.gameManager.AppendWaitingRoom(wr)

	s.log.Debugf("waiting room created. Total rooms: %d", totalRooms)

	// Signal creation (non-blocking).
	select {
	case s.waitingRoomCreated <- struct{}{}:
	default:
	}

	pongWR := wr.Marshal()

	s.notifyallusers(&pong.NtfnStreamResponse{
		NotificationType: pong.NotificationType_ON_WR_CREATED,
		Wr:               pongWR,
	})

	return &pong.CreateWaitingRoomResponse{Wr: pongWR}, nil
}

func (s *Server) JoinWaitingRoom(ctx context.Context, req *pong.JoinWaitingRoomRequest) (*pong.JoinWaitingRoomResponse, error) {
	var uid zkidentity.ShortID
	if err := uid.FromString(req.ClientId); err != nil {
		return nil, err
	}
	player := s.gameManager.PlayerSessions.GetPlayer(uid)
	if player == nil {
		return nil, fmt.Errorf("player not found: %s", req.ClientId)
	}

	// Disallow joining multiple rooms.
	for _, existingWR := range s.gameManager.WaitingRoomsSnapshot() {
		for _, p := range existingWR.Players {
			if p.ID.String() == req.ClientId && p.WR != nil {
				return nil, fmt.Errorf("player %s is already in another waiting room", req.ClientId)
			}
		}
	}

	// F2P mode: allow join without escrow.
	if s.isF2P {
		// Locate room.
		wr := s.gameManager.GetWaitingRoom(req.RoomId)
		if wr == nil {
			return nil, fmt.Errorf("waiting room not found: %s", req.RoomId)
		}

		// Add player without escrow binding.
		wr.AddPlayer(player)
		player.WR = wr

		pwr := wr.Marshal()
		for _, p := range wr.Players {
			if p.NotifierStream == nil {
				s.log.Errorf("player %s has nil NotifierStream", p.ID.String())
				continue
			}
			_ = s.notify(p, &pong.NtfnStreamResponse{
				NotificationType: pong.NotificationType_PLAYER_JOINED_WR,
				Message:          fmt.Sprintf("New player joined Waiting Room: %s", player.Nick),
				PlayerId:         p.ID.String(),
				RoomId:           wr.ID,
				Wr:               pwr,
			})
		}

		return &pong.JoinWaitingRoomResponse{Wr: pwr}, nil
	}

	// Non-F2P: require funded escrow (0-conf OK).
	es, err := s.resolveFundedEscrow(uid, req.EscrowId)
	if err != nil {
		return nil, fmt.Errorf("require funded escrow to join room (0-conf ok): %w", err)
	}

	// Locate room.
	wr := s.gameManager.GetWaitingRoom(req.RoomId)
	if wr == nil {
		return nil, fmt.Errorf("waiting room not found: %s", req.RoomId)
	}

	// Add player and bind escrow to this room.
	wr.AddPlayer(player)
	player.WR = wr

	s.roomEscrowsMu.Lock()
	if s.roomEscrows == nil {
		s.roomEscrows = make(map[zkidentity.ShortID]map[string]string)
	}
	if s.roomEscrows[*player.ID] == nil {
		s.roomEscrows[*player.ID] = make(map[string]string)
	}
	s.roomEscrows[*player.ID][wr.ID] = es.escrowID
	// Notify room.
	s.roomEscrowsMu.Unlock()

	pwr := wr.Marshal()
	for _, p := range wr.Players {
		if p.NotifierStream == nil {
			s.log.Errorf("player %s has nil NotifierStream", p.ID.String())
			continue
		}
		_ = s.notify(p, &pong.NtfnStreamResponse{
			NotificationType: pong.NotificationType_PLAYER_JOINED_WR,
			Message:          fmt.Sprintf("New player joined Waiting Room: %s", player.Nick),
			PlayerId:         p.ID.String(),
			RoomId:           wr.ID,
			Wr:               pwr,
		})
	}

	return &pong.JoinWaitingRoomResponse{Wr: pwr}, nil
}

// LeaveWaitingRoom handles a request from a client to leave a waiting room
func (s *Server) LeaveWaitingRoom(ctx context.Context, req *pong.LeaveWaitingRoomRequest) (*pong.LeaveWaitingRoomResponse, error) {
	s.log.Debugf("LeaveWaitingRoom request from client %s for room %s", req.ClientId, req.RoomId)

	var clientID zkidentity.ShortID
	if err := clientID.FromString(req.ClientId); err != nil {
		return &pong.LeaveWaitingRoomResponse{
			Success: false,
			Message: fmt.Sprintf("invalid client ID: %v", err),
		}, nil
	}

	// Get the waiting room
	wr := s.gameManager.GetWaitingRoom(req.RoomId)
	if wr == nil {
		return &pong.LeaveWaitingRoomResponse{
			Success: false,
			Message: "waiting room not found",
		}, nil
	}

	// Check if player is in the room
	player := wr.GetPlayer(&clientID)
	if player == nil {
		return &pong.LeaveWaitingRoomResponse{
			Success: false,
			Message: "player not in waiting room",
		}, nil
	}

	wr.RLock()
	hostID := wr.HostID
	players := wr.Players
	wr.RUnlock()

	// clean up presign artifacts for all players in the room
	for _, p := range players {
		if es := s.escrowForRoomPlayer(*p.ID, wr.ID); es != nil {
			es.clearPreSigns()
		}
	}

	// If the leaving player is the host, close the room entirely.
	if *hostID == clientID {
		// Cancel the room context if present and remove from manager.
		if wr.Cancel != nil {
			wr.Cancel()
		}
		s.gameManager.RemoveWaitingRoom(wr.ID)

		// Broadcast room removal to all users (for UIs that list rooms).
		s.notifyallusers(&pong.NtfnStreamResponse{
			NotificationType: pong.NotificationType_ON_WR_REMOVED,
			Message:          "host left the waiting room so the room was closed",
			RoomId:           wr.ID,
		})
		return &pong.LeaveWaitingRoomResponse{
			Success: true,
			Message: "host left; room closed",
		}, nil
	}
	// Non-host: remove the player from the waiting room
	wr.RemovePlayer(player)

	// Get remaining players and notify them
	remainingPlayers := ponggame.GetRemainingPlayersInWaitingRoom(wr, clientID)
	for _, remainingPlayer := range remainingPlayers {
		_ = s.notify(remainingPlayer, &pong.NtfnStreamResponse{
			NotificationType: pong.NotificationType_OPPONENT_DISCONNECTED,
			Message:          "player left the waiting room",
			RoomId:           wr.ID,
		})
	}

	return &pong.LeaveWaitingRoomResponse{
		Success: true,
		Message: "successfully left waiting room",
	}, nil
}
