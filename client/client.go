package client

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/decred/dcrd/chaincfg/v3"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/decred/slog"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/vctt94/pongbisonrelay/pongrpc/grpc/pong"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/status"
)

const VERSION = "0.0.1"

type UpdatedMsg struct{}

type PongClient struct {
	sync.RWMutex
	id string

	isReady bool

	betAmt        int64 // bet amt in mAtoms
	playerNumber  int32
	conn          *grpc.ClientConn
	appCfg        *PongConf
	serverVersion string
	serverIsF2P   bool
	// game client
	gc pong.PongGameClient
	// waiting room client
	wr pong.PongWaitingRoomClient
	// referee client
	rc pong.PongRefereeClient

	ntfns *NotificationManager

	log          slog.Logger
	stream       pong.PongGame_StartGameStreamClient
	streamCtx    context.Context
	streamCancel context.CancelFunc
	notifier     pong.PongGame_StartNtfnStreamClient

	ctx    context.Context
	cancel context.CancelFunc

	updatesCh chan tea.Msg
	errorsCh  chan error

	// Settlement session key (in-memory, per-process)
	settlePrivHex    string
	settlePubHex     string
	activeEscrowInfo *EscrowInfo
}

// LoadTLSCreds loads client TLS credentials using the configured cert path,
// with a simple and robust fallback:
// 1) Try cfg.GRPCCertPath
// 2) Try <datadir>/ca.cert
// 3) Create default cert at cfg.GRPCCertPath and retry
func LoadTLSCreds(cfg *PongConf) (credentials.TransportCredentials, error) {
	if cfg == nil {
		return nil, fmt.Errorf("nil TLS config")
	}
	certPath := strings.TrimSpace(cfg.GRPCCertPath)
	if certPath == "" {
		certPath = filepath.Join(cfg.DataDir, "ca.cert")
		cfg.GRPCCertPath = certPath
	}
	if creds, err := credentials.NewClientTLSFromFile(certPath, ""); err == nil {
		return creds, nil
	}
	fallback := filepath.Join(cfg.DataDir, "ca.cert")
	if fallback != certPath {
		if creds, err := credentials.NewClientTLSFromFile(fallback, ""); err == nil {
			cfg.GRPCCertPath = fallback
			return creds, nil
		}
	}
	_ = os.MkdirAll(filepath.Dir(certPath), 0700)
	_ = os.WriteFile(certPath, []byte(defaultServerCertPEM), 0600)
	creds, err := credentials.NewClientTLSFromFile(certPath, "")
	if err != nil {
		return nil, fmt.Errorf("load TLS cert: %w", err)
	}
	return creds, nil
}

func NewPongClient(clientID string, cfg *PongClientCfg) (*PongClient, error) {
	if cfg.LogBackend == nil {
		return nil, fmt.Errorf("client must have logger")
	}
	if cfg.PongConf == nil {
		return nil, fmt.Errorf("client must have PongConf")
	}

	creds, err := LoadTLSCreds(cfg.PongConf)
	if err != nil {
		return nil, err
	}

	dialOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(creds),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:    30 * time.Second,
			Timeout: 10 * time.Second,
		}),
	}

	conn, err := grpc.NewClient(cfg.PongConf.ServerAddr, dialOpts...)
	if err != nil {
		return nil, fmt.Errorf("dial server: %w", err)
	}

	ntfns := cfg.Notifications
	if ntfns == nil {
		ntfns = NewNotificationManager()
	}

	ctx, cancel := context.WithCancel(context.Background())
	pc := &PongClient{
		id:     clientID,
		conn:   conn,
		appCfg: cfg.PongConf,
		gc:     pong.NewPongGameClient(conn),
		wr:     pong.NewPongWaitingRoomClient(conn),
		rc:     pong.NewPongRefereeClient(conn),
		// Larger buffer to absorb bursty game frames without backpressuring producers
		updatesCh: make(chan tea.Msg, 1024),
		errorsCh:  make(chan error, 4),
		log:       cfg.LogBackend.Logger("pongclient"),
		ntfns:     ntfns,
		ctx:       ctx,
		cancel:    cancel,
	}

	err = func() error {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		return pc.initConnection(ctx)
	}()
	if err != nil {
		return nil, fmt.Errorf("init connection: %w", err)
	}

	return pc, nil
}

