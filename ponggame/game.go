package ponggame

import (
	"context"
	"fmt"
	"time"

	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/decred/slog"
	"github.com/ndabAP/ping-pong/engine"
	"github.com/vctt94/bisonbotkit/utils"
	"github.com/vctt94/pongbisonrelay/pongrpc/grpc/pong"
	"google.golang.org/protobuf/proto"
)

const maxScore = 3

// RemovePlayerFromWaitingRoom handles the core logic of removing a player from a waiting room
func (gm *GameManager) RemovePlayerFromWaitingRoom(clientID zkidentity.ShortID) (isHost bool, roomRemoved bool) {
	wr := gm.GetWaitingRoomFromPlayer(clientID)
	if wr == nil {
		return false, false
	}

	// Check if player is the host
	wr.RLock()
	hostID := wr.HostID
	wr.RUnlock()

	isHost = *hostID == clientID

	if isHost {
		// Host is leaving - remove the entire room
		gm.RemoveWaitingRoom(wr.ID)
		return true, true
	} else {
		// Non-host is leaving - remove just the player
		p := gm.PlayerSessions.GetPlayer(clientID)
		wr.RemovePlayer(p)
		return false, false
	}
}

func (gm *GameManager) HandlePlayerInput(clientID zkidentity.ShortID, req *pong.PlayerInput) (*pong.GameUpdate, error) {
	player := gm.PlayerSessions.GetPlayer(clientID)
	if player == nil {
		return nil, fmt.Errorf("player: %s not found", clientID)
	}
	if player.PlayerNumber != 1 && player.PlayerNumber != 2 {
		return nil, fmt.Errorf("player number incorrect, it must be 1 or 2; it is: %d", player.PlayerNumber)
	}

	game := gm.GetPlayerGame(clientID)
	if game == nil {
		return nil, fmt.Errorf("game instance not found for client ID %s", clientID)
	}

	req.PlayerNumber = player.PlayerNumber
	inputBytes, err := proto.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize input: %w", err)
	}

	if !game.Running {
		return nil, fmt.Errorf("game has ended for client ID %s", clientID)
	}

	// Try to send input without blocking, discard old inputs if channel is full
	select {
	case game.Inputch <- inputBytes:
		// success
	default:
		// channel is full; drop the oldest one
		select {
		case <-game.Inputch:
		default:
			// means the channel was emptied in the meantime
		}

		// now that we've popped one off, try once more
		select {
		case game.Inputch <- inputBytes:
			// success
		default:
			// still no room, just drop the input
			gm.Log.Debugf("Input channel full for game %s, dropping input", game.Id)
		}
	}

	return &pong.GameUpdate{}, nil
}

func (g *GameManager) GetWaitingRoomFromPlayer(playerID zkidentity.ShortID) *WaitingRoom {
	g.waitingRoomsMu.RLock()
	rooms := append([]*WaitingRoom(nil), g.WaitingRooms...)
	g.waitingRoomsMu.RUnlock()

	for _, room := range rooms {
		room.RLock()
		for _, p := range room.Players {
			if *p.ID == playerID {
				room.RUnlock()
				return room
			}
		}
		room.RUnlock()
	}
	return nil
}

func (g *GameManager) GetWaitingRoom(roomID string) *WaitingRoom {
	g.waitingRoomsMu.RLock()
	rooms := append([]*WaitingRoom(nil), g.WaitingRooms...)
	g.waitingRoomsMu.RUnlock()

	for _, room := range rooms {
		if room.ID == roomID {
			return room
		}
	}
	return nil
}

func (gm *GameManager) RemoveWaitingRoom(roomID string) error {
	gm.waitingRoomsMu.Lock()
	defer gm.waitingRoomsMu.Unlock()

	for i, room := range gm.WaitingRooms {
		if room.ID == roomID {
			// remove player from waiting room
			for _, p := range room.Players {
				room.RemovePlayer(p)
			}
			// Remove the room by appending the elements before and after it
			gm.WaitingRooms = append(gm.WaitingRooms[:i], gm.WaitingRooms[i+1:]...)
			gm.Log.Debugf("Waiting room %s removed successfully", roomID)
			break
		}
	}
	return nil
}

