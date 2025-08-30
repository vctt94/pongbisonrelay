package serverdb

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"time"

	"github.com/companyzero/bisonrelay/clientrpc/types"
	"github.com/companyzero/bisonrelay/zkidentity"
	bolt "go.etcd.io/bbolt"
)

var (
	receivedTipsBucket    = []byte("receivedTips")
	sendTipProgressBucket = []byte("sendTipsProgress")
	refMatchesBucket      = []byte("refMatches")
	refAllocBucket        = []byte("refAllocByPlayer")
)

// itob converte um uint64 em []byte usando BigEndian.
func itob(v uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, v)
	return b
}

// BoltDB implements the ServerDB interface using bbolt as the backend.
type boltDB struct {
	db *bolt.DB
}

var _ ServerDB = (*boltDB)(nil)

// NewBoltDB initializes the BoltDB instance with a bbolt database.
func NewBoltDB(dbPath string) (ServerDB, error) {
	db, err := bolt.Open(dbPath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, err
	}
	// Initialize the main and status buckets on the first run.
	// Initialize the main and status buckets on the first run.
	if err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(receivedTipsBucket)
		return err
	}); err != nil {
		db.Close()
		return nil, err
	}
	if err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(sendTipProgressBucket)
		return err
	}); err != nil {
		db.Close()
		return nil, err
	}
	// Initialize referee buckets
	if err = db.Update(func(tx *bolt.Tx) error {
		if _, e := tx.CreateBucketIfNotExists(refMatchesBucket); e != nil {
			return e
		}
		if _, e := tx.CreateBucketIfNotExists(refAllocBucket); e != nil {
			return e
		}
		return nil
	}); err != nil {
		db.Close()
		return nil, err
	}
	return &boltDB{db: db}, nil
}

// --- Referee persistence ---

func (b *boltDB) SaveRefMatch(ctx context.Context, rec *RefMatchRecord) error {
	if rec == nil || rec.MatchID == "" {
		return errors.New("invalid ref match record")
	}
	data, err := json.Marshal(rec)
	if err != nil {
		return err
	}
	return b.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(refMatchesBucket)
		if bucket == nil {
			return ErrTipBucketNotFound
		}
		return bucket.Put([]byte(rec.MatchID), data)
	})
}

func (b *boltDB) FetchRefMatch(ctx context.Context, matchID string) (*RefMatchRecord, error) {
	var out *RefMatchRecord
	err := b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(refMatchesBucket)
		if bucket == nil {
			return ErrTipBucketNotFound
		}
		data := bucket.Get([]byte(matchID))
		if data == nil {
			out = nil
			return nil
		}
		var rec RefMatchRecord
		if e := json.Unmarshal(data, &rec); e != nil {
			return e
		}
		out = &rec
		return nil
	})
	return out, err
}

func (b *boltDB) SaveRefAlloc(ctx context.Context, playerID string, matchID string) error {
	if playerID == "" || matchID == "" {
		return errors.New("invalid ref alloc mapping")
	}
	return b.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(refAllocBucket)
		if bucket == nil {
			return ErrTipBucketNotFound
		}
		return bucket.Put([]byte(playerID), []byte(matchID))
	})
}

func (b *boltDB) FetchRefAlloc(ctx context.Context, playerID string) (string, error) {
	var mid string
	err := b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(refAllocBucket)
		if bucket == nil {
			return ErrTipBucketNotFound
		}
		v := bucket.Get([]byte(playerID))
		if v == nil {
			mid = ""
			return nil
		}
		mid = string(v)
		return nil
	})
	return mid, err
}

// StoreUnprocessedTip stores a tip under the Uid sub-bucket in the main tips bucket.
func (b *boltDB) StoreUnprocessedTip(ctx context.Context, tip *types.ReceivedTip) error {
	payload := ReceivedTipWrapper{
		Tip:    tip,
		Status: StatusUnpaid,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return b.db.Update(func(tx *bolt.Tx) error {
		mainBucket := tx.Bucket(receivedTipsBucket)
		if mainBucket == nil {
			return ErrMainBucketNotFound
		}

		userBucket, err := mainBucket.CreateBucketIfNotExists(payload.Tip.Uid)
		if err != nil {
			return err
		}

		sequenceId := make([]byte, 8)
		binary.BigEndian.PutUint64(sequenceId, tip.SequenceId)
		if userBucket.Get(sequenceId) != nil {
			return ErrDuplicateEntry
		}

		return userBucket.Put(sequenceId, data)
	})
}

// MarkTipAsSending updates the status of a specific tip to "sending".
func (b *boltDB) UpdateTipStatus(ctx context.Context, uid []byte, tipID []byte, status TipStatus) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		mainBucket := tx.Bucket(receivedTipsBucket)
		if mainBucket == nil {
			return ErrMainBucketNotFound
		}

		// Locate the sub-bucket for the specified uid.
		userBucket := mainBucket.Bucket(uid)
		if userBucket == nil {
			return ErrUserBucketNotFound
		}

		// Retrieve the tip data by tipID within the user bucket.
		data := userBucket.Get(tipID)
		if data == nil {
			return ErrTipNotFound
		}

		// Decode the data into ReceivedTipWrapper
		wrapper := &ReceivedTipWrapper{}
		if err := json.Unmarshal(data, wrapper); err != nil {
			return err
		}

		// Update the status
		wrapper.Status = status

		// Re-encode the updated wrapper and store it back in the database.
		updatedData, err := json.Marshal(wrapper)
		if err != nil {
			return err
		}

		return userBucket.Put(tipID, updatedData)
	})
}

