package serverdb

import (
	"context"
	"errors"
	"time"

	"github.com/companyzero/bisonrelay/clientrpc/types"
	"github.com/companyzero/bisonrelay/zkidentity"
)

var (
	ErrDuplicateEntry     = errors.New("tip already stored")
	ErrMainBucketNotFound = errors.New("main bucket not found")
	ErrUserBucketNotFound = errors.New("user bucket not found")
	ErrTipNotFound        = errors.New("tip not found")
	ErrTipBucketNotFound  = errors.New("tip bucket not found")
)

type TipStatus string

const (
	StatusUnpaid  TipStatus = "unpaid"
	StatusSending TipStatus = "sending"
	StatusPaid    TipStatus = "paid"
)

type ReceivedTipWrapper struct {
	Tip    *types.ReceivedTip
	Status TipStatus
}

type TipProgressRecord struct {
	ID          uint64               `json:"id"`
	WinnerUID   []byte               `json:"winner_uid"`
	TotalAmount int64                `json:"total_amount"`
	Status      TipStatus            `json:"status"`
	Tips        []*types.ReceivedTip `json:"received_tip"`
	CreatedAt   time.Time            `json:"created_at"`
}

// RefMatchRecord stores minimal state to restore a referee match after restart.
type RefMatchRecord struct {
	MatchID string `json:"match_id"`
	AComp   []byte `json:"a_comp"`
	BComp   []byte `json:"b_comp"`
	XA      []byte `json:"x_a"`
	XB      []byte `json:"x_b"`
	CSV     uint32 `json:"csv"`
	CoeffA  string `json:"coef_a"`
	CoeffB  string `json:"coef_b"`
	CoeffS  string `json:"coef_s"`

	DepositPkScriptHex     string `json:"pk_script_hex"`
	DepositRedeemScriptHex string `json:"redeem_script_hex"`
	RequiredAtoms          uint64 `json:"required_atoms"`
}

type ServerDB interface {
	StoreUnprocessedTip(ctx context.Context, tip *types.ReceivedTip) error
	FetchUnprocessedTips(ctx context.Context) (map[zkidentity.ShortID][]*types.ReceivedTip, error)
	FetchTip(ctx context.Context, tipID uint64) (*ReceivedTipWrapper, error)
	FetchReceivedTipsByUID(ctx context.Context, uid zkidentity.ShortID, status TipStatus) ([]*types.ReceivedTip, error)
	UpdateTipStatus(ctx context.Context, uid []byte, tipID []byte, status TipStatus) error
	FetchAllReceivedTipsByUID(ctx context.Context, uid zkidentity.ShortID) ([]ReceivedTipWrapper, error)

	StoreSendTipProgress(ctx context.Context, winnerUID []byte, totalAmount int64, tips []*types.ReceivedTip, status TipStatus) error
	FetchLatestUncompletedTipProgress(ctx context.Context, winnerUID []byte, totalAmount int64) (*TipProgressRecord, error)
	FetchSendTipProgressByClient(ctx context.Context, clientID []byte) ([]*TipProgressRecord, error)
	UpdateTipProgressStatus(ctx context.Context, recordID uint64, status TipStatus) error
	Close() error

	// --- Referee persistence ---
	SaveRefMatch(ctx context.Context, rec *RefMatchRecord) error
	FetchRefMatch(ctx context.Context, matchID string) (*RefMatchRecord, error)
	SaveRefAlloc(ctx context.Context, playerID string, matchID string) error
	FetchRefAlloc(ctx context.Context, playerID string) (string, error)
}
