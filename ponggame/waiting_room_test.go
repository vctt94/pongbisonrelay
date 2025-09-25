package ponggame

import (
	"context"
	"testing"

	"github.com/companyzero/bisonrelay/client/clientintf"
	"github.com/stretchr/testify/assert"
)

func createTestWaitingRoom() *WaitingRoom {
	ctx, cancel := context.WithCancel(context.Background())
	hostID := &clientintf.UserID{}

	return &WaitingRoom{
		Ctx:       ctx,
		Cancel:    cancel,
		ID:        "test-room-id",
		HostID:    hostID,
		Players:   []*Player{},
		BetAmount: 100,
	}
}

func TestWaitingRoom_AddPlayer(t *testing.T) {
	wr := createTestWaitingRoom()
	players := createTestPlayers()

	// Test adding players to waiting room
	wr.AddPlayer(players[0])
	assert.Equal(t, 1, len(wr.Players))
	assert.Equal(t, players[0], wr.Players[0])

	wr.AddPlayer(players[1])
	assert.Equal(t, 2, len(wr.Players))

	// Test adding the same player again (should not duplicate)
	wr.AddPlayer(players[0])
	assert.Equal(t, 2, len(wr.Players)) // Should still be 2
}

func TestWaitingRoom_GetPlayer(t *testing.T) {
	wr := createTestWaitingRoom()
	players := createTestPlayers()

	// Test getting player that doesn't exist
	player := wr.GetPlayer(players[0].ID)
	assert.Nil(t, player)

	// Add player and test getting it
	wr.AddPlayer(players[0])
	retrievedPlayer := wr.GetPlayer(players[0].ID)
	assert.Equal(t, players[0], retrievedPlayer)

	// Test with nil ID
	player = wr.GetPlayer(nil)
	assert.Nil(t, player)
}

func TestWaitingRoom_RemovePlayer(t *testing.T) {
	wr := createTestWaitingRoom()
	players := createTestPlayers()

	// Add players
	wr.AddPlayer(players[0])
	wr.AddPlayer(players[1])
	assert.Equal(t, 2, len(wr.Players))

	// Remove first player
	wr.RemovePlayer(players[0])
	assert.Equal(t, 1, len(wr.Players))
	assert.Equal(t, players[1], wr.Players[0])

	// Remove second player
	wr.RemovePlayer(players[1])
	assert.Equal(t, 0, len(wr.Players))

	// Test removing non-existent player (should not panic)
	wr.RemovePlayer(nil)
	assert.Equal(t, 0, len(wr.Players))
}

func TestWaitingRoom_Marshal(t *testing.T) {
	wr := createTestWaitingRoom()
	players := createTestPlayers()

	// Test marshaling empty waiting room
	pongWR := wr.Marshal()
	assert.Equal(t, wr.ID, pongWR.Id)
	assert.Equal(t, wr.BetAmount, pongWR.BetAmt)
	assert.Equal(t, 0, len(pongWR.Players))

	// Add players and test marshaling
	wr.AddPlayer(players[0])
	wr.AddPlayer(players[1])

	pongWR = wr.Marshal()
	assert.Equal(t, wr.ID, pongWR.Id)
	assert.Equal(t, wr.BetAmount, pongWR.BetAmt)
	assert.Equal(t, 2, len(pongWR.Players))

	// Verify player data is correctly marshaled
	for i, player := range pongWR.Players {
		assert.Equal(t, players[i].Nick, player.Nick)
		assert.Equal(t, players[i].BetAmt, player.BetAmt)
		assert.Equal(t, players[i].Ready, player.Ready)
	}
}

func TestWaitingRoom_ReadyPlayers(t *testing.T) {
	wr := createTestWaitingRoom()
	ps := createTestPlayers()

	// no players
	rp := wr.ReadyPlayers()
	assert.Len(t, rp, 0)
	assert.False(t, len(rp) == 2)

	// one unready
	ps[0].Ready = false
	wr.AddPlayer(ps[0])
	rp = wr.ReadyPlayers()
	assert.Len(t, rp, 0)
	assert.False(t, len(rp) == 2)

	// two unready
	ps[1].Ready = false
	wr.AddPlayer(ps[1])
	rp = wr.ReadyPlayers()
	assert.Len(t, rp, 0)
	assert.False(t, len(rp) == 2)

	// one ready
	ps[0].Ready = true
	rp = wr.ReadyPlayers()
	assert.Len(t, rp, 1)
	assert.False(t, len(rp) == 2)

	// two ready
	ps[1].Ready = true
	rp = wr.ReadyPlayers()
	assert.Len(t, rp, 2)
	assert.True(t, len(rp) == 2)
}

func TestWaitingRoom_ConcurrentAccess(t *testing.T) {
	wr := createTestWaitingRoom()
	players := createTestPlayers()

	// Test concurrent access to waiting room operations
	done := make(chan bool, 3)

	// Goroutine 1: Add player
	go func() {
		wr.AddPlayer(players[0])
		done <- true
	}()

	// Goroutine 2: Get player
	go func() {
		wr.GetPlayer(players[0].ID)
		done <- true
	}()

	// Goroutine 3: Marshal waiting room
	go func() {
		wr.Marshal()
		done <- true
	}()

	// Wait for all goroutines to complete
	for i := 0; i < 3; i++ {
		<-done
	}
}

func TestWaitingRoom_BetAmountCalculations(t *testing.T) {
	wr := createTestWaitingRoom()
	players := createTestPlayers()

	// Set different bet amounts
	players[0].BetAmt = 100
	players[1].BetAmt = 200

	wr.AddPlayer(players[0])
	wr.AddPlayer(players[1])

	// Test that waiting room tracks bet amount correctly
	assert.Equal(t, int64(100), wr.BetAmount) // Initial bet amount

	// Test total bet calculation
	totalBets := int64(0)
	for _, player := range wr.Players {
		totalBets += player.BetAmt
	}
	assert.Equal(t, int64(300), totalBets)
}

func TestWaitingRoom_StateConsistency(t *testing.T) {
	wr := createTestWaitingRoom()
	players := createTestPlayers()

	// Test that waiting room maintains consistent state
	assert.Equal(t, "test-room-id", wr.ID)
	assert.NotNil(t, wr.Ctx)
	assert.NotNil(t, wr.Cancel)
	assert.Equal(t, int64(100), wr.BetAmount)

	// Add players and verify state
	wr.AddPlayer(players[0])
	wr.AddPlayer(players[1])

	// Marshal and verify consistency
	pongWR := wr.Marshal()
	assert.Equal(t, wr.ID, pongWR.Id)
	assert.Equal(t, len(wr.Players), len(pongWR.Players))

	// Remove player and verify state
	wr.RemovePlayer(players[0])
	assert.Equal(t, 1, len(wr.Players))

	// Verify remaining player is correct
	remaining := wr.GetPlayer(players[1].ID)
	assert.Equal(t, players[1], remaining)
}