func (pc *PongClient) ID() string {
	pc.RLock()
	defer pc.RUnlock()
	return pc.id
}

// SetID updates the client ID. This should be called after wallet authentication
// to ensure the client uses the authenticated identity.
func (pc *PongClient) SetID(newID string) {
	pc.Lock()
	defer pc.Unlock()
	pc.id = newID
}

func (pc *PongClient) IsReady() bool {
	return pc.isReady
}

func (pc *PongClient) initConnection(ctx context.Context) error {
	req := &pong.InitConnectionRequest{ClientVersion: VERSION}
	resp, err := pc.gc.InitConnection(ctx, req)
	if err != nil {
		if status.Code(err) == codes.Unimplemented {
			pc.log.Infof("Server does not support InitConnection RPC; assuming escrow-required mode")
			pc.serverVersion = "unknown"
			pc.serverIsF2P = false
			return nil
		}
		return err
	}
	pc.serverVersion = resp.GetServerVersion()
	pc.serverIsF2P = resp.GetIsF2P()
	pc.log.Infof("Server version: %s (F2P=%v)", pc.serverVersion, pc.serverIsF2P)
	return nil
}

func (pc *PongClient) ServerIsF2P() bool {
	pc.RLock()
	defer pc.RUnlock()
	return pc.serverIsF2P
}

func (pc *PongClient) ServerVersion() string {
	pc.RLock()
	defer pc.RUnlock()
	return pc.serverVersion
}

// ServerAddr returns the configured gRPC server address.
func (pc *PongClient) ServerAddr() string {
	pc.RLock()
	defer pc.RUnlock()
	if pc.appCfg == nil {
		return ""
	}
	return pc.appCfg.ServerAddr
}

// GRPCCertPath returns the configured TLS certificate path for the gRPC server.
func (pc *PongClient) GRPCCertPath() string {
	pc.RLock()
	defer pc.RUnlock()
	if pc.appCfg == nil {
		return ""
	}
	return pc.appCfg.GRPCCertPath
}

// AppConfig returns a copy of the current PongConf pointer for read-only use.
func (pc *PongClient) AppConfig() *PongConf {
	pc.RLock()
	defer pc.RUnlock()
	return pc.appCfg
}

// ResolveClientID starts a short-lived BR RPC client to fetch the local
// user's identity and returns it as a hex string. The internal RPC client
// is stopped before returning.
func ResolveClientID(ctx context.Context, appCfg *PongConf) (string, error) {
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

// sessionKeyFilePath returns the path used to persist the settlement session key.
func (pc *PongClient) sessionKeyFilePath() string {
	if pc == nil || pc.appCfg == nil || strings.TrimSpace(pc.appCfg.DataDir) == "" {
		return ""
	}
	return filepath.Join(pc.appCfg.DataDir, "settlement_session_key.json")
}

// GetChainParams returns the chaincfg.Params for the configured network.
// Returns an error if the network is invalid or config is missing.
func (pc *PongClient) GetChainParams() (*chaincfg.Params, error) {
	if pc == nil || pc.appCfg == nil {
		return nil, fmt.Errorf("client or config is nil")
	}
	return pc.appCfg.GetChainParams()
}

func (pc *PongClient) historicSessionsDir() (string, error) {
	base := strings.TrimSpace(pc.sessionKeyFilePath())
	if base == "" {
		return "", fmt.Errorf("session key storage path not configured")
	}
	return filepath.Join(filepath.Dir(base), "historic_sessions"), nil
}

func (pc *PongClient) sessionDataForEscrow(escrowID string) (*SessionKeyData, error) {
	escrowID = strings.TrimSpace(escrowID)
	if escrowID == "" {
		return nil, fmt.Errorf("escrowID required")
	}
	dir, err := pc.historicSessionsDir()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read historic sessions dir: %w", err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasPrefix(name, "sessionkey_") || !strings.HasSuffix(name, ".json") {
			continue
		}
		path := filepath.Join(dir, name)
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var session SessionKeyData
		if err := json.Unmarshal(data, &session); err != nil {
			continue
		}
		if session.EscrowInfo != nil && session.EscrowInfo.EscrowID == escrowID {
			return &session, nil
		}
	}
	return nil, fmt.Errorf("escrow %s not found in historic sessions", escrowID)
}

