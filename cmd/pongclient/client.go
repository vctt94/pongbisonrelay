package main

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/vctt94/pong-bisonrelay/client"
	"github.com/vctt94/pong-bisonrelay/pongrpc/grpc/pong"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/proto"

	"github.com/decred/dcrd/chaincfg/v3"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/decred/dcrd/dcrutil/v4"
	"github.com/decred/dcrd/txscript/v4"
	"github.com/decred/dcrd/txscript/v4/stdaddr"
	"github.com/decred/dcrd/wire"
	basecfg "github.com/vctt94/bisonbotkit/config"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/companyzero/bisonrelay/clientrpc/types"
	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/decred/dcrd/crypto/blake256"
	"github.com/decred/slog"
	"github.com/vctt94/bisonbotkit/botclient"
	"github.com/vctt94/bisonbotkit/logging"
	"github.com/vctt94/bisonbotkit/utils"
)

type ID = zkidentity.ShortID

type appMode int

var isF2p = false

const (
	gameIdle appMode = iota
	gameMode
	listRooms
	createRoom
	joinRoom
	viewLogs
	settlementMode
)

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

func (m *appstate) listenForUpdates() tea.Cmd {
	return func() tea.Msg {
		// Start a goroutine to listen for updates
		go func() {
			for msg := range m.pc.UpdatesCh {
				m.msgCh <- msg
			}
		}()
		return nil
	}
}

func (m *appstate) listenForErrors() tea.Cmd {
	return func() tea.Msg {
		go func() {
			for err := range m.pc.ErrorsCh {
				m.msgCh <- fmt.Sprintf("Error: %v", err)
			}
		}()
		return nil
	}
}

func (m *appstate) Init() tea.Cmd {
	m.msgCh = make(chan tea.Msg)
	m.viewport = viewport.New(0, 0)
	m.logViewport = viewport.New(0, 0)
	m.logBuffer = make([]string, 0)

	// Set default key release delay to 150ms
	m.keyReleaseDelay = 50 * time.Millisecond

	return tea.Batch(
		m.listenForUpdates(),
		m.listenForErrors(),
		tea.EnterAltScreen,
	)
}

// handleSubmitPreSig triggers the streaming settlement flow (Phase 1 UX).
func (m *appstate) handlePreSig() tea.Cmd {
	return func() tea.Msg {
		m.notification = "Starting presign"
		m.msgCh <- client.UpdatedMsg{}
		go m.preSign()
		return client.UpdatedMsg{}
	}
}

// payoutPubkeyFromConfHex reads payout pubkey from -address: accepts 33B hex or a DCR P2PK address
func (m *appstate) payoutPubkeyFromConfHex() ([]byte, error) {
	if addressFlag != nil && *addressFlag != "" {
		s := strings.TrimSpace(*addressFlag)
		// Try as hex compressed pubkey first
		if b, err := hex.DecodeString(s); err == nil {
			if len(b) == 33 {
				return b, nil
			}
		}
		// Try as Decred address (prefer testnet, then mainnet/simnet/regnet)
		paramsList := []*chaincfg.Params{
			chaincfg.TestNet3Params(),
			chaincfg.MainNetParams(),
			chaincfg.SimNetParams(),
			chaincfg.RegNetParams(),
		}
		var lastErr error
		for _, p := range paramsList {
			addr, err := stdaddr.DecodeAddress(s, p)
			if err != nil {
				lastErr = err
				continue
			}
			// Check if this is a P2PK (raw pubkey) address
			type compressedPubKeyer interface{ SerializedPubKey() []byte }
			if pk, ok := addr.(compressedPubKeyer); ok {
				b := pk.SerializedPubKey()
				if len(b) == 33 {
					return b, nil
				}
				if len(b) == 65 {
					pp, err := secp256k1.ParsePubKey(b)
					if err != nil {
						return nil, fmt.Errorf("invalid pubkey inside address: %v", err)
					}
					return pp.SerializeCompressed(), nil
				}
				return nil, fmt.Errorf("pubkey in address has unexpected length %d", len(b))
			}
			return nil, fmt.Errorf("address is not P2PK; pass a P2PK address or 33-byte pubkey hex")
		}
		return nil, fmt.Errorf("decode -address failed: %v", lastErr)
	}
	return nil, fmt.Errorf("missing -address flag")
}

