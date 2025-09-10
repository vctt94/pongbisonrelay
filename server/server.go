package server

import (
	"context"
	"encoding/hex"
	"fmt"
	"net/http"
	"path/filepath"
	"sync"
	"time"

	"os"

	"github.com/companyzero/bisonrelay/clientrpc/types"
	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/decred/dcrd/rpcclient/v8"
	"github.com/decred/slog"
	"github.com/vctt94/bisonbotkit/logging"
	"github.com/vctt94/pong-bisonrelay/ponggame"
	"github.com/vctt94/pong-bisonrelay/pongrpc/grpc/pong"
	"github.com/vctt94/pong-bisonrelay/server/serverdb"
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
	SPrime32        []byte // 32 bytes
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
	depositAddress  string
	createdAt       time.Time

	// ----------------- runtime state (protected by mu) -------------
	mu        sync.RWMutex
	latest    DepositUpdate      // watcher-pushed snapshot (Confs, UTXOCount, OK, At)
	lastUTXOs []*pong.EscrowUTXO // optional cache for settlement (first UTXO, etc.)
	unsubW    func()             // watcher unsubscribe hook
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
	sync.RWMutex

	log                slog.Logger
	isF2P              bool
	minBetAmt          float64
	waitingRoomCreated chan struct{}

	users       map[zkidentity.ShortID]*ponggame.Player
	gameManager *ponggame.GameManager

	httpServer        *http.Server
	activeNtfnStreams sync.Map
	activeGameStreams sync.Map
	db                serverdb.ServerDB

	appdata string

	// dcrd RPC client
	dcrd *rpcclient.Client

	// chain watcher for tip + mempool
	watcher *chainWatcher

	// Escrow-first funding state
	escrows     map[string]*escrowSession
	roomEscrows map[string]map[string]string // roomID -> owner_uid -> escrow_id
	// v0-min defaults
	pocFeeAtoms uint64
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
	c, err := rpcclient.New(connCfg, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create dcrd rpc client (host=%s user=%s cert=%s): %w", cfg.DcrdHostPort, cfg.DcrdRPCUser, cfg.DcrdRPCCertPath, err)
	}
	s.dcrd = c
	s.log.Infof("Connected to dcrd at %s", cfg.DcrdHostPort)

	// Start chain watcher to keep tip and mempool for watched scripts
	s.watcher = newChainWatcher(s.log, s.dcrd)
	go s.watcher.run(context.Background())

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

	s.Lock()
	delete(s.users, clientID)
	s.Unlock()

	// Only process tips if player exists in sessions AND is not in an active game
	playerSession := s.gameManager.PlayerSessions.GetPlayer(clientID)
	if playerSession != nil {
		s.gameManager.PlayerSessions.RemovePlayer(clientID)
	}

	// These can safely be called multiple times
	s.gameManager.HandleWaitingRoomDisconnection(clientID, s.log)
	s.gameManager.HandleGameDisconnection(clientID, s.log)
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
func (s *Server) notify(p *ponggame.Player, msg string) {
	if p != nil && p.NotifierStream != nil {
		if p.ID != nil {
			s.log.Debugf("notify: to=%s msg=%q", p.ID.String(), msg)
		} else {
			s.log.Debugf("notify: to=<nil-id> msg=%q", msg)
		}
		if err := p.NotifierStream.Send(&pong.NtfnStreamResponse{
			NotificationType: pong.NotificationType_MESSAGE,
			Message:          msg,
		}); err != nil {
			var idStr string
			if p.ID != nil {
				idStr = p.ID.String()
			} else {
				idStr = "<nil-id>"
			}
			s.log.Warnf("notify: failed to deliver to %s: %v", idStr, err)
		}
	} else {
		s.log.Debugf("notify: dropped (player or stream nil) msg=%q", msg)
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

			s.gameManager.Lock()
			for _, wr := range s.gameManager.WaitingRooms {
				if wr.Ctx.Err() == nil { // Only manage rooms with active contexts
					s.log.Debugf("Managing waiting room: %s", wr.ID)
					go s.manageWaitingRoom(wr.Ctx, wr)
				}
			}
			s.gameManager.Unlock()
		}
	}
}

