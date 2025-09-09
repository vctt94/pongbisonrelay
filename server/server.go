package server

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"sync"
	"time"

	"os"

	"github.com/companyzero/bisonrelay/clientrpc/types"
	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/decred/dcrd/dcrutil/v4"
	"github.com/decred/dcrd/rpcclient/v8"
	"github.com/decred/slog"
	"github.com/vctt94/bisonbotkit"
	"github.com/vctt94/bisonbotkit/logging"
	"github.com/vctt94/pong-bisonrelay/ponggame"
	"github.com/vctt94/pong-bisonrelay/pongrpc/grpc/pong"
	"github.com/vctt94/pong-bisonrelay/server/serverdb"
)

const (
	csvBlocks = 64
	name      = "pong"
	version   = "v0.0.0"
)

// BotInterface defines the methods needed by the server
type BotInterface interface {
	Run(ctx context.Context) error
	AckTipProgress(ctx context.Context, sequenceId uint64) error
	AckTipReceived(ctx context.Context, sequenceId uint64) error
	PayTip(ctx context.Context, recipient zkidentity.ShortID, amount dcrutil.Amount, priority int32) error
}

type ServerConfig struct {
	ServerDir string

	Bot *bisonbotkit.Bot

	MinBetAmt             float64
	IsF2P                 bool
	DebugLevel            string
	DebugGameManagerLevel string
	PaymentClient         types.PaymentsServiceClient
	ChatClient            types.ChatServiceClient
	HTTPPort              string
	LogBackend            *logging.LogBackend

	// dcrd RPC connectivity
	DcrdHostPort    string // e.g. 127.0.0.1:19109
	DcrdRPCCertPath string // path to rpc.cert
	DcrdRPCUser     string
	DcrdRPCPass     string
}

type Server struct {
	pong.UnimplementedPongGameServer
	pong.UnimplementedPongRefereeServer
	sync.RWMutex

	bot                BotInterface
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
	pocCSV      uint32
	pocFeeAtoms uint64
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
	InputID          string // "txid:vout"
	RedeemScriptHex  string
	DraftHex         string // exact serialized tx used at presign
	MHex             string // 32-byte sighash for (DraftHex, RedeemScriptHex, idx, SIGHASH_ALL)
	RPrimeCompressed []byte // 33 bytes, even-Y (0x02)
	SPrime32         []byte // 32 bytes
	TCompressed      []byte // 33 bytes (if used in adaptor domain)
	WinnerUID        string // tie to player/session (owner uid)
	Branch           int32  // 0 = A-wins, 1 = B-wins (payout branch)
}

// SettleInput is the only thing the presign needs from funding.
// It binds identity (escrowID, ownerUID) + the exact outpoint and script.
type SettleInput struct {
	EscrowID        string           // for storage of presigs
	OwnerUID        string           // for deterministic (a,b) ordering
	InputID         string           // "txid:vout" (must be the BOUND outpoint)
	UTXO            *pong.EscrowUTXO // must include Txid, Vout, AmountAtoms
	RedeemScriptHex string           // exact redeem used to spend UTXO
	PayoutPubkey    []byte           // 33-byte compressed pubkey for payout path
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

// twoBranchDrafts holds the fully serialized drafts for A-wins and B-wins,
// and their per-input presign data (m and T) for each input using its own redeem.
// Branch mapping: 0 = pay to A, 1 = pay to B.
type twoBranchDrafts struct {
	DraftHexA string
	InputsA   []*pong.NeedPreSigs_PerInput
	DraftHexB string
	InputsB   []*pong.NeedPreSigs_PerInput
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
		bot:                cfg.Bot,
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
		escrows:     make(map[string]*escrowSession),
		roomEscrows: make(map[string]map[string]string),
		pocCSV:      csvBlocks,
		pocFeeAtoms: 0,
	}
	s.gameManager.OnWaitingRoomRemoved = s.handleWaitingRoomRemoved

	// Initialize optional dcrd RPC client (no fallbacks; require explicit values)
	if cfg.DcrdHostPort != "" || cfg.DcrdRPCUser != "" || cfg.DcrdRPCPass != "" || cfg.DcrdRPCCertPath != "" {
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
	} else {
		s.log.Infof("dcrd RPC not configured; FundingStatus will not query chain state")
	}

	if cfg.HTTPPort != "" {
		// Set up HTTP server for db calls
		mux := http.NewServeMux()
		mux.HandleFunc("/received", s.handleFetchTipsByClientIDHandler)
		mux.HandleFunc("/fetchAllUnprocessedTips", s.handleFetchAllUnprocessedTipsHandler)
		mux.HandleFunc("/tipprogress", s.handleGetSendProgressByWinnerHandler)
		s.httpServer = &http.Server{
			Addr:    fmt.Sprintf(":%s", cfg.HTTPPort),
			Handler: mux,
		}

		go func() {
			s.log.Infof("Starting HTTP server on port %s", cfg.HTTPPort)
			if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				s.log.Errorf("HTTP server error: %v", err)
			}
		}()
	}

	return s, nil
}

