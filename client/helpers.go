package client

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// sessionKeyFilePath returns the path used to persist the settlement session key.
func (pc *PongClient) sessionKeyFilePath() string {
	if pc == nil || pc.appCfg == nil || strings.TrimSpace(pc.appCfg.DataDir) == "" {
		return ""
	}
	return filepath.Join(pc.appCfg.DataDir, "settlement_session_key.json")
}

// saveSettlementSessionKey writes the current session keypair to disk (0600) in JSON.
func (pc *PongClient) saveSettlementSessionKey() error {
	path := pc.sessionKeyFilePath()
	if strings.TrimSpace(path) == "" {
		return nil // no datadir configured; skip persistence in POC mode
	}
	type pair struct {
		Priv string `json:"priv"`
		Pub  string `json:"pub"`
	}
	pc.RLock()
	data, err := json.MarshalIndent(pair{Priv: pc.settlePrivHex, Pub: pc.settlePubHex}, "", "  ")
	pc.RUnlock()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

// loadSettlementSessionKey loads a previously saved session keypair from disk.
// It returns (true, nil) if a valid key was loaded and cached in memory.
func (pc *PongClient) loadSettlementSessionKey() (bool, error) {
	path := pc.sessionKeyFilePath()
	if strings.TrimSpace(path) == "" {
		return false, nil
	}
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	type pair struct {
		Priv string `json:"priv"`
		Pub  string `json:"pub"`
	}
	var p pair
	if err := json.Unmarshal(b, &p); err != nil {
		return false, err
	}
	p.Priv = strings.TrimSpace(p.Priv)
	p.Pub = strings.TrimSpace(p.Pub)
	if p.Priv == "" || p.Pub == "" {
		return false, fmt.Errorf("empty session key file")
	}
	if _, err := hex.DecodeString(p.Priv); err != nil {
		return false, fmt.Errorf("bad session privkey in file: %w", err)
	}
	if pubB, err := hex.DecodeString(p.Pub); err != nil || len(pubB) != 33 {
		return false, fmt.Errorf("bad session pubkey in file")
	}
	pc.Lock()
	pc.settlePrivHex = p.Priv
	pc.settlePubHex = p.Pub
	pc.Unlock()
	return true, nil
}

// GenerateNewSettlementSessionKey always creates a new session key and overwrites the cached one.
func (pc *PongClient) GenerateNewSettlementSessionKey() (string, string, error) {
	pc.Lock()
	p, err := secp256k1.GeneratePrivateKey()
	if err != nil {
		pc.Unlock()
		return "", "", err
	}
	pc.settlePrivHex = hex.EncodeToString(p.Serialize())
	pc.settlePubHex = hex.EncodeToString(p.PubKey().SerializeCompressed())
	pc.Unlock()
	if err := pc.saveSettlementSessionKey(); err != nil {
		return "", "", fmt.Errorf("save session key: %w", err)
	}
	return pc.settlePrivHex, pc.settlePubHex, nil
}

// RequireSettlementSessionKey returns the current session key or an error if unset.
func (pc *PongClient) RequireSettlementSessionKey() (string, string, error) {
	pc.RLock()
	priv, pub := pc.settlePrivHex, pc.settlePubHex
	pc.RUnlock()
	if priv == "" || pub == "" {
		if ok, err := pc.loadSettlementSessionKey(); err != nil {
			return "", "", err
		} else if ok {
			pc.RLock()
			priv, pub = pc.settlePrivHex, pc.settlePubHex
			pc.RUnlock()
		}
	}
	if priv == "" || pub == "" {
		return "", "", fmt.Errorf("no settlement session key; generate one with [K] in the UI or via GenerateNewSettlementSessionKey()")
	}
	return priv, pub, nil
}