func (gm *GameManager) GetPlayerGame(clientID zkidentity.ShortID) *GameInstance {
	gm.playerGameMapMu.RLock()
	gi := gm.PlayerGameMap[clientID]
	gm.playerGameMapMu.RUnlock()
	return gi
}

func (s *GameManager) StartGame(ctx context.Context, players []*Player) (*GameInstance, error) {
	gameID, err := utils.GenerateRandomString(16)
	if err != nil {
		return nil, err
	}

	newGameInstance := s.startNewGame(ctx, players, gameID)
	s.gamesMu.Lock()
	s.Games[gameID] = newGameInstance
	s.gamesMu.Unlock()

	return newGameInstance, nil
}

func (gm *GameManager) startNewGame(ctx context.Context, players []*Player, id string) *GameInstance {
	framesch := make(chan []byte, INPUT_BUF_SIZE)
	inputch := make(chan []byte, INPUT_BUF_SIZE)
	roundResult := make(chan int32)
	ctx, cancel := context.WithCancel(ctx)

	// sum of all bets
	betAmt := int64(0)
	for _, player := range players {
		player.Lock()
		player.Score = 0
		betAmt += player.BetAmt
		// Create individual frame buffer for each player with frame dropping capability
		player.FrameCh = make(chan []byte, INPUT_BUF_SIZE/4) // Smaller buffer per player
		player.Unlock()
	}

	newGame := &GameInstance{
		Id:          id,
		Framesch:    framesch,
		Inputch:     inputch,
		roundResult: roundResult,
		Running:     true,
		ctx:         ctx,
		cancel:      cancel,
		Players:     players,
		betAmt:      betAmt,
		log:         gm.Log,

		// Initialize the ready to play fields
		PlayersReady:     make(map[string]bool),
		CountdownStarted: false,
		CountdownValue:   3,
		GameReady:        false,
	}

	// Setup engine
	swidth := 800.0
	sheight := 600.0

	newGame.engine = NewEngine(swidth, sheight, players, gm.Log)

	// Start frame distributor goroutine to distribute frames to individual player channels
	go newGame.distributeFrames()

	// Update PlayerSessions with the correct player numbers after NewEngine assigns them
	for _, player := range players {
		sessionPlayer := gm.PlayerSessions.GetPlayer(*player.ID)
		if sessionPlayer != nil {
			sessionPlayer.PlayerNumber = player.PlayerNumber
		}
	}

	// Map players to this game for easy lookup
	for _, player := range players {
		gm.playerGameMapMu.Lock()
		gm.PlayerGameMap[*player.ID] = newGame
		gm.playerGameMapMu.Unlock()

		// Send initial dimensions via per-player queue (best effort).
		engineState := newGame.engine.State()
		gameUpdate := &pong.GameUpdate{
			GameWidth:  swidth,
			GameHeight: sheight,
			P1Width:    engineState.PaddleWidth,
			P1Height:   engineState.PaddleHeight,
			P2Width:    engineState.PaddleWidth,
			P2Height:   engineState.PaddleHeight,
			BallWidth:  engineState.BallWidth,
			BallHeight: engineState.BallHeight,
		}
		// Set positions and velocities
		gameUpdate.P1X, gameUpdate.P1Y = engineState.P1PosX, engineState.P1PosY
		gameUpdate.P2X, gameUpdate.P2Y = engineState.P2PosX, engineState.P2PosY
		gameUpdate.BallX, gameUpdate.BallY = engineState.BallPosX, engineState.BallPosY
		gameUpdate.P1YVelocity, gameUpdate.P2YVelocity = 0, 0
		gameUpdate.BallXVelocity, gameUpdate.BallYVelocity = engineState.BallVelX, engineState.BallVelY
		// Do not set Fps/Tps here; client controls render FPS
		sendInitialGameState(player, gameUpdate)

		// Notify all players that the game has started (best effort).
		_ = player.EnqueueNotif(&pong.NtfnStreamResponse{
			NotificationType: pong.NotificationType_GAME_READY_TO_PLAY,
			Message:          "Game created! Signal when ready to play.",
			Started:          true,
			GameId:           id,
			PlayerNumber:     player.PlayerNumber,
		})
	}

	return newGame
}

