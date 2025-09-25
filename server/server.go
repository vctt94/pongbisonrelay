package server

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"net/http"
	"path/filepath"
	"sync"
	"time"

	"os"

	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/decred/dcrd/chaincfg/chainhash"
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
)

type ServerConfig struct {
	ServerDir string

	MinBetAmt             float64
	IsF2P                 bool
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
	DraftHex        string // exact serialized tx used at presign
	MHex            string // 32-byte sighash for (DraftHex, RedeemScriptHex, idx, SIGHASH_ALL)
	RLineCompressed []byte // 33 bytes, even-Y (0x02)
	SLine32         []byte // 32 bytes
	TCompressed     []byte // 33 bytes (if used in adaptor domain)
	WinnerUID       string // tie to player/session (owner uid)
	Branch          int32  // 0 = A-wins, 1 = B-wins (payout branch)
}

// escrowSession represents a pre-match funding session for a single player.
type escrowSession struct {
	// ----------------- immutable identity & params -----------------
	boundInputID    string
	boundInput      *pong.EscrowUTXO
	escrowID        string
	ownerUID        string
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

type Server struct {
	pong.UnimplementedPongGameServer
	pong.UnimplementedPongWaitingRoomServer
	pong.UnimplementedPongRefereeServer

	log                slog.Logger
	isF2P              bool
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

	// Only process tips if player exists in sessions AND is not in an active game
	playerSession := s.gameManager.PlayerSessions.GetPlayer(clientID)
	if playerSession != nil {
		s.gameManager.PlayerSessions.RemovePlayer(clientID)
	}

	s.gameManager.HandleWaitingRoomDisconnection(clientID, s.log)
	s.gameManager.HandleGameDisconnection(clientID, s.log)
}

// escrowForRoomPlayer returns the escrow session bound to (wrID, ownerUID)
// via the roomEscrows mapping. It does not attempt to pick a "newest" escrow.
func (s *Server) escrowForRoomPlayer(ownerUID zkidentity.ShortID, wrID string) *escrowSession {
	s.roomEscrowsMu.RLock()
	defer s.roomEscrowsMu.RUnlock()
	m := s.roomEscrows[ownerUID]
	if m == nil {
		return nil
	}
	escrowID := m[wrID]
	if escrowID == "" {
		return nil
	}
	return s.escrows[escrowID]
}

// ensureBoundFunding either binds the canonical funding input if not yet bound
// (requiring exactly one UTXO), or verifies the previously-bound input still
// exists and that no extra deposits were made.
func (s *Server) ensureBoundFunding(es *escrowSession) error {
	es.mu.Lock()
	defer es.mu.Unlock()

	// Must have a watcher snapshot that says "funded".
	if !es.latest.OK || es.latest.UTXOCount == 0 {
		return fmt.Errorf("escrow not yet funded")
	}

	// Normalize current UTXO list.
	var current []*pong.EscrowUTXO
	for _, u := range es.lastUTXOs {
		if u != nil && u.Txid != "" {
			current = append(current, u)
		}
	}
	// If we haven't cached details yet, we can't bind.
	if len(current) == 0 {
		return fmt.Errorf("escrow UTXO details not available yet; wait for index")
	}

	if es.boundInputID == "" {
		// Not yet bound: require an exact-value funding UTXO matching betAtoms.
		var matches []*pong.EscrowUTXO
		for _, u := range current {
			if u.Value == es.betAtoms {
				matches = append(matches, u)
			}
		}
		if len(matches) == 0 {
			return fmt.Errorf("no funding UTXO with exact amount: want %d atoms", es.betAtoms)
		}
		// Fail if there are multiple deposits (of any value) — require exactly one UTXO total.
		if len(current) != 1 || es.latest.UTXOCount != 1 || len(matches) != 1 {
			return fmt.Errorf("unexpected deposits present (total=%d, exact-matches=%d); require a single exact %d atoms deposit", len(current), len(matches), es.betAtoms)
		}
		u := matches[0]
		boundID := fmt.Sprintf("%s:%d", u.Txid, u.Vout)
		es.boundInputID = boundID
		es.boundInput = u
		s.notify(es.player, &pong.NtfnStreamResponse{NotificationType: pong.NotificationType_MESSAGE, Message: fmt.Sprintf("Escrow bound to %s (exact %d atoms).", boundID, es.betAtoms)})
		return nil
	}

	// Already bound: verify the exact input is still present.
	var found *pong.EscrowUTXO
	for _, u := range current {
		if fmt.Sprintf("%s:%d", u.Txid, u.Vout) == es.boundInputID {
			found = u
			break
		}
	}
	if found == nil {
		// The canonical input disappeared (reorg/spent?) -> invalidate.
		return fmt.Errorf("bound funding UTXO %s not present", es.boundInputID)
	}
	es.boundInput = found
	// Enforce: only the bound input must exist and must match the exact amount.
	if len(current) != 1 || es.latest.UTXOCount != 1 {
		return fmt.Errorf("unexpected additional deposits (%d); only the bound %s with %d atoms is allowed", len(current), es.boundInputID, es.betAtoms)
	}
	if es.boundInput != nil && es.boundInput.Value != es.betAtoms {
		return fmt.Errorf("bound funding amount mismatch: have %d want %d", es.boundInput.Value, es.betAtoms)
	}

	return nil
}

func (s *Server) sendGameUpdates(ctx context.Context, player *ponggame.Player, game *ponggame.GameInstance) {
	for {
		select {
		case <-ctx.Done():
			s.handleDisconnect(*player.ID)
			return
		case frame, ok := <-player.FrameCh: // Use individual player channel instead of shared game channel
			if !ok {
				return // Player's frame channel closed, exit
			}
			if player.GameStream == nil {
				// XXX something going on with the stream, should try a reconnect.
				s.log.Errorf("player %s has no game stream", player.ID)
				continue
			}
			err := player.GameStream.Send(&pong.GameUpdateBytes{Data: frame})
			if err != nil {
				s.handleDisconnect(*player.ID)
				return
			}
		}
	}
}

// notify sends a simple text notification if the stream is live.
// (Optional: de-dupe by caching the last message per player to avoid spam.)
func (s *Server) notify(p *ponggame.Player, resp *pong.NtfnStreamResponse) error {
	if p == nil {
		return fmt.Errorf("player nil")
	}
	if p.ID == nil {
		return fmt.Errorf("player id nil")
	}
	if p.NotifierStream == nil {
		return fmt.Errorf("player stream nil")
	}
	if resp == nil {
		return fmt.Errorf("nil response")
	}
	if err := p.NotifierStream.Send(resp); err != nil {
		s.log.Warnf("notify: failed to deliver to %s: %v", p.ID.String(), err)
		return err
	}
	return nil
}

func (s *Server) notifyallusers(resp *pong.NtfnStreamResponse) {
	// Notify all users.
	s.usersMu.RLock()
	usersSnap := make([]*ponggame.Player, 0, len(s.users))
	for _, u := range s.users {
		usersSnap = append(usersSnap, u)
	}
	s.usersMu.RUnlock()

	for _, user := range usersSnap {
		if user.NotifierStream == nil {
			s.log.Errorf("user %s without NotifierStream", user.ID)
			continue
		}
		_ = s.notify(user, resp)
	}
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

func (s *Server) handleGameEnd(ctx context.Context, game *ponggame.GameInstance, wr *ponggame.WaitingRoom) {
	players := wr.ReadyPlayers()
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

	// Finalize winner branch and broadcast on server (skip in F2P)
	if winner != nil && !s.isF2P {
		// Determine branch index anchored to room host: branch 0 pays host (a), 1 pays non-host (b).
		winnerBranch := int32(0)
		var hostUID string
		if len(players) >= 2 {
			if players[0].WR != nil && players[0].WR.HostID != nil {
				hostUID = players[0].WR.HostID.String()
			} else if players[1].WR != nil && players[1].WR.HostID != nil {
				hostUID = players[1].WR.HostID.String()
			}
			if hostUID != "" && winnerID != hostUID {
				winnerBranch = 1
			}
		}
		s.log.Debugf("finalize: computed winnerBranch=%d host=%s winner=%s", winnerBranch, hostUID, winnerID)

		// Look up the winner's escrow session bound to this room and gather presigs for the branch.
		var wrID string
		for _, p := range players {
			if p != nil && p.WR != nil {
				wrID = p.WR.ID
				break
			}
		}
		s.roomEscrowsMu.RLock()
		var es *escrowSession
		if wrID != "" && s.roomEscrows != nil {
			if m := s.roomEscrows[*winner]; m != nil {
				if eid := m[wrID]; eid != "" {
					s.escrowsMu.RLock()
					es = s.escrows[eid]
					s.escrowsMu.RUnlock()
				}
			}
		}
		s.roomEscrowsMu.RUnlock()
		if es == nil {
			s.log.Errorf("finalize: no room-bound escrow session found for winner %s in wr %s", winnerID, wrID)
			return
		}
		if es.preSign == nil || len(es.preSign) == 0 {
			s.log.Warnf("finalize: no presign contexts stored for winner %s", winnerID)
			return
		}

		// Choose any context for the winner branch to anchor the draft hex.
		var chosen *PreSignCtx
		var branches []int32
		for _, ctxp := range es.preSign {
			branches = append(branches, ctxp.Branch)
			if ctxp.Branch == winnerBranch {
				chosen = ctxp
			}
		}
		if chosen == nil {
			s.log.Warnf("finalize: no presign context for branch %d; have branches=%v", winnerBranch, branches)
			return
		}

		// Build inputs/presigs for the chosen draft from stored contexts.
		inputs := make([]*pong.NeedPreSigs_PerInput, 0, len(es.preSign))
		presigs := make(map[string]*pong.PreSig)
		for id, ctxp := range es.preSign {
			if ctxp.Branch != winnerBranch || ctxp.DraftHex != chosen.DraftHex {
				continue
			}
			inputs = append(inputs, &pong.NeedPreSigs_PerInput{InputId: id, RedeemScriptHex: ctxp.RedeemScriptHex})
			presigs[id] = &pong.PreSig{InputId: id, RLineCompressed: append([]byte(nil), ctxp.RLineCompressed...), SLine32: append([]byte(nil), ctxp.SLine32...)}
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
		// After game finishes, and tx was successfully broadcasted,
		// remove the associated waiting room  and escrows.
		if wr != nil {
			// Cancel room context to stop any related goroutines.
			if wr.Cancel != nil {
				wr.Cancel()
			}

			// Remove from game manager's waiting rooms list.
			s.gameManager.RemoveWaitingRoom(wr.ID)
			// Clean up any escrow bookkeeping for this room to avoid leaks.
			s.roomEscrowsMu.Lock()
			if s.roomEscrows != nil {
				for _, p := range wr.Players {
					delete(s.roomEscrows, *p.ID)
				}
			}
			s.roomEscrowsMu.Unlock()

			s.notifyallusers(&pong.NtfnStreamResponse{
				NotificationType: pong.NotificationType_ON_WR_REMOVED,
				Message:          "Waiting room has been removed",
				RoomId:           wrID,
			})

			s.log.Debugf("Waiting room %s removed after game end", wr.ID)
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

// ManageWaitingRoom
func (s *Server) manageWaitingRoom(ctx context.Context, wr *ponggame.WaitingRoom) error {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.log.Infof("Exited ManageWaitingRoom: %s (context cancelled)", wr.ID)
			return nil

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

			go s.handleGameLifecycle(ctx, wr)
			return nil
		}
	}
}

func (s *Server) handleGameLifecycle(ctx context.Context, wr *ponggame.WaitingRoom) {
	players := wr.ReadyPlayers()[:2]
	game, err := s.gameManager.StartGame(ctx, players)
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
	}()

	game.Run()

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
			s.sendGameUpdates(ctx, player, game)
		}(player)
	}

	wg.Wait() // Wait for both players' streams to finish

	s.handleGameEnd(ctx, game, wr)
}
