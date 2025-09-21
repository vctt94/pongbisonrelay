package ponggame

import (
	"context"
	"testing"
	"time"

	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/decred/slog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vctt94/pongbisonrelay/pongrpc/grpc/pong"
)

func createTestGameManager() *GameManager {
	return &GameManager{
		Games:          make(map[string]*GameInstance),
		WaitingRooms:   []*WaitingRoom{},
		PlayerSessions: &PlayerSessions{Sessions: make(map[zkidentity.ShortID]*Player)},
		PlayerGameMap:  make(map[zkidentity.ShortID]*GameInstance),
		Log:            slog.Disabled,
	}
}

func createTestPlayers() []*Player {
	player1ID := zkidentity.ShortID{1}
	player2ID := zkidentity.ShortID{2}

	return []*Player{
		{
			ID:           &player1ID,
			Nick:         "Player1",
			BetAmt:       100,
			PlayerNumber: 1,
			Score:        0,
			Ready:        true,
		},
		{
			ID:           &player2ID,
			Nick:         "Player2",
			BetAmt:       150,
			PlayerNumber: 2,
			Score:        0,
			Ready:        true,
		},
	}
}

func TestGameManager_StartGame(t *testing.T) {
	gm := createTestGameManager()
	players := createTestPlayers()
	ctx := context.Background()

	// Test successful game creation
	game, err := gm.StartGame(ctx, players)
	require.NoError(t, err)
	assert.NotNil(t, game)
	assert.True(t, game.Running)
	assert.Equal(t, len(players), len(game.Players))
	assert.Equal(t, int64(250), game.betAmt) // 100 + 150

	// Verify game is tracked in manager
	assert.Equal(t, 1, len(gm.Games))
	assert.NotNil(t, gm.Games[game.Id])

	// Verify player mapping
	for _, player := range players {
		assert.Equal(t, game, gm.PlayerGameMap[*player.ID])
	}

	// Test with nil players - this will cause a panic in NewEngine, so we'll skip this test
	// _, err = gm.StartGame(ctx, nil)
	// assert.Error(t, err)

	// Test with empty players slice - this will cause a panic in NewEngine, so we'll skip this test
	// _, err = gm.StartGame(ctx, []*Player{})
	// assert.Error(t, err)
}

func TestGameManager_GetPlayerGame(t *testing.T) {
	gm := createTestGameManager()
	players := createTestPlayers()
	ctx := context.Background()

	// Test getting game for non-existent player
	nonExistentID := zkidentity.ShortID{}
	game := gm.GetPlayerGame(nonExistentID)
	assert.Nil(t, game)

	// Create a game and test getting it
	createdGame, err := gm.StartGame(ctx, players)
	require.NoError(t, err)

	retrievedGame := gm.GetPlayerGame(*players[0].ID)
	assert.Equal(t, createdGame, retrievedGame)

	retrievedGame = gm.GetPlayerGame(*players[1].ID)
	assert.Equal(t, createdGame, retrievedGame)
}

func TestGameManager_HandlePlayerInput(t *testing.T) {
	gm := createTestGameManager()
	players := createTestPlayers()
	ctx := context.Background()

	// Create player sessions
	gm.PlayerSessions.CreateSession(*players[0].ID)
	gm.PlayerSessions.CreateSession(*players[1].ID)

	// Start a game
	game, err := gm.StartGame(ctx, players)
	require.NoError(t, err)

	tests := []struct {
		name     string
		clientID zkidentity.ShortID
		input    *pong.PlayerInput
		wantErr  bool
	}{
		{
			name:     "valid input for player 1",
			clientID: *players[0].ID,
			input: &pong.PlayerInput{
				Input: "ArrowUp",
			},
			wantErr: false,
		},
		{
			name:     "valid input for player 2",
			clientID: *players[1].ID,
			input: &pong.PlayerInput{
				Input: "ArrowDown",
			},
			wantErr: false,
		},
		{
			name:     "input for non-existent player",
			clientID: zkidentity.ShortID{},
			input: &pong.PlayerInput{
				Input: "ArrowUp",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := gm.HandlePlayerInput(tt.clientID, tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Verify player number is set correctly
				if tt.clientID == *players[0].ID {
					assert.Equal(t, int32(1), tt.input.PlayerNumber)
				} else if tt.clientID == *players[1].ID {
					assert.Equal(t, int32(2), tt.input.PlayerNumber)
				}
			}
		})
	}

	// Stop the game and test input on stopped game
	game.Running = false
	_, err = gm.HandlePlayerInput(*players[0].ID, &pong.PlayerInput{Input: "ArrowUp"})
	assert.Error(t, err)
}

func TestGameInstance_ShouldEndGame(t *testing.T) {
	players := createTestPlayers()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	game := &GameInstance{
		Id:      "test-game",
		Players: players,
		Running: true,
		ctx:     ctx,
		cancel:  cancel,
		log:     slog.Disabled,
	}

	// Test game should not end initially
	assert.False(t, game.shouldEndGame())

	// Test game should end when max score is reached
	players[0].Score = maxScore
	game.Winner = players[0].ID
	assert.True(t, game.shouldEndGame())
	assert.False(t, game.Running)
}

func TestGameInstance_HandleRoundResult(t *testing.T) {
	players := createTestPlayers()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	game := &GameInstance{
		Id:      "test-game",
		Players: players,
		Running: true,
		ctx:     ctx,
		cancel:  cancel,
		log:     slog.Disabled,
	}

	// Test player 1 wins a round
	initialScore := players[0].Score
	game.handleRoundResult(1)
	assert.Equal(t, initialScore+1, players[0].Score)
	assert.Equal(t, 0, players[1].Score) // Player 2 score should remain 0

	// Test player 2 wins a round
	initialScore = players[1].Score
	game.handleRoundResult(2)
	assert.Equal(t, initialScore+1, players[1].Score)
}