// Close closes the bbolt database.
func (b *boltDB) Close() error {
	return b.db.Close()
}

// FetchReceivedTipsByUID retrieves all tips with a specific status for a specific user by Uid.
func (b *boltDB) FetchReceivedTipsByUID(ctx context.Context, uid zkidentity.ShortID, status TipStatus) ([]*types.ReceivedTip, error) {
	var unprocessedTips []*types.ReceivedTip

	err := b.db.View(func(tx *bolt.Tx) error {
		mainBucket := tx.Bucket(receivedTipsBucket)
		if mainBucket == nil {
			return ErrTipBucketNotFound
		}

		// Retrieve the sub-bucket for the specified Uid
		userBucket := mainBucket.Bucket(uid[:])
		if userBucket == nil {
			// No tips found for this Uid
			return nil
		}

		// Iterate through each tip in the Uid's sub-bucket
		return userBucket.ForEach(func(tipID, v []byte) error {

			// Decode the protobuf payload into a ReceivedTip struct
			tip := &ReceivedTipWrapper{}
			if err := json.Unmarshal(v, tip); err != nil {
				return err
			}

			// Filter based on the specified status
			if tip.Status == status {
				// Add the decoded tip to the unprocessed tips slice
				unprocessedTips = append(unprocessedTips, tip.Tip)
			}
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	return unprocessedTips, nil
}

// FetchAllReceivedTipsByUID retrieves all tips for a specific user by Uid, regardless of status.
func (b *boltDB) FetchAllReceivedTipsByUID(ctx context.Context, uid zkidentity.ShortID) ([]ReceivedTipWrapper, error) {
	var allTips []ReceivedTipWrapper

	err := b.db.View(func(tx *bolt.Tx) error {
		mainBucket := tx.Bucket(receivedTipsBucket)
		if mainBucket == nil {
			return errors.New("tip bucket not found")
		}

		// Retrieve the sub-bucket for the specified Uid
		userBucket := mainBucket.Bucket(uid[:])
		if userBucket == nil {
			// No tips found for this Uid
			return nil
		}

		// Iterate through each tip in the Uid's sub-bucket
		return userBucket.ForEach(func(tipID, v []byte) error {
			// Decode the JSON payload into a ReceivedTipWrapper struct
			tip := &ReceivedTipWrapper{}
			if err := json.Unmarshal(v, tip); err != nil {
				return err
			}

			// Add the decoded tip to the allTips slice
			allTips = append(allTips, *tip)
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	return allTips, nil
}

// FetchUnprocessedTips retrieves all unprocessed tips for all users.
func (b *boltDB) FetchUnprocessedTips(ctx context.Context) (map[zkidentity.ShortID][]*types.ReceivedTip, error) {
	unprocessedTips := make(map[zkidentity.ShortID][]*types.ReceivedTip)

	err := b.db.View(func(tx *bolt.Tx) error {
		mainBucket := tx.Bucket(receivedTipsBucket)
		if mainBucket == nil {
			return ErrMainBucketNotFound
		}

		// Iterate over each UID's sub-bucket within the main bucket
		return mainBucket.ForEach(func(uid, _ []byte) error {
			userBucket := mainBucket.Bucket(uid)
			if userBucket == nil {
				// Skip if there's no bucket for this UID
				return nil
			}

			// Convert uid bytes to ShortID
			var userID zkidentity.ShortID
			copy(userID[:], uid)

			// Iterate through each tip in the user's sub-bucket
			return userBucket.ForEach(func(_, v []byte) error {
				// Decode the JSON-encoded `ReceivedTipWrapper` object
				var wrapper ReceivedTipWrapper
				if err := json.Unmarshal(v, &wrapper); err != nil {
					return err
				}

				// Only append tips with StatusUnprocessed
				if wrapper.Status == StatusUnpaid {
					unprocessedTips[userID] = append(unprocessedTips[userID], wrapper.Tip)
				}
				return nil
			})
		})
	})
	if err != nil {
		return nil, err
	}

	return unprocessedTips, nil
}

func (b *boltDB) FetchTip(ctx context.Context, tipID uint64) (*ReceivedTipWrapper, error) {
	var tip *ReceivedTipWrapper
	var found bool

	err := b.db.View(func(tx *bolt.Tx) error {
		mainBucket := tx.Bucket(receivedTipsBucket)
		if mainBucket == nil {
			return ErrMainBucketNotFound
		}

		// Iterate through all user buckets
		return mainBucket.ForEach(func(uid, _ []byte) error {
			if found { // Skip remaining iterations once found
				return nil
			}

			userBucket := mainBucket.Bucket(uid)
			if userBucket == nil {
				return nil
			}

			tipIDBytes := make([]byte, 8)
			binary.BigEndian.PutUint64(tipIDBytes, tipID)
			data := userBucket.Get(tipIDBytes)
			if data != nil {
				var wrapper *ReceivedTipWrapper
				if err := json.Unmarshal(data, &wrapper); err != nil {
					return err
				}
				tip = wrapper
				found = true
			}
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, nil
	}
	return tip, nil
}

func (b *boltDB) StoreSendTipProgress(ctx context.Context, winnerUID []byte, totalAmount int64, tips []*types.ReceivedTip, status TipStatus) error {
	err := b.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(sendTipProgressBucket)
		if bucket == nil {
			return ErrTipBucketNotFound
		}

		record := TipProgressRecord{
			WinnerUID:   winnerUID,
			TotalAmount: totalAmount,
			Tips:        tips,
			CreatedAt:   time.Now(),
			Status:      status,
		}

		// Generate sequence ID
		id, _ := bucket.NextSequence()
		record.ID = id

		data, err := json.Marshal(record)
		if err != nil {
			return err
		}

		return bucket.Put(itob(id), data)
	})
	return err
}

func (b *boltDB) FetchLatestUncompletedTipProgress(ctx context.Context, winnerUID []byte, totalAmount int64) (*TipProgressRecord, error) {
	var latestRecord *TipProgressRecord

	err := b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(sendTipProgressBucket)
		if bucket == nil {
			return ErrTipBucketNotFound
		}

		return bucket.ForEach(func(k, v []byte) error {
			var record TipProgressRecord
			if err := json.Unmarshal(v, &record); err != nil {
				return err
			}

			// Check if the record matches the given winnerUID and totalAmount and is uncompleted.
			if bytes.Equal(record.WinnerUID, winnerUID) && record.TotalAmount == totalAmount && record.Status != StatusPaid {
				// Update latestRecord if this record is newer.
				if latestRecord == nil || record.CreatedAt.After(latestRecord.CreatedAt) {
					tmp := record // create a copy to get its address
					latestRecord = &tmp
				}
			}
			return nil
		})
	})

	if err != nil {
		return nil, err
	}

	if latestRecord == nil {
		return nil, ErrTipNotFound
	}

	return latestRecord, nil
}

func (b *boltDB) FetchSendTipProgressByClient(ctx context.Context, clientID []byte) ([]*TipProgressRecord, error) {
	var results []*TipProgressRecord

	err := b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(sendTipProgressBucket)
		if bucket == nil {
			return ErrTipBucketNotFound
		}

		return bucket.ForEach(func(_, v []byte) error {
			var record TipProgressRecord
			if err := json.Unmarshal(v, &record); err != nil {
				return err
			}

			if bytes.Equal(record.WinnerUID, clientID) {
				// Clone the record to avoid referencing loop variable
				recordCopy := record
				results = append(results, &recordCopy)
			}
			return nil
		})
	})

	if err != nil {
		return nil, err
	}

	return results, nil
}

func (b *boltDB) UpdateTipProgressStatus(ctx context.Context, recordID uint64, status TipStatus) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(sendTipProgressBucket)
		if bucket == nil {
			return ErrTipBucketNotFound
		}

		key := itob(recordID)
		data := bucket.Get(key)
		if data == nil {
			return ErrTipNotFound
		}

		var record TipProgressRecord
		if err := json.Unmarshal(data, &record); err != nil {
			return err
		}

		record.Status = status
		updatedData, err := json.Marshal(record)
		if err != nil {
			return err
		}

		return bucket.Put(key, updatedData)
	})
}
