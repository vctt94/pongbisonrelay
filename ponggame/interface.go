package ponggame

import (
	"context"
	"sync"

	"github.com/companyzero/bisonrelay/client/clientintf"
	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/decred/slog"
	"github.com/ndabAP/ping-pong/engine"
	"github.com/vctt94/pongbisonrelay/pongrpc/grpc/pong"
)

// Rect represents a bounding box with center position and half-dimensions
type Rect struct {
	Cx    float64 // Center X
	Cy    float64 // Center Y
	HalfW float64 // Half-width
	HalfH float64 // Half-height
}

type Vec2 struct {
	X, Y float64
}

func (v Vec2) Add(w Vec2) Vec2 {
	return Vec2{v.X + w.X, v.Y + w.Y}
}

func (v Vec2) Scale(s float64) Vec2 {
	return Vec2{v.X * s, v.Y * s}
}

type Player struct {
	sync.RWMutex
	ID *zkidentity.ShortID

	Nick           string
	BetAmt         int64
	PlayerNumber   int32 // 1 for player 1, 2 for player 2
	Score          int
	GameStream     pong.PongGame_StartGameStreamServer
	NotifierStream pong.PongGame_StartNtfnStreamServer
	Ready          bool

	// Per-player frame buffer to prevent one slow client from affecting others
	FrameCh chan []byte

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
	p.GameStream = nil
	p.Score = 0
	p.PlayerNumber = 0
	p.BetAmt = 0
	p.Ready = false
	if p.FrameCh != nil {
		close(p.FrameCh)
		p.FrameCh = nil
	}
	p.WR = nil
}

type GameInstance struct {
	sync.RWMutex
	Id          string
	engine      *CanvasEngine
	Framesch    chan []byte
	Inputch     chan []byte
	roundResult chan int32
	Players     []*Player
	cleanedUp   bool
	Running     bool
	ctx         context.Context
	cancel      context.CancelFunc
	Winner      *zkidentity.ShortID

	// betAmt sum of total bets
	betAmt int64

	// Ready to play state
	PlayersReady     map[string]bool
	CountdownStarted bool
	CountdownValue   int
	GameReady        bool

	log slog.Logger
}

type WaitingRoom struct {
	sync.RWMutex
	Ctx       context.Context
	Cancel    context.CancelFunc
	ID        string
	HostID    *clientintf.UserID
	Players   []*Player
	BetAmount int64
}

func (wr *WaitingRoom) Marshal() *pong.WaitingRoom {
	wr.RLock()
	defer wr.RUnlock()
	players := make([]*pong.Player, len(wr.Players))
	for i, player := range wr.Players {
		players[i] = player.Marshal()
	}
	return &pong.WaitingRoom{
		Id:      wr.ID,
		HostId:  wr.HostID.String(),
		Players: players,
		BetAmt:  wr.BetAmount,
	}
}

type GameManager struct {
	ID *zkidentity.ShortID

	PlayerSessions *PlayerSessions

	Games   map[string]*GameInstance
	gamesMu sync.RWMutex

	waitingRoomsMu sync.RWMutex
	WaitingRooms   []*WaitingRoom

	playerGameMapMu sync.RWMutex
	PlayerGameMap   map[zkidentity.ShortID]*GameInstance

	Log slog.Logger
}

// WaitingRoomsSnapshot returns a shallow copy of the waiting rooms slice.
func (g *GameManager) WaitingRoomsSnapshot() []*WaitingRoom {
	g.waitingRoomsMu.RLock()
	defer g.waitingRoomsMu.RUnlock()
	return append([]*WaitingRoom(nil), g.WaitingRooms...)
}

// AppendWaitingRoom appends a waiting room and returns the total count.
func (g *GameManager) AppendWaitingRoom(wr *WaitingRoom) int {
	g.waitingRoomsMu.Lock()
	g.WaitingRooms = append(g.WaitingRooms, wr)
	total := len(g.WaitingRooms)
	g.waitingRoomsMu.Unlock()
	return total
}

// GamesSnapshot returns a shallow copy of the games map.
func (g *GameManager) GamesSnapshot() map[string]*GameInstance {
	g.gamesMu.RLock()
	defer g.gamesMu.RUnlock()
	out := make(map[string]*GameInstance, len(g.Games))
	for k, v := range g.Games {
		out[k] = v
	}
	return out
}

// DeleteGame removes a game by id.
func (g *GameManager) DeleteGame(id string) {
	g.gamesMu.Lock()
	delete(g.Games, id)
	g.gamesMu.Unlock()
}

// RemovePlayerGame removes the player->game mapping.
func (g *GameManager) RemovePlayerGame(clientID zkidentity.ShortID) {
	g.playerGameMapMu.Lock()
	delete(g.PlayerGameMap, clientID)
	g.playerGameMapMu.Unlock()
}

// CancelAllWaitingRooms cancels and clears all waiting rooms.
func (g *GameManager) CancelAllWaitingRooms() {
	g.waitingRoomsMu.Lock()
	for _, wr := range g.WaitingRooms {
		if wr != nil {
			wr.Cancel()
		}
	}
	g.WaitingRooms = nil
	g.waitingRoomsMu.Unlock()
}

// CanvasEngine is a ping-pong engine for browsers with Canvas support
type CanvasEngine struct {
	// Static
	FPS, TPS float64

	Game engine.Game

	// State
	P1Score, P2Score int

	BallPos, BallVel Vec2
	// Replace individual position/velocity fields with vectors
	P1Pos, P2Pos Vec2
	P1Vel, P2Vel Vec2

	// Velocity multiplier that increases over time
	VelocityMultiplier float64
	VelocityIncrease   float64

	// Error of the current tick
	Err error

	// Engine debug state
	log slog.Logger

	mu sync.RWMutex
}

// StartGameStreamRequest encapsulates the data needed to start a game stream.
type StartGameStreamRequest struct {
	ClientID zkidentity.ShortID
	Stream   pong.PongGame_StartGameStreamServer
	MinBet   float64
	IsF2P    bool
	Log      slog.Logger
}
