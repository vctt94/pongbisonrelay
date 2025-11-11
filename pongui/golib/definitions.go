package golib

import (
	"github.com/companyzero/bisonrelay/client/clientintf"
	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/vctt94/pongbisonrelay/pongrpc/grpc/pong"
)

type initClient struct {
	ClientID       string `json:"client_id"` // Wallet-authenticated clientID (required)
	ServerAddr     string `json:"server_addr"`
	GRPCCertPath   string `json:"grpc_cert_path"`
	DBRoot         string `json:"dbroot"`
	DataDir        string `json:"datadir"`
	DownloadsDir   string `json:"downloads_dir"`
	LogFile        string `json:"log_file"`
	DebugLevel     string `json:"debug_level"`
	LogPings       bool   `json:"log_pings"`
	PingIntervalMs int64  `json:"ping_interval_ms"`
}

type createWaitingRoom struct {
	ClientID string `json:"client_id"`
	BetAmt   int64  `json:"bet_amt"`
	EscrowId string `json:"escrow_id"`
}

type localInfo struct {
	ID            clientintf.UserID `json:"id"`
	Nick          string            `json:"nick"`
	ServerVersion string            `json:"server_version,omitempty"`
	ServerIsF2P   bool              `json:"server_is_f2p,omitempty"`
}

// Settlement/escrow payloads
type openEscrowReq struct {
	// Payout may be a 33/65-byte pubkey hex or a Decred pubkey address (P2PK).
	Payout    string `json:"payout"`
	BetAtoms  uint64 `json:"bet_atoms"`
	CSVBlocks uint32 `json:"csv_blocks"`
}

type preSignReq struct {
	MatchID string `json:"match_id"` // "<wrID>|<hostId>"
}

type refundEscrowReq struct {
	EscrowID  string `json:"escrow_id"`
	DestAddr  string `json:"dest_addr"`
	FeeAtoms  uint64 `json:"fee_atoms"`
	CSVBlocks uint32 `json:"csv_blocks"`
}

type refundEscrowRes struct {
	RefundTxHex string `json:"refund_tx_hex"`
	UTXOTxid    string `json:"utxo_txid"`
	UTXOVout    uint32 `json:"utxo_vout"`
	UTXOValue   uint64 `json:"utxo_value"`
	RedeemHex   string `json:"redeem_hex"`
	CSVBlocks   uint32 `json:"csv_blocks"`
	CanRefund   bool   `json:"can_refund"`
	Reason      string `json:"reason,omitempty"`
}

type deleteHistoricEscrowReq struct {
	EscrowID string `json:"escrow_id"`
}

type joinWaitingRoom struct {
	RoomID   string `json:"room_id"`
	EscrowId string `json:"escrow_id"`
}

type waitingRoom struct {
	ID      string    `json:"id"`
	BetAmt  int64     `json:"bet_amt"`
	HostID  string    `json:"host_id"`
	Players []*player `json:"players"`
}

type player struct {
	UID    string `json:"uid"`
	Nick   string `json:"nick"`
	BetAmt int64  `json:"bet_amt"`
	Ready  bool   `json:"ready"`
}

func playerFromServer(p *pong.Player) (*player, error) {
	var id zkidentity.ShortID
	err := id.FromString(p.Uid)
	if err != nil {
		return nil, err
	}
	return &player{
		UID:    id.String(),
		Nick:   p.Nick,
		BetAmt: p.BetAmt,
		Ready:  p.Ready,
	}, nil
}


type runState struct {
	DcrlndRunning bool `json:"dcrlnd_running"`
	ClientRunning bool `json:"client_running"`
}

// Wallet-auth request payloads for golib (used before InitClient).
type requestNonceArgs struct {
	ServerAddr   string `json:"server_addr"`
	GRPCCertPath string `json:"grpc_cert_path"`
}

type verifyLoginArgs struct {
	ServerAddr   string `json:"server_addr"`
	GRPCCertPath string `json:"grpc_cert_path"`
	Address      string `json:"address"`
	Nonce        string `json:"nonce"`
	Signature    string `json:"signature"`
}
