package server

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/decred/dcrd/chaincfg/chainhash"
	"github.com/decred/dcrd/chaincfg/v3"
	"github.com/decred/dcrd/dcrutil/v4"
	"github.com/decred/dcrd/rpcclient/v8"
	"github.com/decred/dcrd/wire"
	"github.com/decred/slog"
	"github.com/vctt94/bisonbotkit/logging"
	pongbisonrelay "github.com/vctt94/pongbisonrelay"
	"github.com/vctt94/pongbisonrelay/chainwatcher"
	"github.com/vctt94/pongbisonrelay/ponggame"
	"github.com/vctt94/pongbisonrelay/pongrpc/grpc/pong"
	"github.com/vctt94/pongbisonrelay/server/serverdb"
)

const (
	name    = "pong"
	version = "v0.0.0"

	WarnSendBlock = 50 * time.Millisecond
	ErrSendBlock  = 250 * time.Millisecond
)

type ServerConfig struct {
	ServerDir string

	MinBetAmt             float64
	IsF2P                 bool
	ReadyTimeoutSeconds   int
	DebugLevel            string
	DebugGameManagerLevel string
	LogBackend            *logging.LogBackend

	// dcrd RPC connectivity
	DcrdHostPort    string // e.g. 127.0.0.1:19109
	DcrdRPCCertPath string // path to rpc.cert
	DcrdRPCUser     string
	DcrdRPCPass     string

	// Adaptor secret seed (hex, 32 bytes recommended). Used to deterministically
	// derive per-branch adaptor secrets bound to match/input/sighash.
	// For POC, if empty, a built-in default will be used.
	AdaptorSecret string

	// Network specifies the Decred network to use: "mainnet" or "testnet"
	// Defaults to "testnet" if empty
	Network string
}

// PreSignCtx stores all artifacts needed to finalize using the exact same
// draft and message digest that were used during the presign phase.
//
// It binds the presign to:
// - the specific input (txid:vout)
// - the redeem script and its version (implicitly v0 here)
// - the exact serialized draft transaction
// - the sighash digest used to compute s'
// - the adaptor point used during presign
//
// The winner can then finalize with s = s' + gamma (mod n) using the same m.
type PreSignCtx struct {
	InputID         string // "txid:vout"
	RedeemScriptHex string
	DraftHex        string             // exact serialized tx used at presign
	MHex            string             // 32-byte sighash for (DraftHex, RedeemScriptHex, idx, SIGHASH_ALL)
	RLineCompressed []byte             // 33 bytes, even-Y (0x02)
	SLine32         []byte             // 32 bytes
	TCompressed     []byte             // 33 bytes (if used in adaptor domain)
	WinnerUID       zkidentity.ShortID // tie to player/session (owner uid)
	Branch          int32              // 0 = A-wins, 1 = B-wins (payout branch)
}

// escrowSession represents a pre-match funding session for a single player.
type escrowSession struct {
	// ----------------- immutable identity & params -----------------
	boundInputID    string
	boundInput      *pong.EscrowUTXO
	escrowID        string
	ownerUID        zkidentity.ShortID
	compPubkey      []byte // 33 bytes
	payoutPubkey    []byte // 33 bytes
	betAtoms        uint64
	csvBlocks       uint32
	redeemScriptHex string
	pkScriptHex     string

	// ----------------- runtime state (protected by mu) -------------
	mu        sync.RWMutex
	latest    chainwatcher.DepositUpdate // watcher-pushed snapshot (Confs, UTXOCount, OK, At)
	lastUTXOs []*pong.EscrowUTXO         // optional cache for settlement (first UTXO, etc.)
	unsubW    func()                     // watcher unsubscribe hook
	// cancelTrack cancels the background trackEscrow goroutine associated
	// with this escrow session.
	cancelTrack context.CancelFunc

	player  *ponggame.Player       // optional: current player binding
	preSign map[string]*PreSignCtx // presign artifacts by input_id "txid:vout"
}

