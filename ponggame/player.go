package ponggame

import (
	"fmt"
	"sync"

	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/vctt94/pongbisonrelay/pongrpc/grpc/pong"
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

// SendNotif serializes sends on the player's NotifierStream to avoid
// concurrent Send races on the same gRPC stream.
func (p *Player) SendNotif(resp *pong.NtfnStreamResponse) error {
	if p == nil {
		return fmt.Errorf("nil player")
	}
	if p.NotifierStream == nil {
		return fmt.Errorf("player stream nil")
	}
	p.ntfnSendMu.Lock()
	defer p.ntfnSendMu.Unlock()
	return p.NotifierStream.Send(resp)
}

// SendGameBytes serializes sends on the player's GameStream to avoid
// concurrent Send races on the same gRPC stream.
func (p *Player) SendGameBytes(b []byte) error {
	if p == nil {
		return fmt.Errorf("nil player")
	}
	if p.GameStream == nil {
		return fmt.Errorf("player game stream nil")
	}
	p.gameSendMu.Lock()
	defer p.gameSendMu.Unlock()
	return p.GameStream.Send(&pong.GameUpdateBytes{Data: b})
}