func (m *appstate) startSettlement() {
	// Ensure session key A_c
	if m.genPrivHex == "" {
		p, _ := secp256k1.GeneratePrivateKey()
		m.genPrivHex = hex.EncodeToString(p.Serialize())
		m.genPubHex = hex.EncodeToString(p.PubKey().SerializeCompressed())
		m.settle.aCompHex = m.genPubHex
		m.notification = "Generated session key A_c (in-memory)."
		m.msgCh <- client.UpdatedMsg{}
	}

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
	pubBytes, _ := hex.DecodeString(m.settle.aCompHex)
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
		res, err := m.pc.RefOpenEscrow(m.pc.ID, pubBytes, payoutBytes, m.settle.betAtoms, m.settle.csvBlocks)
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

// preSign performs HELLO -> REQ -> VERIFY_OK -> SERVER_OK and returns.
func (m *appstate) preSign() {
	ctx := m.ctx
	stream, err := m.pc.RefStartSettlementStream(ctx)
	if err != nil {
		m.notification = fmt.Sprintf("stream error: %v", err)
		m.msgCh <- client.UpdatedMsg{}
		return
	}
	// match_id: "<wrID>|<hostId>"
	if m.currentWR == nil || m.currentWR.Id == "" {
		m.notification = "Join or create a waiting room and bind escrow before presigning"
		m.msgCh <- client.UpdatedMsg{}
		_ = stream.CloseSend()
		return
	}
	// Use room-scoped match_id anchored to the host so branch mapping is stable (host = A/0).
	m.settle.matchID = fmt.Sprintf("%s|%s", m.currentWR.Id, m.currentWR.HostId)

	pubBytes, _ := hex.DecodeString(m.settle.aCompHex)
	_ = stream.Send(&pong.ClientMsg{MatchId: m.settle.matchID, Kind: &pong.ClientMsg_Hello{Hello: &pong.Hello{MatchId: m.settle.matchID, CompPubkey: pubBytes, ClientVersion: "poc"}}})
	// Expect one or more REQ rounds (A-branch, B-branch)
	m.settle.draftInputs = make(map[string][]*pong.NeedPreSigs_PerInput)
	m.settle.draftPresigs = make(map[string]map[string]*pong.PreSig)
	for {
		in, err := stream.Recv()
		if err != nil {
			m.notification = fmt.Sprintf("handshake recv error: %v", err)
			m.msgCh <- client.UpdatedMsg{}
			_ = stream.CloseSend()
			return
		}
		if req := in.GetReq(); req != nil {
			m.notification = "Verifying draft & building pre-sigs…"
			m.msgCh <- client.UpdatedMsg{}

			verify, err := m.buildVerifyOk(req)
			if err != nil {
				m.notification = "presig build failed: " + err.Error()
				m.msgCh <- client.UpdatedMsg{}
				_ = stream.CloseSend()
				return
			}
			// Record per-draft artifacts
			m.settle.draftInputs[req.DraftTxHex] = req.Inputs
			if m.settle.draftPresigs[req.DraftTxHex] == nil {
				m.settle.draftPresigs[req.DraftTxHex] = make(map[string]*pong.PreSig)
			}
			for _, p := range verify.Presigs {
				m.settle.draftPresigs[req.DraftTxHex][p.InputId] = p
			}
			_ = stream.Send(&pong.ClientMsg{MatchId: m.settle.matchID, Kind: &pong.ClientMsg_VerifyOk{VerifyOk: verify}})
			continue
		}
		if ok := in.GetOk(); ok != nil {
			// Pick last seen draft as the immediate finalization target for UI preview
			var lastDraft string
			for dh := range m.settle.draftInputs {
				lastDraft = dh
			}
			m.settle.lastDraftHex = lastDraft
			m.settle.lastInputs = m.settle.draftInputs[lastDraft]
			m.settle.lastPresigs = m.settle.draftPresigs[lastDraft]
			_ = stream.CloseSend()
			m.notification = "Verified & exchanged pre-sigs for both branches (server OK)."
			m.msgCh <- client.UpdatedMsg{}
			return
		}
	}
}

// buildVerifyOk validates draft and inputs and constructs PreSig list and ack digest.
func (m *appstate) buildVerifyOk(req *pong.NeedPreSigs) (*pong.VerifyOk, error) {
	txb, err := hex.DecodeString(req.DraftTxHex)
	if err != nil {
		return nil, err
	}
	var tx wire.MsgTx
	if err := tx.Deserialize(bytes.NewReader(txb)); err != nil {
		return nil, err
	}
	inputs := make([]*pong.NeedPreSigs_PerInput, len(req.Inputs))
	copy(inputs, req.Inputs)
	sort.Slice(inputs, func(i, j int) bool { return inputs[i].InputId < inputs[j].InputId })
	var concat []byte
	concat = append(concat, []byte(req.DraftTxHex)...)
	for _, in := range inputs {
		concat = append(concat, []byte(in.InputId)...)
		concat = append(concat, []byte(in.MHex)...)
		concat = append(concat, in.TCompressed...)
	}
	ack := blake256.Sum256(concat)
	pres := make([]*pong.PreSig, 0, len(inputs))
	for _, in := range inputs {
		redeem, err := hex.DecodeString(in.RedeemScriptHex)
		if err != nil {
			return nil, fmt.Errorf("bad redeem: %w", err)
		}
		idx, err := client.FindInputIndex(&tx, in.InputId)
		if err != nil {
			return nil, err
		}
		mBytes, err := txscript.CalcSignatureHash(redeem, txscript.SigHashAll, &tx, idx, nil)
		if err != nil || len(mBytes) != 32 {
			return nil, fmt.Errorf("sighash")
		}
		mHex := hex.EncodeToString(mBytes)
		if mHex != in.MHex {
			return nil, fmt.Errorf("m mismatch")
		}
		rComp, sPrime, err := client.ComputePreSigMinusFull(m.genPrivHex, mHex, hex.EncodeToString(in.TCompressed))
		if err != nil {
			return nil, err
		}
		pres = append(pres, &pong.PreSig{InputId: in.InputId, RprimeCompressed: rComp, Sprime32: sPrime})
	}
	return &pong.VerifyOk{AckDigest: ack[:], Presigs: pres}, nil
}

func (m *appstate) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Lock()
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height
		m.logViewport.Width = msg.Width
		m.logViewport.Height = msg.Height - 6
		m.Unlock()
		return m, nil
	case client.UpdatedMsg:
		// Simply return the model to refresh the view
		return m, m.waitForMsg()
	case *pong.NtfnStreamResponse:
		// Handle specific notification types
		switch msg.NotificationType {
		case pong.NotificationType_GAME_READY_TO_PLAY:
			m.notification = "=== GAME CREATED! === Press 'r' or SPACE to signal you're ready to play!"
			m.currentGameId = msg.GameId
		case pong.NotificationType_COUNTDOWN_UPDATE:
			m.notification = msg.Message
		case pong.NotificationType_ON_PLAYER_READY:
			if msg.PlayerId != m.pc.ID {
				m.notification = "Opponent is ready to play"
			}
		case pong.NotificationType_MESSAGE:
			m.notification = msg.Message
			// If we saw escrow funding in mempool or confirmed, reflect the escrow bet in UI.
			if strings.Contains(strings.ToLower(msg.Message), "deposit seen in mempool") ||
				strings.Contains(strings.ToLower(msg.Message), "deposit confirmed") {
				if m.settle.betAtoms > 0 {
					m.betAmount = float64(m.settle.betAtoms) / 1e8
				}
			}
			// Fetch finalize bundle via RPC and produce fully signed tx
			if strings.Contains(msg.Message, "Gamma received; finalizing…") {
				if m.settle.matchID != "" {
					go func(matchID string) {
						bundle, err := m.pc.RefGetFinalizeBundle(matchID)
						if err != nil {
							m.notification = "Finalize bundle fetch failed: " + err.Error()
							m.msgCh <- client.UpdatedMsg{}
							return
						}
						// Merge presigs for the returned draft into local store and finalize
						presigs := make(map[string]*pong.PreSig)
						inputs := make([]*pong.NeedPreSigs_PerInput, 0, len(bundle.Inputs))
						for _, fin := range bundle.Inputs {
							presigs[strings.ToLower(fin.InputId)] = &pong.PreSig{InputId: fin.InputId, RprimeCompressed: fin.RprimeCompressed, Sprime32: fin.Sprime32}
							inputs = append(inputs, &pong.NeedPreSigs_PerInput{InputId: fin.InputId, RedeemScriptHex: fin.RedeemScriptHex})
						}
						if m.settle.draftPresigs == nil {
							m.settle.draftPresigs = make(map[string]map[string]*pong.PreSig)
						}
						m.settle.draftPresigs[bundle.DraftTxHex] = presigs
						if m.settle.draftInputs == nil {
							m.settle.draftInputs = make(map[string][]*pong.NeedPreSigs_PerInput)
						}
						m.settle.draftInputs[bundle.DraftTxHex] = inputs
						hexGamma := hex.EncodeToString(bundle.Gamma32)
						hexTx, err := m.finalizeWinner(hexGamma, bundle.DraftTxHex)
						if err != nil {
							m.notification = "error finalizing winner: " + err.Error()
							m.msgCh <- client.UpdatedMsg{}
							return
						}
						m.notification = "Finalized (not broadcast): " + hexTx
						m.msgCh <- client.UpdatedMsg{}
					}(m.settle.matchID)
				}
			}
		}
		return m, m.waitForMsg()
	case tea.KeyMsg:
		// Add Ctrl+C handling
		if msg.Type == tea.KeyCtrlC {
			m.cancel()
			return m, tea.Quit
		}

		switch msg.String() {
		case "l":
			// Switch to list rooms mode
			m.mode = listRooms
			m.listWaitingRooms()
			return m, nil
		case "c":
			// Escrow-first: create room when an escrow exists (server will enforce 0-conf funding)
			if m.settle.activeEscrowID == "" && !isF2p {
				m.notification = "Open escrow first ([X] -> [E]) before creating a room."
				return m, nil
			}
			err := m.createRoom()
			if err != nil {
				m.notification = fmt.Sprintf("Error creating room: %v", err)
			}
			return m, nil
		case "j":
			// Switch to join room mode
			m.mode = joinRoom
			m.selectedRoomIndex = 0
			m.listWaitingRooms()
			return m, nil
		case "p":
			// Allow presign from any mode when appropriate
			return m, m.handlePreSig()
		case "w", "up":
			if m.mode == gameMode {
				return m, m.handleGameInput(msg)
			} else if m.mode == joinRoom && m.selectedRoomIndex > 0 {
				m.selectedRoomIndex--
			}
			return m, nil
		case "s", "down":
			if m.mode == gameMode {
				return m, m.handleGameInput(msg)
			} else if m.mode == joinRoom && m.selectedRoomIndex < len(m.waitingRooms)-1 {
				m.selectedRoomIndex++
			}
			return m, nil
		case "enter":
			if m.mode == joinRoom && len(m.waitingRooms) > 0 {
				selectedRoom := m.waitingRooms[m.selectedRoomIndex]
				err := m.joinRoom(selectedRoom.Id)
				if err != nil {
					m.notification = fmt.Sprintf("Error joining room: %v", err)
				}
			}
			return m, nil
		case "q":
			// Leave the current waiting room
			if m.currentWR != nil && !m.isGameRunning {
				err := m.leaveRoom()
				if err != nil {
					m.notification = fmt.Sprintf("Error leaving room: %v", err)
				}
				return m, nil
			}
		case "v":
			if m.mode == gameIdle {
				m.mode = viewLogs
				// Get the last logs directly from the LogBackend
				if lines := m.logBackend.LastLogLines(100); len(lines) > 0 {
					m.logBuffer = lines
					m.logViewport.SetContent(strings.Join(lines, "\n"))
					m.logViewport.GotoBottom()
				}
				return m, nil
			}
		case "x":
			// Enter settlement mode and start streaming settlement
			if m.mode == gameIdle {
				m.mode = settlementMode
				m.notification = "Settlement: stream driven. [p]=retry stream, [Esc]=back"
				go func() { m.startSettlement(); m.msgCh <- client.UpdatedMsg{} }()
				return m, nil
			}

		case "esc":
			if m.mode == viewLogs {
				m.mode = gameIdle
				return m, nil
			}
			if m.mode == settlementMode {
				// Cancel any active funding status stream
				if m.fundingCancel != nil {
					m.fundingCancel()
					m.fundingCancel = nil
				}
				m.mode = gameIdle

				return m, nil
			}

		case "e":
			if m.mode == settlementMode {
				if m.settle.aCompHex == "" {
					m.notification = "Generate A_c first with [K]"
					return m, nil
				}
				// Default bet amount for POC simplicity
				if m.settle.betAtoms == 0 {
					m.settle.betAtoms = client.DefaultBetAtoms
				}
				if m.settle.csvBlocks == 0 {
					m.settle.csvBlocks = 64
				}
				payoutBytes, perr := m.payoutPubkeyFromConfHex()
				if perr != nil {
					m.notification = "address error: " + perr.Error()
					m.msgCh <- client.UpdatedMsg{}
					return m, nil
				}
				go func() {
					pubBytes, _ := hex.DecodeString(m.settle.aCompHex)
					res, err := m.pc.RefOpenEscrow(m.pc.ID, pubBytes, payoutBytes, m.settle.betAtoms, m.settle.csvBlocks)
					if err != nil {
						m.notification = err.Error()
						m.msgCh <- client.UpdatedMsg{}
						return
					}
					b, _ := json.MarshalIndent(res, "", "  ")
					m.settle.lastJSON = string(b)
					m.settle.activeEscrowID = res.EscrowId
					m.msgCh <- client.UpdatedMsg{}
				}()
				return m, nil
			}

		case "g":
			if m.mode == settlementMode {

				go func() {
					// Cancel previous stream if any
					if m.fundingCancel != nil {
						m.fundingCancel()
					}
					ctx, cancel := context.WithCancel(context.Background())
					m.fundingCancel = cancel
					// FundingStatus removed: migrate to server WaitFunding stream (both A and B)
					m.notification = "Use WaitFunding (server streams both A and B)."
					m.msgCh <- client.UpdatedMsg{}
					_ = ctx
				}()
				return m, nil
			}
		case "k":
			if m.mode == settlementMode {
				// Generate a fresh secp256k1 key and set k for this session
				p, _ := secp256k1.GeneratePrivateKey()
				m.genPrivHex = hex.EncodeToString(p.Serialize())
				m.genPubHex = hex.EncodeToString(p.PubKey().SerializeCompressed())
				m.settle.aCompHex = m.genPubHex
				m.notification = "Generated A_c; private key kept in-memory (copy from logs if you need persistence)"
				return m, nil
			}

		case "+", "=":
			if m.keyReleaseDelay < 500*time.Millisecond {
				m.keyReleaseDelay += 25 * time.Millisecond
				m.notification = fmt.Sprintf("Key release delay: %d ms", m.keyReleaseDelay/time.Millisecond)
			}
			return m, nil
		case "-", "_":
			if m.keyReleaseDelay > 50*time.Millisecond {
				m.keyReleaseDelay -= 25 * time.Millisecond
				m.notification = fmt.Sprintf("Key release delay: %d ms", m.keyReleaseDelay/time.Millisecond)
			}
			return m, nil
		}

		if msg.Type == tea.KeyF2 {
			m.mode = gameMode
			return m, nil
		}
		if msg.Type == tea.KeySpace {

			if m.isGameRunning {
				// When in game, space signals ready to play
				err := m.signalReadyToPlay()
				if err != nil {
					m.notification = fmt.Sprintf("Error signaling ready: %v", err)
				}
				return m, nil
			} else if m.pc.IsReady {
				// If already ready in waiting room, set to unready
				err := m.makeClientUnready()
				if err != nil {
					m.notification = fmt.Sprintf("Error signaling unreadiness: %v", err)
					return m, nil
				}
			} else {
				// If not ready in waiting room, only require being in a WR (server enforces escrow funding)
				if m.currentWR == nil {
					m.notification = "Join or create a waiting room before getting ready."
					return m, nil
				}
				m.mode = gameMode
				err := m.makeClientReady()
				if err != nil {
					m.notification = fmt.Sprintf("Error signaling readiness: %v", err)
					return m, nil
				}
			}
			return m, nil
		}
		if msg.Type == tea.KeyEsc {
			if m.mode == gameMode {
				// When exiting game mode, stop all paddle movement
				if m.upKeyPressed {
					m.pc.SendInput("ArrowUpStop")
					m.upKeyPressed = false
				}
				if m.downKeyPressed {
					m.pc.SendInput("ArrowDownStop")
					m.downKeyPressed = false
				}
			}

			m.mode = gameIdle
			return m, nil
		}

	case *pong.GameUpdateBytes:
		var gameUpdate pong.GameUpdate
		// Use Protocol Buffers unmarshaling instead of JSON
		if err := proto.Unmarshal(msg.Data, &gameUpdate); err != nil {
			m.err = err
			return m, nil
		}
		m.Lock()
		m.gameState = &gameUpdate
		m.Unlock()

		return m, m.waitForMsg()
	case string:
		if strings.HasPrefix(msg, "Error:") {
			m.notification = msg
			return m, nil
		}
	}
	return m, m.waitForMsg()
}

