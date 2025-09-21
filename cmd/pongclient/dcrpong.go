package main

import (
	"context"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	pongbisonrelay "github.com/vctt94/pongbisonrelay"
	"github.com/vctt94/pongbisonrelay/client"
	"github.com/vctt94/pongbisonrelay/pongrpc/grpc/pong"
	"golang.org/x/sync/errgroup"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/companyzero/bisonrelay/clientrpc/types"
	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/decred/slog"
	"github.com/vctt94/bisonbotkit/botclient"
	"github.com/vctt94/bisonbotkit/logging"
	"github.com/vctt94/bisonbotkit/utils"
)

type ID = zkidentity.ShortID

// UI mode definitions moved to ui.go

var (
	// serverAddr = flag.String("server_addr", "104.131.180.29:50051", "The server address in the format of host:port")
	serverAddr         = flag.String("server_addr", "", "The server address in the format of host:port")
	datadir            = flag.String("datadir", "", "Directory to load config file from")
	flagURL            = flag.String("url", "", "URL of the websocket endpoint")
	flagServerCertPath = flag.String("servercert", "", "Path to rpc.cert file")
	flagClientCertPath = flag.String("clientcert", "", "Path to rpc-client.cert file")
	flagClientKeyPath  = flag.String("clientkey", "", "Path to rpc-client.key file")
	rpcUser            = flag.String("rpcuser", "", "RPC user for basic authentication")
	rpcPass            = flag.String("rpcpass", "", "RPC password for basic authentication")
	grpcServerCert     = flag.String("grpcservercert", "", "Path to grpc server.cert file")
	refHTTP            = flag.String("refhttp", "", "Referee HTTP base URL, e.g. http://localhost:8080")
	addressFlag        = flag.String("address", "", "33-byte compressed pubkey hex for winner payout")
)

type appstate struct {
	sync.Mutex
	mode          appMode
	gameState     *pong.GameUpdate
	currentGameId string
	ctx           context.Context
	err           error
	cancel        context.CancelFunc
	pc            *client.PongClient
	dataDir       string

	selectedRoomIndex int
	msgCh             chan tea.Msg
	viewport          viewport.Model
	createdWRChan     chan struct{}
	betAmtChangedChan chan struct{}

	isGameRunning bool
	log           slog.Logger
	logBackend    *logging.LogBackend
	players       []*pong.Player

	// player current bet amt
	betAmount float64

	currentWR *pong.WaitingRoom

	waitingRooms []*pong.WaitingRoom

	notification string

	logBuffer   []string
	logViewport viewport.Model

	// Track which keys are pressed for paddle movement
	upKeyPressed   bool
	downKeyPressed bool

	// Auto key release timer
	keyReleaseDelay time.Duration
	upKeyTimer      *time.Timer
	downKeyTimer    *time.Timer

	// Settlement (Referee) state
	settle struct {
		matchID    string
		aCompHex   string
		bCompHex   string
		escrowPath string
		preSigPath string
		lastJSON   string
		betAtoms   uint64
		csvBlocks  uint32

		activeEscrowID string

		// Handshake artifacts for finalization
		lastDraftHex string
		lastInputs   []*pong.NeedPreSigs_PerInput
		lastPresigs  map[string]*pong.PreSig // input_id -> presig (for lastDraftHex)

		// Per-draft contexts accumulated during handshake rounds
		draftInputs  map[string][]*pong.NeedPreSigs_PerInput // draftHex -> inputs
		draftPresigs map[string]map[string]*pong.PreSig      // draftHex -> (input_id -> presig)
	}

	// In-memory generated key (optional helper)
	genPrivHex string
	genPubHex  string

	// Funding status stream management
	fundingCancel context.CancelFunc
}

// payoutPubkeyFromConfHex delegates to client package helper
func (m *appstate) payoutPubkeyFromConfHex() ([]byte, error) {
	if addressFlag == nil || *addressFlag == "" {
		return nil, fmt.Errorf("missing -address flag")
	}
	return pongbisonrelay.PayoutPubkeyFromConfHex(*addressFlag)
}

