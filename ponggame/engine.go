package ponggame

import (
	"context"
	"errors"
	"time"

	"github.com/decred/slog"
	"google.golang.org/protobuf/proto"

	"github.com/vctt94/pongbisonrelay/pongrpc/grpc/pong"

	"github.com/ndabAP/ping-pong/engine"
)

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
				return
			case <-frameTimer.C:
				// Apply any pending inputs for this frame, then advance physics once.
				// Drain input channel without blocking to coalesce inputs.
				drained := false
				for !drained {
					select {
					case key, ok := <-inputch:
						if !ok {
							// Input channel closed; stop draining
							inputch = nil
							drained = true
							continue
						}
						if len(key) == 0 {
							continue
						}
						in := &pong.PlayerInput{}
						if err := proto.Unmarshal(key, in); err != nil {
							continue
						}
						if in.PlayerNumber == 1 {
							switch in.Input {
							case "ArrowUp":
								e.P1Vel = Vec2{0, -y_vel_ratio * e.Game.Height}
							case "ArrowDown":
								e.P1Vel = Vec2{0, y_vel_ratio * e.Game.Height}
							case "ArrowUpStop":
								if e.P1Vel.Y < 0 {
									e.P1Vel.Y = 0
								}
							case "ArrowDownStop":
								if e.P1Vel.Y > 0 {
									e.P1Vel.Y = 0
								}
							}
						} else if in.PlayerNumber == 2 {
							switch in.Input {
							case "ArrowUp":
								e.P2Vel = Vec2{0, -y_vel_ratio * e.Game.Height}
							case "ArrowDown":
								e.P2Vel = Vec2{0, y_vel_ratio * e.Game.Height}
							case "ArrowUpStop":
								if e.P2Vel.Y < 0 {
									e.P2Vel.Y = 0
								}
							case "ArrowDownStop":
								if e.P2Vel.Y > 0 {
									e.P2Vel.Y = 0
								}
							}
						}
					default:
						drained = true
					}
				}
				e.mu.Lock()
				e.tick()
				// round end check: read Err under lock
				p1win := errors.Is(e.Err, engine.ErrP1Win)
				p2win := errors.Is(e.Err, engine.ErrP2Win)

				var winner int32
				if p1win || p2win {
					if p1win {
						e.P1Score++
						winner = 1
					} else {
						e.P2Score++
						winner = 2
					}

					e.mu.Unlock()

					// Send round result and exit goroutine
					select {
					case roundResult <- winner:
					case <-ctx.Done():
						return
					}
					return // Exit the goroutine after round ends
				}

				// Build the frame snapshot under lock
				var gu pong.GameUpdate
				gu.GameWidth, gu.GameHeight = e.Game.Width, e.Game.Height
				gu.P1Width, gu.P1Height = e.Game.P1.Width, e.Game.P1.Height
				gu.P2Width, gu.P2Height = e.Game.P2.Width, e.Game.P2.Height
				gu.BallWidth, gu.BallHeight = e.Game.Ball.Width, e.Game.Ball.Height
				gu.P1Score, gu.P2Score = int32(e.P1Score), int32(e.P2Score)
				gu.BallX, gu.BallY = e.BallPos.X, e.BallPos.Y
				gu.P1X, gu.P1Y = e.P1Pos.X, e.P1Pos.Y
				gu.P2X, gu.P2Y = e.P2Pos.X, e.P2Pos.Y
				gu.P1YVelocity, gu.P2YVelocity = e.P1Vel.Y, e.P2Vel.Y
				gu.BallXVelocity, gu.BallYVelocity = e.BallVel.X, e.BallVel.Y
				// Do not send server FPS/TPS; clients control render FPS

				e.mu.Unlock()

				// marshal (no lock held)
				b, err := proto.Marshal(&gu)
				if err == nil {
					select {
					case framesch <- b:
						// Frame sent successfully
					case <-ctx.Done():
						return
					default:
						// Channel is full, drop this frame to prevent blocking
					}
				}
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
