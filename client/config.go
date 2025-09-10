package client

import (
	"fmt"

	"github.com/companyzero/bisonrelay/clientrpc/types"
	"github.com/decred/slog"
	brconfig "github.com/vctt94/bisonbotkit/config"
	"github.com/vctt94/bisonbotkit/utils"
)

type PongClientCfg struct {
	AppCfg        *AppConfig  // Consolidated app config (single source of truth)
	Log           slog.Logger // Application's logger
	ChatClient    types.ChatServiceClient
	PaymentClient types.PaymentsServiceClient

	// Notifications tracks handlers for client events. If nil, the client
	// will initialize a new notification manager.
	Notifications *NotificationManager
}

// ConfigOverrides carries optional CLI/runtime overrides for config values.
type ConfigOverrides struct {
	RPCURL          string
	BRClientCert    string
	BRClientRPCCert string
	BRClientRPCKey  string
	RPCUser         string
	RPCPass         string

	// Pong-specific (stored under ExtraConfig in the .conf)
	ServerAddr     string
	GRPCServerCert string
	Address        string
}

// AppConfig is the consolidated configuration used by the pong client app.
type AppConfig struct {
	// Absolute directory where the config/logs live.
	DataDir string
	// BR holds the loaded bisonbotkit client configuration.
	BR *brconfig.ClientConfig
	// Extracted Pong gRPC settings (also persisted in BR.ExtraConfig).
	ServerAddr   string
	GRPCCertPath string
	Address      string
}

// LoadAppConfig loads pongclient configuration from disk, applies overrides,
// and returns a consolidated AppConfig. If datadir is empty, it uses the
// default application data dir for "pongclient".
func LoadAppConfig(datadir string, ov ConfigOverrides) (*AppConfig, error) {
	if datadir == "" {
		datadir = utils.AppDataDir("pongclient", false)
	}

	cfg, err := brconfig.LoadClientConfig(datadir, "pongclient.conf")
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	// Apply BR RPC/TLS overrides
	if ov.RPCURL != "" {
		cfg.RPCURL = ov.RPCURL
	}
	if ov.BRClientCert != "" {
		cfg.BRClientCert = ov.BRClientCert
	}
	if ov.BRClientRPCCert != "" {
		cfg.BRClientRPCCert = ov.BRClientRPCCert
	}
	if ov.BRClientRPCKey != "" {
		cfg.BRClientRPCKey = ov.BRClientRPCKey
	}
	if ov.RPCUser != "" {
		cfg.RPCUser = ov.RPCUser
	}
	if ov.RPCPass != "" {
		cfg.RPCPass = ov.RPCPass
	}

	// Pong gRPC settings live in ExtraConfig; let overrides win but persist in cfg
	srvAddr := cfg.GetString("serveraddr")
	if ov.ServerAddr != "" {
		srvAddr = ov.ServerAddr
		cfg.SetString("serveraddr", srvAddr)
	}
	grpcCert := cfg.GetString("grpcservercert")
	if ov.GRPCServerCert != "" {
		grpcCert = ov.GRPCServerCert
		cfg.SetString("grpcservercert", grpcCert)
	}

	// Payout address
	addr := cfg.GetString("address")
	if ov.Address != "" {
		addr = ov.Address
		cfg.SetString("address", addr)
	}

	return &AppConfig{
		DataDir:      datadir,
		BR:           cfg,
		ServerAddr:   srvAddr,
		GRPCCertPath: grpcCert,
		Address:      addr,
	}, nil
}
