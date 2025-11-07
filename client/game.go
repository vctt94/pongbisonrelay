package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/vctt94/pongbisonrelay/pongrpc/grpc/pong"
)

// SignalReadyToPlay signals that the player is ready to start playing
func (pc *PongClient) RefSignalReadyToPlay(gameID string) error {
	ctx := context.Background()

	resp, err := pc.gc.SignalReadyToPlay(ctx, &pong.SignalReadyToPlayRequest{
		ClientId: pc.id,
		GameId:   gameID,
	})
	if err != nil {
		return fmt.Errorf("error signaling ready to play: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("server rejected ready signal: %s", resp.Message)
	}

	return nil
}

// SignalUnready tells the server that the player is no longer ready to play
func (pc *PongClient) RefUnreadyGameStream() error {

	ctx := context.Background()

	// Call the unready RPC method
	_, err := pc.gc.UnreadyGameStream(ctx, &pong.UnreadyGameStreamRequest{
		ClientId: pc.id,
	})
	if err != nil {
		return fmt.Errorf("error signaling not ready: %w", err)
	}

	pc.stopGameStream()

	// Notify UI of state change
	pc.isReady = false
	pc.updatesCh <- UpdatedMsg{}

	return nil
}

// RefStartNtfnStream starts the server->client notification stream.
// gRPC handles reconnects under the hood (when configured), so we just
// read until ctx is canceled or the stream returns a terminal error.
func (pc *PongClient) RefStartNtfnStream(ctx context.Context) error {
	stream, err := pc.gc.StartNtfnStream(ctx, &pong.StartNtfnStreamRequest{
		ClientId: pc.id,
	})
	if err != nil {
		return fmt.Errorf("start ntfn stream: %w", err)
	}
	pc.notifier = stream

	go pc.runNtfnRecv(ctx, stream)
	return nil
}

func (pc *PongClient) runNtfnRecv(ctx context.Context, stream pong.PongGame_StartNtfnStreamClient) {
	pc.log.Infof("ntfn stream started")
	defer pc.log.Infof("ntfn stream stopped")

	for {
		ntfn, err := stream.Recv()
		if err != nil {
			// If our ctx is done, it's a graceful stop.
			select {
			case <-ctx.Done():
				return
			default:
			}
			// Propagate the error once and exit. If gRPC reconnects, the call
			// that created this stream should be re-run by the caller.
			pc.errorsCh <- fmt.Errorf("ntfn stream recv: %w", err)
			return
		}
		pc.handleNtfn(ntfn)
	}
}

func (pc *PongClient) handleNtfn(ntfn *pong.NtfnStreamResponse) {
	switch ntfn.NotificationType {
	case pong.NotificationType_ON_WR_CREATED:
		pc.ntfns.notifyOnWRCreated(ntfn.Wr, time.Now())
		// Forward to updates channel for structured UI updates
		pc.updatesCh <- ntfn

	case pong.NotificationType_MESSAGE:
		pc.updatesCh <- ntfn

	case pong.NotificationType_PLAYER_JOINED_WR:
		pc.ntfns.notifyPlayerJoinedWR(ntfn.Wr, time.Now())
		pc.updatesCh <- ntfn

	case pong.NotificationType_GAME_START:
		if ntfn.Started {
			pc.ntfns.notifyGameStarted(ntfn.GameId, time.Now())
		}
		// Forward start event so UI can transition to initialized state
		pc.updatesCh <- ntfn

	case pong.NotificationType_GAME_END:
		pc.ntfns.notifyGameEnded(ntfn.GameId, ntfn.Message, time.Now())
		pc.log.Infof("%s", ntfn.Message)
		// Forward to updates channel so UI layers (golib plugin) can
		// emit a "game_end" notification to Flutter. Without this,
		// the UI never transitions out of the playing state.
		pc.updatesCh <- ntfn

	case pong.NotificationType_OPPONENT_DISCONNECTED:
		pc.ntfns.notifyPlayerLeftWR(ntfn.Wr, ntfn.PlayerId, time.Now())
		pc.updatesCh <- ntfn

	case pong.NotificationType_BET_AMOUNT_UPDATE:
		if ntfn.PlayerId == pc.id {
			// If pc.BetAmt is read elsewhere concurrently, guard with a mutex.
			pc.betAmt = ntfn.BetAmt
			pc.ntfns.notifyBetAmtChanged(ntfn.PlayerId, ntfn.BetAmt, time.Now())
		}
		// Forward for UI if you want:
		pc.updatesCh <- ntfn

	case pong.NotificationType_ON_WR_REMOVED:
		// Forward room removal event for UI state updates
		pc.updatesCh <- ntfn

	case pong.NotificationType_ON_PLAYER_READY:
		if ntfn.PlayerId == pc.id {
			pc.isReady = ntfn.Ready
			pc.updatesCh <- true
		}
		pc.updatesCh <- ntfn

	case pong.NotificationType_COUNTDOWN_UPDATE,
		pong.NotificationType_GAME_READY_TO_PLAY:
		pc.updatesCh <- ntfn

	default:
		// no-op
	}
}

func (pc *PongClient) RefStartGameStream() error {
	pc.stopGameStream()

	ctx, cancel := context.WithCancel(pc.ctx)

	// Signal readiness after stream is initialized
	stream, err := pc.gc.StartGameStream(ctx, &pong.StartGameStreamRequest{
		ClientId: pc.id,
	})
	if err != nil {
		cancel()
		return fmt.Errorf("error signaling readiness: %w", err)
	}

	pc.Lock()
	pc.stream = stream
	pc.streamCtx = ctx
	pc.streamCancel = cancel
	pc.Unlock()

	go pc.handleGameStream(ctx, stream, cancel)

	return nil
}

func (pc *PongClient) handleGameStream(ctx context.Context, stream pong.PongGame_StartGameStreamClient, cancel context.CancelFunc) {
	defer cancel()
	defer pc.clearStreamRefs(ctx)

	currentStream := stream

mainLoop:
	for {
		update, err := currentStream.Recv()
		if err != nil {
			select {
			case <-ctx.Done():
				pc.log.Infof("game stream stopped: %v", ctx.Err())
				return
			default:
			}

			if errors.Is(err, io.EOF) {
				pc.log.Infof("game stream ended")
				return
			}

			if strings.Contains(err.Error(), "transport is closing") {
				backoff := 500 * time.Millisecond
				maxBackoff := 30 * time.Second
				for {
					select {
					case <-ctx.Done():
						pc.log.Infof("game stream restart canceled")
						return
					case <-time.After(backoff):
						ns, nerr := pc.gc.StartGameStream(ctx, &pong.StartGameStreamRequest{ClientId: pc.id})
						if nerr == nil {
							pc.Lock()
							pc.stream = ns
							pc.Unlock()
							currentStream = ns
							pc.log.Infof("game stream restarted")
							continue mainLoop
						}
						if backoff < maxBackoff {
							backoff *= 2
							if backoff > maxBackoff {
								backoff = maxBackoff
							}
						}
					}
				}
			}

			pc.errorsCh <- fmt.Errorf("game stream error: %v", err)
			return
		}

		// Forward updates to UpdatesCh without spawning a goroutine per frame.
		// If the channel is full, drop the frame to avoid backpressure and lag.
		select {
		case pc.updatesCh <- update:
		default:
			pc.log.Warnf("updates channel full, dropping frame")
		}
	}
}

func (pc *PongClient) RefSendInput(input string) error {
	ctx := context.Background()

	_, err := pc.gc.SendInput(ctx, &pong.PlayerInput{
		Input:        input,
		PlayerId:     pc.id,
		PlayerNumber: pc.playerNumber,
	})
	if err != nil {
		return fmt.Errorf("error sending input: %w", err)
	}
	return nil
}

func (pc *PongClient) stopGameStream() {
	pc.Lock()
	cancel := pc.streamCancel
	pc.streamCancel = nil
	pc.streamCtx = nil
	pc.stream = nil
	pc.Unlock()

	if cancel != nil {
		cancel()
	}
}

func (pc *PongClient) clearStreamRefs(ctx context.Context) {
	pc.Lock()
	if pc.streamCtx == ctx {
		pc.streamCancel = nil
		pc.streamCtx = nil
		pc.stream = nil
	}
	pc.Unlock()
}