func (m *appstate) waitForMsg() tea.Cmd {
	return func() tea.Msg {
		return <-m.msgCh
	}
}

func (m *appstate) listWaitingRooms() error {
	wr, err := m.pc.GetWaitingRooms()
	if err != nil {
		m.log.Errorf("Failed to get waiting rooms: %v", err)
		return err
	}
	m.waitingRooms = wr
	return nil
}

func (m *appstate) createRoom() error {
	var err error
	// Require escrow betAtoms and use it as the room bet for UI/consistency
	if m.settle.betAtoms == 0 {
		m.notification = "Set bet atoms first ([X] -> [E] or prefill bet) before creating a room"
		return nil
	}
	wr, err := m.pc.CreateWaitingRoom(m.pc.ID, int64(m.settle.betAtoms), m.settle.activeEscrowID)
	if err != nil {
		m.log.Errorf("Error creating room: %v", err)
		return err
	}

	m.currentWR = wr
	return nil
}

func (m *appstate) joinRoom(roomID string) error {
	res, err := m.pc.JoinWaitingRoom(roomID, m.settle.activeEscrowID)
	if err != nil {
		m.log.Errorf("Failed to join room %s: %v", roomID, err)
		return err
	}
	m.currentWR = res.Wr

	m.mode = gameMode
	return nil
}

