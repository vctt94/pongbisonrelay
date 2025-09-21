package golib

import (
	"github.com/companyzero/bisonrelay/client"
	"github.com/companyzero/bisonrelay/client/clientintf"
	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/vctt94/pongbisonrelay/pongrpc/grpc/pong"
)

type initClient struct {
	ServerAddr     string `json:"server_addr"`
	GRPCCertPath   string `json:"grpc_cert_path"`
	DBRoot         string `json:"dbroot"`
	DataDir        string `json:"datadir"`
	DownloadsDir   string `json:"downloads_dir"`
	LogFile        string `json:"log_file"`
	DebugLevel     string `json:"debug_level"`
	WantsLogNtfns  bool   `json:"wants_log_ntfns"`
	LogPings       bool   `json:"log_pings"`
	PingIntervalMs int64  `json:"ping_interval_ms"`

	// New fields for RPC configuration
	RPCWebsocketURL   string `json:"rpc_websocket_url"`
	RPCCertPath       string `json:"rpc_cert_path"`
	RPCCLientCertPath string `json:"rpc_client_cert_path"`
	RPCCLientKeyPath  string `json:"rpc_client_key_path"`
	RPCUser           string `json:"rpc_user"`
	RPCPass           string `json:"rpc_pass"`
}

type createWaitingRoom struct {
	ClientID string `json:"client_id"`
	BetAmt   int64  `json:"bet_amt"`
}

type localInfo struct {
	ID   clientintf.UserID `json:"id"`
	Nick string            `json:"nick"`
}

type waitingRoom struct {
	ID      string    `json:"id"`
	BetAmt  int64     `json:"bet_amt"`
	HostID  string    `json:"host_id"`
	Players []*player `json:"players"`
}

type player struct {
	UID    client.UserID `json:"uid"`
	Nick   string        `json:"nick"`
	BetAmt int64         `json:"bet_amt"`
	Ready  bool          `json:"ready"`
}

func playerFromServer(p *pong.Player) (*player, error) {
	var id zkidentity.ShortID
	err := id.FromString(p.Uid)
	if err != nil {
		return nil, err
	}
	return &player{
		UID:    id,
		Nick:   p.Nick,
		BetAmt: p.BetAmt,
		Ready:  p.Ready,
	}, nil
}

const (
	ConnStateOffline = 0
	ConnStateOnline  = 1
)

type remoteUser struct {
	UID  string `json:"uid"`
	Nick string `json:"nick"`
}

func remoteUserFromPII(pii *zkidentity.PublicIdentity) remoteUser {
	return remoteUser{
		UID:  pii.Identity.String(),
		Nick: pii.Nick,
	}
}

func remoteUserFromRU(ru *client.RemoteUser) remoteUser {
	if ru == nil {
		return remoteUser{}
	}
	return remoteUser{
		UID:  ru.ID().String(),
		Nick: ru.Nick(),
	}
}

type runState struct {
	DcrlndRunning bool `json:"dcrlnd_running"`
	ClientRunning bool `json:"client_running"`
}
