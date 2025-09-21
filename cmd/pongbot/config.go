package main

import (
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/vctt94/bisonbotkit/config"
)

type PongBotConfig struct {
	*config.BotConfig // Embed the base BotConfig

	// Additional pong-specific fields
	IsF2P     bool
	MinBetAmt float64
	GRPCHost  string
	GRPCPort  string

	// dcrd connectivity (optional)
	DcrdHost string
	DcrdCert string
	DcrdUser string
	DcrdPass string

	// Schnorr adaptor secret (32-byte hex string)
	AdaptorSecret string
}

// Load config function
func LoadPongBotConfig(dataDir, configFile string) (*PongBotConfig, error) {
	// First load the base bot config
	baseConfig, err := config.LoadBotConfig(dataDir, configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load base config: %w", err)
	}

	minBetAmt, err := strconv.ParseFloat(baseConfig.ExtraConfig["minbetamt"], 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse minbetamt: %w", err)
	}
	// Create the combined config
	cfg := &PongBotConfig{
		BotConfig:     baseConfig,
		IsF2P:         false,
		MinBetAmt:     minBetAmt,
		GRPCHost:      baseConfig.ExtraConfig["grpchost"],
		GRPCPort:      baseConfig.ExtraConfig["grpcport"],
		DcrdHost:      baseConfig.ExtraConfig["dcrdhost"],
		DcrdCert:      baseConfig.ExtraConfig["dcrdcert"],
		DcrdUser:      baseConfig.ExtraConfig["dcrduser"],
		DcrdPass:      baseConfig.ExtraConfig["dcrdpass"],
		AdaptorSecret: baseConfig.ExtraConfig["adaptorsecret"],
	}

	// Validate adaptor secret: must be present and 32 bytes of hex (64 chars)
	if cfg.AdaptorSecret == "" {
		return nil, fmt.Errorf("missing adaptorsecret in %s", configFile)
	}
	sb, err := hex.DecodeString(cfg.AdaptorSecret)
	if err != nil || len(sb) != 32 {
		return nil, fmt.Errorf("invalid adaptorsecret: expected 64 hex chars (32 bytes)")
	}

	// Load the config file if it exists
	configPath := filepath.Join(dataDir, configFile)
	if _, err := os.Stat(configPath); err == nil {
		_, err := os.ReadFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}

	}

	return cfg, nil
}
