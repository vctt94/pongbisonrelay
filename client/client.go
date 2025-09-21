package client

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/companyzero/bisonrelay/clientrpc/types"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/decred/slog"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/vctt94/bisonbotkit/botclient"
	"github.com/vctt94/bisonbotkit/logging"
	"github.com/vctt94/pongbisonrelay/pongrpc/grpc/pong"
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

	ctx    context.Context
	cancel context.CancelFunc

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
	if cfg.AppCfg == nil {
		return nil, fmt.Errorf("client must have AppCfg")
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
	ctx, cancel := context.WithCancel(context.Background())
	pc := &PongClient{
		ID:        clientID,
		conn:      pongConn,
		appCfg:    cfg.AppCfg,
		gc:        pong.NewPongGameClient(pongConn),
		wr:        pong.NewPongWaitingRoomClient(pongConn),
		rc:        pong.NewPongRefereeClient(pongConn),
		UpdatesCh: make(chan tea.Msg, 64),
		ErrorsCh:  make(chan error, 4),
		log:       cfg.Log,
		ntfns:     ntfns,
		ctx:       ctx,
		cancel:    cancel,
	}

	// Quick gRPC sanity check (fail fast on bad addr/cert)
	if _, err := pc.RefGetWaitingRooms(); err != nil {
		_ = pongConn.Close()
		cancel()
		return nil, fmt.Errorf("gRPC server connection failed: %w", err)
	}

	// Start notifier stream in background tied to client lifecycle
	go func() {
		if err := pc.RefStartNtfnStream(pc.ctx); err != nil {
			select {
			case pc.ErrorsCh <- fmt.Errorf("ntfn stream error: %w", err):
			default:
			}
		}
	}()

	return pc, nil
}

// ResolveClientID starts a short-lived BR RPC client to fetch the local
// user's identity and returns it as a hex string. The internal RPC client
// is stopped before returning.
func ResolveClientID(ctx context.Context, appCfg *AppConfig) (string, error) {
	if appCfg == nil || appCfg.BR == nil {
		return "", fmt.Errorf("missing BR config in AppConfig")
	}

	c, err := botclient.NewClient(appCfg.BR)
	if err != nil {
		return "", fmt.Errorf("create botclient: %w", err)
	}

	// Run the RPC client in the background while we query identity.
	runCtx, runCancel := context.WithCancel(context.Background())
	defer runCancel()
	go func() { _ = c.RPCClient.Run(runCtx) }()

	// Retry identity query until the RPC is ready, up to a short deadline.
	deadline := time.Now().Add(10 * time.Second)
	var pii types.PublicIdentity
	for {
		// Use a short per-attempt timeout, but honor the parent ctx.
		attemptCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		err = c.Chat.UserPublicIdentity(attemptCtx, &types.PublicIdentityReq{}, &pii)
		cancel()
		if err == nil {
			break
		}
		if time.Now().After(deadline) || attemptCtx.Err() != nil {
			return "", fmt.Errorf("get public identity: %w", err)
		}
		time.Sleep(200 * time.Millisecond)
	}

	return hex.EncodeToString(pii.Identity[:]), nil
}

func SetupLogging(appCfg *AppConfig, appName string) (*logging.LogBackend, slog.Logger, error) {
	if appCfg == nil || appCfg.BR == nil {
		return nil, nil, fmt.Errorf("missing AppConfig or BR config")
	}
	useStdout := false
	lb, err := logging.NewLogBackend(logging.LogConfig{
		LogFile:        filepath.Join(appCfg.DataDir, "logs", fmt.Sprintf("%s.log", strings.TrimSpace(appName))),
		DebugLevel:     appCfg.BR.Debug,
		MaxLogFiles:    10,
		MaxBufferLines: 1000,
		UseStdout:      &useStdout,
	})
	if err != nil {
		return nil, nil, err
	}
	return lb, lb.Logger("BotClient"), nil
}