type Server struct {
	pong.UnimplementedPongGameServer
	pong.UnimplementedPongWaitingRoomServer
	pong.UnimplementedPongRefereeServer
	pong.UnimplementedPongAuthServer

	log                slog.Logger
	isF2P              bool
	network            string // Decred network: "mainnet" or "testnet"
	minBetAmt          float64
	waitingRoomCreated chan struct{}

	usersMu sync.RWMutex
	users   map[zkidentity.ShortID]*ponggame.Player

	gameManager *ponggame.GameManager

	httpServer        *http.Server
	activeNtfnStreams sync.Map
	activeGameStreams sync.Map
	db                serverdb.ServerDB

	appdata string

	// dcrd RPC client
	dcrd *rpcclient.Client

	// chain watcher for tip + mempool
	watcher *chainwatcher.ChainWatcher

	// Escrow-first funding state
	escrowsMu sync.RWMutex
	escrows   map[string]*escrowSession

	roomEscrowsMu sync.RWMutex
	roomEscrows   map[zkidentity.ShortID]map[string]string // owner_uid -> roomID ->  escrow_id
	// v0-min defaults
	pocFeeAtoms uint64

	// Secret seed for adaptor gamma derivation.
	adaptorSecret string

	// Chain parameters for address generation
	params *chaincfg.Params

	// In-memory auth/session state and HTTP auth server
	auth authState
}