func (m *appstate) startSettlement() {
	// Ensure session key A_c
	privHex, pubHex, err := m.pc.RequireSettlementSessionKey()
	if err != nil {
		m.notification = err.Error()
		m.msgCh <- client.UpdatedMsg{}
		return
	}
	m.genPrivHex = privHex
	m.genPubHex = pubHex
	m.settle.aCompHex = pubHex

	// Already have an active escrow
	if m.settle.activeEscrowID != "" {
		return
	}

	// Defaults
	if m.settle.betAtoms == 0 {
		m.settle.betAtoms = client.DefaultBetAtoms
	}
	if m.settle.csvBlocks == 0 {
		m.settle.csvBlocks = 64
	}

	// Inputs
	payoutBytes, perr := m.payoutPubkeyFromConfHex()
	if perr != nil {
		m.notification = "address error: " + perr.Error()
		m.msgCh <- client.UpdatedMsg{}
		return
	}

	m.notification = "Opening escrow…"
	m.msgCh <- client.UpdatedMsg{}

	// Open escrow (no waiting for funding; server will notify on updates).
	go func() {
		res, err := m.pc.OpenEscrowWithSession(m.ctx, payoutBytes, m.settle.betAtoms, m.settle.csvBlocks)
		if err != nil {
			m.notification = "OpenEscrow failed: " + err.Error()
			m.msgCh <- client.UpdatedMsg{}
			return
		}
		m.settle.activeEscrowID = res.EscrowId

		// UI message; actual funding/confirm status will come from server notifications.
		m.notification = "Escrow created. Fund the address; you’ll be notified when deposits are seen/confirmed."
		m.msgCh <- client.UpdatedMsg{}
	}()
}

func (m *appstate) preSign() {
	ctx := m.ctx
	// match_id: "<wrID>|<hostId>"
	if m.currentWR == nil || m.currentWR.Id == "" {
		m.notification = "Join or create a waiting room and bind escrow before presigning"
		m.msgCh <- client.UpdatedMsg{}
		return
	}
	// Use room-scoped match_id anchored to the host so branch mapping is stable (host = A/0).
	m.settle.matchID = fmt.Sprintf("%s|%s", m.currentWR.Id, m.currentWR.HostId)
	if err := m.pc.RefStartSettlementHandshake(ctx, m.settle.matchID); err != nil {
		m.notification = fmt.Sprintf("handshake error: %v", err)
		m.msgCh <- client.UpdatedMsg{}
		return
	}
	// Note: if you still want to cache artifacts for UI preview, plumb a callback through client
	m.notification = "Verified & exchanged pre-sigs for both branches (server OK)."
	m.msgCh <- client.UpdatedMsg{}
}