func (m *appstate) makeClientReady() error {
	err := m.pc.SignalReady()
	if err != nil {
		m.log.Errorf("Failed to signal ready state: %v", err)
		return err
	}
	return nil
}

func (m *appstate) makeClientUnready() error {
	err := m.pc.SignalUnready()
	if err != nil {
		m.log.Errorf("Failed to signal unready state: %v", err)
		return err
	}
	return nil
}

func (m *appstate) handleGameInput(msg tea.KeyMsg) tea.Cmd {
	return func() tea.Msg {
		var input string

		switch msg.String() {
		case "w", "up":
			m.Lock()
			// Only send if not already pressed
			if !m.upKeyPressed {
				input = "ArrowUp"
				m.upKeyPressed = true

				// Cancel existing timer if any
				if m.upKeyTimer != nil {
					m.upKeyTimer.Stop()
				}

				// Set timer to automatically release the key
				m.upKeyTimer = time.AfterFunc(m.keyReleaseDelay, func() {
					m.Lock()
					if m.upKeyPressed {
						m.upKeyPressed = false
						err := m.pc.SendInput("ArrowUpStop")
						if err != nil {
							m.log.Errorf("Error auto-releasing up key: %v", err)
						}
					}
					m.Unlock()
				})

				// If down was pressed, release it
				if m.downKeyPressed {
					m.pc.SendInput("ArrowDownStop")
					m.downKeyPressed = false
					if m.downKeyTimer != nil {
						m.downKeyTimer.Stop()
					}
				}
			}
			m.Unlock()
		case "s", "down":
			m.Lock()
			// Only send if not already pressed
			if !m.downKeyPressed {
				input = "ArrowDown"
				m.downKeyPressed = true

				// Cancel existing timer if any
				if m.downKeyTimer != nil {
					m.downKeyTimer.Stop()
				}

				// Set timer to automatically release the key
				m.downKeyTimer = time.AfterFunc(m.keyReleaseDelay, func() {
					m.Lock()
					if m.downKeyPressed {
						m.downKeyPressed = false
						err := m.pc.SendInput("ArrowDownStop")
						if err != nil {
							m.log.Errorf("Error auto-releasing down key: %v", err)
						}
					}
					m.Unlock()
				})

				// If up was pressed, release it
				if m.upKeyPressed {
					m.pc.SendInput("ArrowUpStop")
					m.upKeyPressed = false
					if m.upKeyTimer != nil {
						m.upKeyTimer.Stop()
					}
				}
			}
			m.Unlock()
		case "esc":
			// When exiting game mode, stop all movement
			m.Lock()
			if m.upKeyPressed {
				m.pc.SendInput("ArrowUpStop")
				m.upKeyPressed = false
			}
			if m.downKeyPressed {
				m.pc.SendInput("ArrowDownStop")
				m.downKeyPressed = false
			}
			m.Unlock()
		}

		if input != "" {
			err := m.pc.SendInput(input)
			if err != nil {
				m.log.Errorf("Error sending game input: %v", err)
				return err
			}
		}
		return nil
	}
}

