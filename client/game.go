package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/vctt94/pong-bisonrelay/pongrpc/grpc/pong"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
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

func (pc *PongClient) reconnect() error {
	pc.reconnectMu.Lock()
	if pc.reconnecting {
		pc.reconnectMu.Unlock()
		return nil // Already reconnecting
	}
	pc.reconnecting = true
	pc.reconnectMu.Unlock()

	defer func() {
		pc.reconnectMu.Lock()
		pc.reconnecting = false
		pc.reconnectMu.Unlock()
	}()

	pc.log.Infof("Attempting to reconnect to server...")

	// Load credentials
	creds, err := credentials.NewClientTLSFromFile(pc.appCfg.GRPCCertPath, "")
	if err != nil {
		return fmt.Errorf("failed to load credentials for reconnection: %w", err)
	}

	// Close existing connection if it's still around
	if pc.conn != nil {
		pc.conn.Close()
	}

	// Implement exponential backoff for reconnection attempts
	backoff := 1 * time.Second
	maxBackoff := 30 * time.Second
	for i := 0; i < 10; i++ { // Try 10 times before giving up
		// Check if context was canceled
		if pc.ctx.Err() != nil {
			return pc.ctx.Err()
		}

		// Attempt to reconnect
		pongConn, err := grpc.Dial(pc.appCfg.ServerAddr,
			grpc.WithTransportCredentials(creds),
			grpc.WithBlock(),
			grpc.WithKeepaliveParams(keepalive.ClientParameters{
				Time:                30 * time.Second, // Send pings every 60 seconds instead of 10
				Timeout:             10 * time.Second, // Wait 20 seconds for ping ack
				PermitWithoutStream: false,            // Allow pings when there are no active streams
			}),
		)

		if err != nil {
			// Sleep with backoff before retrying below
		} else {
			// Successfully reconnected
			pc.conn = pongConn
			pc.gc = pong.NewPongGameClient(pongConn)
			pc.rc = pong.NewPongRefereeClient(pongConn)
			pc.wr = pong.NewPongWaitingRoomClient(pongConn)

			// Re-establish streams
			err = pc.RefStartNtfnStream(pc.ctx)
			if err != nil {
				pc.log.Errorf("Failed to restart notifier after reconnection: %v", err)
				// Close connection and try again
				pongConn.Close()
			} else {
				// If we were in a game, we need to re-establish the game stream
				if pc.stream != nil {
					err = pc.RefStartGameStream()
					if err != nil {
						pc.log.Errorf("Failed to restart game stream after reconnection: %v", err)
						// Continue with the reconnected client even if we couldn't restart the game stream
					}
				}

				pc.log.Infof("Successfully reconnected to server")
				// Send notification that we've reconnected
				pc.UpdatesCh <- UpdatedMsg{}
				return nil
			}
		}

		// Sleep with backoff before retrying
		select {
		case <-pc.ctx.Done():
			return pc.ctx.Err()
		case <-time.After(backoff):
			// Increase backoff for next attempt, but cap it
			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
		}
	}

	return fmt.Errorf("failed to reconnect after multiple attempts")
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

						// Try to reconnect
						reconnectErr := pc.reconnect()
						if reconnectErr != nil {
							pc.ErrorsCh <- fmt.Errorf("failed to reconnect: %v", reconnectErr)
						}
						return // This goroutine ends, but a new one will be started by reconnect()
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
					return
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