// Shutdown forcefully shuts down the server, closing HTTP server, database, waiting rooms, and games.
func (s *Server) Shutdown(ctx context.Context) error {
	// Stop chain watcher first so background RPCs stop
	if s.watcher != nil {
		s.watcher.stop()
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
	s.gameManager.Lock()
	for id, game := range s.gameManager.Games {
		s.log.Debugf("Forcefully terminating game: %s", id)
		// Close the frame channel to signal goroutines to exit
		game.Cleanup()
	}
	s.gameManager.Unlock()

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

	s.gameManager.Lock()
	for _, wr := range s.gameManager.WaitingRooms {
		wr.Cancel() // Cancel each waiting room context
	}
	s.gameManager.WaitingRooms = nil // Clear all waiting rooms
	s.gameManager.Unlock()

	// Close database LAST after all operations are done
	s.log.Info("Closing database...")
	if err := s.db.Close(); err != nil {
		s.log.Errorf("Error closing database: %v", err)
	}

	s.log.Info("Server shut down completed.")
	return nil
}

func (s *Server) handleGameLifecycle(ctx context.Context, players []*ponggame.Player, tips []*types.ReceivedTip) {
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
		delete(s.gameManager.Games, game.Id)
		s.log.Debugf("Game %s cleaned up", game.Id)
	}()

	game.Run()

	var wg sync.WaitGroup
	for _, player := range players {
		wg.Add(1)
		go func(player *ponggame.Player) {
			defer wg.Done()
			if player.NotifierStream != nil {
				err := player.NotifierStream.Send(&pong.NtfnStreamResponse{
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

	s.handleGameEnd(ctx, game, players, tips)
}

func (s *Server) handleGameEnd(ctx context.Context, game *ponggame.GameInstance, players []*ponggame.Player, tips []*types.ReceivedTip) {
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
		player.NotifierStream.Send(&pong.NtfnStreamResponse{
			NotificationType: pong.NotificationType_GAME_END,
			Message:          message,
			GameId:           game.Id,
		})
		// delete player from gameManager PlayerGameMap
		delete(s.gameManager.PlayerGameMap, *player.ID)
	}

	// POC: deliver gamma to winner via NtfnStream for finalization
	if winner != nil {
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

		// Look up the winner's escrow session bound to this room and find the presign context
		// for the computed winner branch.
		var wrID string
		for _, p := range players {
			if p != nil && p.WR != nil {
				wrID = p.WR.ID
				break
			}
		}
		s.RLock()
		var es *escrowSession
		if wrID != "" && s.roomEscrows != nil {
			if m := s.roomEscrows[wrID]; m != nil {
				if eid := m[winnerID]; eid != "" {
					es = s.escrows[eid]
				}
			}
		}
		s.RUnlock()
		if es == nil {
			s.log.Errorf("finalize: no room-bound escrow session found for winner %s in wr %s", winnerID, wrID)
			return
		}
		if es == nil {
			s.log.Warnf("finalize: no escrow session found for winner %s", winnerID)
			return
		}
		if es.preSign == nil {
			s.log.Warnf("finalize: no presign contexts stored for winner %s", winnerID)
			return
		}
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
		// Derive gamma using the same branch-bound domain separation used during presign.
		const pocServerPrivHex = "11ee22dd33cc44bb55aa66ff77ee88dd99cc00bbaa11223344556677889900aa"
		branchTag := fmt.Sprintf("branch-%d", chosen.Branch)
		gammaHex, _ := deriveAdaptorGamma("", branchTag, chosen.Branch, branchTag, pocServerPrivHex)
		if gb, err := hex.DecodeString(gammaHex); err == nil && len(gb) == 32 {
			if w := s.gameManager.PlayerSessions.GetPlayer(*winner); w != nil && w.NotifierStream != nil {
				_ = w.NotifierStream.Send(&pong.NtfnStreamResponse{
					NotificationType: pong.NotificationType_MESSAGE,
					Message:          fmt.Sprintf("Gamma received; finalizing… (%x)\nResult draft tx hex: %s", gb, chosen.DraftHex),
				})
			}
		} else if err != nil {
			s.log.Warnf("failed to decode gamma hex: %v", err)
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

			allBound := true
			for _, p := range players {
				// Resolve the exact escrow tied to this room for this player.
				es := s.escrowForRoomPlayer(wr.ID, p.ID.String())
				if es == nil {
					allBound = false
					s.notify(p, "No escrow bound to this room for you. Bind a funded escrow first.")
					continue
				}

				// Enforce identity: must have a single, canonical bound input.
				if err := s.ensureBoundFunding(es); err != nil {
					allBound = false
					s.notify(p, fmt.Sprintf("Waiting for exact funding input: %v", err))
					continue
				}
			}

			if !allBound {
				continue
			}

			// Require both players to have completed presign handshakes for the
			// branch where they win: winner's escrow must hold presigs for BOTH
			// inputs (their own and opponent's) and all contexts must share the
			// same Branch value.
			esA := s.escrowForRoomPlayer(wr.ID, players[0].ID.String())
			esB := s.escrowForRoomPlayer(wr.ID, players[1].ID.String())
			isComplete := func(winES, loseES *escrowSession) bool {
				if winES == nil || loseES == nil {
					return false
				}
				winES.mu.RLock()
				defer winES.mu.RUnlock()
				loseES.mu.RLock()
				loseBound := loseES.boundInputID
				loseES.mu.RUnlock()
				if len(winES.preSign) < 2 {
					return false
				}
				winBound := winES.boundInputID
				if winBound == "" || loseBound == "" {
					return false
				}
				var branch *int32
				haveWin := false
				haveLose := false
				for _, ctx := range winES.preSign {
					if branch == nil {
						b := ctx.Branch
						branch = &b
					} else if ctx.Branch != *branch {
						return false // mixed branches; not acceptable
					}
					if ctx.InputID == winBound {
						haveWin = true
					} else if ctx.InputID == loseBound {
						haveLose = true
					}
				}
				return haveWin && haveLose
			}

			if !isComplete(esA, esB) || !isComplete(esB, esA) {
				if !isComplete(esA, esB) {
					s.notify(players[0], "Waiting: complete presign ([P]) for both inputs.")
				}
				if !isComplete(esB, esA) {
					s.notify(players[1], "Waiting: complete presign ([P]) for both inputs.")
				}
				continue
			}

			// At this point: both players ready, both have a canonical bound UTXO,
			// and both have completed presigning. Start the game.

			s.log.Infof("Game starting with players: %s and %s", players[0].ID, players[1].ID)

			go s.handleGameLifecycle(ctx, players, wr.ReservedTips)
			return nil
		}
	}
}
