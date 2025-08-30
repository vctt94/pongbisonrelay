package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/vctt94/pong-bisonrelay/client"
	"github.com/vctt94/pong-bisonrelay/pongrpc/grpc/pong"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/proto"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/companyzero/bisonrelay/clientrpc/types"
	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/decred/slog"
	"github.com/vctt94/bisonbotkit/botclient"
	"github.com/vctt94/bisonbotkit/config"
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
)

// cachedInput stores per-input data for both branches returned by the server.
type cachedInput struct {
	owner           string
	redeemScriptHex string
	mAwins          string
	tAwins          string
	mBwins          string
	tBwins          string
}

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
		branch     pong.Branch
		escrowPath string
		preSigPath string
		lastJSON   string
		betAtoms   uint64
		csvBlocks  uint32
		xA         string // winner A payout pubkey (compressed hex)
		xB         string // winner B payout pubkey (compressed hex)

		// Drafts and per-input cache for both branches
		draftAHex string
		draftBHex string
		inputs    map[string]cachedInput
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
			// Switch to create room mode if player has a bet
			if m.betAmount > 0 || isF2p {
				err := m.createRoom()
				if err != nil {
					m.notification = fmt.Sprintf("Error creating room: %v", err)
				}
				return m, nil
			} else {
				m.notification = "Bet amount must be > 0 to create a room."
			}
		case "j":
			// Switch to join room mode
			m.mode = joinRoom
			m.selectedRoomIndex = 0
			m.listWaitingRooms()
			return m, nil
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
		case "p":
			if m.mode == settlementMode {
				go func() { m.startSettlementStream(); m.msgCh <- client.UpdatedMsg{} }()
				return m, nil
			}
		case "x":
			// Enter settlement mode and start streaming settlement
			if m.mode == gameIdle {
				m.mode = settlementMode
				m.notification = "Settlement: stream driven. [p]=retry stream, [Esc]=back"
				go func() { m.startSettlementStream(); m.msgCh <- client.UpdatedMsg{} }()
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
		case "r":
			if m.isGameRunning {
				err := m.signalReadyToPlay()
				if err != nil {
					m.notification = fmt.Sprintf("Error signaling ready: %v", err)
				}
			}
			return m, nil

		case "e":
			if m.mode == settlementMode {
				if m.settle.aCompHex == "" {
					m.notification = "Generate A_c first with [K]"
					return m, nil
				}
				// Default bet amount for POC simplicity
				if m.settle.betAtoms == 0 {
					m.settle.betAtoms = 100000000
				}
				if m.settle.csvBlocks == 0 {
					m.settle.csvBlocks = 64
				}
				go func() {
					res, err := m.pc.RefAllocateEscrow(m.pc.ID, m.settle.aCompHex, m.settle.betAtoms, m.settle.csvBlocks)
					if err != nil {
						m.notification = err.Error()
						m.msgCh <- client.UpdatedMsg{}
						return
					}
					b, _ := json.MarshalIndent(res, "", "  ")
					m.settle.lastJSON = string(b)
					m.settle.matchID = res.MatchId
					m.settle.xA = res.XA
					m.settle.xB = res.XB
					// Defer reflecting bet amount until funding is observed (mempool/blocks)
					m.msgCh <- client.UpdatedMsg{}
				}()
				return m, nil
			}
		case "A":
			if m.mode == settlementMode {
				m.settle.branch = pong.Branch_BRANCH_A
				m.notification = "Selected branch: A"
				return m, nil
			}
		case "B":
			if m.mode == settlementMode {
				m.settle.branch = pong.Branch_BRANCH_B
				m.notification = "Selected branch: B"
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
				// If not ready in waiting room, set to ready
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
	_, err = m.pc.CreateWaitingRoom(m.pc.ID, int64(m.settle.betAtoms))
	if err != nil {
		m.log.Errorf("Error creating room: %v", err)
		return err
	}

	m.mode = gameMode
	return nil
}

func (m *appstate) joinRoom(roomID string) error {
	res, err := m.pc.JoinWaitingRoom(roomID)
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
			headerLines := countLines(b.String())
			footerLines := 2 // For the score and any additional messages

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
		b.WriteString("Auto: presign both branches when you enter. [p]=retry, [r]=reveal, [Esc]=back\n\n")
		b.WriteString(fmt.Sprintf("MatchID: %s\n", m.settle.matchID))
		b.WriteString(fmt.Sprintf("A_c: %s\n", m.settle.aCompHex))
		b.WriteString(fmt.Sprintf("B_c: %s\n", m.settle.bCompHex))
		b.WriteString(fmt.Sprintf("Branch: %v\n", m.settle.branch))
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
	cfg, err := config.LoadClientConfig(*datadir, "pongclient.conf")
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