func NewServer(id *zkidentity.ShortID, cfg ServerConfig) (*Server, error) {
	dbPath := filepath.Join(cfg.ServerDir, "server.db")
	db, err := serverdb.NewBoltDB(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if cfg.LogBackend == nil {
		return nil, fmt.Errorf("log is nil")
	}
	bknd, err := logging.NewLogBackend(logging.LogConfig{
		LogFile:        filepath.Join(cfg.ServerDir, "logs", "gamemanager.log"),
		DebugLevel:     cfg.DebugGameManagerLevel,
		MaxLogFiles:    10,
		MaxBufferLines: 1000,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize game manager logger: %w", err)
	}
	logGM := bknd.Logger("GM")
	s := &Server{
		appdata:            cfg.ServerDir,
		log:                cfg.LogBackend.Logger("Server"),
		db:                 db,
		isF2P:              cfg.IsF2P,
		minBetAmt:          cfg.MinBetAmt,
		waitingRoomCreated: make(chan struct{}, 1),
		users:              make(map[zkidentity.ShortID]*ponggame.Player),
		gameManager: &ponggame.GameManager{
			ID:             id,
			Games:          make(map[string]*ponggame.GameInstance),
			WaitingRooms:   []*ponggame.WaitingRoom{},
			PlayerSessions: &ponggame.PlayerSessions{Sessions: make(map[zkidentity.ShortID]*ponggame.Player)},
			Log:            logGM,
			PlayerGameMap:  make(map[zkidentity.ShortID]*ponggame.GameInstance),
		},
		adaptorSecret: cfg.AdaptorSecret,
	}

	if cfg.ReadyTimeoutSeconds <= 0 {
		return nil, fmt.Errorf("readytimeoutseconds cfg param must be greater than 0")
	}
	s.gameManager.ReadyTimeoutSeconds = cfg.ReadyTimeoutSeconds

	// Initialize chain parameters based on network config
	params, err := initChainParams(cfg.Network)
	if err != nil {
		return nil, err
	}
	s.params = params
	s.network = cfg.Network
	if s.network == "" {
		s.network = "mainnet" // default
	}
	s.log.Infof("Using %s chain parameters", s.params.Name)

	// Log F2P status as early as possible.
	if cfg.IsF2P {
		s.log.Infof("Free-to-Play mode ENABLED (no escrow required)")
	} else {
		s.log.Infof("Free-to-Play mode DISABLED (escrow required)")
	}

	if cfg.DcrdHostPort == "" || cfg.DcrdRPCUser == "" || cfg.DcrdRPCPass == "" || cfg.DcrdRPCCertPath == "" {
		return nil, fmt.Errorf("incomplete dcrd config: host=%q user=%q pass_set=%t cert=%q", cfg.DcrdHostPort, cfg.DcrdRPCUser, cfg.DcrdRPCPass != "", cfg.DcrdRPCCertPath)
	}
	b, err := os.ReadFile(cfg.DcrdRPCCertPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read dcrd rpc cert at %s: %w", cfg.DcrdRPCCertPath, err)
	}
	s.log.Infof("Connecting to dcrd host=%s user=%s cert=%s endpoint=ws", cfg.DcrdHostPort, cfg.DcrdRPCUser, cfg.DcrdRPCCertPath)
	connCfg := &rpcclient.ConnConfig{
		Host:         cfg.DcrdHostPort,
		User:         cfg.DcrdRPCUser,
		Pass:         cfg.DcrdRPCPass,
		Endpoint:     "ws",
		Certificates: b,
	}
	// Enable event-driven notifications from dcrd for fast 0-conf updates.
	ntfnHandlers := &rpcclient.NotificationHandlers{
		OnTxAccepted: func(hash *chainhash.Hash, _ dcrutil.Amount) {
			if s.watcher != nil && hash != nil {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				go func() { defer cancel(); s.watcher.ProcessTxAcceptedHash(ctx, hash) }()
			}
		},
		OnBlockConnected: func(_ []byte, _ [][]byte) {
			if s.watcher != nil {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				go func() { defer cancel(); s.watcher.ProcessBlockConnected(ctx) }()
			}
		},
	}

	c, err := rpcclient.New(connCfg, ntfnHandlers)
	if err != nil {
		return nil, fmt.Errorf("failed to create dcrd rpc client (host=%s user=%s cert=%s): %w", cfg.DcrdHostPort, cfg.DcrdRPCUser, cfg.DcrdRPCCertPath, err)
	}
	s.dcrd = c
	s.log.Infof("Connected to dcrd at %s", cfg.DcrdHostPort)

	// Start chain watcher to keep tip and mempool for watched scripts
	s.watcher = chainwatcher.NewChainWatcher(s.log, s.dcrd)

	// Subscribe to dcrd notifications for tx/mempool.
	// Non-verbose is sufficient; we also hooked verbose variant above.
	if err := s.dcrd.NotifyNewTransactions(context.Background(), false); err != nil {
		s.log.Warnf("dcrd: NotifyNewTransactions failed: %v", err)
	}
	if err := s.dcrd.NotifyBlocks(context.Background()); err != nil {
		s.log.Warnf("dcrd: NotifyBlocks failed: %v", err)
	}

	return s, nil
}

// cleanupEscrowSessionsForPlayers cleans up all escrow sessions for the given players.
// This includes canceling trackEscrow goroutines, unsubscribing from chain watcher,
// clearing presign artifacts, and removing sessions from memory.
func (s *Server) cleanupEscrowSessionsForPlayers(players []*ponggame.Player) {
	s.escrowsMu.Lock()
	var escrowsToDelete []string
	for _, p := range players {
		// Find and clean up escrow sessions for this player
		for escrowID, es := range s.escrows {
			if es != nil && es.ownerUID == *p.ID {
				// Clear player binding first (while holding lock) to stop notifications
				// This must happen before canceling context so trackEscrow sees nil player
				es.mu.Lock()
				es.player = nil
				es.mu.Unlock()
				// Cancel the trackEscrow goroutine
				if es.cancelTrack != nil {
					es.cancelTrack()
				}
				// Unsubscribe from chain watcher
				if es.unsubW != nil {
					es.unsubW()
				}
				// Clear presign artifacts
				es.clearPreSigns()
				// Mark for deletion
				escrowsToDelete = append(escrowsToDelete, escrowID)
			}
		}
	}
	// Remove escrow sessions from the map
	for _, escrowID := range escrowsToDelete {
		delete(s.escrows, escrowID)
	}
	s.escrowsMu.Unlock()

	// Clean up room escrow mappings
	s.roomEscrowsMu.Lock()
	for _, p := range players {
		delete(s.roomEscrows, *p.ID)
	}
	s.roomEscrowsMu.Unlock()
}

func (s *Server) handleDisconnect(clientID zkidentity.ShortID) {
	// Cancel any active streams for this client
	if cancel, ok := s.activeNtfnStreams.Load(clientID); ok {
		if cancelFn, isCancel := cancel.(context.CancelFunc); isCancel {
			cancelFn()
		}
	}
	if cancel, ok := s.activeGameStreams.Load(clientID); ok {
		if cancelFn, isCancel := cancel.(context.CancelFunc); isCancel {
			cancelFn()
		}
	}

	s.usersMu.Lock()
	delete(s.users, clientID)
	s.usersMu.Unlock()

	player := s.gameManager.PlayerSessions.GetPlayer(clientID)
	if player != nil {
		// Clean up escrow sessions for this player
		s.cleanupEscrowSessionsForPlayers([]*ponggame.Player{player})
		s.gameManager.PlayerSessions.RemovePlayer(clientID)
	}

	s.gameManager.RemovePlayerFromWaitingRoom(clientID)
}

func (s *Server) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			// Call the server's Shutdown method
			if err := s.Shutdown(ctx); err != nil {
				s.log.Errorf("Error during server shutdown: %v", err)
			}

			return nil

		case <-s.waitingRoomCreated:
			s.log.Debugf("New waiting room created")

			for _, wr := range s.gameManager.WaitingRoomsSnapshot() {
				if wr.Ctx.Err() == nil { // Only manage rooms with active contexts
					s.log.Debugf("Managing waiting room: %s", wr.ID)
					go s.manageWaitingRoom(wr.Ctx, wr)
				}
			}
		}
	}
}

