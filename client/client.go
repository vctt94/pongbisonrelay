package client

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/decred/slog"

	tea "github.com/charmbracelet/bubbletea"

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

	creds, err := credentials.NewClientTLSFromFile(cfg.AppCfg.GRPCCertPath, "")
	if err != nil {
		return nil, fmt.Errorf("load TLS cert: %w", err)
	}

	dialOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(creds),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:    30 * time.Second,
			Timeout: 10 * time.Second,
		}),
	}

	conn, err := grpc.NewClient(cfg.AppCfg.ServerAddr, dialOpts...)
	if err != nil {
		return nil, fmt.Errorf("dial server: %w", err)
	}

	ntfns := cfg.Notifications
	if ntfns == nil {
		ntfns = NewNotificationManager()
	}

	ctx, cancel := context.WithCancel(context.Background())
	pc := &PongClient{
		ID:        clientID,
		conn:      conn,
		appCfg:    cfg.AppCfg,
		gc:        pong.NewPongGameClient(conn),
		wr:        pong.NewPongWaitingRoomClient(conn),
		rc:        pong.NewPongRefereeClient(conn),
		UpdatesCh: make(chan tea.Msg, 64),
		ErrorsCh:  make(chan error, 4),
		log:       cfg.Log,
		ntfns:     ntfns,
		ctx:       ctx,
		cancel:    cancel,
	}

	// quick check (fail fast on bad addr/cert), with timeout.
	if func() error {
		cctx, ccancel := context.WithTimeout(ctx, 5*time.Second)
		defer ccancel()
		_, err := pc.RefGetWaitingRooms(cctx) // or your existing method with ctx
		return err
	}() != nil {
		cancel()
		_ = conn.Close()
		return nil, fmt.Errorf("gRPC server connection failed")
	}

	return pc, nil
}

// ResolveClientID starts a short-lived BR RPC client to fetch the local
// user's identity and returns it as a hex string. The internal RPC client
// is stopped before returning.
func ResolveClientID(ctx context.Context, appCfg *AppConfig) (string, error) {
	// if appCfg == nil || appCfg.BR == nil {
	// 	return "", fmt.Errorf("missing BR config in AppConfig")
	// }

	// c, err := botclient.NewClient(appCfg.BR)
	// if err != nil {
	// 	return "", fmt.Errorf("create botclient: %w", err)
	// }

	// // Run the RPC client in the background while we query identity.
	// runCtx, runCancel := context.WithCancel(context.Background())
	// defer runCancel()
	// go func() { _ = c.RPCClient.Run(runCtx) }()

	// // Retry identity query until the RPC is ready, up to a short deadline.
	// deadline := time.Now().Add(10 * time.Second)
	// var pii types.PublicIdentity
	// for {
	// 	// Use a short per-attempt timeout, but honor the parent ctx.
	// 	attemptCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	// 	err = c.Chat.UserPublicIdentity(attemptCtx, &types.PublicIdentityReq{}, &pii)
	// 	cancel()
	// 	if err == nil {
	// 		break
	// 	}
	// 	if time.Now().After(deadline) || attemptCtx.Err() != nil {
	// 		return "", fmt.Errorf("get public identity: %w", err)
	// 	}
	// 	time.Sleep(200 * time.Millisecond)
	// }

	// return hex.EncodeToString(pii.Identity[:]), nil

	// For now, avoid BR identity and generate a random 32-byte hex ID.
	var b [32]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("generate random id: %w", err)
	}
	return hex.EncodeToString(b[:]), nil
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

// allow letters/digits and -_.; map everything else (incl. '/', '\', '|', quotes) to '_'
func sanitize(matchID string) string {
	matchID = strings.TrimSpace(matchID)
	if matchID == "" {
		return ""
	}
	mapped := strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z',
			r >= 'A' && r <= 'Z',
			r >= '0' && r <= '9',
			r == '-', r == '_', r == '.':
			return r
		default:
			return '_'
		}
	}, matchID)
	// avoid hidden/awkward names
	mapped = strings.Trim(mapped, "._")
	if mapped == "" {
		return ""
	}
	return mapped
}

// ArchiveSettlementSessionKey moves the current session key file to a historical
// directory, namespaced by match ID, and clears in-memory keys.
func (pc *PongClient) ArchiveSettlementSessionKey(matchID string) error {
	// Clear cached keys in memory
	pc.Lock()
	pc.settlePrivHex, pc.settlePubHex = "", ""
	pc.Unlock()

	base := strings.TrimSpace(pc.sessionKeyFilePath())
	if base == "" {
		return nil
	}
	if _, err := os.Stat(base); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	// Ensure destination dir
	dir := filepath.Join(filepath.Dir(base), "historic_sessions")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}

	// Safe, portable name (fallback to timestamp)
	safe := sanitize(matchID)
	if safe == "" {
		safe = fmt.Sprintf("unknown_%d", time.Now().Unix())
	}
	dst := filepath.Join(dir, fmt.Sprintf("sessionkey_%s.json", safe))

	// If a file with same name exists, add a short timestamp suffix
	if _, err := os.Stat(dst); err == nil {
		dst = filepath.Join(dir, fmt.Sprintf("sessionkey_%s_%s.json",
			safe, time.Now().Format("20060102-150405")))
	}

	// Move: try rename; if cross-device, copy then remove
	if err := os.Rename(base, dst); err != nil {
		srcF, rerr := os.Open(base)
		if rerr != nil {
			return rerr
		}
		defer srcF.Close()

		// 0600: only current user can read the archived key
		dstF, werr := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
		if werr != nil {
			return werr
		}
		if _, cerr := io.Copy(dstF, srcF); cerr != nil {
			dstF.Close()
			_ = os.Remove(dst) // best effort cleanup
			return cerr
		}
		if cerr := dstF.Close(); cerr != nil {
			return cerr
		}
		if derr := os.Remove(base); derr != nil {
			return derr
		}
	}
	return nil
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
