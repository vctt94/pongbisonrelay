package ponggame

import (
	"context"
	"fmt"
	"time"

	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/decred/slog"
	"github.com/ndabAP/ping-pong/engine"
	"github.com/vctt94/bisonbotkit/utils"
	"github.com/vctt94/pong-bisonrelay/pongrpc/grpc/pong"
	"google.golang.org/protobuf/proto"
)

const maxScore = 3

// HandleWaitingRoomDisconnection handles player disconnection from a waiting room.
func (gm *GameManager) HandleWaitingRoomDisconnection(clientID zkidentity.ShortID, log slog.Logger) {
	wr := gm.GetWaitingRoomFromPlayer(clientID)
	if wr == nil {
		return
	}

	// If host disconnected, remove wr
	if clientID == *wr.HostID {
		remainingPlayers := GetRemainingPlayersInWaitingRoom(wr, clientID)
		for _, player := range remainingPlayers {
			if player.NotifierStream != nil {
				player.NotifierStream.Send(&pong.NtfnStreamResponse{
					NotificationType: pong.NotificationType_OPPONENT_DISCONNECTED,
					Message:          "Host left the waiting room. Room closed.",
					Started:          false,
				})
			}
		}

		log.Debugf("Player %s disconnected; removing waiting room %s", clientID, wr.ID)
		wr.Cancel()
		return
	}

	// Handle regular player disconnection
	wr.RemovePlayer(clientID)
	remainingPlayers := GetRemainingPlayersInWaitingRoom(wr, clientID)

	// Marshal updated waiting room
	pongwr := wr.Marshal()
	// Notify remaining players
	for _, player := range remainingPlayers {
		if player.NotifierStream != nil {
			player.NotifierStream.Send(&pong.NtfnStreamResponse{
				NotificationType: pong.NotificationType_OPPONENT_DISCONNECTED,
				Message:          "Player left the waiting room",
				Wr:               pongwr,
			})
		}
	}

	log.Debugf("Player %s disconnected from waiting room %s", clientID, wr.ID)
}