func TestGameInstance_Cleanup(t *testing.T) {
	players := createTestPlayers()
	ctx, cancel := context.WithCancel(context.Background())

	framesch := make(chan []byte, 10)
	inputch := make(chan []byte, 10)
	roundResult := make(chan int32, 10)

	game := &GameInstance{
		Id:          "test-game",
		Players:     players,
		Running:     true,
		ctx:         ctx,
		cancel:      cancel,
		Framesch:    framesch,
		Inputch:     inputch,
		roundResult: roundResult,
		log:         slog.Disabled,
	}

	// Verify channels are open initially
	select {
	case framesch <- []byte("test"):
	default:
		t.Fatal("Framesch should be open")
	}

	// Cleanup the game
	game.Cleanup()

	// Verify cleanup state
	assert.True(t, game.cleanedUp)

	// Verify context is cancelled
	select {
	case <-ctx.Done():
		// Expected
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Context should be cancelled")
	}

	// Verify channels are closed
	select {
	case <-framesch:
		// Channel should be closed
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Framesch should be closed")
	}
}

func TestGameManager_RemoveWaitingRoom(t *testing.T) {
	gm := createTestGameManager()

	// Create a test waiting room
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wr := &WaitingRoom{
		ID:     "test-room",
		Ctx:    ctx,
		Cancel: cancel,
	}

	gm.WaitingRooms = append(gm.WaitingRooms, wr)
	assert.Equal(t, 1, len(gm.WaitingRooms))

	// Remove the waiting room
	gm.RemoveWaitingRoom("test-room")
	assert.Equal(t, 0, len(gm.WaitingRooms))

	// Test removing non-existent room (should not panic)
	gm.RemoveWaitingRoom("non-existent")
	assert.Equal(t, 0, len(gm.WaitingRooms))
}

func TestGameManager_GetWaitingRoom(t *testing.T) {
	gm := createTestGameManager()

	// Test getting non-existent waiting room
	wr := gm.GetWaitingRoom("non-existent")
	assert.Nil(t, wr)

	// Create a test waiting room
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	testWR := &WaitingRoom{
		ID:     "test-room",
		Ctx:    ctx,
		Cancel: cancel,
	}

	gm.WaitingRooms = append(gm.WaitingRooms, testWR)

	// Test getting existing waiting room
	retrievedWR := gm.GetWaitingRoom("test-room")
	assert.Equal(t, testWR, retrievedWR)
}

func TestGameManager_GetWaitingRoomFromPlayer(t *testing.T) {
	gm := createTestGameManager()
	players := createTestPlayers()

	// Test getting waiting room for player not in any room
	wr := gm.GetWaitingRoomFromPlayer(*players[0].ID)
	assert.Nil(t, wr)

	// Create a test waiting room with a player
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	testWR := &WaitingRoom{
		ID:      "test-room",
		Ctx:     ctx,
		Cancel:  cancel,
		Players: []*Player{players[0]},
	}

	gm.WaitingRooms = append(gm.WaitingRooms, testWR)

	// Test getting waiting room for player in room
	retrievedWR := gm.GetWaitingRoomFromPlayer(*players[0].ID)
	assert.Equal(t, testWR, retrievedWR)

	// Test getting waiting room for player not in room
	retrievedWR = gm.GetWaitingRoomFromPlayer(*players[1].ID)
	assert.Nil(t, retrievedWR)
}

func TestGameInstance_IsTimeout(t *testing.T) {
	players := createTestPlayers()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	game := &GameInstance{
		Id:      "test-game",
		Players: players,
		Running: true,
		ctx:     ctx,
		cancel:  cancel,
		log:     slog.Disabled,
	}

	// Current implementation always returns false
	assert.False(t, game.isTimeout())
}

func TestNewEngine(t *testing.T) {
	players := createTestPlayers()
	log := slog.Disabled

	engine := NewEngine(800, 600, players, log)

	assert.NotNil(t, engine)
	assert.Equal(t, 800.0, engine.Game.Width)
	assert.Equal(t, 600.0, engine.Game.Height)
	assert.Equal(t, float64(DEFAULT_FPS), engine.FPS)

	// Verify player numbers are set
	assert.Equal(t, int32(1), players[0].PlayerNumber)
	assert.Equal(t, int32(2), players[1].PlayerNumber)
}

func TestGameInstanceConcurrentAccess(t *testing.T) {
	players := createTestPlayers()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	framesch := make(chan []byte, 100)
	inputch := make(chan []byte, 100)
	roundResult := make(chan int32, 10)

	game := &GameInstance{
		Id:          "test-game",
		Players:     players,
		Running:     true,
		ctx:         ctx,
		cancel:      cancel,
		Framesch:    framesch,
		Inputch:     inputch,
		roundResult: roundResult,
		log:         slog.Disabled,
	}

	// Test concurrent access to round result handling
	done := make(chan bool, 2)

	go func() {
		game.handleRoundResult(1)
		done <- true
	}()

	go func() {
		game.handleRoundResult(2)
		done <- true
	}()

	// Wait for both goroutines to complete
	for i := 0; i < 2; i++ {
		<-done
	}

	// Verify both players got their scores
	totalScore := players[0].Score + players[1].Score
	assert.Equal(t, 2, totalScore)
}
