package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/decred/slog"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/vctt94/pong-bisonrelay/pongrpc/grpc/pong"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
)

type UpdatedMsg struct{}

type PongClient struct {
	sync.RWMutex
	ID string

	IsReady bool

	BetAmt       int64 // bet amt in mAtoms
	playerNumber int32
	conn         *grpc.ClientConn
	appCfg       *AppConfig
	// game client
	gc pong.PongGameClient
	// waiting room client
	wr pong.PongWaitingRoomClient
	// referee client
	rc pong.PongRefereeClient

	ntfns *NotificationManager

	log      slog.Logger
	stream   pong.PongGame_StartGameStreamClient
	notifier pong.PongGame_StartNtfnStreamClient

	UpdatesCh chan tea.Msg
	GameCh    chan *pong.GameUpdateBytes
	ErrorsCh  chan error

	// For reconnection handling
	ctx          context.Context
	cancelFunc   context.CancelFunc
	reconnecting bool
	reconnectMu  sync.Mutex

	// Settlement session key (in-memory, per-process)
	settlePrivHex string
	settlePubHex  string
}

// RequireSettlementSessionKey returns the current session key or an error if unset.
func (pc *PongClient) RequireSettlementSessionKey() (string, string, error) {
	pc.RLock()
	defer pc.RUnlock()
	if pc.settlePrivHex == "" || pc.settlePubHex == "" {
		return "", "", fmt.Errorf("no settlement session key; generate one with [K] in the UI or via GenerateNewSettlementSessionKey()")
	}
	return pc.settlePrivHex, pc.settlePubHex, nil
}

// GetSettlementSessionKey returns the currently cached session key (may be empty strings).
func (pc *PongClient) GetSettlementSessionKey() (string, string) {
	pc.RLock()
	defer pc.RUnlock()
	return pc.settlePrivHex, pc.settlePubHex
}

func (pc *PongClient) StartNotifier(ctx context.Context) error {
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

func (pc *PongClient) SignalReady() error {
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

func (pc *PongClient) SendInput(input string) error {
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

func (pc *PongClient) GetWaitingRooms() ([]*pong.WaitingRoom, error) {
	ctx := context.Background()

	res, err := pc.wr.GetWaitingRooms(ctx, &pong.WaitingRoomsRequest{})
	if err != nil {
		return nil, fmt.Errorf("error getting wr: %w", err)
	}
	go func() { pc.UpdatesCh <- res.Wr }()

	return res.Wr, nil
}

func (pc *PongClient) GetWRPlayers() ([]*pong.Player, error) {
	ctx := context.Background()

	res, err := pc.wr.GetWaitingRoom(ctx, &pong.WaitingRoomRequest{})
	if err != nil {
		return nil, fmt.Errorf("error getting wr players: %w", err)
	}
	return res.Wr.Players, nil
}

func (pc *PongClient) CreateWaitingRoom(clientId string, betAmt int64, escrowID string) (*pong.WaitingRoom, error) {
	ctx := context.Background()
	req := &pong.CreateWaitingRoomRequest{HostId: clientId, BetAmt: betAmt}
	if escrowID != "" {
		req.EscrowId = escrowID
	}
	res, err := pc.wr.CreateWaitingRoom(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("error creating wr: %w", err)
	}
	return res.Wr, nil
}

func (pc *PongClient) JoinWaitingRoom(roomID string, escrowID string) (*pong.JoinWaitingRoomResponse, error) {
	ctx := context.Background()
	req := &pong.JoinWaitingRoomRequest{ClientId: pc.ID, RoomId: roomID}
	if escrowID != "" {
		req.EscrowId = escrowID
	}
	res, err := pc.wr.JoinWaitingRoom(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("error joining wr: %w", err)
	}
	return res, nil
}

func (pc *PongClient) LeaveWaitingRoom(roomID string) error {
	ctx := context.Background()
	res, err := pc.wr.LeaveWaitingRoom(ctx, &pong.LeaveWaitingRoomRequest{
		ClientId: pc.ID,
		RoomId:   roomID,
	})
	if err != nil {
		return fmt.Errorf("error leaving waiting room: %w", err)
	}

	if !res.Success {
		return fmt.Errorf("failed to leave waiting room: %s", res.Message)
	}

	return nil
}

func NewPongClient(clientID string, cfg *PongClientCfg) (*PongClient, error) {
	if cfg.Log == nil {
		return nil, fmt.Errorf("client must have logger")
	}

	// Create a cancelable context for the client
	ctx, cancel := context.WithCancel(context.Background())

	// Load the credentials from the certificate file
	creds, err := credentials.NewClientTLSFromFile(cfg.AppCfg.GRPCCertPath, "")
	if err != nil {
		cancel() // Clean up the context
		log.Fatalf("Failed to load credentials: %v", err)
	}

	// Add connection options with healthchecking to detect disconnection faster
	dialOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(creds),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:    30 * time.Second, // Send pings every 60 seconds instead of 10
			Timeout: 10 * time.Second, // Wait 20 seconds for ping ack
		}),
	}

	// Dial the gRPC server with TLS credentials
	pongConn, err := grpc.Dial(cfg.AppCfg.ServerAddr, dialOpts...)
	if err != nil {
		cancel() // Clean up the context
		log.Fatalf("Failed to connect to server: %v", err)
	}

	ntfns := cfg.Notifications
	if ntfns == nil {
		ntfns = NewNotificationManager()
	}

	// Initialize the pongClient instance
	pc := &PongClient{
		ID:         clientID,
		conn:       pongConn,
		appCfg:     cfg.AppCfg,
		gc:         pong.NewPongGameClient(pongConn),
		wr:         pong.NewPongWaitingRoomClient(pongConn),
		rc:         pong.NewPongRefereeClient(pongConn),
		UpdatesCh:  make(chan tea.Msg, 64),
		ErrorsCh:   make(chan error),
		log:        cfg.Log,
		ntfns:      ntfns,
		ctx:        ctx,
		cancelFunc: cancel,
	}

	return pc, nil
}

func (pc *PongClient) Cleanup() {
	if pc.cancelFunc != nil {
		pc.cancelFunc()
	}
	if pc.conn != nil {
		pc.conn.Close()
	}
}

// SignalUnready tells the server that the player is no longer ready to play
func (pc *PongClient) SignalUnready() error {

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

// SignalReadyToPlay signals that the player is ready to start playing
func (pc *PongClient) SignalReadyToPlay(gameID string) error {
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