// HandleGameDisconnection handles player disconnection from an active game.
func (gm *GameManager) HandleGameDisconnection(clientID zkidentity.ShortID, log slog.Logger) {
	game := gm.GetPlayerGame(clientID)
	if game == nil {
		return
	}

	remainingPlayer := GetRemainingPlayerInGame(game, clientID)
	if remainingPlayer != nil && remainingPlayer.NotifierStream != nil {
		remainingPlayer.NotifierStream.Send(&pong.NtfnStreamResponse{
			NotificationType: pong.NotificationType_OPPONENT_DISCONNECTED,
			Message:          "Opponent disconnected. Game over.",
			Started:          false,
		})
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
	g.RLock()
	defer g.RUnlock()

	for _, room := range g.WaitingRooms {
		for _, p := range room.Players {
			if *p.ID == playerID {
				return room
			}
		}
	}
	return nil
}

func (g *GameManager) GetWaitingRoom(roomID string) *WaitingRoom {
	g.RLock()
	defer g.RUnlock()

	for _, room := range g.WaitingRooms {
		if room.ID == roomID {
			return room
		}
	}
	return nil
}

func (gm *GameManager) RemoveWaitingRoom(roomID string) {
	gm.Lock()
	defer gm.Unlock()

	for i, room := range gm.WaitingRooms {
		if room.ID == roomID {
			// Remove the room by appending the elements before and after it
			gm.WaitingRooms = append(gm.WaitingRooms[:i], gm.WaitingRooms[i+1:]...)
			gm.Log.Debugf("Waiting room %s removed successfully", roomID)

			break
		}
	}
}

func (gm *GameManager) GetPlayerGame(clientID zkidentity.ShortID) *GameInstance {
	gm.RLock()
	defer gm.RUnlock()
	return gm.PlayerGameMap[clientID]
}

func (s *GameManager) StartGame(ctx context.Context, players []*Player) (*GameInstance, error) {
	s.Lock()
	defer s.Unlock()
	gameID, err := utils.GenerateRandomString(16)
	if err != nil {
		return nil, err
	}

	newGameInstance := s.startNewGame(ctx, players, gameID)
	s.Games[gameID] = newGameInstance

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
		player.Score = 0
		betAmt += player.BetAmt
		// Create individual frame buffer for each player with frame dropping capability
		player.FrameCh = make(chan []byte, INPUT_BUF_SIZE/4) // Smaller buffer per player
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
		gm.PlayerGameMap[*player.ID] = newGame
		if player.GameStream != nil {
			// Send initial dimensions
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

			// Set paddle positions
			gameUpdate.P1X = engineState.P1PosX
			gameUpdate.P1Y = engineState.P1PosY
			gameUpdate.P2X = engineState.P2PosX
			gameUpdate.P2Y = engineState.P2PosY

			// Set ball position
			gameUpdate.BallX = engineState.BallPosX
			gameUpdate.BallY = engineState.BallPosY

			// Set velocities
			gameUpdate.P1YVelocity = 0
			gameUpdate.P2YVelocity = 0
			gameUpdate.BallXVelocity = engineState.BallVelX
			gameUpdate.BallYVelocity = engineState.BallVelY

			// Set FPS and TPS
			gameUpdate.Fps = engineState.FPS
			gameUpdate.Tps = engineState.TPS

			sendInitialGameState(player, gameUpdate)
		}

		// Notify all players that the game has started
		if player.NotifierStream != nil {
			player.NotifierStream.Send(&pong.NtfnStreamResponse{
				NotificationType: pong.NotificationType_GAME_READY_TO_PLAY,
				Message:          "Game created! Signal when ready to play.",
				Started:          true,
				GameId:           id,
				PlayerNumber:     player.PlayerNumber,
			})
		}
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
				if allPlayersReady && !g.CountdownStarted && !g.GameReady {
					g.CountdownStarted = true
					g.Unlock()

					// Start the countdown
					go g.startCountdown()
				} else {
					g.Unlock()
				}

				// If game is ready, start the actual gameplay
				if g.GameReady {
					// Start actual gameplay
					go func() {
						// Run a new round only if the game is still running
						if g.Running {
							g.engine.NewRound(g.ctx, g.Framesch, g.Inputch, g.roundResult)
						}
					}()

					go func() {
						for winnerNumber := range g.roundResult {
							if !g.Running {
								break
							}

							// Handle the result of each round
							g.handleRoundResult(winnerNumber)

							// Check if the game should continue or end
							if g.shouldEndGame() {
								// clean up the game after ending
								g.Cleanup()
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
			g.Lock()

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
				Fps:           engineState.FPS,
				Tps:           engineState.TPS,
			}

			// Send countdown notification to all players
			for _, player := range g.Players {
				if player.NotifierStream != nil {
					player.NotifierStream.Send(&pong.NtfnStreamResponse{
						NotificationType: pong.NotificationType_COUNTDOWN_UPDATE,
						Message:          fmt.Sprintf("Game starting in %d...", g.CountdownValue),
						GameId:           g.Id,
					})
				}

				// Send current game state to all players during countdown
				if player.GameStream != nil {
					sendInitialGameState(player, gameUpdate)
				}
			}

			g.CountdownValue--

			// Check if countdown has finished
			if g.CountdownValue < 0 {
				g.GameReady = true
				g.CountdownStarted = false

				// Notify players that the game is starting
				for _, player := range g.Players {
					if player.NotifierStream != nil {
						player.NotifierStream.Send(&pong.NtfnStreamResponse{
							NotificationType: pong.NotificationType_GAME_START,
							Message:          "Game is starting now!",
							Started:          true,
							GameId:           g.Id,
						})
					}
				}

				g.Unlock()
				return // Exit the countdown goroutine
			}

			g.Unlock()
		}
	}
}

func (g *GameInstance) handleRoundResult(winner int32) {
	// update player score
	for _, player := range g.Players {
		if player.PlayerNumber == winner {
			player.Score++
		}
	}
}

func (g *GameInstance) Cleanup() {
	g.cleanedUp = true
	g.cancel()
	close(g.Framesch)
	close(g.Inputch)
	close(g.roundResult)

	// Close individual player frame channels
	for _, player := range g.Players {
		if player.FrameCh != nil {
			close(player.FrameCh)
			player.FrameCh = nil
		}
	}
}

func (g *GameInstance) shouldEndGame() bool {
	for _, player := range g.Players {
		// Check if any player has reached the max score
		if player.Score >= maxScore {
			g.log.Infof("Game ending: Player %s reached the maximum score of %d", player.ID, player.Score)
			g.Winner = player.ID
			g.Running = false
			return true
		}
	}

	// Add other conditions as needed, e.g., time limit or disconnection
	if g.isTimeout() {
		g.log.Info("Game ending: Timeout reached")
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

	players[0].PlayerNumber = 1
	players[1].PlayerNumber = 2

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
				for _, player := range g.Players {
					if player.FrameCh != nil {
						close(player.FrameCh)
					}
				}
				return
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
							g.log.Debugf("Dropping frame for player %s (buffer full)", player.ID)
						}
					}
				}
			}
		}
	}
}

// Fix the code that was causing "bytes declared and not used" and "select case must be send or receive" errors
func sendInitialGameState(player *Player, gameUpdate *pong.GameUpdate) {
	if player.GameStream == nil {
		return
	}

	bytes, err := proto.Marshal(gameUpdate)
	if err != nil {
		return
	}

	// Use the bytes variable by sending it
	player.GameStream.Send(&pong.GameUpdateBytes{Data: bytes})
}
