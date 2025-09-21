package ponggame

import (
	"context"
	"sync"

	"github.com/companyzero/bisonrelay/client/clientintf"
	"github.com/companyzero/bisonrelay/clientrpc/types"
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
	p.RLock()
	defer p.RUnlock()
	return &pong.Player{
		Uid:    p.ID.String(),
		Nick:   p.Nick,
		BetAmt: p.BetAmt,
		Number: p.PlayerNumber,
		Score:  int32(p.Score),
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
	Ctx          context.Context
	Cancel       context.CancelFunc
	ID           string
	HostID       *clientintf.UserID
	Players      []*Player
	BetAmount    int64
	ReservedTips []*types.ReceivedTip
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
	sync.RWMutex

	ID             *zkidentity.ShortID
	Games          map[string]*GameInstance
	WaitingRooms   []*WaitingRoom
	PlayerSessions *PlayerSessions
	PlayerGameMap  map[zkidentity.ShortID]*GameInstance

	Log slog.Logger
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
