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
	escrowByKey map[string]string            // owner|comp|bet|csv -> escrowID
	roomEscrows map[string]map[string]string // roomID -> owner_uid -> escrow_id
	// v0-min defaults
	pocCSV      uint32
	pocFeeAtoms uint64
}

// escrowSession represents a pre-match funding session for a single player.
type escrowSession struct {
	escrowID        string
	ownerUID        string
	compPubkey      []byte // 33 bytes
	betAtoms        uint64
	csvBlocks       uint32
	redeemScriptHex string
	pkScriptHex     string
	depositAddress  string
	confs           uint32
	createdAt       time.Time
	updatedAt       time.Time
}

// pickConfirmedEscrow returns a CONFIRMED escrow session for the given owner.
// If escrowID is provided, validates and returns it. Otherwise, returns the
// most recently updated CONFIRMED escrow for that owner. Returns error if none.
func (s *Server) pickConfirmedEscrow(ownerUID string, escrowID string) (*escrowSession, error) {
	s.RLock()
	defer s.RUnlock()
	ensureConfirmed := func(es *escrowSession, tag string) *escrowSession {
		if es == nil {
			return nil
		}

		// Require watcher to see a UTXO (0-conf unlock). If no watcher or no pkScript, treat as not yet funded.
		if s.watcher == nil || es.pkScriptHex == "" {
			return nil
		}
		// Query watcher and accept when chain/mempool shows a UTXO for this script (0-conf OK)
		s.watcher.registerDeposit(es.pkScriptHex, es.redeemScriptHex, tag)
		utxos, confs, ok := s.watcher.queryDeposit(es.pkScriptHex)
		if ok && len(utxos) > 0 {
			// Promote and update timestamps/confs
			s.RUnlock()
			s.Lock()
			es.confs = confs
			es.updatedAt = time.Now()
			s.Unlock()
			s.RLock()
			return es
		}
		return nil
	}
	if escrowID != "" {
		es := ensureConfirmed(s.escrows[escrowID], escrowID)
		if es == nil {
			return nil, fmt.Errorf("escrow not yet funded (0-conf required)")
		}
		if es.ownerUID != ownerUID {
			return nil, fmt.Errorf("escrow not owned by user")
		}

		return es, nil
	}
	var best *escrowSession
	for id, es := range s.escrows {
		es = ensureConfirmed(es, id)
		if es == nil || es.ownerUID != ownerUID {
			continue
		}
		if best == nil || es.updatedAt.After(best.updatedAt) {
			best = es
		}
	}
	if best == nil {
		return nil, fmt.Errorf("no funded escrow for user (0-conf)")
	}
	return best, nil
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
		escrowByKey: make(map[string]string),
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
	if player.GameStream != nil {
		return fmt.Errorf("game stream is already set for id %s", clientID)
	}

	// Escrow-first: readiness decoupled from legacy match-bound funding checks

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

	s.Lock()
	s.users[clientID] = player
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

func (s *Server) ManageWaitingRoom(ctx context.Context, wr *ponggame.WaitingRoom) error {
	for {
		select {
		case <-ctx.Done():
			s.log.Infof("Exited ManageWaitingRoom: %s (context cancelled)", wr.ID)
			return nil

		case <-time.After(time.Second):
			players, ready := wr.ReadyPlayers()
			if ready {
				// In paid mode, ensure presigs are complete before starting
				if !s.isF2P {
					// Derive matchID from any player mapping (A/B mapping uses player->match)
					var matchID string

					if matchID == "" {
						// No mapping; requeue and continue waiting
						continue
					}

				}

				s.log.Infof("Game starting with players: %v and %v", players[0].ID, players[1].ID)

				s.gameManager.RemoveWaitingRoom(wr.ID)
				go s.handleGameLifecycle(ctx, players, wr.ReservedTips) // Start game lifecycle in a goroutine
				return nil
			}
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
	err := uid.FromString(req.ClientId)
	if err != nil {
		return nil, err
	}
	player := s.gameManager.PlayerSessions.GetPlayer(uid)
	if player == nil {
		return nil, fmt.Errorf("player not found: %s", req.ClientId)
	}

	// Check if player is already in another waiting room
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

	// Escrow-first gating: require a funded escrow (0-conf) for the joining player
	if _, err := s.pickConfirmedEscrow(uid.String(), req.EscrowId); err != nil {
		return nil, fmt.Errorf("require funded escrow to join room: %w", err)
	}

	wr := s.gameManager.GetWaitingRoom(req.RoomId)
	if wr == nil {
		return nil, fmt.Errorf("waiting room not found: %s", req.RoomId)
	}

	wr.AddPlayer(player)
	// Optional: record player's escrow selection for this room
	if req.EscrowId != "" {
		s.Lock()
		if s.roomEscrows == nil {
			s.roomEscrows = make(map[string]map[string]string)
		}
		if s.roomEscrows[wr.ID] == nil {
			s.roomEscrows[wr.ID] = make(map[string]string)
		}
		s.roomEscrows[wr.ID][player.ID.String()] = req.EscrowId
		s.Unlock()
	}
	player.WR = wr

	pwr, err := wr.Marshal()
	if err != nil {
		return nil, err
	}
	for _, p := range wr.Players {
		if p.NotifierStream == nil {
			s.log.Errorf("player %s has nil NotifierStream", p.ID.String())
			continue
		}
		p.NotifierStream.Send(&pong.NtfnStreamResponse{
			NotificationType: pong.NotificationType_PLAYER_JOINED_WR,
			Message:          fmt.Sprintf("New player joined Waiting Room: %s", player.Nick),
			PlayerId:         p.ID.String(),
			RoomId:           wr.ID,
			Wr:               pwr,
		})
	}

	return &pong.JoinWaitingRoomResponse{
		Wr: pwr,
	}, nil
}

func (s *Server) CreateWaitingRoom(ctx context.Context, req *pong.CreateWaitingRoomRequest) (*pong.CreateWaitingRoomResponse, error) {
	var hostID zkidentity.ShortID
	err := hostID.FromString(req.HostId)
	if err != nil {
		return nil, err
	}

	hostPlayer := s.gameManager.PlayerSessions.GetPlayer(hostID)
	if hostPlayer == nil {
		return nil, fmt.Errorf("player not found: %s", req.HostId)
	}
	// Decouple room creation from tip-based BetAmt; escrow lobbies will set funding separately.
	// Allow zero bet for escrow-based lobbies; enforce min bet only when positive
	if !s.isF2P && req.BetAmt > 0 && float64(req.BetAmt)/1e11 < s.minBetAmt {
		return nil, fmt.Errorf("bet needs to be higher than %.8f", s.minBetAmt)
	}
	if hostPlayer.WR != nil {
		return nil, fmt.Errorf("player %s is already in a waiting room", hostID.String())
	}

	s.log.Debugf("creating waiting room. Host ID: %s", hostID)

	// Escrow-first gating: require a funded escrow (0-conf) for the host.
	es, err := s.pickConfirmedEscrow(hostID.String(), req.EscrowId)
	if err != nil {
		return nil, fmt.Errorf("require funded escrow to create room: %w", err)
	}

	// Create waiting room; betAmt comes from host's escrow value expectation
	wr, err := ponggame.NewWaitingRoom(hostPlayer, int64(es.betAtoms))
	if err != nil {
		return nil, fmt.Errorf("failed to create waiting room: %v", err)
	}

	hostPlayer.WR = wr

	// Optional: store escrow selection by room and owner when provided
	if req.EscrowId != "" {
		s.Lock()
		if s.roomEscrows == nil {
			s.roomEscrows = make(map[string]map[string]string)
		}
		if s.roomEscrows[wr.ID] == nil {
			s.roomEscrows[wr.ID] = make(map[string]string)
		}
		s.roomEscrows[wr.ID][hostID.String()] = req.EscrowId
		s.Unlock()
	}

	// Optional: store escrow selection by room and owner when provided
	if req.EscrowId != "" {
		s.Lock()
		if s.roomEscrows == nil {
			s.roomEscrows = make(map[string]map[string]string)
		}
		if s.roomEscrows[wr.ID] == nil {
			s.roomEscrows[wr.ID] = make(map[string]string)
		}
		s.roomEscrows[wr.ID][hostID.String()] = req.EscrowId
		s.Unlock()
	}

	// append to WaitingRooms slice
	s.gameManager.Lock()
	s.gameManager.WaitingRooms = append(s.gameManager.WaitingRooms, wr)
	totalRooms := len(s.gameManager.WaitingRooms)
	s.gameManager.Unlock()

	s.log.Debugf("waiting room created. Total rooms: %d", totalRooms)

	// Signal that a new Waiting Room has been created
	select {
	case s.waitingRoomCreated <- struct{}{}:
	default:
		// Non-blocking send to avoid deadlock in case of rapid room creations
	}

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
		user.NotifierStream.Send(&pong.NtfnStreamResponse{
			Wr:               pongWR,
			NotificationType: pong.NotificationType_ON_WR_CREATED,
		})
	}
	s.RUnlock()

	return &pong.CreateWaitingRoomResponse{
		Wr: pongWR,
	}, nil
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
					PlayerId:         p.ID.String(),
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