func (m *appstate) leaveRoom() error {
	if m.currentWR == nil {
		return fmt.Errorf("not in a waiting room")
	}

	err := m.pc.LeaveWaitingRoom(m.currentWR.Id)
	if err != nil {
		m.log.Errorf("Failed to leave room %s: %v", m.currentWR.Id, err)
		return err
	}

	m.currentWR = nil
	m.mode = gameIdle
	m.notification = "Successfully left the waiting room"
	return nil
}

func (m *appstate) signalReadyToPlay() error {
	if !m.isGameRunning {
		return fmt.Errorf("no active game to signal readiness")
	}

	if m.currentGameId == "" {
		return fmt.Errorf("game ID not available")
	}

	err := m.pc.SignalReadyToPlay(m.currentGameId)
	if err != nil {
		return fmt.Errorf("failed to signal ready to play: %v", err)
	}

	m.notification = "*** YOU ARE READY TO PLAY! *** Waiting for opponent..."
	return nil
}

// finalizeWinner computes s = s' + gamma (mod n), builds sig65 = r||s||0x01,
// attaches winner-path scriptSig = <sig65> OP_1 <redeem> for each presigned input,
// and validates locally via the script engine (P2SH wrapper) to catch stack/order issues.
func (m *appstate) finalizeWinner(gammaHex, draftHex string) (string, error) {
	tryFinalize := func(dh string) (string, error) {
		presigsForDraft := m.settle.draftPresigs[dh]
		inputsForDraft := m.settle.draftInputs[dh]
		if presigsForDraft == nil || len(presigsForDraft) == 0 || inputsForDraft == nil || len(inputsForDraft) == 0 {
			return "", fmt.Errorf("no presigs stored for this draft")
		}

		// Decode the exact draft that was used during presign.
		raw, err := hex.DecodeString(strings.TrimSpace(dh))
		if err != nil {
			return "", fmt.Errorf("decode draft hex: %w", err)
		}
		var tx wire.MsgTx
		if err := tx.Deserialize(bytes.NewReader(raw)); err != nil {
			return "", fmt.Errorf("deserialize draft: %w", err)
		}

		// Map draft inputs -> index for deterministic lookup.
		idxByID := make(map[string]int, len(tx.TxIn))
		have := make([]string, 0, len(tx.TxIn))
		for i, ti := range tx.TxIn {
			k := strings.ToLower(fmt.Sprintf("%s:%d", ti.PreviousOutPoint.Hash.String(), ti.PreviousOutPoint.Index))
			idxByID[k] = i
			have = append(have, k)
		}
		sort.Strings(have)

		// Build redeem lookup (input_id -> redeem hex) from what you stored at presign for this draft.
		redeemByID := make(map[string]string, len(inputsForDraft))
		for _, in := range inputsForDraft {
			redeemByID[strings.ToLower(in.InputId)] = in.RedeemScriptHex
		}

		// Parse gamma.
		gb, err := hex.DecodeString(strings.TrimSpace(gammaHex))
		if err != nil || len(gb) != 32 {
			return "", fmt.Errorf("bad gamma")
		}
		var gamma secp256k1.ModNScalar
		if overflow := gamma.SetByteSlice(gb); overflow {
			return "", fmt.Errorf("gamma overflow")
		}

		// Require a presig for every input in the draft to avoid half-signed txs.
		for id := range idxByID {
			if _, ok := presigsForDraft[id]; !ok {
				return "", fmt.Errorf("missing presig for input %s; cannot finalize", id)
			}
		}

		// Finalize each presigned input for this draft.
		for id, ps := range presigsForDraft {
			key := strings.ToLower(strings.TrimSpace(id))

			// Ensure the draft actually has this input.
			idx, ok := idxByID[key]
			if !ok {
				return "", fmt.Errorf("input %s not found in draft. draft has: %v", id, have)
			}

			// Get redeem script for this input.
			redHex, ok := redeemByID[key]
			if !ok {
				return "", fmt.Errorf("no redeem script stored for input %s", id)
			}
			redeem, err := hex.DecodeString(redHex)
			if err != nil {
				return "", fmt.Errorf("decode redeem for %s: %w", id, err)
			}

			// Sanity: presig sizes and even-Y on R'.
			if len(ps.RprimeCompressed) != 33 || len(ps.Sprime32) != 32 {
				return "", fmt.Errorf("bad presig sizes for %s", id)
			}
			if ps.RprimeCompressed[0] != 0x02 {
				return "", fmt.Errorf("R' must have even Y (0x02) for %s, got 0x%02x", id, ps.RprimeCompressed[0])
			}

			// s = s' + gamma (mod n)
			var sPrime secp256k1.ModNScalar
			if overflow := sPrime.SetByteSlice(ps.Sprime32); overflow {
				return "", fmt.Errorf("s' overflow for %s", id)
			}
			sPrime.Add(&gamma)
			sBytes := sPrime.Bytes()

			// sig65 = r||s || 0x01 (SigHashAll)
			rX := ps.RprimeCompressed[1:33]
			sig65 := make([]byte, 0, 65)
			sig65 = append(sig65, rX...)
			sig65 = append(sig65, sBytes[:]...)
			sig65 = append(sig65, byte(txscript.SigHashAll))

			// Winner scriptSig: <sig65> OP_1 <redeem>
			sb := txscript.NewScriptBuilder()
			sb.AddData(sig65)
			sb.AddOp(txscript.OP_1)
			sb.AddData(redeem)
			sigScript, err := sb.Script()
			if err != nil {
				return "", fmt.Errorf("build scriptSig for %s: %w", id, err)
			}
			tx.TxIn[idx].SignatureScript = sigScript

			// --- Local VM verify (P2SH wrapper) ---
			// Execute against OP_HASH160 <Hash160(redeem)> OP_EQUAL, not the redeem itself.
			sh := dcrutil.Hash160(redeem)
			pkScript, err := txscript.NewScriptBuilder().
				AddOp(txscript.OP_HASH160).
				AddData(sh).
				AddOp(txscript.OP_EQUAL).
				Script()
			if err != nil {
				return "", fmt.Errorf("build pkScript for %s: %w", id, err)
			}

			// scriptVersion 0, no special flags/sigcache while iterating.
			vm, err := txscript.NewEngine(pkScript, &tx, idx, 0, 0, nil)
			if err != nil {
				return "", fmt.Errorf("engine init for %s: %w", id, err)
			}
			if err := vm.Execute(); err != nil {
				return "", fmt.Errorf("local VM verify failed on %s: %w", id, err)
			}
		}

		// Serialize finalized tx.
		var out bytes.Buffer
		_ = tx.Serialize(&out)
		return hex.EncodeToString(out.Bytes()), nil
	}

	if txHex, err := tryFinalize(draftHex); err == nil {
		return txHex, nil
	}
	// Fallback: try all stored drafts with the given gamma, pick the first that verifies
	for dh := range m.settle.draftPresigs {
		if txHex, err := tryFinalize(dh); err == nil {
			return txHex, nil
		}
	}
	return "", fmt.Errorf("no presigs stored for this draft")
}

