package server

import (
	"context"
	"fmt"

	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/vctt94/pong-bisonrelay/ponggame"
	"github.com/vctt94/pong-bisonrelay/pongrpc/grpc/pong"
)

//GetWaitingRoom ??

func (s *Server) GetWaitingRoom(ctx context.Context, req *pong.WaitingRoomRequest) (*pong.WaitingRoomResponse, error) {
	s.Lock()
	defer s.Unlock()

	wr := s.gameManager.GetWaitingRoom(req.RoomId)
	if wr == nil {
		return nil, fmt.Errorf("waiting room not found: %s", req.RoomId)
	}

	return &pong.WaitingRoomResponse{
		Wr: wr.MarshalProto(),
	}, nil
}

func (s *Server) GetWaitingRooms(ctx context.Context, req *pong.WaitingRoomsRequest) (*pong.WaitingRoomsResponse, error) {
	s.Lock()
	defer s.Unlock()

	pongWaitingRooms := make([]*pong.WaitingRoom, len(s.gameManager.WaitingRooms))
	for i, room := range s.gameManager.WaitingRooms {
		wr, err := room.Marshal()
		if err != nil {
			return nil, err
		}
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

	// Require funded escrow (0-conf OK).
	es, err := s.resolveFundedEscrow(hostID.String(), req.EscrowId)
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
	s.Lock()
	if s.roomEscrows == nil {
		s.roomEscrows = make(map[string]map[string]string)
	}
	if s.roomEscrows[wr.ID] == nil {
		s.roomEscrows[wr.ID] = make(map[string]string)
	}
	s.roomEscrows[wr.ID][hostID.String()] = es.escrowID
	s.Unlock()

	// Add to list of rooms.
	s.gameManager.Lock()
	s.gameManager.WaitingRooms = append(s.gameManager.WaitingRooms, wr)
	totalRooms := len(s.gameManager.WaitingRooms)
	s.gameManager.Unlock()

	s.log.Debugf("waiting room created. Total rooms: %d", totalRooms)

	// Signal creation (non-blocking).
	select {
	case s.waitingRoomCreated <- struct{}{}:
	default:
	}

	// Notify all users.
	pongWR, err := wr.Marshal()
	if err != nil {
		return nil, err
	}
	s.RLock()
	for _, user := range s.users {
		if user.NotifierStream == nil {
			s.log.Errorf("user %s without NotifierStream", user.ID)
			continue
		}
		_ = user.NotifierStream.Send(&pong.NtfnStreamResponse{
			Wr:               pongWR,
			NotificationType: pong.NotificationType_ON_WR_CREATED,
		})
	}
	s.RUnlock()

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
	s.gameManager.Lock()
	for _, existingWR := range s.gameManager.WaitingRooms {
		for _, p := range existingWR.Players {
			if p.ID.String() == req.ClientId && p.WR != nil {
				s.gameManager.Unlock()
				return nil, fmt.Errorf("player %s is already in another waiting room", req.ClientId)
			}
		}
	}
	s.gameManager.Unlock()

	// Require funded escrow (0-conf OK).
	es, err := s.resolveFundedEscrow(uid.String(), req.EscrowId)
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

	s.Lock()
	if s.roomEscrows == nil {
		s.roomEscrows = make(map[string]map[string]string)
	}
	if s.roomEscrows[wr.ID] == nil {
		s.roomEscrows[wr.ID] = make(map[string]string)
	}
	s.roomEscrows[wr.ID][player.ID.String()] = es.escrowID
	s.Unlock()

	// Notify room.
	pwr, err := wr.Marshal()
	if err != nil {
		return nil, err
	}
	for _, p := range wr.Players {
		if p.NotifierStream == nil {
			s.log.Errorf("player %s has nil NotifierStream", p.ID.String())
			continue
		}
		_ = p.NotifierStream.Send(&pong.NtfnStreamResponse{
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

	// Remove the player from the waiting room
	wr.RemovePlayer(clientID)

	// If player was the host and there are other players, assign a new host
	if wr.HostID.String() == clientID.String() && len(wr.Players) > 0 {
		wr.Lock()
		wr.HostID = wr.Players[0].ID
		wr.Unlock()
	}

	// If the room is now empty, remove it
	if len(wr.Players) == 0 {
		s.gameManager.RemoveWaitingRoom(wr.ID)
	} else {
		// Notify remaining players that someone left
		pwrMarshaled, err := wr.Marshal()
		if err == nil {
			for _, p := range wr.Players {
				// Send notification to remaining players
				p.NotifierStream.Send(&pong.NtfnStreamResponse{
					NotificationType: pong.NotificationType_PLAYER_LEFT_WR,
					RoomId:           wr.ID,
					Wr:               pwrMarshaled,
					PlayerId:         req.ClientId,
				})
			}
		}
	}

	// Reset the player's waiting room reference
	player.WR = nil

	return &pong.LeaveWaitingRoomResponse{
		Success: true,
		Message: "successfully left waiting room",
	}, nil
}
