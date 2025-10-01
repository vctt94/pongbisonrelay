package ponggame

import (
	"sync"

	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/vctt94/pongbisonrelay/pongrpc/grpc/pong"
)

type Player struct {
	sync.RWMutex
	ID *zkidentity.ShortID

	Nick         string
	BetAmt       int64
	PlayerNumber int32 // 1 for player 1, 2 for player 2
	Score        int
	Ready        bool

	// Per-player frame buffer to prevent one slow client from affecting others
	FrameCh chan []byte

	// Per-player send queues; dedicated server goroutines consume and call
	// stream.Send() to avoid concurrent writes on the same gRPC stream.
	ntfnQ chan *pong.NtfnStreamResponse
	gameQ chan []byte

	WR *WaitingRoom
}

func (p *Player) Marshal() *pong.Player {
	if p == nil || p.ID == nil {
		return nil
	}
	p.RLock()
	defer p.RUnlock()
	return &pong.Player{
		Uid:    p.ID.String(),
		Nick:   p.Nick,
		BetAmt: p.BetAmt,
		Number: p.PlayerNumber,
		Score:  int32(p.Score),
		Ready:  p.Ready,
	}
}

func (p *Player) ResetPlayer() {
	p.Lock()
	defer p.Unlock()

	p.Score = 0
	p.PlayerNumber = 0
	p.BetAmt = 0
	p.Ready = false
	p.WR = nil
}

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
		// Initialize per-player send queues.
		player.ntfnQ = make(chan *pong.NtfnStreamResponse, 64)
		player.gameQ = make(chan []byte, 64)
		ps.Sessions[clientID] = player
	}

	return player
}

// EnqueueNotif enqueues a notification to this player's notification queue.
// Returns false if the queue is unavailable or full.
func (p *Player) EnqueueNotif(resp *pong.NtfnStreamResponse) bool {
	if p == nil || resp == nil || p.ntfnQ == nil {
		return false
	}
	select {
	case p.ntfnQ <- resp:
		return true
	default:
		// Drop if queue is full to avoid blocking producers.
		return false
	}
}

// EnqueueGameBytes enqueues a game frame for this player.
// Returns false if the queue is unavailable or full.
func (p *Player) EnqueueGameBytes(b []byte) bool {
	if p == nil || b == nil || p.gameQ == nil {
		return false
	}
	select {
	case p.gameQ <- b:
		return true
	default:
		// Drop if queue is full.
		return false
	}
}

// NtfnQueue exposes a receive-only view of the player's notification queue.
func (p *Player) NtfnQueue() <-chan *pong.NtfnStreamResponse {
	return p.ntfnQ
}

// GameQueue exposes a receive-only view of the player's game bytes queue.
func (p *Player) GameQueue() <-chan []byte {
	return p.gameQ
}