// sessionKeyFilePath returns the path used to persist the settlement session key.
func (pc *PongClient) sessionKeyFilePath() string {
	if pc == nil || pc.appCfg == nil || strings.TrimSpace(pc.appCfg.DataDir) == "" {
		return ""
	}
	return filepath.Join(pc.appCfg.DataDir, "settlement_session_key.json")
}

// saveSettlementSessionKey writes the current session keypair to disk (0600) in JSON.
func (pc *PongClient) saveSettlementSessionKey() error {
	path := pc.sessionKeyFilePath()
	if strings.TrimSpace(path) == "" {
		return nil // no datadir configured; skip persistence in POC mode
	}
	type pair struct {
		Priv string `json:"priv"`
		Pub  string `json:"pub"`
	}
	pc.RLock()
	data, err := json.MarshalIndent(pair{Priv: pc.settlePrivHex, Pub: pc.settlePubHex}, "", "  ")
	pc.RUnlock()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

// loadSettlementSessionKey loads a previously saved session keypair from disk.
// It returns (true, nil) if a valid key was loaded and cached in memory.
func (pc *PongClient) loadSettlementSessionKey() (bool, error) {
	path := pc.sessionKeyFilePath()
	if strings.TrimSpace(path) == "" {
		return false, nil
	}
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	type pair struct {
		Priv string `json:"priv"`
		Pub  string `json:"pub"`
	}
	var p pair
	if err := json.Unmarshal(b, &p); err != nil {
		return false, err
	}
	p.Priv = strings.TrimSpace(p.Priv)
	p.Pub = strings.TrimSpace(p.Pub)
	if p.Priv == "" || p.Pub == "" {
		return false, fmt.Errorf("empty session key file")
	}
	if _, err := hex.DecodeString(p.Priv); err != nil {
		return false, fmt.Errorf("bad session privkey in file: %w", err)
	}
	if pubB, err := hex.DecodeString(p.Pub); err != nil || len(pubB) != 33 {
		return false, fmt.Errorf("bad session pubkey in file")
	}
	pc.Lock()
	pc.settlePrivHex = p.Priv
	pc.settlePubHex = p.Pub
	pc.Unlock()
	return true, nil
}

// GenerateNewSettlementSessionKey always creates a new session key and overwrites the cached one.
func (pc *PongClient) GenerateNewSettlementSessionKey() (string, string, error) {
	pc.Lock()
	p, err := secp256k1.GeneratePrivateKey()
	if err != nil {
		pc.Unlock()
		return "", "", err
	}
	pc.settlePrivHex = hex.EncodeToString(p.Serialize())
	pc.settlePubHex = hex.EncodeToString(p.PubKey().SerializeCompressed())
	pc.Unlock()
	if err := pc.saveSettlementSessionKey(); err != nil {
		return "", "", fmt.Errorf("save session key: %w", err)
	}
	return pc.settlePrivHex, pc.settlePubHex, nil
}

// currentOrLoadSettlementSessionKey returns the cached session keypair if present,
// otherwise attempts to load it from disk into memory and returns it. The boolean
// indicates whether a key was found (either cached or loaded).
func (pc *PongClient) currentOrLoadSettlementSessionKey() (string, string, bool, error) {
	pc.RLock()
	priv, pub := pc.settlePrivHex, pc.settlePubHex
	pc.RUnlock()
	if priv != "" && pub != "" {
		return priv, pub, true, nil
	}
	ok, err := pc.loadSettlementSessionKey()
	if err != nil {
		return "", "", false, err
	}
	if ok {
		pc.RLock()
		priv, pub = pc.settlePrivHex, pc.settlePubHex
		pc.RUnlock()
		if priv != "" && pub != "" {
			return priv, pub, true, nil
		}
	}
	return "", "", false, nil
}

// Close terminates background streams and closes the gRPC connection.
func (pc *PongClient) Close() error {
	if pc == nil {
		return nil
	}
	if pc.cancel != nil {
		pc.cancel()
	}
	if pc.conn != nil {
		return pc.conn.Close()
	}
	return nil
}