// saveSettlementSessionKey writes the current session keypair to disk (0600) in JSON.
func (pc *PongClient) saveSettlementSessionKey() error {
	path := pc.sessionKeyFilePath()
	if strings.TrimSpace(path) == "" {
		return nil // no datadir configured; skip persistence in POC mode
	}
	pc.RLock()
	data := SessionKeyData{
		Priv: pc.settlePrivHex,
		Pub:  pc.settlePubHex,
	}
	if pc.activeEscrowInfo != nil {
		copyInfo := *pc.activeEscrowInfo
		data.EscrowInfo = &copyInfo
	}
	pc.RUnlock()
	blob, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	return os.WriteFile(path, blob, 0600)
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
	var data SessionKeyData
	if err := json.Unmarshal(b, &data); err != nil {
		return false, err
	}
	data.Priv = strings.TrimSpace(data.Priv)
	data.Pub = strings.TrimSpace(data.Pub)
	if data.Priv == "" || data.Pub == "" {
		return false, fmt.Errorf("empty session key file")
	}
	if _, err := hex.DecodeString(data.Priv); err != nil {
		return false, fmt.Errorf("bad session privkey in file: %w", err)
	}
	if pubB, err := hex.DecodeString(data.Pub); err != nil || len(pubB) != 33 {
		return false, fmt.Errorf("bad session pubkey in file")
	}
	pc.Lock()
	pc.settlePrivHex = data.Priv
	pc.settlePubHex = data.Pub
	if data.EscrowInfo != nil {
		copyInfo := *data.EscrowInfo
		pc.activeEscrowInfo = &copyInfo
	} else {
		pc.activeEscrowInfo = nil
	}
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

// EscrowInfo represents the data we need to store about an escrow for potential refund
type EscrowInfo struct {
	EscrowID        string `json:"escrow_id"`
	FundingTxid     string `json:"funding_txid"`
	FundingVout     uint32 `json:"funding_vout"`
	FundedAmount    uint64 `json:"funded_amount"`
	RedeemScriptHex string `json:"redeem_script_hex"`
	PKScriptHex     string `json:"pk_script_hex"`
	CSVBlocks       uint32 `json:"csv_blocks"`
	ArchivedAt      int64  `json:"archived_at"`
	FundingVoutSet  bool   `json:"-"`
	FundedAmountSet bool   `json:"-"`
	CSVBlocksSet    bool   `json:"-"`
}

// SessionKeyData includes both the keypair and escrow info for archiving
type SessionKeyData struct {
	Priv       string      `json:"priv"`
	Pub        string      `json:"pub"`
	EscrowInfo *EscrowInfo `json:"escrow_info,omitempty"`
}

// ArchiveSettlementSessionKeyWithEscrow moves the current session key file to a historical
// directory, includes escrow information, and clears in-memory keys.
func (pc *PongClient) ArchiveSettlementSessionKeyWithEscrow(matchID string, escrowInfo *EscrowInfo) error {
	// Clear cached keys in memory
	pc.Lock()
	priv, pub := pc.settlePrivHex, pc.settlePubHex
	pc.settlePrivHex, pc.settlePubHex = "", ""
	pc.activeEscrowInfo = nil
	pc.Unlock()

	if priv == "" || pub == "" {
		return fmt.Errorf("no session key to archive")
	}

	if escrowInfo == nil {
		escrowInfo = &EscrowInfo{}
	}
	base := strings.TrimSpace(pc.sessionKeyFilePath())
	if base == "" {
		return nil
	}

	data := SessionKeyData{Priv: priv, Pub: pub, EscrowInfo: escrowInfo}
	if err := pc.writeHistoricSession(matchID, data, false); err != nil {
		return err
	}
	// Remove the original session key file
	_ = os.Remove(base)

	return nil
}

// ArchiveSettlementSessionKey moves the current session key file to a historical
// directory, namespaced by match ID, and clears in-memory keys.
// This is the legacy version without escrow info - use ArchiveSettlementSessionKeyWithEscrow instead.
func (pc *PongClient) ArchiveSettlementSessionKey(matchID string) error {
	// Clear cached keys in memory
	pc.Lock()
	pc.settlePrivHex, pc.settlePubHex = "", ""
	pc.activeEscrowInfo = nil
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

	raw, err := os.ReadFile(base)
	if err != nil {
		return err
	}
	var data SessionKeyData
	if err := json.Unmarshal(raw, &data); err != nil {
		return err
	}
	if err := pc.writeHistoricSession(matchID, data, false); err != nil {
		return err
	}
	return os.Remove(base)
}

func (pc *PongClient) writeHistoricSession(matchID string, data SessionKeyData, allowOverwrite bool) error {
	base := strings.TrimSpace(pc.sessionKeyFilePath())
	if base == "" {
		return nil
	}
	dir := filepath.Join(filepath.Dir(base), "historic_sessions")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	safe := sanitize(matchID)
	if safe == "" {
		safe = fmt.Sprintf("unknown_%d", time.Now().Unix())
	}
	dst := filepath.Join(dir, fmt.Sprintf("sessionkey_%s.json", safe))
	if !allowOverwrite {
		if _, err := os.Stat(dst); err == nil {
			dst = filepath.Join(dir, fmt.Sprintf("sessionkey_%s_%s.json", safe, time.Now().Format("20060102-150405")))
		}
	}
	blob, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(dst, blob, 0o600)
}

func (pc *PongClient) snapshotActiveSession(name string) error {
	pc.RLock()
	priv, pub := pc.settlePrivHex, pc.settlePubHex
	info := pc.activeEscrowInfo
	pc.RUnlock()
	if priv == "" || pub == "" {
		return fmt.Errorf("no settlement session key present")
	}
	var copyInfo *EscrowInfo
	if info != nil {
		ci := *info
		copyInfo = &ci
	}
	data := SessionKeyData{Priv: priv, Pub: pub, EscrowInfo: copyInfo}
	return pc.writeHistoricSession(name, data, true)
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

// UpdatesCh returns the updates channel for receiving UI updates.
func (pc *PongClient) UpdatesCh() <-chan tea.Msg {
	return pc.updatesCh
}

// ErrorsCh returns the errors channel for receiving error messages.
func (pc *PongClient) ErrorsCh() <-chan error {
	return pc.errorsCh
}

// Close terminates background streams and closes the gRPC connection.
func (pc *PongClient) Close() error {
	if pc == nil {
		return nil
	}
	pc.stopGameStream()
	if pc.cancel != nil {
		pc.cancel()
	}
	if pc.conn != nil {
		return pc.conn.Close()
	}
	return nil
}

// GetSettlementPrivKeyForEscrow returns the private key recorded alongside the
// archived session for the supplied escrow ID.
func (pc *PongClient) GetSettlementPrivKeyForEscrow(escrowID string) (string, error) {
	session, err := pc.sessionDataForEscrow(escrowID)
	if err != nil {
		return "", err
	}
	priv := strings.TrimSpace(session.Priv)
	if priv == "" {
		return "", fmt.Errorf("historic session missing private key for escrow %s", escrowID)
	}
	return priv, nil
}

// CurrentSettlementPubKey returns the currently cached settlement session pubkey.
func (pc *PongClient) CurrentSettlementPubKey() (string, error) {
	pc.RLock()
	defer pc.RUnlock()
	if pc.settlePubHex == "" {
		return "", fmt.Errorf("no settlement session key present")
	}
	return pc.settlePubHex, nil
}

// EscrowDetails contains information about a funded escrow
type EscrowDetails struct {
	EscrowID        string
	FundingTxHash   string
	FundingOutpoint string
	FundingVout     uint32
	FundedAmount    uint64
	RedeemScriptHex string
	PKScriptHex     string
	CSVBlocks       uint32
}

// GetEscrowDetails returns details for a given escrow ID
func (pc *PongClient) GetEscrowDetails(escrowID string) (*EscrowDetails, error) {
	session, err := pc.sessionDataForEscrow(escrowID)
	if err != nil {
		return nil, err
	}
	info := session.EscrowInfo
	if info == nil {
		return nil, fmt.Errorf("historic session missing escrow info for %s", escrowID)
	}
	if strings.TrimSpace(info.FundingTxid) == "" {
		return nil, fmt.Errorf("historic escrow %s missing funding txid", escrowID)
	}
	return &EscrowDetails{
		EscrowID:        info.EscrowID,
		FundingTxHash:   info.FundingTxid,
		FundingOutpoint: fmt.Sprintf("%s:%d", info.FundingTxid, info.FundingVout),
		FundingVout:     info.FundingVout,
		FundedAmount:    info.FundedAmount,
		RedeemScriptHex: info.RedeemScriptHex,
		PKScriptHex:     info.PKScriptHex,
		CSVBlocks:       info.CSVBlocks,
	}, nil
}

// CacheEscrowInfo merges the provided escrow info into the active session file
// so refunds remain possible even if the UI crashes before archiving.
func (pc *PongClient) CacheEscrowInfo(info *EscrowInfo) error {
	if info == nil || strings.TrimSpace(info.EscrowID) == "" {
		return fmt.Errorf("escrow info requires escrow_id")
	}
	pc.Lock()
	if pc.activeEscrowInfo == nil {
		pc.activeEscrowInfo = &EscrowInfo{}
	}
	mergeEscrowInfo(pc.activeEscrowInfo, info)
	pc.Unlock()
	if err := pc.saveSettlementSessionKey(); err != nil {
		return err
	}
	name := fmt.Sprintf("escrow_%s_pending", info.EscrowID)
	return pc.snapshotActiveSession(name)
}

func mergeEscrowInfo(dst, src *EscrowInfo) {
	if dst == nil || src == nil {
		return
	}
	if src.EscrowID != "" {
		dst.EscrowID = src.EscrowID
	}
	if src.FundingTxid != "" {
		dst.FundingTxid = src.FundingTxid
	}
	if src.FundingVoutSet || src.FundingVout != 0 {
		dst.FundingVout = src.FundingVout
	}
	if src.FundedAmountSet || src.FundedAmount != 0 {
		dst.FundedAmount = src.FundedAmount
	}
	if src.RedeemScriptHex != "" {
		dst.RedeemScriptHex = src.RedeemScriptHex
	}
	if src.PKScriptHex != "" {
		dst.PKScriptHex = src.PKScriptHex
	}
	if src.CSVBlocksSet || src.CSVBlocks != 0 {
		dst.CSVBlocks = src.CSVBlocks
	}
	if src.ArchivedAt != 0 {
		dst.ArchivedAt = src.ArchivedAt
	}
}

// LoadHistoricEscrows loads all escrow information from historic session key files
func (pc *PongClient) LoadHistoricEscrows() ([]*EscrowInfo, error) {
	base := strings.TrimSpace(pc.sessionKeyFilePath())
	if base == "" {
		return nil, nil
	}

	dir := filepath.Join(filepath.Dir(base), "historic_sessions")
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No historic sessions yet
		}
		return nil, fmt.Errorf("failed to read historic sessions dir: %w", err)
	}

	var escrows []*EscrowInfo
	skippedLegacy := 0
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasPrefix(entry.Name(), "sessionkey_") || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue // Skip unreadable files
		}

		var sessionData SessionKeyData
		if err := json.Unmarshal(data, &sessionData); err != nil {
			continue // Skip invalid JSON
		}

		if sessionData.EscrowInfo != nil {
			escrows = append(escrows, sessionData.EscrowInfo)
			continue
		}
		skippedLegacy++
		if pc.log != nil {
			pc.log.Warnf("LoadHistoricEscrows: %s missing escrow_info (legacy format)", entry.Name())
		}
	}
	if pc.log != nil {
		pc.log.Infof("LoadHistoricEscrows: loaded %d escrows (skipped %d legacy files)", len(escrows), skippedLegacy)
	}

	return escrows, nil
}

