package ponggame

import (
	"sync"

	"github.com/companyzero/bisonrelay/zkidentity"
)

type PlayerSessions struct {
	sync.RWMutex
	Sessions map[zkidentity.ShortID]*Player
}

func (ps *PlayerSessions) RemovePlayer(clientID zkidentity.ShortID) {
	ps.Lock()
	defer ps.Unlock()
	delete(ps.Sessions, clientID)
}

func (ps *PlayerSessions) GetPlayer(clientID zkidentity.ShortID) *Player {
	ps.RLock()
	defer ps.RUnlock()
	player := ps.Sessions[clientID]
	return player
}

func (ps *PlayerSessions) CreateSession(clientID zkidentity.ShortID) *Player {
	ps.Lock()
	defer ps.Unlock()

	player := ps.Sessions[clientID]
	if player == nil {
		clientIDCopy := clientID
		player = &Player{
			ID:    &clientIDCopy,
			Score: 0,
		}
		ps.Sessions[clientID] = player
	}

	return player
}