// Shutdown forcefully shuts down the server, closing HTTP server, database, waiting rooms, and games.
func (s *Server) Shutdown(ctx context.Context) error {
	// Stop chain watcher first so background RPCs stop
	if s.watcher != nil {
		s.watcher.Stop()
	}
	// Stop HTTP server first
	if s.httpServer != nil {
		s.log.Info("Shutting down HTTP server...")
		if err := s.httpServer.Shutdown(ctx); err != nil {
			s.log.Errorf("Error shutting down HTTP server: %v", err)
		}
	}

	// Forcefully terminate all active games
	s.log.Info("Terminating all active games...")
	for id, game := range s.gameManager.GamesSnapshot() {
		s.log.Debugf("Forcefully terminating game: %s", id)
		// Close the frame channel to signal goroutines to exit
		game.Cleanup()
	}

	// Cancel all active streams before cleaning up resources
	s.log.Info("Canceling all active streams...")
	s.activeNtfnStreams.Range(func(key, value interface{}) bool {
		if cancel, ok := value.(context.CancelFunc); ok {
			cancel()
		}
		return true
	})
	s.activeGameStreams.Range(func(key, value interface{}) bool {
		if cancel, ok := value.(context.CancelFunc); ok {
			cancel()
		}
		return true
	})

	// Give a moment for goroutines to clean up
	time.Sleep(200 * time.Millisecond)

	// Clean up game resources before closing database
	s.log.Info("Shutting down waiting rooms and games...")
	s.gameManager.CancelAllWaitingRooms()

	// Close database LAST after all operations are done
	s.log.Info("Closing database...")
	if err := s.db.Close(); err != nil {
		s.log.Errorf("Error closing database: %v", err)
	}

	s.log.Info("Server shut down completed.")
	return nil
}

