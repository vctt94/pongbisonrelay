package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	pongbisonrelay "github.com/vctt94/pongbisonrelay"
	"github.com/vctt94/pongbisonrelay/client"
	"github.com/vctt94/pongbisonrelay/pongrpc/grpc/pong"
	"google.golang.org/protobuf/proto"
)

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
							presigs[strings.ToLower(fin.InputId)] = &pong.PreSig{InputId: fin.InputId, RLineCompressed: fin.RLineCompressed, SLine32: fin.SLine32}
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
						hexTx, err := pongbisonrelay.FinalizeWinner(hexGamma, bundle.DraftTxHex, inputs, presigs)
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

		case "r":
			// Refund via CSV to the same payout address configured
			if m.mode == settlementMode {
				if m.settle.activeEscrowID == "" {
					m.notification = "No active escrow to refund"
					return m, nil
				}
				payoutBytes, err := m.payoutPubkeyFromConfHex()
				if err != nil || len(payoutBytes) != 33 {
					m.notification = "Invalid payout address for refund"
					return m, nil
				}
				// Build refund asynchronously
				go func() {
					m.notification = "Building CSV refund..."
					m.msgCh <- client.UpdatedMsg{}
					// Fetch escrow utxo from server stream
					utxo, err := m.pc.GetEscrowUTXO(m.settle.activeEscrowID)
					if err != nil || utxo == nil {
						m.notification = "Refund failed: funding not found or not indexed yet"
						m.msgCh <- client.UpdatedMsg{}
						return
					}
					// Use the session private key for the refund (same key controls the deposit)
					priv := m.genPrivHex
					if strings.TrimSpace(priv) == "" {
						m.notification = "Refund failed: session private key not available (generate with [K])"
						m.msgCh <- client.UpdatedMsg{}
						return
					}
					fee := uint64(20000)
					csv := m.settle.csvBlocks
					if csv == 0 {
						csv = 64
					}
					// Destination: reuse the configured payout address text
					dest := *addressFlag
					xHex, err := client.BuildCSVRefundTx(
						priv,
						utxo.Txid,
						utxo.Vout,
						utxo.Value,
						utxo.RedeemScriptHex,
						dest,
						fee,
						csv,
					)
					if err != nil {
						m.notification = "Refund build error: " + err.Error()
						m.msgCh <- client.UpdatedMsg{}
						return
					}
					m.notification = "Refund tx hex (broadcast with your node): " + xHex
					m.msgCh <- client.UpdatedMsg{}
				}()
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
					// FundingStatus removed: migrate to server WaitFunding stream (both A and B)
					m.notification = "Use WaitFunding (server streams both A and B)."
					m.msgCh <- client.UpdatedMsg{}
				}()
				return m, nil
			}
		case "k":
			if m.mode == settlementMode {
				priv, pub, err := m.pc.GenerateNewSettlementSessionKey()
				if err != nil {
					m.notification = "error generating A_c: " + err.Error()
					return m, nil
				}
				m.genPrivHex = priv
				m.genPubHex = pub
				m.settle.aCompHex = pub
				m.notification = "Generated A_c; session key saved to disk (POC)."
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
					m.pc.RefSendInput("ArrowUpStop")
					m.upKeyPressed = false
				}
				if m.downKeyPressed {
					m.pc.RefSendInput("ArrowDownStop")
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
	wr, err := m.pc.RefGetWaitingRooms()
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
	wr, err := m.pc.RefCreateWaitingRoom(m.pc.ID, int64(m.settle.betAtoms), m.settle.activeEscrowID)
	if err != nil {
		m.log.Errorf("Error creating room: %v", err)
		return err
	}

	m.currentWR = wr
	return nil
}

func (m *appstate) joinRoom(roomID string) error {
	res, err := m.pc.RefJoinWaitingRoom(roomID, m.settle.activeEscrowID)
	if err != nil {
		m.log.Errorf("Failed to join room %s: %v", roomID, err)
		return err
	}
	m.currentWR = res.Wr

	m.mode = gameMode
	return nil
}

func (m *appstate) makeClientReady() error {
	err := m.pc.RefStartGameStream()
	if err != nil {
		m.log.Errorf("Failed to signal ready state: %v", err)
		return err
	}
	return nil
}

func (m *appstate) makeClientUnready() error {
	err := m.pc.RefUnreadyGameStream()
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
						err := m.pc.RefSendInput("ArrowUpStop")
						if err != nil {
							m.log.Errorf("Error auto-releasing up key: %v", err)
						}
					}
					m.Unlock()
				})

				// If down was pressed, release it
				if m.downKeyPressed {
					m.pc.RefSendInput("ArrowDownStop")
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
						err := m.pc.RefSendInput("ArrowDownStop")
						if err != nil {
							m.log.Errorf("Error auto-releasing down key: %v", err)
						}
					}
					m.Unlock()
				})

				// If up was pressed, release it
				if m.upKeyPressed {
					m.pc.RefSendInput("ArrowUpStop")
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
				m.pc.RefSendInput("ArrowUpStop")
				m.upKeyPressed = false
			}
			if m.downKeyPressed {
				m.pc.RefSendInput("ArrowDownStop")
				m.downKeyPressed = false
			}
			m.Unlock()
		}

		if input != "" {
			err := m.pc.RefSendInput(input)
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

	err := m.pc.RefLeaveWaitingRoom(m.currentWR.Id)
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

	err := m.pc.RefSignalReadyToPlay(m.currentGameId)
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
