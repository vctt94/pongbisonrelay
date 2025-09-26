package client

import (
	"context"
	"fmt"

	"github.com/vctt94/pongbisonrelay/pongrpc/grpc/pong"
)

func (pc *PongClient) RefGetWaitingRoom(roomID string) (*pong.WaitingRoom, error) {
	ctx := context.Background()
	res, err := pc.wr.GetWaitingRoom(ctx, &pong.WaitingRoomRequest{})
	if err != nil {
		return nil, fmt.Errorf("error getting wr: %w", err)
	}
	return res.Wr, nil
}

func (pc *PongClient) RefGetWaitingRooms(ctx context.Context) ([]*pong.WaitingRoom, error) {

	res, err := pc.wr.GetWaitingRooms(ctx, &pong.WaitingRoomsRequest{})
	if err != nil {
		return nil, fmt.Errorf("error getting wr: %w", err)
	}
	go func() { pc.updatesCh <- res.Wr }()

	return res.Wr, nil
}

func (pc *PongClient) RefCreateWaitingRoom(clientId string, betAmt int64, escrowID string) (*pong.WaitingRoom, error) {
	ctx := context.Background()
	req := &pong.CreateWaitingRoomRequest{HostId: clientId, BetAmt: betAmt}
	if escrowID != "" {
		req.EscrowId = escrowID
	}
	res, err := pc.wr.CreateWaitingRoom(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("error creating wr: %w", err)
	}
	return res.Wr, nil
}

func (pc *PongClient) RefJoinWaitingRoom(roomID string, escrowID string) (*pong.JoinWaitingRoomResponse, error) {
	ctx := context.Background()
	req := &pong.JoinWaitingRoomRequest{ClientId: pc.id, RoomId: roomID}
	if escrowID != "" {
		req.EscrowId = escrowID
	}
	res, err := pc.wr.JoinWaitingRoom(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("error joining wr: %w", err)
	}
	return res, nil
}

func (pc *PongClient) RefLeaveWaitingRoom(roomID string) error {
	ctx := context.Background()
	res, err := pc.wr.LeaveWaitingRoom(ctx, &pong.LeaveWaitingRoomRequest{
		ClientId: pc.id,
		RoomId:   roomID,
	})
	if err != nil {
		return fmt.Errorf("error leaving waiting room: %w", err)
	}

	if !res.Success {
		return fmt.Errorf("failed to leave waiting room: %s", res.Message)
	}

	return nil
}