func (s *Server) handleGameEnd(ctx context.Context, game *ponggame.GameInstance) {
	// Use the game's waiting room reference
	gameWR := game.WaitingRoom

	players := gameWR.ReadyPlayers()
	winner := game.Winner
	var winnerID string
	if winner != nil {
		winnerID = winner.String()
		s.log.Infof("Game ended. Winner: %s", winnerID)
	} else {
		s.log.Infof("Game ended in a draw.")
	}

	// Notify players of game outcome (no transfers)
	for _, player := range players {
		message := "Game ended in a draw."
		if player.ID == winner {
			message = "Congratulations, you won!"
		} else if winner != nil {
			message = "Sorry, you lost."
		}
		_ = s.notify(player, &pong.NtfnStreamResponse{
			NotificationType: pong.NotificationType_GAME_END,
			Message:          message,
			GameId:           game.Id,
		})
		// delete player from gameManager PlayerGameMap
		s.gameManager.RemovePlayerGame(*player.ID)
	}

	// Ensure all players are marked not ready before the next match cycle and notify listeners.
	for _, player := range players {
		if player == nil || player.ID == nil {
			continue
		}

		player.Lock()
		player.Ready = false
		player.Unlock()

		pwr := gameWR.Marshal()
		for _, p := range gameWR.Players {
			_ = s.notify(p, &pong.NtfnStreamResponse{
				NotificationType: pong.NotificationType_ON_PLAYER_READY,
				Message:          fmt.Sprintf("Player %s is not ready", player.Nick),
				PlayerId:         player.ID.String(),
				RoomId:           gameWR.ID,
				Wr:               pwr,
				Ready:            false,
			})
		}

		// Cancel the active game stream so clients observe EOF.
		clientID := *player.ID
		if cancel, ok := s.activeGameStreams.Load(clientID); ok {
			if cancelFn, isCancel := cancel.(context.CancelFunc); isCancel {
				cancelFn()
			}
			s.activeGameStreams.Delete(clientID)
		}
		player.GameStream = nil
	}

	if winner != nil {
		// Determine branch index anchored to room host: branch 0 pays host (a), 1 pays non-host (b).
		wrID := gameWR.ID
		host := gameWR.HostID
		if host == nil {
			s.log.Errorf("handleGameEnd: waiting room host nil for wr %s", wrID)
			return
		}
		winnerBranch := int32(0)
		if host.String() != winnerID {
			winnerBranch = 1
		}

		// Look up the winner's escrow session bound to this room and gather presigs for the branch.
		s.roomEscrowsMu.RLock()
		var es *escrowSession
		if m := s.roomEscrows[*winner]; m != nil {
			if eid := m[wrID]; eid != "" {
				s.escrowsMu.RLock()
				es = s.escrows[eid]
				s.escrowsMu.RUnlock()
			}
		}
		s.roomEscrowsMu.RUnlock()
		if es == nil {
			s.log.Errorf("finalize: no room-bound escrow session found for winner %s in wr %s", winnerID, wrID)
			return
		}
		if len(es.preSign) == 0 {
			s.log.Warnf("finalize: no presign contexts stored for winner %s", winnerID)
			return
		}

		// Choose any context for the winner branch to anchor the draft hex.
		var chosen *PreSignCtx
		var branches []int32
		for _, ctx := range es.preSign {
			branches = append(branches, ctx.Branch)
			if ctx.Branch == winnerBranch {
				chosen = ctx
			}
		}
		if chosen == nil {
			s.log.Warnf("finalize: no presign context for branch %d; have branches=%v", winnerBranch, branches)
			return
		}

		// Build inputs/presigs for the chosen draft from stored contexts.
		inputs := make([]*pong.NeedPreSigs_PerInput, 0, len(es.preSign))
		presigs := make(map[string]*pong.PreSig)
		for id, ctx := range es.preSign {
			if ctx.Branch != winnerBranch || ctx.DraftHex != chosen.DraftHex {
				continue
			}
			inputs = append(inputs, &pong.NeedPreSigs_PerInput{InputId: id, RedeemScriptHex: ctx.RedeemScriptHex})
			presigs[id] = &pong.PreSig{InputId: id, RLineCompressed: append([]byte(nil), ctx.RLineCompressed...), SLine32: append([]byte(nil), ctx.SLine32...)}
		}
		if len(inputs) == 0 || len(presigs) == 0 {
			s.log.Warnf("finalize: missing presigs/inputs for winner branch %d", winnerBranch)
			return
		}

		// Derive gamma using the configured adaptor secret (same domain separation as presign).
		serverSecret := s.adaptorSecret
		if serverSecret == "" {
			s.log.Warnf("finalize: server adaptor secret not configured; cannot finalize")
			return
		}
		branchTag := fmt.Sprintf("branch-%d", winnerBranch)
		gammaHex, _ := pongbisonrelay.DeriveAdaptorGamma("", branchTag, winnerBranch, branchTag, serverSecret)

		// Finalize winner transaction.
		hexTx, err := pongbisonrelay.FinalizeWinner(gammaHex, chosen.DraftHex, inputs, presigs)
		if err != nil {
			s.log.Warnf("finalize: failed to finalize tx: %v", err)
			if w := s.gameManager.PlayerSessions.GetPlayer(*winner); w != nil && w.NotifierStream != nil {
				_ = s.notify(w, &pong.NtfnStreamResponse{NotificationType: pong.NotificationType_MESSAGE, Message: "Settlement failed to finalize; please contact support."})
			}
			return
		}

		// Broadcast the transaction via dcrd.
		raw, err := hex.DecodeString(hexTx)
		if err != nil {
			s.log.Warnf("finalize: bad hex for tx: %v", err)
			return
		}
		var tx wire.MsgTx
		if err := tx.Deserialize(bytes.NewReader(raw)); err != nil {
			s.log.Warnf("finalize: deserialize tx failed: %v", err)
			return
		}
		ctxBroadcast, cancel := context.WithTimeout(ctx, 20*time.Second)
		defer cancel()
		h, err := s.dcrd.SendRawTransaction(ctxBroadcast, &tx, false)
		if err != nil {
			s.log.Warnf("broadcast failed: %v", err)
			// Include hex for manual broadcast/debugging.
			if w := s.gameManager.PlayerSessions.GetPlayer(*winner); w != nil && w.NotifierStream != nil {
				_ = s.notify(w, &pong.NtfnStreamResponse{NotificationType: pong.NotificationType_MESSAGE, Message: fmt.Sprintf("Settlement broadcast failed: %v. You may broadcast manually with this hex: %s", err, hexTx)})
			}
			return
		}
		txid := h.String()
		// Notify both players of settlement broadcast.
		for _, p := range players {
			_ = s.notify(p, &pong.NtfnStreamResponse{
				NotificationType: pong.NotificationType_MESSAGE,
				Message:          fmt.Sprintf("Settlement broadcasted. txid=%s", txid),
			})
		}
	}
}