func (s *Server) StartGameStream(req *pong.StartGameStreamRequest, stream pong.PongGame_StartGameStreamServer) error {
	ctx, cancel := context.WithCancel(stream.Context())
	defer cancel()

	var clientID zkidentity.ShortID
	clientID.FromString(req.ClientId)
	defer s.activeGameStreams.Delete(clientID)

	// Store the cancel function
	s.activeGameStreams.Store(clientID, cancel)

	s.log.Debugf("Client %s called StartGameStream", req.ClientId)

	player := s.gameManager.PlayerSessions.GetPlayer(clientID)
	if player == nil {
		return fmt.Errorf("player not found for client ID %s", clientID)
	}
	if player.NotifierStream == nil {
		return fmt.Errorf("player notifier nil %s", clientID)
	}

	// Require being in a waiting room
	if player.WR == nil {
		return fmt.Errorf("not in a waiting room")
	}
	var escrowID string
	s.RLock()
	if m, ok := s.roomEscrows[player.WR.ID]; ok {
		escrowID = m[player.ID.String()]
	}
	var es *escrowSession
	if escrowID != "" {
		es = s.escrows[escrowID]
	}
	s.RUnlock()
	if es == nil {
		return fmt.Errorf("no escrow bound to waiting room for player")
	}
	// Do not check presign readiness here; waiting room gating ensures bound funding only.

	player.GameStream = stream
	player.Ready = true

	// Notify all players in the waiting room that this player is ready
	if player.WR != nil {
		// Marshal the waiting room state to include in notifications
		pwr, err := player.WR.Marshal()
		if err != nil {
			return err
		}
		for _, p := range player.WR.Players {
			p.NotifierStream.Send(&pong.NtfnStreamResponse{
				NotificationType: pong.NotificationType_ON_PLAYER_READY,
				Message:          fmt.Sprintf("Player %s is ready", player.Nick),
				PlayerId:         player.ID.String(),
				Wr:               pwr,
				Ready:            true,
			})
		}
	}

	// Wait for context to end and handle disconnection
	<-ctx.Done()
	s.log.Debugf("Client %s disconnected from game stream", clientID)
	return nil
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

func (s *Server) StartNtfnStream(req *pong.StartNtfnStreamRequest, stream pong.PongGame_StartNtfnStreamServer) error {
	ctx, cancel := context.WithCancel(stream.Context())
	defer cancel()
	defer s.activeNtfnStreams.Delete(req.ClientId)

	var clientID zkidentity.ShortID
	clientID.FromString(req.ClientId)
	s.log.Debugf("StartNtfnStream called by client %s", clientID)

	// Add to active streams
	s.activeNtfnStreams.Store(clientID, cancel)

	// Create player session
	player := s.gameManager.PlayerSessions.CreateSession(clientID)
	player.NotifierStream = stream

	// Bind any existing escrow sessions for this owner to this player so
	// watcher notifications reach the active notifier stream.
	s.Lock()
	for _, es := range s.escrows {
		if es != nil && es.ownerUID == clientID.String() {
			// Replace player binding
			es.player = player
			// If we already have a latest funding snapshot, nudge the player once so UI updates.
			es.mu.RLock()
			u := es.latest
			es.mu.RUnlock()
			if u.OK && u.UTXOCount > 0 {
				if u.Confs == 0 {
					s.log.Debugf("rebinding escrow: owner=%s pk=%s has mempool funding; nudging", es.ownerUID, u.PkScriptHex)
					s.notify(player, "Deposit seen in mempool. Waiting confirmations.")
				} else {
					s.log.Debugf("rebinding escrow: owner=%s pk=%s has %d confs; nudging", es.ownerUID, u.PkScriptHex, u.Confs)
					s.notify(player, "Deposit confirmed. You can presign now ([P]).")
				}
			}
		}
	}
	s.Unlock()

	// Escrow-first: remove legacy tips-based bet sync
	// Wait for disconnection
	<-ctx.Done()
	s.log.Debugf("Client %s disconnected", clientID)
	s.handleDisconnect(clientID)
	return ctx.Err()
}

func (s *Server) SendInput(ctx context.Context, req *pong.PlayerInput) (*pong.GameUpdate, error) {
	var clientID zkidentity.ShortID
	clientID.FromString(req.PlayerId)
	return s.gameManager.HandlePlayerInput(clientID, req)
}

// ManageWaitingRoom now gates ONLY on (a) both players "ready" and
// (b) each player's room-bound escrow having a **bound** funding input
// (exact txid:vout). No presign/confirm logic here.
func (s *Server) ManageWaitingRoom(ctx context.Context, wr *ponggame.WaitingRoom) error {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.log.Infof("Exited ManageWaitingRoom: %s (context cancelled)", wr.ID)
			return nil

		case <-ticker.C:
			players, ready := wr.ReadyPlayers()
			if !ready {
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

			// At this point: both players ready AND both have a canonical bound UTXO.
			// Start the game.

			s.log.Infof("Game starting with players: %s and %s", players[0].ID, players[1].ID)

			go s.handleGameLifecycle(ctx, players, wr.ReservedTips)
			return nil
		}
	}
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

func (s *Server) Run(ctx context.Context) error {
	go s.bot.Run(ctx)

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
					go s.ManageWaitingRoom(wr.Ctx, wr)
				}
			}
			s.gameManager.Unlock()
		}
	}
}

