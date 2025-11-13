package client

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/vctt94/bisonbotkit/logging"
)

// defaultServerCertPEM is written to <datadir>/ca.cert on first run when creating
// a default config, so the UI has a usable TLS cert path out of the box.
const defaultServerCertPEM = `-----BEGIN CERTIFICATE-----
MIIBizCCATKgAwIBAgIQbtFxrgQfuhUSaHsw+tbNoDAKBggqhkjOPQQDAjAmMREw
DwYDVQQKEwhnZW5jZXJ0czERMA8GA1UEAxMIZ2VuY2VydHMwHhcNMjUxMTA4MTU1
MjQzWhcNMzUxMTA3MTU1MjQzWjAmMREwDwYDVQQKEwhnZW5jZXJ0czERMA8GA1UE
AxMIZ2VuY2VydHMwWTATBgcqhkjOPQIBBggqhkjOPQMBBwNCAAQoLsfKo3eU1B1c
+GuDgatRBnI889XhmVet8aIGlew+A4hsUyduD8LfP1k7aZ3bHNIq+4H5LLg3sVj8
hNseJ/cFo0IwQDAOBgNVHQ8BAf8EBAMCAoQwDwYDVR0TAQH/BAUwAwEB/zAdBgNV
HQ4EFgQURzfqDTuTTKzRYgMTW1IZiUhFjRIwCgYIKoZIzj0EAwIDRwAwRAIgfTUP
ufQQaHv0dXYDwWfYgL2ry5vLM7xPy9l2iDxWRDcCIADhyHCj1r+M3p6/5yaJNZxd
TLq8HnLRGlOPhEKOCgit
-----END CERTIFICATE-----`

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

	LogFile         string
	Debug           string
	MaxLogFiles     int
	MaxBufferLines  int
	ShowPerfOverlay bool
}

// parseClientConfigFile parses the config file at the given path into a ClientConfig struct.
func parseClientConfigFile(configPath string, appName string) (*PongConf, error) {
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	cfg := &PongConf{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
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
		case "logfile":
			cfg.LogFile = value
			if cfg.LogFile == "" {
				cfg.LogFile = filepath.Join(cfg.DataDir, "logs", appName+".log")
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
			// Ignore unknown keys to preserve forward-compatibility with older configs.
			continue
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

	// Require explicit configPath; callers must provide the sandboxed dir.
	if strings.TrimSpace(configPath) == "" {
		return nil, fmt.Errorf("configPath is required")
	}

	// Ensure the config directory exists
	if err := os.MkdirAll(configPath, 0700); err != nil {
		return nil, err
	}

	// Try to load existing config
	fullPath := filepath.Join(configPath, fileName)
	if _, err := os.Stat(fullPath); err == nil {
		return parseClientConfigFile(fullPath, appName)
	}

	// Create default config
	cfg := &PongConf{
		DataDir:        configPath,
		GRPCCertPath:   filepath.Join(configPath, "ca.cert"),
		ServerAddr:     "178.156.178.191:50051",
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

// WriteClientConfigFile writes the configuration to a file.
func WriteClientConfigFile(cfg *PongConf, configPath string) error {
	configData := fmt.Sprintf(
		`datadir=%s
serveraddress=%s
grpccertpath=%s
logfile=%s
debug=%s
maxlogfiles=%d
maxbufferlines=%d
showperfoverlay=%t
`,
		cfg.DataDir,
		cfg.ServerAddr,
		cfg.GRPCCertPath,
		cfg.LogFile,
		cfg.Debug,
		cfg.MaxLogFiles,
		cfg.MaxBufferLines,
		cfg.ShowPerfOverlay,
	)

	return os.WriteFile(configPath, []byte(configData), 0600)
}

// writeClientConfigFile is a wrapper for backward compatibility.
func writeClientConfigFile(cfg *PongConf, configPath string) error {
	return WriteClientConfigFile(cfg, configPath)
}

// LoadAppConfig loads pongclient configuration from disk, applies overrides,
// and returns a consolidated AppConfig. If datadir is empty, it uses the
// default application data dir for "pongclient".
func LoadAppConfig(datadir string, appName string) (*PongClientCfg, error) {
	if strings.TrimSpace(datadir) == "" {
		return nil, fmt.Errorf("datadir is required")
	}

	cfg, err := loadClientConf(datadir, appName+".conf")
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	// Use the log file from config, or default to appName.log if not set
	logFile := cfg.LogFile
	if logFile == "" {
		logFile = filepath.Join(cfg.DataDir, "logs", appName+".log")
	}
	// Ensure the logs directory exists before creating the log backend
	logsDir := filepath.Dir(logFile)
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create logs directory %s: %w", logsDir, err)
	}
	logBackend, err := logging.NewLogBackend(logging.LogConfig{
		LogFile:        logFile,
		DebugLevel:     cfg.Debug,
		MaxLogFiles:    cfg.MaxLogFiles,
		MaxBufferLines: cfg.MaxBufferLines,
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