func (m *appstate) View() string {
	var b strings.Builder

	// Show the header and controls only if the game is not in game mode
	if !m.isGameRunning {
		// Build the header
		b.WriteString("========== Pong Game Client ==========\n\n")

		if m.notification != "" {
			b.WriteString(fmt.Sprintf("🔔 Notification: %s\n\n", m.notification))
		} else {
			b.WriteString("🔔 No new notifications.\n\n")
		}

		b.WriteString(fmt.Sprintf("👤 Player ID: %s\n", m.pc.ID))
		b.WriteString(fmt.Sprintf("💵 Bet Amount: %.8f\n", m.betAmount))
		b.WriteString(fmt.Sprintf("✅ Status Ready: %t\n", m.pc.IsReady))

		// Display the current room or show a placeholder if not in a room
		if m.currentWR != nil {
			b.WriteString(fmt.Sprintf("🏠 Current Room: %s\n\n", m.currentWR.Id))
		} else {
			b.WriteString("🏠 Current Room: None\n\n")
		}

		// Instructions
		b.WriteString("===== Controls =====\n")
		b.WriteString("Use the following keys to navigate:\n")
		b.WriteString("[L] - List rooms\n")
		b.WriteString("[C] - Create room\n")
		b.WriteString("[J] - Join room\n")
		b.WriteString("[Q] - Leave current room\n")
		b.WriteString("[V] - View logs\n")
		b.WriteString("[X] - Settlement (escrow/referee) menu\n")
		b.WriteString("[P] - Presign drafts when notified\n")
		b.WriteString("[Ctrl+C] - Exit\n")
		b.WriteString("====================\n\n")

		if !m.isGameRunning && m.currentWR != nil {
			if m.pc.IsReady {
				b.WriteString("[Space] - Toggle ready status (currently READY)\n")
			} else {
				b.WriteString("[Space] - Toggle ready status (currently NOT READY)\n")
			}
		}

		// Add key release delay info
		b.WriteString(fmt.Sprintf("⏱️ Key Release Delay: %d ms\n", m.keyReleaseDelay/time.Millisecond))
	}

	// Switch based on the current mode
	switch m.mode {
	case gameIdle:
		b.WriteString("\n[Idle Mode]\n")

	case gameMode:
		b.WriteString("\n[Game Mode]\n")
		b.WriteString("Press 'Esc' to return to the main menu.\n")
		b.WriteString("Use W/S or Arrow Keys to move.\n")
		b.WriteString(fmt.Sprintf("Use +/- to adjust key release delay (current: %d ms).\n\n", m.keyReleaseDelay/time.Millisecond))

		if m.gameState != nil {
			var gameView strings.Builder

			// Calculate header and footer sizes
			headerLines := strings.Count(b.String(), "\n") + 1 // +1 because the last line might not have a newline character
			footerLines := 2                                   // For the score and any additional messages

			// Calculate available space
			availableHeight := m.viewport.Height - headerLines - footerLines
			availableWidth := m.viewport.Width

			// Minimum game size constraints
			const minGameHeight = 5
			const minGameWidth = 10

			if availableHeight < minGameHeight || availableWidth < minGameWidth {
				b.WriteString("\n[Warning] Terminal window is too small to display the game.\n")
				b.WriteString("Please resize your window or use a larger terminal.\n")
				return b.String()
			}

			// Original game dimensions
			gameHeight := int(m.gameState.GameHeight)
			gameWidth := int(m.gameState.GameWidth)

			// Calculate scaling factors for width and height
			scaleY := float64(availableHeight) / float64(gameHeight)
			scaleX := float64(availableWidth) / float64(gameWidth)

			// Use the smaller scaling factor to ensure the game fits in both dimensions
			scale := math.Min(scaleX, scaleY)
			scale = math.Min(scale, 1.0) // Prevent upscaling

			// Scale the game elements
			scaledGameHeight := int(float64(gameHeight) * scale)
			scaledGameWidth := int(float64(gameWidth) * scale)

			// Ensure scaled dimensions do not exceed available space
			if scaledGameHeight > availableHeight {
				scaledGameHeight = availableHeight
			}
			if scaledGameWidth > availableWidth {
				scaledGameWidth = availableWidth
			}

			// Scale ball position
			ballX := int(math.Round(float64(m.gameState.BallX) * scale))
			ballY := int(math.Round(float64(m.gameState.BallY) * scale))

			// Scale paddle positions and sizes
			p1Y := int(math.Round(float64(m.gameState.P1Y) * scale))
			p1Height := int(math.Round(float64(m.gameState.P1Height) * scale))

			p2Y := int(math.Round(float64(m.gameState.P2Y) * scale))
			p2Height := int(math.Round(float64(m.gameState.P2Height) * scale))

			// Ensure positions are within bounds
			if ballX >= scaledGameWidth {
				ballX = scaledGameWidth - 1
			}
			if ballY >= scaledGameHeight {
				ballY = scaledGameHeight - 1
			}
			if p1Y+p1Height > scaledGameHeight {
				p1Height = scaledGameHeight - p1Y
			}
			if p2Y+p2Height > scaledGameHeight {
				p2Height = scaledGameHeight - p2Y
			}

			// Drawing the game
			for y := 0; y < scaledGameHeight; y++ {
				for x := 0; x < scaledGameWidth; x++ {
					switch {
					case x == ballX && y == ballY:
						gameView.WriteString("O")
					case x == 0 && y >= p1Y && y < p1Y+p1Height:
						gameView.WriteString("|")
					case x == scaledGameWidth-1 && y >= p2Y && y < p2Y+p2Height:
						gameView.WriteString("|")
					default:
						gameView.WriteString(" ")
					}
				}
				gameView.WriteString("\n")
			}

			// Append the score
			gameView.WriteString(fmt.Sprintf("Score: %d - %d\n", m.gameState.P1Score, m.gameState.P2Score))

			// Add ready status information with clear visibility
			if m.pc.IsReady {
				gameView.WriteString("\n*** You are READY to play! ***\n")
			} else {
				gameView.WriteString("\n*** Press 'r' or SPACE to signal you're ready to play ***\n")
			}

			b.WriteString(gameView.String())
		} else {
			b.WriteString("Waiting for game to start... Not all players are ready.\nHit [Space] to get ready\n")
		}

	case listRooms:
		b.WriteString("\n[List Rooms Mode]\n")
		if len(m.waitingRooms) > 0 {
			for i, room := range m.waitingRooms {
				b.WriteString(fmt.Sprintf("%d: Room ID %s - Bet Price: %.8f\n", i+1, room.Id, float64(room.BetAmt)/1e8))
			}
		} else {
			b.WriteString("No rooms available.\n")
		}
		b.WriteString("Press 'esc' to go back to the main menu.\n")

	case createRoom:
		b.WriteString("\n[Create Room Mode]\n")
		b.WriteString("Creating a new room...\n")

	case joinRoom:
		b.WriteString("\n[Join Room Mode]\n")
		b.WriteString("Select a room to join. Use [up]/[down] to navigate and [enter] to join.\n")
		b.WriteString("Press [esc] to go back to the main menu.\n")

		if len(m.waitingRooms) > 0 {
			for i, room := range m.waitingRooms {
				indicator := " " // Indicator for selected room
				if i == m.selectedRoomIndex {
					indicator = ">" // Mark the selected room
				}
				b.WriteString(fmt.Sprintf("%s %d: Room ID %s - Bet Price: %.8f\n", indicator, i+1, room.Id, float64(room.BetAmt)/1e8))
			}
		} else {
			b.WriteString("No rooms available.\n")
		}

	case viewLogs:
		b.WriteString("=============== Log Viewer ===============\n\n")
		if len(m.logBuffer) == 0 {
			b.WriteString("No logs available.\n")
		} else {
			// Set viewport height to leave room for header and footer
			m.logViewport.Height = m.viewport.Height - 6
			m.logViewport.Width = m.viewport.Width
			b.WriteString(m.logViewport.View())
		}
		b.WriteString("\n\n")
		b.WriteString("Press 'Esc' to return • ↑/↓ to scroll • PgUp/PgDn for pages • Home/End for top/bottom")

	case settlementMode:
		b.WriteString("\n[Settlement Mode]\n")
		b.WriteString("Auto: presign assigned branch. [p]=retry, [r]=reveal, [Esc]=back\n\n")
		b.WriteString(fmt.Sprintf("MatchID: %s\n", m.settle.matchID))
		b.WriteString(fmt.Sprintf("A_c: %s\n", m.settle.aCompHex))
		b.WriteString(fmt.Sprintf("B_c: %s\n", m.settle.bCompHex))
		b.WriteString(fmt.Sprintf("Escrows JSON: %s  PreSig JSON: %s\n\n", m.settle.escrowPath, m.settle.preSigPath))
		if m.settle.lastJSON != "" {
			b.WriteString("Last result:\n")
			b.WriteString(m.settle.lastJSON)
			b.WriteString("\n")
			var alloc struct {
				DepositAddress string `json:"deposit_address"`
				PkScriptHex    string `json:"pk_script_hex"`
			}
			if json.Unmarshal([]byte(m.settle.lastJSON), &alloc) == nil {
				if alloc.DepositAddress != "" {
					b.WriteString(fmt.Sprintf("Server deposit address: %s\n", alloc.DepositAddress))
				}
				if alloc.PkScriptHex != "" {
					b.WriteString(fmt.Sprintf("Deposit pkScript:       %s\n", alloc.PkScriptHex))
					b.WriteString("Use your node to derive the address from pkScript if needed.\n")
				}
				b.WriteString("\n")
			}
		}

	default:
		b.WriteString("\nUnknown mode.\n")
	}

	// Set the viewport content to the built string
	m.viewport.SetContent(b.String())

	// Return the viewport's view
	return m.viewport.View()
}

