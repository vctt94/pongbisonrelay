package ponggame

import (
	"context"
	"fmt"

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
		p.RLock()
		ready := p.Ready
		p.RUnlock()
		if ready {
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

func (wr *WaitingRoom) RemovePlayer(player *Player) bool {
	if player == nil {
		return true
	}
	wr.Lock()
	defer wr.Unlock()

	for i, p := range wr.Players {
		if p == player {
			// Remove from slice
			wr.Players = append(wr.Players[:i], wr.Players[i+1:]...)
			// Clear back-ref
			p.WR = nil
			return true
		}
	}
	return false
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
