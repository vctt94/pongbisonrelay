package client

import (
	"fmt"
	"log"
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

	// Settlement session key (in-memory, per-process)
	settlePrivHex string
	settlePubHex  string
}

func NewPongClient(clientID string, cfg *PongClientCfg) (*PongClient, error) {
	if cfg.Log == nil {
		return nil, fmt.Errorf("client must have logger")
	}
	// Load the credentials from the certificate file
	creds, err := credentials.NewClientTLSFromFile(cfg.AppCfg.GRPCCertPath, "")
	if err != nil {
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
		log.Fatalf("Failed to connect to server: %v", err)
	}

	ntfns := cfg.Notifications
	if ntfns == nil {
		ntfns = NewNotificationManager()
	}

	// Initialize the pongClient instance
	pc := &PongClient{
		ID:        clientID,
		conn:      pongConn,
		appCfg:    cfg.AppCfg,
		gc:        pong.NewPongGameClient(pongConn),
		wr:        pong.NewPongWaitingRoomClient(pongConn),
		rc:        pong.NewPongRefereeClient(pongConn),
		UpdatesCh: make(chan tea.Msg, 64),
		ErrorsCh:  make(chan error),
		log:       cfg.Log,
		ntfns:     ntfns,
	}

	return pc, nil
}