// UpdateHistoricEscrow merges the provided escrow info into an existing historic session file.
func (pc *PongClient) UpdateHistoricEscrow(info *EscrowInfo) error {
	if info == nil || strings.TrimSpace(info.EscrowID) == "" {
		return fmt.Errorf("escrow info requires escrow_id")
	}
	dir, err := pc.historicSessionsDir()
	if err != nil {
		return err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read historic sessions dir: %w", err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasPrefix(entry.Name(), "sessionkey_") || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var session SessionKeyData
		if err := json.Unmarshal(data, &session); err != nil {
			continue
		}
		if session.EscrowInfo == nil || session.EscrowInfo.EscrowID != info.EscrowID {
			continue
		}
		mergeEscrowInfo(session.EscrowInfo, info)
		if session.EscrowInfo.ArchivedAt == 0 {
			session.EscrowInfo.ArchivedAt = time.Now().Unix()
		}
		blob, err := json.MarshalIndent(session, "", "  ")
		if err != nil {
			return err
		}
		if err := os.WriteFile(path, blob, 0o600); err != nil {
			return err
		}
		if pc.log != nil {
			pc.log.Infof("UpdateHistoricEscrow: updated %s", entry.Name())
		}
		return nil
	}
	return fmt.Errorf("historic escrow %s not found", info.EscrowID)
}

