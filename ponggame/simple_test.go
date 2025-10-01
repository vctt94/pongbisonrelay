package ponggame

import (
	"context"
	"testing"

	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/decred/slog"
	"github.com/stretchr/testify/assert"
)

// Basic working tests for core functionality

func TestBasic_PlayerSessions(t *testing.T) {
	ps := &PlayerSessions{
		Sessions: make(map[zkidentity.ShortID]*Player),
	}

	clientID := zkidentity.ShortID{}

	// Test creating a session
	player := ps.CreateSession(clientID)
	assert.NotNil(t, player)
	assert.Equal(t, 0, player.Score)

	// Test getting the session
	retrievedPlayer := ps.GetPlayer(clientID)
	assert.Equal(t, player, retrievedPlayer)

	// Test removing the session
	ps.RemovePlayer(clientID)
	removedPlayer := ps.GetPlayer(clientID)
	assert.Nil(t, removedPlayer)
}

func TestBasic_PlayerMarshal(t *testing.T) {
	// Test valid player marshaling
	shortID := zkidentity.ShortID{}
	player := &Player{
		ID:           &shortID,
		Nick:         "TestPlayer",
		BetAmt:       100,
		PlayerNumber: 1,
		Score:        5,
		Ready:        true,
	}

	pongPlayer := player.Marshal()
	assert.Equal(t, "TestPlayer", pongPlayer.Nick)
	assert.Equal(t, int64(100), pongPlayer.BetAmt)
	assert.Equal(t, int32(1), pongPlayer.Number)
	assert.Equal(t, int32(5), pongPlayer.Score)
	assert.True(t, pongPlayer.Ready)

}

func TestBasic_PlayerReset(t *testing.T) {
	shortID := zkidentity.ShortID{}
	player := &Player{
		ID:           &shortID,
		Nick:         "TestPlayer",
		BetAmt:       100,
		PlayerNumber: 1,
		Score:        5,
		Ready:        true,
	}

	player.ResetPlayer()

	// Check that fields are reset
	assert.Equal(t, int64(0), player.BetAmt)
	assert.Equal(t, int32(0), player.PlayerNumber)
	assert.Equal(t, 0, player.Score)
	assert.False(t, player.Ready)
}

func TestBasic_WaitingRoomOperations(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wr := &WaitingRoom{
		ID:        "test-room",
		Ctx:       ctx,
		Cancel:    cancel,
		Players:   []*Player{},
		BetAmount: 100,
	}

	// Test adding players
	player1ID := zkidentity.ShortID{}
	player1 := &Player{
		ID:   &player1ID,
		Nick: "Player1",
	}

	wr.AddPlayer(player1)
	assert.Equal(t, 1, len(wr.Players))

	// Test getting players
	players := wr.GetPlayers()
	assert.Equal(t, 1, len(players))
	assert.Equal(t, player1, players[0])

	// Test removing players
	wr.RemovePlayer(player1)
	assert.Equal(t, 0, len(wr.Players))
}

func TestBasic_WaitingRoomMarshal(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	hostID := zkidentity.ShortID{}
	wr := &WaitingRoom{
		ID:        "test-room",
		Ctx:       ctx,
		Cancel:    cancel,
		HostID:    &hostID,
		Players:   []*Player{},
		BetAmount: 100,
	}

	pongWR := wr.Marshal()
	assert.Equal(t, "test-room", pongWR.Id)
	assert.Equal(t, int64(100), pongWR.BetAmt)
	assert.Equal(t, 0, len(pongWR.Players))
}

func TestBasic_GameManagerOperations(t *testing.T) {
	gm := &GameManager{
		Games:          make(map[string]*GameInstance),
		WaitingRooms:   []*WaitingRoom{},
		PlayerSessions: &PlayerSessions{Sessions: make(map[zkidentity.ShortID]*Player)},
		PlayerGameMap:  make(map[zkidentity.ShortID]*GameInstance),
		Log:            slog.Disabled,
	}

	// Test basic game manager setup
	assert.NotNil(t, gm.Games)
	assert.NotNil(t, gm.PlayerSessions)
	assert.Equal(t, 0, len(gm.Games))
	assert.Equal(t, 0, len(gm.WaitingRooms))

	// Test getting non-existent player game
	nonExistentID := zkidentity.ShortID{}
	game := gm.GetPlayerGame(nonExistentID)
	assert.Nil(t, game)
}

func TestBasic_PhysicsOperations(t *testing.T) {
	// Test basic vector operations
	v1 := Vec2{X: 3, Y: 4}
	v2 := Vec2{X: 1, Y: 2}

	// Test Add
	result := v1.Add(v2)
	assert.Equal(t, Vec2{X: 4, Y: 6}, result)

	// Test Scale
	scaled := v1.Scale(2.0)
	assert.Equal(t, Vec2{X: 6, Y: 8}, scaled)

	// Test intersection function
	rect1 := Rect{Cx: 100, Cy: 100, HalfW: 50, HalfH: 50}
	rect2 := Rect{Cx: 120, Cy: 120, HalfW: 50, HalfH: 50}
	rect3 := Rect{Cx: 200, Cy: 200, HalfW: 20, HalfH: 20}

	assert.True(t, intersects(rect1, rect2))  // Overlapping
	assert.False(t, intersects(rect1, rect3)) // Non-overlapping
}

func TestBasic_GameInstanceLifecycle(t *testing.T) {
	// Create unique IDs for players
	player1ID := zkidentity.ShortID{1}
	player2ID := zkidentity.ShortID{2}

	players := []*Player{
		{ID: &player1ID, Nick: "Player1", Score: 0, PlayerNumber: 1},
		{ID: &player2ID, Nick: "Player2", Score: 0, PlayerNumber: 2},
	}

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

	// Test initial state
	assert.True(t, game.Running)
	assert.Equal(t, 2, len(game.Players))
	assert.Equal(t, "test-game", game.Id)

	// Test round result handling
	game.handleRoundResult(1)
	assert.Equal(t, 1, players[0].Score)
	assert.Equal(t, 0, players[1].Score)

	// Test cleanup
	game.Cleanup()
	assert.True(t, game.cleanedUp)
}

func TestBasic_UtilityFunctions(t *testing.T) {
	// Create unique IDs for players
	player1ID := zkidentity.ShortID{1}
	player2ID := zkidentity.ShortID{2}

	players := []*Player{
		{ID: &player1ID, Nick: "Player1"},
		{ID: &player2ID, Nick: "Player2"},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wr := &WaitingRoom{
		ID:      "test-room",
		Ctx:     ctx,
		Cancel:  cancel,
		Players: players,
	}

	// Test GetRemainingPlayersInWaitingRoom
	remaining := GetRemainingPlayersInWaitingRoom(wr, *players[0].ID)
	assert.Equal(t, 1, len(remaining))
	assert.Equal(t, players[1], remaining[0])

	// Test GetRemainingPlayerInGame
	game := &GameInstance{
		Id:      "test-game",
		Players: players,
	}

	remainingPlayer := GetRemainingPlayerInGame(game, *players[0].ID)
	assert.Equal(t, players[1], remainingPlayer)
}