func (g *GameInstance) Run() {
	g.Running = true

	// Wait for players to be ready before starting the actual game
	go func() {
		// Check every 500ms if both players are ready
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-g.ctx.Done():
				return
			case <-ticker.C:
				g.Lock()

				// Check if all players are ready
				allPlayersReady := len(g.PlayersReady) == len(g.Players)

				// If all players are ready and countdown hasn't started yet, start countdown
				gameReady := g.GameReady
				if allPlayersReady && !g.CountdownStarted && !gameReady {
					g.CountdownStarted = true
					g.Unlock()

					// Start the countdown
					go g.startCountdown()
				} else {
					g.Unlock()
				}

				// If game is ready, start the actual gameplay
				if gameReady {
					// Start actual gameplay - single goroutine to handle all rounds
					go func() {
						// Start the first round
						if g.Running {
							g.engine.NewRound(g.ctx, g.Framesch, g.Inputch, g.roundResult)
						}

						for winnerNumber := range g.roundResult {
							if !g.Running {
								break
							}

							// Handle the result of each round
							g.handleRoundResult(winnerNumber)

							// Check if the game should continue or end
							if g.shouldEndGame() {
								break
							} else {
								g.engine.NewRound(g.ctx, g.Framesch, g.Inputch, g.roundResult)
							}
						}
					}()

					return // Exit this goroutine once the game has started
				}
			}
		}
	}()
}

// startCountdown initiates and manages the countdown before the game starts
func (g *GameInstance) startCountdown() {
	countdownTicker := time.NewTicker(1 * time.Second)
	defer countdownTicker.Stop()

	for {
		select {
		case <-g.ctx.Done():
			return
		case <-countdownTicker.C:
			// Snapshot state
			g.RLock()
			countdownVal := g.CountdownValue
			gameID := g.Id
			playersSnap := append([]*Player(nil), g.Players...)
			g.RUnlock()

			// Fetch engine state independently (CanvasEngine has its own locking).
			engineState := g.engine.State()
			gameUpdate := &pong.GameUpdate{
				GameWidth:     g.engine.Game.Width,
				GameHeight:    g.engine.Game.Height,
				P1Width:       engineState.PaddleWidth,
				P1Height:      engineState.PaddleHeight,
				P2Width:       engineState.PaddleWidth,
				P2Height:      engineState.PaddleHeight,
				BallWidth:     engineState.BallWidth,
				BallHeight:    engineState.BallHeight,
				P1X:           engineState.P1PosX,
				P1Y:           engineState.P1PosY,
				P2X:           engineState.P2PosX,
				P2Y:           engineState.P2PosY,
				BallX:         engineState.BallPosX,
				BallY:         engineState.BallPosY,
				P1YVelocity:   0,
				P2YVelocity:   0,
				BallXVelocity: 0,
				BallYVelocity: 0,
				// Omit Fps/Tps in updates; clients drive rendering FPS
			}

			// Send countdown updates and current state without holding g's lock.
			for _, player := range playersSnap {
				_ = player.EnqueueNotif(&pong.NtfnStreamResponse{
					NotificationType: pong.NotificationType_COUNTDOWN_UPDATE,
					Message:          fmt.Sprintf("Game starting in %d...", countdownVal),
					GameId:           gameID,
				})
				sendInitialGameState(player, gameUpdate)
			}

			// Apply countdown state change under lock.
			g.Lock()
			g.CountdownValue--
			finished := g.CountdownValue < 0
			var startPlayers []*Player
			if finished {
				g.GameReady = true
				g.CountdownStarted = false
				startPlayers = append([]*Player(nil), g.Players...)
			}
			g.Unlock()

			if finished {
				// Notify players game is starting now (no g lock held).
				for _, p := range startPlayers {
					_ = p.EnqueueNotif(&pong.NtfnStreamResponse{
						NotificationType: pong.NotificationType_GAME_START,
						Message:          "Game is starting now!",
						Started:          true,
						GameId:           gameID,
					})
				}
				return
			}
		}
	}
}

// handleRoundResult updates the score of the player who won the round
// does not hold locks to avoid deadlocks
func (g *GameInstance) handleRoundResult(winner int32) {
	// update player score
	for _, player := range g.Players {
		if player.PlayerNumber == winner {
			player.Score++
		}
	}
}