func realMain() error {
	flag.Parse()

	if *datadir == "" {
		*datadir = utils.AppDataDir("pongclient", false)
	}

	// Load consolidated app config and apply overrides from flags
	appCfg, err := client.LoadAppConfig(*datadir, client.ConfigOverrides{
		RPCURL:          *flagURL,
		BRClientCert:    *flagServerCertPath,
		BRClientRPCCert: *flagClientCertPath,
		BRClientRPCKey:  *flagClientKeyPath,
		RPCUser:         *rpcUser,
		RPCPass:         *rpcPass,
		ServerAddr:      *serverAddr,
		GRPCServerCert:  *grpcServerCert,
	})
	if err != nil {
		fmt.Println("Error loading configuration:", err)
		os.Exit(1)
	}

	// Determine payout address from flag or config file
	addrValue := ""
	if addressFlag != nil && *addressFlag != "" {
		addrValue = *addressFlag
	} else {
		addrValue = appCfg.Address
	}
	if strings.TrimSpace(addrValue) == "" {
		return fmt.Errorf("missing payout address: pass -address flag or set address= in %s", filepath.Join(appCfg.DataDir, "pongclient.conf"))
	}
	if _, err := pongbisonrelay.PayoutPubkeyFromConfHex(addrValue); err != nil {
		return fmt.Errorf("invalid payout address: %v", err)
	}
	*addressFlag = addrValue

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	g, gctx := errgroup.WithContext(ctx)

	// Logging
	useStdout := false
	lb, err := logging.NewLogBackend(logging.LogConfig{
		LogFile:        filepath.Join(appCfg.DataDir, "logs", "pongclient.log"),
		DebugLevel:     appCfg.BR.Debug,
		MaxLogFiles:    10,
		MaxBufferLines: 1000,
		UseStdout:      &useStdout,
	})
	if err != nil {
		return err
	}

	log := lb.Logger("BotClient")

	c, err := botclient.NewClient(appCfg.BR)
	if err != nil {
		return err
	}
	g.Go(func() error { return c.RPCClient.Run(gctx) })

	// Identify ourselves
	var zkShortID zkidentity.ShortID
	req := &types.PublicIdentityReq{}
	var publicIdentity types.PublicIdentity
	if err := c.Chat.UserPublicIdentity(ctx, req, &publicIdentity); err != nil {
		return fmt.Errorf("failed to get user public identity: %v", err)
	}
	clientID := hex.EncodeToString(publicIdentity.Identity[:])
	if idBytes, decErr := hex.DecodeString(clientID); decErr == nil && len(idBytes) >= len(zkShortID) {
		copy(zkShortID[:], idBytes[:len(zkShortID)])
	}

	as := &appstate{
		ctx:        ctx,
		cancel:     cancel,
		log:        log,
		logBackend: lb,
		mode:       gameIdle,
	}
	as.dataDir = appCfg.DataDir

	// Notifications
	ntfns := client.NewNotificationManager()
	ntfns.RegisterSync(client.OnWRCreatedNtfn(func(wr *pong.WaitingRoom, ts time.Time) {
		as.Lock()
		as.waitingRooms = append(as.waitingRooms, wr)
		for _, p := range as.players {
			if p.Uid == clientID {
				as.currentWR = wr
				as.betAmount = float64(wr.BetAmt) / 1e8
				as.mode = gameMode
			}
		}
		as.Unlock()
		as.notification = fmt.Sprintf("New waiting room created: %s", wr.Id)

		go func() {
			as.msgCh <- client.UpdatedMsg{}
			select {
			case as.createdWRChan <- struct{}{}:
			case <-as.ctx.Done():
			}
		}()
	}))
	ntfns.Register(client.OnBetAmtChangedNtfn(func(playerID string, betAmt int64, ts time.Time) {
		if clientID == playerID {
			as.notification = "bet amount updated"
			as.betAmount = float64(betAmt) / 1e8
			as.msgCh <- client.UpdatedMsg{}
		}
		for i, p := range as.players {
			if p.Uid == playerID {
				as.Lock()
				as.players[i].BetAmt = betAmt
				as.Unlock()
				break
			}
		}
		go func() {
			select {
			case as.betAmtChangedChan <- struct{}{}:
			case <-as.ctx.Done():
			}
		}()
	}))
	ntfns.Register(client.OnGameStartedNtfn(func(id string, ts time.Time) {
		as.mode = gameMode
		as.isGameRunning = true
		as.notification = fmt.Sprintf("game started with ID %s", id)
		go func() { as.msgCh <- client.UpdatedMsg{} }()
	}))
	ntfns.Register(client.OnPlayerJoinedNtfn(func(wr *pong.WaitingRoom, ts time.Time) {
		as.currentWR = wr
		as.notification = "new player joined your waiting room"
		go func() { as.msgCh <- client.UpdatedMsg{} }()
	}))
	ntfns.Register(client.OnGameEndedNtfn(func(gameID, msg string, ts time.Time) {
		as.notification = fmt.Sprintf("game %s ended\n%s", gameID, msg)
		as.betAmount = 0
		as.isGameRunning = false
		as.mode = gameIdle
		go func() { as.msgCh <- client.UpdatedMsg{} }()
	}))
	ntfns.Register(client.OnPlayerLeftNtfn(func(wr *pong.WaitingRoom, playerID string, ts time.Time) {
		if playerID == clientID {
			as.currentWR = nil
			as.notification = "You left the waiting room"
		} else {
			as.currentWR = wr
			as.notification = fmt.Sprintf("Player %s left the waiting room", playerID)
		}
		go func() { as.msgCh <- client.UpdatedMsg{} }()
	}))

	// Create Pong client (use appCfg values pulled from config/flags)
	pc, err := client.NewPongClient(clientID, &client.PongClientCfg{
		AppCfg:        appCfg,
		Notifications: ntfns,
		Log:           log,
	})
	if err != nil {
		return fmt.Errorf("failed to create pong client: %v", err)
	}
	as.pc = pc

	log.Infof("Connected to server at %s with ID %s", appCfg.ServerAddr, clientID)

	// Quick gRPC sanity check
	if _, err := pc.RefGetWaitingRooms(); err != nil {
		return fmt.Errorf("gRPC server connection failed: %v", err)
	}

	// Start notifier
	g.Go(func() error { return pc.RefStartNtfnStream(ctx) })

	defer as.cancel()

	p := tea.NewProgram(as)
	_, err = p.Run()
	if err != nil {
		return err
	}

	return g.Wait()
}

func main() {
	err := realMain()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