func (s *Server) removeWaitingRoom(wr *ponggame.WaitingRoom, msg string) {
	if wr == nil {
		return
	}
	if msg == "" {
		msg = "Waiting room removed"
	}
	// Snapshot players BEFORE removing the room, otherwise the room removal
	// clears wr.Players and escrow cleanup won't find any sessions to stop.
	wr.RLock()
	playersSnapshot := append([]*ponggame.Player(nil), wr.Players...)
	wr.RUnlock()
	if wr.Cancel != nil {
		wr.Cancel()
	}
	// Clean up all escrow sessions for players in this room before removing it.
	s.cleanupEscrowSessionsForPlayers(playersSnapshot)
	// Now safely remove the room from the game manager.
	s.gameManager.RemoveWaitingRoom(wr.ID)
	s.notifyallusers(&pong.NtfnStreamResponse{
		NotificationType: pong.NotificationType_ON_WR_REMOVED,
		Message:          msg,
		RoomId:           wr.ID,
	})
	s.log.Debugf("Waiting room %s removed (%s)", wr.ID, msg)
}

// preSignSnapshot returns a consistent snapshot of the presign state while
// holding a read lock only once. It includes the bound input id, the list of
// input ids present in presign contexts, whether all contexts agree on the
// same branch, and the count of presign contexts.
func (es *escrowSession) preSignSnapshot() (bound string, inputs []string, consistent bool) {
	es.mu.RLock()
	defer es.mu.RUnlock()

	bound = es.boundInputID
	consistent = true
	var haveBranch bool
	var branch int32
	for _, ctx := range es.preSign {
		if !haveBranch {
			branch = ctx.Branch
			haveBranch = true
		} else if ctx.Branch != branch {
			consistent = false
		}
		inputs = append(inputs, ctx.InputID)
	}
	return
}

