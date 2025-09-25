package ponggame

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/decred/slog"
	"google.golang.org/protobuf/proto"

	"github.com/vctt94/pongbisonrelay/pongrpc/grpc/pong"

	"github.com/ndabAP/ping-pong/engine"
)

// Pool for GameUpdate objects to reduce allocations
var gameUpdatePool = sync.Pool{
	New: func() interface{} {
		return &pong.GameUpdate{}
	},
}

// New returns a new Canvas engine for browsers with Canvas support
func New(g engine.Game) *CanvasEngine {
	e := new(CanvasEngine)
	e.Game = g
	e.FPS = DEFAULT_FPS
	e.TPS = 1000.0 / e.FPS
	e.VelocityIncrease = DEFAULT_VEL_INCR

	return e
}

// SetDebug sets the Canvas engines debug state
func (e *CanvasEngine) SetLogger(log slog.Logger) *CanvasEngine {
	e.log = log
	return e
}

// SetFPS sets the Canvas engines frames per second
func (e *CanvasEngine) SetFPS(fps uint) *CanvasEngine {
	if fps <= 0 {
		panic("fps must be greater zero")
	}
	e.log.Debugf("fps %d", fps)
	e.FPS = float64(fps)
	e.TPS = 1000.0 / e.FPS
	return e
}

// Error returns the Canvas engines error
func (e *CanvasEngine) Error() error {
	return e.Err
}

// NewRound resets the ball, players and starts a new round. It accepts
// a frames channel to write into and input channel to read from
func (e *CanvasEngine) NewRound(ctx context.Context, framesch chan<- []byte, inputch <-chan []byte, roundResult chan<- int32) {
	time.Sleep(time.Second)
	e.reset()

	// Calculates and writes frames
	go func() {
		frameTimer := time.NewTicker(time.Duration(1000.0/e.FPS) * time.Millisecond)
		defer frameTimer.Stop()

		for {
			select {
			case <-ctx.Done():
				e.log.Debug("exiting")
				return
			case <-frameTimer.C:
				e.tick()

				if errors.Is(e.Err, engine.ErrP1Win) {
					e.P1Score += 1

					// Send the winner's ID through the roundResult channel
					select {
					case roundResult <- 1:
					case <-ctx.Done():
						return
					}

					return
				} else if errors.Is(e.Err, engine.ErrP2Win) {
					e.P2Score += 1

					// Send the winner's ID through the roundResult channel
					select {
					case roundResult <- 2:
					case <-ctx.Done():
						return
					}

					return
				}

				// Use pooled object to reduce allocations
				gameUpdateFrame := gameUpdatePool.Get().(*pong.GameUpdate)

				// Reset the object
				gameUpdateFrame.Reset()

				// Populate the frame data
				gameUpdateFrame.GameWidth = e.Game.Width
				gameUpdateFrame.GameHeight = e.Game.Height
				gameUpdateFrame.P1Width = e.Game.P1.Width
				gameUpdateFrame.P1Height = e.Game.P1.Height
				gameUpdateFrame.P2Width = e.Game.P2.Width
				gameUpdateFrame.P2Height = e.Game.P2.Height
				gameUpdateFrame.BallWidth = e.Game.Ball.Width
				gameUpdateFrame.BallHeight = e.Game.Ball.Height
				gameUpdateFrame.P1Score = int32(e.P1Score)
				gameUpdateFrame.P2Score = int32(e.P2Score)
				gameUpdateFrame.BallX = e.BallPos.X
				gameUpdateFrame.BallY = e.BallPos.Y
				gameUpdateFrame.P1X = e.P1Pos.X
				gameUpdateFrame.P1Y = e.P1Pos.Y
				gameUpdateFrame.P2X = e.P2Pos.X
				gameUpdateFrame.P2Y = e.P2Pos.Y
				gameUpdateFrame.P1YVelocity = e.P1Vel.Y
				gameUpdateFrame.P2YVelocity = e.P2Vel.Y
				gameUpdateFrame.BallXVelocity = e.BallVel.X
				gameUpdateFrame.BallYVelocity = e.BallVel.Y
				gameUpdateFrame.Fps = e.FPS
				gameUpdateFrame.Tps = e.TPS

				protoTick, err := proto.Marshal(gameUpdateFrame)
				if err != nil {
					e.log.Errorf("Error marshaling protobuf: %v", err)
				}
				// Return object to pool
				gameUpdatePool.Put(gameUpdateFrame)

				select {
				case framesch <- protoTick:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	// Reads user input and moves player one according to it
	go func() {
		for {
			select {
			case key, ok := <-inputch:
				if !ok {
					// Input channel is closed; exit goroutine
					e.log.Debug("Input channel closed; exiting input reader goroutine")
					return
				}
				if len(key) == 0 {
					// Empty input received; possibly due to closed channel
					e.log.Debug("Received empty input data; exiting")
					return
				}

				in := &pong.PlayerInput{}
				err := proto.Unmarshal(key, in)
				if err != nil {
					e.log.Errorf("Failed to unmarshal input: %v", err)
					// Decide whether to continue or exit; here we'll continue
					continue
				}

				// Process the valid input
				if in.PlayerNumber == int32(1) {
					switch k := in.Input; k {
					case "ArrowUp":
						e.p1Up()
					case "ArrowDown":
						e.p1Down()
					case "ArrowUpStop":
						// Stop upward movement
						if e.P1Vel.Y < 0 {
							e.P1Vel.Y = 0
						}
					case "ArrowDownStop":
						// Stop downward movement
						if e.P1Vel.Y > 0 {
							e.P1Vel.Y = 0
						}
					}
				} else {
					switch k := in.Input; k {
					case "ArrowUp":
						e.p2Up()
					case "ArrowDown":
						e.p2Down()
					case "ArrowUpStop":
						// Stop upward movement
						if e.P2Vel.Y < 0 {
							e.P2Vel.Y = 0
						}
					case "ArrowDownStop":
						// Stop downward movement
						if e.P2Vel.Y > 0 {
							e.P2Vel.Y = 0
						}
					}
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}

// State returns the current state of the canvas engine
func (e *CanvasEngine) State() struct {
	PaddleWidth, PaddleHeight float64
	BallWidth, BallHeight     float64
	P1PosX, P1PosY            float64
	P2PosX, P2PosY            float64
	BallPosX, BallPosY        float64
	BallVelX, BallVelY        float64
	FPS, TPS                  float64
} {
	return struct {
		PaddleWidth, PaddleHeight float64
		BallWidth, BallHeight     float64
		P1PosX, P1PosY            float64
		P2PosX, P2PosY            float64
		BallPosX, BallPosY        float64
		BallVelX, BallVelY        float64
		FPS, TPS                  float64
	}{
		PaddleWidth:  e.Game.P1.Width,
		PaddleHeight: e.Game.P1.Height,
		BallWidth:    e.Game.Ball.Width,
		BallHeight:   e.Game.Ball.Height,
		P1PosX:       e.P1Pos.X,
		P1PosY:       e.P1Pos.Y,
		P2PosX:       e.P2Pos.X,
		P2PosY:       e.P2Pos.Y,
		BallPosX:     e.BallPos.X,
		BallPosY:     e.BallPos.Y,
		BallVelX:     e.BallVel.X,
		BallVelY:     e.BallVel.Y,
		FPS:          e.FPS,
		TPS:          e.TPS,
	}
}
