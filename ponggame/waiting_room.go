package ponggame

import (
	"bytes"
	"context"
	"fmt"

	"github.com/companyzero/bisonrelay/clientrpc/types"
	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/vctt94/bisonbotkit/utils"
)

func (wr *WaitingRoom) AddPlayer(player *Player) {
	wr.Lock()
	defer wr.Unlock()
	for _, p := range wr.Players {
		// don't add repeated players
		if p.ID == player.ID {
			return
		}
	}
	wr.Players = append(wr.Players, player)
}

// ReadyPlayers returns all ready players in the waiting room.
func (wr *WaitingRoom) ReadyPlayers() []*Player {
	wr.RLock()
	defer wr.RUnlock()

	// Collect all ready players without mutating room state.
	var selected []*Player
	for _, p := range wr.Players {
		if p.Ready {
			selected = append(selected, p)
		}
	}
	return selected
}

func (wr *WaitingRoom) GetPlayer(clientID *zkidentity.ShortID) *Player {
	wr.RLock()
	defer wr.RUnlock()

	// Check for nil clientID to prevent panic
	if clientID == nil {
		return nil
	}

	for _, player := range wr.Players {
		if player.ID.String() == clientID.String() {
			return player
		}
	}
	return nil
}

func (wr *WaitingRoom) GetPlayers() []*Player {
	wr.RLock()
	defer wr.RUnlock()
	return wr.Players
}

func (wr *WaitingRoom) RemovePlayer(clientID zkidentity.ShortID) {
	wr.Lock()
	defer wr.Unlock()

	// Remove player from Players slice
	for i, player := range wr.Players {
		if *player.ID == clientID {
			wr.Players = append(wr.Players[:i], wr.Players[i+1:]...)
			break
		}
	}

	// Remove all reserved tips for this player
	filteredTips := make([]*types.ReceivedTip, 0, len(wr.ReservedTips))
	for _, tip := range wr.ReservedTips {
		if !bytes.Equal(tip.Uid, clientID.Bytes()) {
			filteredTips = append(filteredTips, tip)
		}
	}
	wr.ReservedTips = filteredTips
}

func (wr *WaitingRoom) length() int {
	wr.RLock()
	defer wr.RUnlock()
	return len(wr.Players)
}

// NewWaitingRoom creates and initializes a new waiting room.
func NewWaitingRoom(hostPlayer *Player, betAmount int64) (*WaitingRoom, error) {
	id, err := utils.GenerateRandomString(16)
	if err != nil {
		return nil, fmt.Errorf("failed to generate waiting room ID: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	return &WaitingRoom{
		Ctx:       ctx,
		Cancel:    cancel,
		ID:        id,
		HostID:    hostPlayer.ID,
		BetAmount: betAmount,
		Players:   []*Player{hostPlayer},
	}, nil
}