// ManageWaitingRoom
func (s *Server) manageWaitingRoom(ctx context.Context, wr *ponggame.WaitingRoom) error {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.log.Infof("Exited ManageWaitingRoom: %s (context cancelled)", wr.ID)
			return ctx.Err()

		case <-ticker.C:
			players := wr.ReadyPlayers()
			if len(players) < 2 {
				continue
			}

			// F2P: start immediately once both players are ready.
			if s.isF2P {
				s.log.Infof("Game starting with players: %s and %s", players[0].ID, players[1].ID)
				go s.handleGameLifecycle(ctx, wr)
				return nil
			}

			// Non-F2P: require escrow bound and funded for both players.
			escrowOK := true
			for _, p := range players {
				es := s.escrowForRoomPlayer(*p.ID, wr.ID)
				if es == nil {
					escrowOK = false
					s.notify(p, &pong.NtfnStreamResponse{NotificationType: pong.NotificationType_MESSAGE, Message: "No escrow bound to this room for you. Bind a funded escrow first."})
					continue
				}
				if err := s.ensureBoundFunding(es); err != nil {
					escrowOK = false
					s.notify(p, &pong.NtfnStreamResponse{NotificationType: pong.NotificationType_MESSAGE, Message: fmt.Sprintf("Waiting for exact funding input: %v", err)})
					continue
				}
			}
			if !escrowOK {
				continue
			}

			// Require both players to have completed presign handshakes for the same branch.
			esA := s.escrowForRoomPlayer(*players[0].ID, wr.ID)
			esB := s.escrowForRoomPlayer(*players[1].ID, wr.ID)
			isComplete := func(winES, loseES *escrowSession) bool {
				if winES == nil || loseES == nil {
					return false
				}
				winBound, winInputs, winConsistent := winES.preSignSnapshot()
				loseBound, _, _ := loseES.preSignSnapshot()
				if len(winInputs) < 2 || winBound == "" || loseBound == "" || !winConsistent {
					return false
				}
				haveWin := false
				haveLose := false
				for _, in := range winInputs {
					if in == winBound {
						haveWin = true
					}
					if in == loseBound {
						haveLose = true
					}
				}
				return haveWin && haveLose
			}

			if !isComplete(esA, esB) || !isComplete(esB, esA) {
				if !isComplete(esA, esB) {
					s.notify(players[0], &pong.NtfnStreamResponse{NotificationType: pong.NotificationType_MESSAGE, Message: "Waiting: complete presign ([P]) for both inputs."})
				}
				if !isComplete(esB, esA) {
					s.notify(players[1], &pong.NtfnStreamResponse{NotificationType: pong.NotificationType_MESSAGE, Message: "Waiting: complete presign ([P]) for both inputs."})
				}
				continue
			}

			// At this point: both players have completed escrow & presign checks.
			// and both have completed presigning. Start the game.

			s.log.Infof("Game starting with players: %s and %s", players[0].ID, players[1].ID)

			// Require an active game stream for all ready players before starting.
			streamsOK := true
			for _, p := range players {
				if p == nil || p.GameStream == nil {
					streamsOK = false
					// Inform the player via notifier stream if available.
					_ = s.notify(p, &pong.NtfnStreamResponse{
						NotificationType: pong.NotificationType_MESSAGE,
						Message:          "Waiting: your game stream is not active. Please toggle ready or reconnect.",
					})
				}
			}
			if !streamsOK {
				continue
			}
			go s.handleGameLifecycle(ctx, wr)
			return nil
		}
	}
}