func (g *GameInstance) Cleanup() {
	g.closeOnce.Do(func() {
		g.cleanedUp = true
		g.cancel()
		// distributeFrames() will close per-player FrameChs.
		// Close roundResult channel to signal the round handler goroutine to exit
		close(g.roundResult)
		// Do NOT close g.Inputch here if others may still send/recv.
	})
}

func (g *GameInstance) shouldEndGame() bool {
	for _, player := range g.Players {
		// Check if any player has reached the max score
		if player.Score >= maxScore {
			g.log.Infof("Game ending: Player %s reached the maximum score of %d", player.ID, player.Score)
			g.Winner = player.ID
			g.Running = false
			// Signal that the game has ended by closing the frame channel
			// This will cause distributeFrames to exit and close player channels
			if g.Framesch != nil {
				close(g.Framesch)
			}
			return true
		}
	}

	// Add other conditions as needed, e.g., time limit or disconnection
	if g.isTimeout() {
		g.log.Info("Game ending: Timeout reached")
		g.Running = false
		// Signal that the game has ended by closing the frame channel
		if g.Framesch != nil {
			close(g.Framesch)
		}
		return true
	}

	// Return false if none of the end conditions are met
	return false
}

// isTimeout checks if the game duration has exceeded a set limit
func (g *GameInstance) isTimeout() bool {
	// For example, a simple time limit check
	// const maxGameDuration = 10 * time.Minute
	// return time.Since(g.startTime) >= maxGameDuration
	return false
}

// NewEngine creates a new CanvasEngine
func NewEngine(width, height float64, players []*Player, log slog.Logger) *CanvasEngine {
	// Create game with dimensions that match the display
	game := engine.NewGame(
		width, height,
		engine.NewPlayer(10, 75),
		engine.NewPlayer(10, 75),
		engine.NewBall(15, 15),
	)

	players[0].Lock()
	players[0].PlayerNumber = 1
	players[0].Unlock()
	players[1].Lock()
	players[1].PlayerNumber = 2
	players[1].Unlock()

	canvasEngine := New(game)
	canvasEngine.SetLogger(log).SetFPS(DEFAULT_FPS)

	canvasEngine.reset()

	return canvasEngine
}

// distributeFrames distributes frames from the main channel to individual player channels
// This prevents one slow client from affecting others by implementing frame dropping
func (g *GameInstance) distributeFrames() {
	for {
		select {
		case <-g.ctx.Done():
			return
		case frame, ok := <-g.Framesch:
			if !ok {
				// Main frame channel closed, close all player channels
				for _, p := range g.Players {
					if p.FrameCh != nil {
						close(p.FrameCh)
					}
				}
				return
			}

			// Coalesce backlog: keep only the latest frame if multiple are queued.
			for drain := 0; drain < 8; drain++ {
				select {
				case newer, ok2 := <-g.Framesch:
					if !ok2 {
						for _, player := range g.Players {
							if player.FrameCh != nil {
								close(player.FrameCh)
							}
						}
						return
					}
					frame = newer
				default:
					// No more queued frames; proceed with latest.
					break
				}
			}

			// Distribute frame to each player with non-blocking send and frame dropping
			for _, player := range g.Players {
				if player.FrameCh != nil {
					select {
					case player.FrameCh <- frame:
						// Frame sent successfully
					default:
						// Player's buffer is full, drop oldest frame and try again
						select {
						case <-player.FrameCh:
							// Dropped oldest frame
						default:
							// Channel was somehow emptied in the meantime
						}

						// Try to send the new frame
						select {
						case player.FrameCh <- frame:
							// Frame sent successfully after dropping old one
						default:
							// Still full, just drop this frame
						}
					}
				}
			}
		}
	}
}

// Fix the code that was causing "bytes declared and not used" and "select case must be send or receive" errors
func sendInitialGameState(player *Player, gameUpdate *pong.GameUpdate) {
	bytes, err := proto.Marshal(gameUpdate)
	if err != nil {
		return
	}
	// Enqueue bytes for sending via the per-player game sender.
	_ = player.EnqueueGameBytes(bytes)
}