func (s *Server) GetWaitingRooms(ctx context.Context, req *pong.WaitingRoomsRequest) (*pong.WaitingRoomsResponse, error) {
	s.Lock()
	defer s.Unlock()

	pongWaitingRooms := make([]*pong.WaitingRoom, len(s.gameManager.WaitingRooms))
	for i, room := range s.gameManager.WaitingRooms {
		wr, err := room.Marshal()
		if err != nil {
			return nil, err
		}
		pongWaitingRooms[i] = wr
	}

	return &pong.WaitingRoomsResponse{
		Wr: pongWaitingRooms,
	}, nil
}

func (s *Server) JoinWaitingRoom(ctx context.Context, req *pong.JoinWaitingRoomRequest) (*pong.JoinWaitingRoomResponse, error) {
	var uid zkidentity.ShortID
	if err := uid.FromString(req.ClientId); err != nil {
		return nil, err
	}
	player := s.gameManager.PlayerSessions.GetPlayer(uid)
	if player == nil {
		return nil, fmt.Errorf("player not found: %s", req.ClientId)
	}

	// Disallow joining multiple rooms.
	s.gameManager.Lock()
	for _, existingWR := range s.gameManager.WaitingRooms {
		for _, p := range existingWR.Players {
			if p.ID.String() == req.ClientId && p.WR != nil {
				s.gameManager.Unlock()
				return nil, fmt.Errorf("player %s is already in another waiting room", req.ClientId)
			}
		}
	}
	s.gameManager.Unlock()

	// Require funded escrow (0-conf OK).
	es, err := s.resolveFundedEscrow(uid.String(), req.EscrowId)
	if err != nil {
		return nil, fmt.Errorf("require funded escrow to join room (0-conf ok): %w", err)
	}

	// Locate room.
	wr := s.gameManager.GetWaitingRoom(req.RoomId)
	if wr == nil {
		return nil, fmt.Errorf("waiting room not found: %s", req.RoomId)
	}

	// Add player and bind escrow to this room.
	wr.AddPlayer(player)
	player.WR = wr

	s.Lock()
	if s.roomEscrows == nil {
		s.roomEscrows = make(map[string]map[string]string)
	}
	if s.roomEscrows[wr.ID] == nil {
		s.roomEscrows[wr.ID] = make(map[string]string)
	}
	s.roomEscrows[wr.ID][player.ID.String()] = es.escrowID
	s.Unlock()

	// Notify room.
	pwr, err := wr.Marshal()
	if err != nil {
		return nil, err
	}
	for _, p := range wr.Players {
		if p.NotifierStream == nil {
			s.log.Errorf("player %s has nil NotifierStream", p.ID.String())
			continue
		}
		_ = p.NotifierStream.Send(&pong.NtfnStreamResponse{
			NotificationType: pong.NotificationType_PLAYER_JOINED_WR,
			Message:          fmt.Sprintf("New player joined Waiting Room: %s", player.Nick),
			PlayerId:         p.ID.String(),
			RoomId:           wr.ID,
			Wr:               pwr,
		})
	}

	return &pong.JoinWaitingRoomResponse{Wr: pwr}, nil
}

