package client

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/decred/dcrd/chaincfg/v3"
	"github.com/vctt94/bisonbotkit/logging"
	"github.com/vctt94/bisonbotkit/utils"
)

// PongClientCfg is the pong config used on the pong client
type PongClientCfg struct {
	PongConf   *PongConf           // Consolidated app config (single source of truth)
	LogBackend *logging.LogBackend // Application's logger

	// Notifications tracks handlers for client events. If nil, the client
	// will initialize a new notification manager.
	Notifications *NotificationManager
}

// PongConfig is the config loaded from our .conf
type PongConf struct {
	// Absolute directory where the config/logs live.
	DataDir string

	// Extracted Pong gRPC settings (also persisted in BR.ExtraConfig).
	ServerAddr   string
	GRPCCertPath string

	// Network specifies the Decred network: "mainnet" or "testnet" (defaults to "testnet")
	Network string

	LogFile         string
	Debug           string
	MaxLogFiles     int
	MaxBufferLines  int
	ShowPerfOverlay bool
}

// parseClientConfigFile parses the config file at the given path into a ClientConfig struct.
func parseClientConfigFile(configPath string) (*PongConf, error) {
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	fmt.Println("parseClientConfigFile: ", configPath)
	cfg := &PongConf{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "datadir":
			cfg.DataDir = value
		case "serveraddress":
			cfg.ServerAddr = value
		case "grpccertpath":
			cfg.GRPCCertPath = value
		case "network":
			cfg.Network = value
		case "logfile":
			cfg.LogFile = value
			if cfg.LogFile == "" {
				cfg.LogFile = filepath.Join(cfg.DataDir, "logs", "pongclient.log")
			}
		case "debug":
			cfg.Debug = value
			if cfg.Debug == "" {
				cfg.Debug = "info"
			}
		case "maxlogfiles":
			fmt.Sscanf(value, "%d", &cfg.MaxLogFiles)
			if cfg.MaxLogFiles == 0 {
				cfg.MaxLogFiles = 5
			}
		case "maxbufferlines":
			fmt.Sscanf(value, "%d", &cfg.MaxBufferLines)
			if cfg.MaxBufferLines == 0 {
				cfg.MaxBufferLines = 1000
			}
		case "showperfoverlay":
			var v bool
			if value == "true" {
				v = true
			} else {
				v = false
			}
			cfg.ShowPerfOverlay = v
		default:
			return nil, fmt.Errorf("unknown key: %s", key)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	var missing []string
	// Check all required fields after parsing (in case keys were missing entirely)
	if strings.TrimSpace(cfg.DataDir) == "" {
		missing = append(missing, "datadir")
	}
	if strings.TrimSpace(cfg.ServerAddr) == "" {
		missing = append(missing, "serveraddress")
	}
	if strings.TrimSpace(cfg.GRPCCertPath) == "" {
		missing = append(missing, "grpccertpath")
	}
	if strings.TrimSpace(cfg.Network) == "" {
		missing = append(missing, "network")
	}

	if len(missing) > 0 {
		return nil, fmt.Errorf("missing required fields in client config: %s", strings.Join(missing, ", "))
	}

	return cfg, nil
}

// LoadClientConfig attempts to load the client config (.conf) from the default locations.
func loadClientConf(configPath string, fileName string) (*PongConf, error) {
	// Check if fileName has .conf extension
	if !strings.HasSuffix(fileName, ".conf") {
		return nil, fmt.Errorf("filename must have .conf extension, got: %s", fileName)
	}

	// Get app name by removing .conf extension
	appName := strings.TrimSuffix(fileName, ".conf")

	defaultConfigPath := utils.AppDataDir(fileName, false)
	// If configPath is empty, use defaultConfigPath
	if configPath == "" {
		configPath = defaultConfigPath
	}

	// Ensure the config directory exists
	if err := os.MkdirAll(configPath, 0700); err != nil {
		return nil, err
	}

	// Try to load existing config
	fullPath := filepath.Join(configPath, fileName)
	if _, err := os.Stat(fullPath); err == nil {
		return parseClientConfigFile(fullPath)
	}

	// Create default config
	cfg := &PongConf{
		DataDir:        configPath,
		GRPCCertPath:   filepath.Join(configPath, "grpc.cert"),
		ServerAddr:     "178.156.178.191:50051",
		Network:        "mainnet",
		LogFile:        filepath.Join(configPath, "logs", appName+".log"),
		Debug:          "info",
		MaxLogFiles:    5,
		MaxBufferLines: 1000,
	}

	// Write default config
	if err := writeClientConfigFile(cfg, fullPath); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Write the configuration to a file.
func writeClientConfigFile(cfg *PongConf, configPath string) error {
	configData := fmt.Sprintf(
		`datadir=%s
serveraddr=%s
grpccertpath=%s
network=%s
logfile=%s
debug=%s
maxlogfiles=%d
maxbufferlines=%d
`,
		cfg.DataDir,
		cfg.ServerAddr,
		cfg.GRPCCertPath,
		cfg.Network,
		cfg.LogFile,
		cfg.Debug,
		cfg.MaxLogFiles,
		cfg.MaxBufferLines,
	)

	return os.WriteFile(configPath, []byte(configData), 0600)
}

// LoadAppConfig loads pongclient configuration from disk, applies overrides,
// and returns a consolidated AppConfig. If datadir is empty, it uses the
// default application data dir for "pongclient".
func LoadAppConfig(datadir string, appName string) (*PongClientCfg, error) {
	if datadir == "" {
		datadir = utils.AppDataDir(appName, false)
	}

	cfg, err := loadClientConf(datadir, appName)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	logBackend, err := logging.NewLogBackend(logging.LogConfig{
		LogFile:        filepath.Join(cfg.DataDir, "logs", "pongclient.log"),
		DebugLevel:     cfg.Debug,
		MaxLogFiles:    5,
		MaxBufferLines: 1000,
	})
	if err != nil {
		return nil, fmt.Errorf("NewLogBackend failed: %w", err)
	}
	return &PongClientCfg{
		PongConf:      cfg,
		LogBackend:    logBackend,
		Notifications: NewNotificationManager(),
	}, nil
}

// GetChainParams returns the chaincfg.Params for the configured network.
// Returns an error if the network is invalid.
func (cfg *PongConf) GetChainParams() (*chaincfg.Params, error) {
	network := strings.ToLower(strings.TrimSpace(cfg.Network))
	if network == "" {
		network = "mainnet"
	}
	switch network {
	case "mainnet":
		return chaincfg.MainNetParams(), nil
	case "testnet":
		return chaincfg.TestNet3Params(), nil
	case "simnet":
		return chaincfg.SimNetParams(), nil
	case "regnet":
		return chaincfg.RegNetParams(), nil
	default:
		return nil, fmt.Errorf("invalid network: %s (must be 'mainnet' or 'testnet')", cfg.Network)
	}
}