// DeleteHistoricEscrow removes the historic session file associated with the
// provided escrow ID. This allows users to clean up refunded escrows.
func (pc *PongClient) DeleteHistoricEscrow(escrowID string) error {
	escrowID = strings.TrimSpace(escrowID)
	if escrowID == "" {
		return fmt.Errorf("escrow_id required")
	}
	dir, err := pc.historicSessionsDir()
	if err != nil {
		return err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("historic sessions directory not found")
		}
		return fmt.Errorf("read historic sessions dir: %w", err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasPrefix(name, "sessionkey_") || !strings.HasSuffix(name, ".json") {
			continue
		}
		path := filepath.Join(dir, name)
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var session SessionKeyData
		if err := json.Unmarshal(data, &session); err != nil {
			continue
		}
		if session.EscrowInfo == nil || session.EscrowInfo.EscrowID != escrowID {
			continue
		}
		if err := os.Remove(path); err != nil {
			return fmt.Errorf("remove historic escrow file: %w", err)
		}
		if pc.log != nil {
			pc.log.Infof("DeleteHistoricEscrow: removed %s (escrow_id=%s)", name, escrowID)
		}
		return nil
	}
	return fmt.Errorf("historic escrow %s not found", escrowID)
}

// LoadHistoricEscrowsFromDir loads all escrow information from historic session key files in the given directory
func LoadHistoricEscrowsFromDir(dataDir string) ([]*EscrowInfo, error) {
	if dataDir == "" {
		return nil, nil
	}

	dir := filepath.Join(dataDir, "historic_sessions")
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No historic sessions yet
		}
		return nil, fmt.Errorf("failed to read historic sessions dir: %w", err)
	}

	var escrows []*EscrowInfo
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasPrefix(entry.Name(), "sessionkey_") || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue // Skip unreadable files
		}

		var sessionData SessionKeyData
		if err := json.Unmarshal(data, &sessionData); err != nil {
			continue // Skip invalid JSON
		}

		if sessionData.EscrowInfo != nil {
			escrows = append(escrows, sessionData.EscrowInfo)
		}
	}

	return escrows, nil
}