func (s *Server) CreateWaitingRoom(ctx context.Context, req *pong.CreateWaitingRoomRequest) (*pong.CreateWaitingRoomResponse, error) {
	var hostID zkidentity.ShortID
	if err := hostID.FromString(req.HostId); err != nil {
		return nil, err
	}
	hostPlayer := s.gameManager.PlayerSessions.GetPlayer(hostID)
	if hostPlayer == nil {
		return nil, fmt.Errorf("player not found: %s", req.HostId)
	}
	// Allow F2P; otherwise enforce min bet when req.BetAmt > 0.
	if !s.isF2P && req.BetAmt > 0 && float64(req.BetAmt)/1e11 < s.minBetAmt {
		return nil, fmt.Errorf("bet needs to be higher than %.8f", s.minBetAmt)
	}
	if hostPlayer.WR != nil {
		return nil, fmt.Errorf("player %s is already in a waiting room", hostID.String())
	}

	s.log.Debugf("creating waiting room. Host ID: %s", hostID)

	// Require funded escrow (0-conf OK).
	es, err := s.resolveFundedEscrow(hostID.String(), req.EscrowId)
	if err != nil {
		return nil, fmt.Errorf("require funded escrow to create room (0-conf ok): %w", err)
	}

	// Create room; bet uses host escrow's betAtoms.
	wr, err := ponggame.NewWaitingRoom(hostPlayer, int64(es.betAtoms))
	if err != nil {
		return nil, fmt.Errorf("failed to create waiting room: %v", err)
	}
	hostPlayer.WR = wr

	// Bind host escrow to room.
	s.Lock()
	if s.roomEscrows == nil {
		s.roomEscrows = make(map[string]map[string]string)
	}
	if s.roomEscrows[wr.ID] == nil {
		s.roomEscrows[wr.ID] = make(map[string]string)
	}
	s.roomEscrows[wr.ID][hostID.String()] = es.escrowID
	s.Unlock()

	// Add to list of rooms.
	s.gameManager.Lock()
	s.gameManager.WaitingRooms = append(s.gameManager.WaitingRooms, wr)
	totalRooms := len(s.gameManager.WaitingRooms)
	s.gameManager.Unlock()

	s.log.Debugf("waiting room created. Total rooms: %d", totalRooms)

	// Signal creation (non-blocking).
	select {
	case s.waitingRoomCreated <- struct{}{}:
	default:
	}

	// Notify all users.
	pongWR, err := wr.Marshal()
	if err != nil {
		return nil, err
	}
	s.RLock()
	for _, user := range s.users {
		if user.NotifierStream == nil {
			s.log.Errorf("user %s without NotifierStream", user.ID)
			continue
		}
		_ = user.NotifierStream.Send(&pong.NtfnStreamResponse{
			Wr:               pongWR,
			NotificationType: pong.NotificationType_ON_WR_CREATED,
		})
	}
	s.RUnlock()

	return &pong.CreateWaitingRoomResponse{Wr: pongWR}, nil
}

