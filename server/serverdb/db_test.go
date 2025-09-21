package serverdb_test

import (
	"context"
	"encoding/binary"
	"path/filepath"
	"testing"
	"time"

	"github.com/companyzero/bisonrelay/clientrpc/types"
	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/vctt94/pongbisonrelay/server/serverdb"
)

// testPongServerDBInterface tests the ServerDB interface for tips management.
func testPongServerDBInterface(t *testing.T, db serverdb.ServerDB) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	// Generate a test client ID
	var clientID zkidentity.ShortID
	err := clientID.FromString("74657374436c69656e7431323334353674657374436c69656e74313233343536")
	if err != nil {
		t.Fatalf("Error generating client ID: %v", err)
	}

	// Create a test tip entry
	amount := 0.0001
	amountMatoms := int64(amount * 1e11)

	sequenceID := uint64(6546546516)
	tipID := make([]byte, 8)
	binary.BigEndian.PutUint64(tipID, 6546546516)

	// Store the tip and ensure it's retrievable
	err = db.StoreUnprocessedTip(ctx, &types.ReceivedTip{
		Uid:          clientID.Bytes(),
		AmountMatoms: amountMatoms,
		SequenceId:   sequenceID,
	})
	if err != nil {
		t.Fatalf("Failed to store tip: %v", err)
	}

	// Fetch the stored tip and verify its content
	tips, err := db.FetchReceivedTipsByUID(ctx, clientID, serverdb.StatusUnpaid)
	if err != nil {
		t.Fatalf("Failed to fetch tips: %v", err)
	}
	if len(tips) != 1 || tips[0].AmountMatoms != amountMatoms {
		t.Fatalf("Unexpected tip data: %+v", tips)
	}

	// Update the tip status to 'sending'
	err = db.UpdateTipStatus(ctx, clientID.Bytes(), tipID, serverdb.StatusSending)
	if err != nil {
		t.Fatalf("Failed to update tip status: %v", err)
	}

	// Verify the updated status
	updatedTips, err := db.FetchReceivedTipsByUID(ctx, clientID, serverdb.StatusSending)
	if err != nil {
		t.Fatalf("Failed to fetch updated tips: %v", err)
	}
	if len(updatedTips) != 1 {
		t.Fatalf("Unexpected tip data after status update: %+v", updatedTips)
	}

	// Remove the tip and confirm it's no longer in the database
	// err = db.RemoveTip(ctx, clientID.Bytes(), updatedTips[0].Tip.SequenceId)
	// if err != nil {
	// 	t.Fatalf("Failed to remove tip: %v", err)
	// }
	// finalTips, err := db.FetchReceivedTipsByUID(ctx, clientID, serverdb.StatusSending)
	// if err != nil {
	// 	t.Fatalf("Failed to fetch tips after removal: %v", err)
	// }
	// if len(finalTips) != 0 {
	// 	t.Fatalf("Expected no tips after removal, but found: %+v", finalTips)
	// }
}

// TestFSDB runs the database test using an FSDB instance.
func TestFSDB(t *testing.T) {
	dir := t.TempDir()

	db, err := serverdb.NewBoltDB(filepath.Join(dir, "tips.db"))
	if err != nil {
		t.Fatalf("Failed to initialize FSDB: %v", err)
	}
	defer db.Close()

	testPongServerDBInterface(t, db)
}