// ValidateHistoricRefundSession verifies that there exists a historic session entry
// for the given escrow that contains a usable private key and the minimum escrow
// metadata needed to later build a refund (once funding txid/vout are known).
// Returns (true, "") if valid; otherwise (false, reason).
func (pc *PongClient) ValidateHistoricRefundSession(escrowID string) (bool, string) {
	escrowID = strings.TrimSpace(escrowID)
	if escrowID == "" {
		return false, "escrow_id required"
	}
	session, err := pc.sessionDataForEscrow(escrowID)
	if err != nil {
		return false, err.Error()
	}
	// Validate private key presence and encoding.
	priv := strings.TrimSpace(session.Priv)
	if priv == "" {
		return false, "historic session missing private key"
	}
	if _, err := hex.DecodeString(priv); err != nil {
		return false, "invalid private key encoding in historic session"
	}
	// Validate escrow info presence and critical fields.
	if session.EscrowInfo == nil {
		return false, "historic session missing escrow_info"
	}
	info := session.EscrowInfo
	if strings.TrimSpace(info.RedeemScriptHex) == "" {
		return false, "redeem_script_hex not recorded"
	}
	if strings.TrimSpace(info.PKScriptHex) == "" {
		return false, "pk_script_hex not recorded"
	}
	// CSV is necessary to construct refund path later.
	if info.CSVBlocks == 0 {
		return false, "csv_blocks not recorded"
	}
	// FundedAmount at this stage should reflect the intended bet amount.
	if info.FundedAmount == 0 {
		return false, "funded_amount not recorded"
	}
	// Funding txid/vout can be missing pre-deposit; that's acceptable here.
	return true, ""
}