// LeaveWaitingRoom handles a request from a client to leave a waiting room
func (s *Server) LeaveWaitingRoom(ctx context.Context, req *pong.LeaveWaitingRoomRequest) (*pong.LeaveWaitingRoomResponse, error) {
	s.log.Debugf("LeaveWaitingRoom request from client %s for room %s", req.ClientId, req.RoomId)

	var clientID zkidentity.ShortID
	if err := clientID.FromString(req.ClientId); err != nil {
		return &pong.LeaveWaitingRoomResponse{
			Success: false,
			Message: fmt.Sprintf("invalid client ID: %v", err),
		}, nil
	}

	// Get the waiting room
	wr := s.gameManager.GetWaitingRoom(req.RoomId)
	if wr == nil {
		return &pong.LeaveWaitingRoomResponse{
			Success: false,
			Message: "waiting room not found",
		}, nil
	}

	// Check if player is in the room
	player := wr.GetPlayer(&clientID)
	if player == nil {
		return &pong.LeaveWaitingRoomResponse{
			Success: false,
			Message: "player not in waiting room",
		}, nil
	}

	// Remove the player from the waiting room
	wr.RemovePlayer(clientID)

	// If player was the host and there are other players, assign a new host
	if wr.HostID.String() == clientID.String() && len(wr.Players) > 0 {
		wr.Lock()
		wr.HostID = wr.Players[0].ID
		wr.Unlock()
	}

	// If the room is now empty, remove it
	if len(wr.Players) == 0 {
		s.gameManager.RemoveWaitingRoom(wr.ID)
	} else {
		// Notify remaining players that someone left
		pwrMarshaled, err := wr.Marshal()
		if err == nil {
			for _, p := range wr.Players {
				// Send notification to remaining players
				p.NotifierStream.Send(&pong.NtfnStreamResponse{
					NotificationType: pong.NotificationType_PLAYER_LEFT_WR,
					RoomId:           wr.ID,
					Wr:               pwrMarshaled,
					PlayerId:         req.ClientId,
				})
			}
		}
	}

	// Reset the player's waiting room reference
	player.WR = nil

	return &pong.LeaveWaitingRoomResponse{
		Success: true,
		Message: "successfully left waiting room",
	}, nil
}

// UnreadyGameStream handles a request from a client who wants to signal they are no longer ready
func (s *Server) UnreadyGameStream(ctx context.Context, req *pong.UnreadyGameStreamRequest) (*pong.UnreadyGameStreamResponse, error) {
	var clientID zkidentity.ShortID
	clientID.FromString(req.ClientId)

	s.log.Debugf("Client %s called UnreadyGameStream", req.ClientId)

	// Find the player
	player := s.gameManager.PlayerSessions.GetPlayer(clientID)
	if player == nil {
		return nil, fmt.Errorf("player not found: %s", req.ClientId)
	}

	// Check if the player is in a waiting room
	if player.WR != nil {
		player.Ready = false

		// First get the cancel function and call it before deleting
		if cancel, ok := s.activeGameStreams.Load(clientID); ok {
			if cancelFn, isCancel := cancel.(context.CancelFunc); isCancel {
				cancelFn()
			}
		}

		// Then delete the entry
		s.activeGameStreams.Delete(clientID)
		player.GameStream = nil

		// Notify other players in the waiting room
		pwr, err := player.WR.Marshal()
		if err == nil {
			for _, p := range player.WR.Players {
				p.NotifierStream.Send(&pong.NtfnStreamResponse{
					NotificationType: pong.NotificationType_ON_PLAYER_READY,
					Message:          fmt.Sprintf("Player %s is not ready", player.Nick),
					PlayerId:         player.ID.String(),
					RoomId:           player.WR.ID,
					Wr:               pwr,
					Ready:            false,
				})
			}
		}
	}

	return &pong.UnreadyGameStreamResponse{}, nil
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

	s.Lock()
	s.users = nil
	s.Unlock()

	// Close database LAST after all operations are done
	s.log.Info("Closing database...")
	if err := s.db.Close(); err != nil {
		s.log.Errorf("Error closing database: %v", err)
	}

	s.log.Info("Server shut down completed.")
	return nil
}