func (s *Server) sendGameUpdates(ctx context.Context, player *ponggame.Player, game *ponggame.GameInstance) error {
	ch := player.FrameCh
	if ch == nil {
		return fmt.Errorf("nil FrameCh for player %v", player.ID)
	}

	lastSend := time.Now()
	for {
		select {
		case <-ctx.Done():
			s.log.Warnf("sendGameUpdates: context done for player %s with error: %v", player.ID, ctx.Err())
			return ctx.Err()
		case frame, ok := <-ch:
			if !ok {
				// Producer closed: normal shutdown
				return io.EOF
			}
			if player.GameStream == nil {
				s.log.Errorf("player %s has no game stream", player.ID)
				continue
			}
			// Keep it for debug reasons for now
			now := time.Now()
			gap := now.Sub(lastSend)
			if gap >= ponggame.ErrGap {
				s.log.Warnf("sendGameUpdates: gap_since_last_send=%s (>=500ms) player=%s", gap.Truncate(time.Millisecond), player.ID)
			} else if gap >= ponggame.WarnGap {
				s.log.Debugf("sendGameUpdates: gap_since_last_send=%s (>=100ms) player=%s", gap.Truncate(time.Millisecond), player.ID)
			}

			t0 := time.Now()
			err := player.GameStream.Send(&pong.GameUpdateBytes{Data: frame})
			if err != nil {
				return err
			}
			// Keep it for debug reasons for now
			sendDur := time.Since(t0)
			if sendDur >= ErrSendBlock {
				s.log.Warnf("sendGameUpdates: Send blocked for %s (>=250ms) player=%s", sendDur.Truncate(time.Millisecond), player.ID)
			} else if sendDur >= WarnSendBlock {
				s.log.Debugf("sendGameUpdates: Send blocked for %s (>=50ms) player=%s", sendDur.Truncate(time.Millisecond), player.ID)
			}
			lastSend = time.Now()
		}
	}
}

func (s *Server) handleGameLifecycle(ctx context.Context, wr *ponggame.WaitingRoom) {
	game, err := s.gameManager.StartGame(ctx, wr)
	if err != nil {
		s.log.Errorf("Failed to start game: %v", err)
		return
	}

	defer func() {
		// reset player status
		for _, player := range game.Players {
			player.ResetPlayer()
		}
		// remove game from gameManager after it ended
		s.gameManager.DeleteGame(game.Id)
		s.log.Debugf("Game %s cleaned up", game.Id)
		// remove waiting room when game is cleaning up
		s.removeWaitingRoom(wr, "Waiting room removed after game cleanup")
	}()

	if err := game.Run(); err != nil {
		s.log.Errorf("Failed to run game %s: %v", game.Id, err)
		return
	}

	players := game.Players
	var wg sync.WaitGroup
	for _, player := range players {
		wg.Add(1)
		go func(player *ponggame.Player) {
			defer wg.Done()
			if player.NotifierStream != nil {
				err := s.notify(player, &pong.NtfnStreamResponse{
					NotificationType: pong.NotificationType_GAME_START,
					Message:          "Game started with ID: " + game.Id,
					Started:          true,
					GameId:           game.Id,
				})
				if err != nil {
					s.log.Warnf("Failed to notify player %s: %v", player.ID, err)
				}
			}
			err := s.sendGameUpdates(ctx, player, game)
			if err != nil {
				s.log.Errorf("Failed to send game updates to player %s: %v", player.ID, err)
			}
		}(player)
	}

	wg.Wait() // Wait for both players' streams to finish

	// Clean up the game after all streams have finished
	game.Cleanup()

	s.handleGameEnd(ctx, game)
}