func realMain() error {
	flag.Parse()
	if *datadir == "" {
		*datadir = utils.AppDataDir("pongclient", false)
	}
	cfg, err := basecfg.LoadClientConfig(*datadir, "pongclient.conf")
	if err != nil {
		fmt.Println("Error loading configuration:", err)
		os.Exit(1)
	}

	// Apply overrides from flags
	if *flagURL != "" {
		cfg.RPCURL = *flagURL
	}
	if *flagServerCertPath != "" {
		cfg.ServerCertPath = *flagServerCertPath
	}
	if *flagClientCertPath != "" {
		cfg.ClientCertPath = *flagClientCertPath
	}
	if *flagClientKeyPath != "" {
		cfg.ClientKeyPath = *flagClientKeyPath
	}
	if *rpcUser != "" {
		cfg.RPCUser = *rpcUser
	}
	if *rpcPass != "" {
		cfg.RPCPass = *rpcPass
	}
	if *serverAddr != "" {
		cfg.ServerAddr = *serverAddr
	}
	if *grpcServerCert != "" {
		cfg.GRPCServerCert = *grpcServerCert
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	g, gctx := errgroup.WithContext(ctx)

	useStdout := false
	lb, err := logging.NewLogBackend(logging.LogConfig{
		LogFile:        filepath.Join(*datadir, "logs", "pongclient.log"),
		DebugLevel:     cfg.Debug,
		MaxLogFiles:    10,
		MaxBufferLines: 1000,
		UseStdout:      &useStdout,
	})
	if err != nil {
		return err
	}
	log := lb.Logger("BotClient")
	c, err := botclient.NewClient(cfg, lb)
	if err != nil {
		return err
	}
	g.Go(func() error { return c.RPCClient.Run(gctx) })

	var zkShortID zkidentity.ShortID
	req := &types.PublicIdentityReq{}
	var publicIdentity types.PublicIdentity
	err = c.Chat.UserPublicIdentity(ctx, req, &publicIdentity)
	if err != nil {
		return fmt.Errorf("failed to get user public identity: %v", err)
	}

	clientID := hex.EncodeToString(publicIdentity.Identity[:])
	copy(zkShortID[:], clientID)
	as := &appstate{
		ctx:        ctx,
		cancel:     cancel,
		log:        log,
		logBackend: lb,
		mode:       gameIdle,
	}
	as.dataDir = *datadir
	// Setup notification handlers.
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
		// Update bet amount for the player in the local state (e.g., as.Players).
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
		go func() {
			as.msgCh <- client.UpdatedMsg{}
		}()
	}))

	ntfns.Register(client.OnPlayerJoinedNtfn(func(wr *pong.WaitingRoom, ts time.Time) {
		as.currentWR = wr
		as.notification = "new player joined your waiting room"
		go func() {
			as.msgCh <- client.UpdatedMsg{}
		}()
	}))

	ntfns.Register(client.OnGameEndedNtfn(func(gameID, msg string, ts time.Time) {
		as.notification = fmt.Sprintf("game %s ended\n%s", gameID, msg)
		as.betAmount = 0
		as.isGameRunning = false
		as.mode = gameIdle
		go func() {
			as.msgCh <- client.UpdatedMsg{}
		}()
	}))

	ntfns.Register(client.OnPlayerLeftNtfn(func(wr *pong.WaitingRoom, playerID string, ts time.Time) {
		if playerID == clientID {
			as.currentWR = nil
			as.notification = "You left the waiting room"
		} else {
			as.currentWR = wr
			as.notification = fmt.Sprintf("Player %s left the waiting room", playerID)
		}
		go func() {
			as.msgCh <- client.UpdatedMsg{}
		}()
	}))

	pc, err := client.NewPongClient(clientID, &client.PongClientCfg{
		ServerAddr:    cfg.ServerAddr,
		Notifications: ntfns,
		Log:           log,
		GRPCCertPath:  cfg.GRPCServerCert,
	})
	if err != nil {
		return fmt.Errorf("failed to create pong client: %v", err)
	}
	as.pc = pc

	log.Infof("Connected to server at %s with ID %s", cfg.ServerAddr, clientID)

	// Test the connection immediately after creating the client
	_, err = pc.GetWaitingRooms()
	if err != nil {
		return fmt.Errorf("gRPC server connection failed: %v", err)
	}

	// Start the notifier in a goroutine
	g.Go(func() error { return pc.StartNotifier(ctx) })

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
