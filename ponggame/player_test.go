package ponggame

import (
	"testing"

	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/stretchr/testify/assert"
	"github.com/vctt94/pongbisonrelay/pongrpc/grpc/pong"
)

func TestPlayer_Marshal(t *testing.T) {
	short := func() *zkidentity.ShortID {
		s := zkidentity.ShortID{}
		return &s
	}

	tests := []struct {
		name      string
		player    *Player
		want      *pong.Player // nil means we expect Marshal() to return nil
		expectNil bool
	}{
		{
			name:      "nil player",
			player:    nil,
			expectNil: true,
		},
		{
			name: "player with nil ID",
			player: &Player{
				ID:           nil,
				Nick:         "test",
				BetAmt:       100,
				PlayerNumber: 1,
				Score:        5,
				Ready:        true,
			},
			expectNil: true,
		},
		{
			name: "valid player",
			player: &Player{
				ID:           short(),
				Nick:         "TestPlayer",
				BetAmt:       250,
				PlayerNumber: 1,
				Score:        3,
				Ready:        true,
			},
			want: &pong.Player{
				Nick:   "TestPlayer",
				BetAmt: 250,
				Number: 1,
				Score:  3,
				Ready:  true,
			},
		},
		{
			name: "player with zero values",
			player: &Player{
				ID:           short(),
				Nick:         "",
				BetAmt:       0,
				PlayerNumber: 0,
				Score:        0,
				Ready:        false,
			},
			want: &pong.Player{
				Nick:   "",
				BetAmt: 0,
				Number: 0,
				Score:  0,
				Ready:  false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.player.Marshal()

			if tt.expectNil {
				assert.Nil(t, got)
				return
			}

			if assert.NotNil(t, got) {
				assert.Equal(t, tt.want.Nick, got.Nick)
				assert.Equal(t, tt.want.BetAmt, got.BetAmt)
				assert.EqualValues(t, tt.want.Number, got.Number)
				assert.EqualValues(t, tt.want.Score, got.Score)
				assert.Equal(t, tt.want.Ready, got.Ready)
				assert.NotEmpty(t, got.Uid) // derived from player ID
			}
		})
	}
}

func TestPlayer_ResetPlayer(t *testing.T) {
	// Create a player with all fields set
	shortID := zkidentity.ShortID{}

	player := &Player{
		ID:           &shortID,
		Nick:         "TestPlayer",
		BetAmt:       100,
		PlayerNumber: 1,
		Score:        5,
		Ready:        true,
		WR:           &WaitingRoom{},
	}

	// Reset the player
	player.ResetPlayer()

	// Verify all fields are reset except ID and Nick
	assert.Equal(t, &shortID, player.ID)       // ID should remain
	assert.Equal(t, "TestPlayer", player.Nick) // Nick should remain
	assert.Equal(t, int64(0), player.BetAmt)
	assert.Equal(t, int32(0), player.PlayerNumber)
	assert.Equal(t, 0, player.Score)
	assert.False(t, player.Ready)
	assert.Nil(t, player.GameStream)
	assert.Nil(t, player.WR)
}

func TestPlayerSessions_CreateSession(t *testing.T) {
	ps := &PlayerSessions{
		Sessions: make(map[zkidentity.ShortID]*Player),
	}

	clientID := zkidentity.ShortID{}

	// Test creating a new session
	player := ps.CreateSession(clientID)
	assert.NotNil(t, player)
	assert.Equal(t, &clientID, player.ID)
	assert.Equal(t, 0, player.Score)

	// Test getting existing session
	existingPlayer := ps.CreateSession(clientID)
	assert.Equal(t, player, existingPlayer) // Should return the same player instance
}

func TestPlayerSessions_GetPlayer(t *testing.T) {
	ps := &PlayerSessions{
		Sessions: make(map[zkidentity.ShortID]*Player),
	}

	clientID := zkidentity.ShortID{}

	// Test getting non-existent player
	player := ps.GetPlayer(clientID)
	assert.Nil(t, player)

	// Create a session and test getting it
	createdPlayer := ps.CreateSession(clientID)
	retrievedPlayer := ps.GetPlayer(clientID)
	assert.Equal(t, createdPlayer, retrievedPlayer)
}

func TestPlayerSessions_RemovePlayer(t *testing.T) {
	ps := &PlayerSessions{
		Sessions: make(map[zkidentity.ShortID]*Player),
	}

	clientID := zkidentity.ShortID{}

	// Create a session
	ps.CreateSession(clientID)
	assert.NotNil(t, ps.GetPlayer(clientID))

	// Remove the player
	ps.RemovePlayer(clientID)
	assert.Nil(t, ps.GetPlayer(clientID))

	// Test removing non-existent player (should not panic)
	ps.RemovePlayer(clientID)
}

func TestPlayerSessions_ConcurrentAccess(t *testing.T) {
	ps := &PlayerSessions{
		Sessions: make(map[zkidentity.ShortID]*Player),
	}

	clientID := zkidentity.ShortID{}

	// Test concurrent access doesn't cause race conditions
	done := make(chan bool, 3)

	// Goroutine 1: Create session
	go func() {
		ps.CreateSession(clientID)
		done <- true
	}()

	// Goroutine 2: Get player
	go func() {
		ps.GetPlayer(clientID)
		done <- true
	}()

	// Goroutine 3: Remove player
	go func() {
		ps.RemovePlayer(clientID)
		done <- true
	}()

	// Wait for all goroutines to complete
	for i := 0; i < 3; i++ {
		<-done
	}
}
