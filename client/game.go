package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/vctt94/pong-bisonrelay/pongrpc/grpc/pong"
)

// SignalReadyToPlay signals that the player is ready to start playing
func (pc *PongClient) RefSignalReadyToPlay(gameID string) error {
	ctx := context.Background()

	resp, err := pc.gc.SignalReadyToPlay(ctx, &pong.SignalReadyToPlayRequest{
		ClientId: pc.ID,
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
		ClientId: pc.ID,
	})
	if err != nil {
		return fmt.Errorf("error signaling not ready: %w", err)
	}

	// If we have an active game stream, close it
	if pc.stream != nil {
		pc.stream = nil
	}

	// Notify UI of state change
	pc.IsReady = false
	pc.UpdatesCh <- UpdatedMsg{}

	return nil
}

func (pc *PongClient) RefStartNtfnStream(ctx context.Context) error {
	// Creates game start stream so we can notify when the game starts
	gameStartedStream, err := pc.gc.StartNtfnStream(ctx, &pong.StartNtfnStreamRequest{
		ClientId: pc.ID,
	})
	if err != nil {
		return fmt.Errorf("error creating notifier stream: %w", err)
	}
	pc.notifier = gameStartedStream

	go func() {
		for {
			select {
			case <-ctx.Done():
				pc.log.Infof("ntfn stream closed")
				return
			default:
				ntfn, err := pc.notifier.Recv()
				if err != nil {
					if errors.Is(err, io.EOF) || strings.Contains(err.Error(), "transport is closing") ||
						strings.Contains(err.Error(), "connection is being forcefully terminated") {

						// Recreate notifier stream with backoff and continue
						backoff := 500 * time.Millisecond
						maxBackoff := 30 * time.Second
						for {
							select {
							case <-ctx.Done():
								pc.log.Infof("ntfn stream restart canceled")
								return
							case <-time.After(backoff):
								ns, nerr := pc.gc.StartNtfnStream(ctx, &pong.StartNtfnStreamRequest{ClientId: pc.ID})
								if nerr == nil {
									pc.notifier = ns
									pc.log.Infof("ntfn stream restarted")
									// Successfully restarted; continue outer loop
									continue
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

					pc.ErrorsCh <- fmt.Errorf("notifier stream error: %v", err)
					return
				}

				// Handle notifications based on NotificationType
				switch ntfn.NotificationType {
				case pong.NotificationType_ON_WR_CREATED:
					pc.ntfns.notifyOnWRCreated(ntfn.Wr, time.Now())
				case pong.NotificationType_MESSAGE:
					pc.UpdatesCh <- ntfn
				case pong.NotificationType_PLAYER_JOINED_WR:
					pc.ntfns.notifyPlayerJoinedWR(ntfn.Wr, time.Now())
				case pong.NotificationType_PLAYER_LEFT_WR:
					pc.ntfns.notifyPlayerLeftWR(ntfn.Wr, ntfn.PlayerId, time.Now())
				case pong.NotificationType_GAME_START:
					if ntfn.Started {
						pc.ntfns.notifyGameStarted(ntfn.GameId, time.Now())
					}
				case pong.NotificationType_GAME_END:
					pc.ntfns.notifyGameEnded(ntfn.GameId, ntfn.Message, time.Now())
					pc.log.Infof("%s", ntfn.Message)
				case pong.NotificationType_OPPONENT_DISCONNECTED:
				case pong.NotificationType_BET_AMOUNT_UPDATE:
					if ntfn.PlayerId == pc.ID {
						pc.BetAmt = ntfn.BetAmt
						pc.ntfns.notifyBetAmtChanged(ntfn.PlayerId, ntfn.BetAmt, time.Now())
					}
				case pong.NotificationType_ON_PLAYER_READY:
					if ntfn.PlayerId == pc.ID {
						pc.IsReady = ntfn.Ready
						pc.UpdatesCh <- true
					}
					// Forward notification to UI for any player ready event
					pc.UpdatesCh <- ntfn
				case pong.NotificationType_COUNTDOWN_UPDATE:
					// Forward countdown updates to UI
					pc.UpdatesCh <- ntfn
				case pong.NotificationType_GAME_READY_TO_PLAY:
					// Forward game ready to play notifications to UI
					pc.UpdatesCh <- ntfn
				default:
				}
			}
		}
	}()

	return nil
}

func (pc *PongClient) RefStartGameStream() error {
	ctx := context.Background()

	// Signal readiness after stream is initialized
	stream, err := pc.gc.StartGameStream(ctx, &pong.StartGameStreamRequest{
		ClientId: pc.ID,
	})
	if err != nil {
		return fmt.Errorf("error signaling readiness: %w", err)
	}

	// Set the stream before starting the goroutine
	pc.stream = stream

	// Use a separate goroutine to handle the stream
	go func() {
		for {
			update, err := pc.stream.Recv()
			if err != nil {
				if errors.Is(err, io.EOF) || strings.Contains(err.Error(), "transport is closing") {
					// Recreate game stream with backoff and continue
					backoff := 500 * time.Millisecond
					maxBackoff := 30 * time.Second
					for {
						select {
						case <-ctx.Done():
							pc.log.Infof("game stream restart canceled")
							return
						case <-time.After(backoff):
							ns, nerr := pc.gc.StartGameStream(context.Background(), &pong.StartGameStreamRequest{ClientId: pc.ID})
							if nerr == nil {
								pc.stream = ns
								pc.log.Infof("game stream restarted")
								// Successfully restarted; continue outer loop to Recv again
								continue
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

				pc.ErrorsCh <- fmt.Errorf("game stream error: %v", err)
				return
			}
			// Forward updates to UpdatesCh
			go func() { pc.UpdatesCh <- update }()
		}
	}()

	return nil
}

func (pc *PongClient) RefSendInput(input string) error {
	ctx := context.Background()

	_, err := pc.gc.SendInput(ctx, &pong.PlayerInput{
		Input:        input,
		PlayerId:     pc.ID,
		PlayerNumber: pc.playerNumber,
	})
	if err != nil {
		return fmt.Errorf("error sending input: %w", err)
	}
	return nil
}
