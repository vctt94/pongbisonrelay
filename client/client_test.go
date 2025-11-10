package client

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func writeSessionFile(t *testing.T, dir, name string, data SessionKeyData) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatalf("mkdir historic dir: %v", err)
	}
	blob, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		t.Fatalf("marshal session data: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, name), blob, 0o600); err != nil {
		t.Fatalf("write session file: %v", err)
	}
}

func TestGetSettlementPrivKeyAndDetails(t *testing.T) {
	dataDir := t.TempDir()
	pc := &PongClient{appCfg: &AppConfig{DataDir: dataDir}}
	historicDir := filepath.Join(dataDir, "historic_sessions")
	session := SessionKeyData{
		Priv: "deadbeef",
		Pub:  "cafebabe",
		EscrowInfo: &EscrowInfo{
			EscrowID:        "escrow-123",
			FundingTxid:     "abcd1234",
			FundingVout:     2,
			FundedAmount:    50,
			RedeemScriptHex: "51",
			PKScriptHex:     "ac",
			CSVBlocks:       144,
		},
	}
	writeSessionFile(t, historicDir, "sessionkey_test.json", session)

	priv, err := pc.GetSettlementPrivKeyForEscrow("escrow-123")
	if err != nil {
		t.Fatalf("expected privkey, got error: %v", err)
	}
	if priv != "deadbeef" {
		t.Fatalf("unexpected privkey %s", priv)
	}

	details, err := pc.GetEscrowDetails("escrow-123")
	if err != nil {
		t.Fatalf("expected escrow details, got error: %v", err)
	}
	if details.FundingTxHash != "abcd1234" || details.FundingOutpoint != "abcd1234:2" {
		t.Fatalf("unexpected funding info: %+v", details)
	}
	if details.CSVBlocks != 144 || details.FundedAmount != 50 {
		t.Fatalf("unexpected numeric fields: %+v", details)
	}

	if _, err := pc.GetSettlementPrivKeyForEscrow("unknown"); err == nil {
		t.Fatalf("expected error for unknown escrow")
	}
}

func TestCacheEscrowInfoPersists(t *testing.T) {
	dataDir := t.TempDir()
	pc := &PongClient{appCfg: &AppConfig{DataDir: dataDir}}
	pc.settlePrivHex = "aa"
	pc.settlePubHex = "bb"
	if err := pc.CacheEscrowInfo(&EscrowInfo{EscrowID: "esc-1", PKScriptHex: "aa", CSVBlocks: 32}); err != nil {
		t.Fatalf("cache escrow info failed: %v", err)
	}
	loaded, err := os.ReadFile(pc.sessionKeyFilePath())
	if err != nil {
		t.Fatalf("read session key: %v", err)
	}
	if !bytes.Contains(loaded, []byte("esc-1")) {
		t.Fatalf("expected escrow info in session key file: %s", string(loaded))
	}
	if err := pc.CacheEscrowInfo(&EscrowInfo{
		EscrowID:        "esc-1",
		FundingTxid:     "abcd",
		FundingVout:     2,
		FundedAmount:    123,
		RedeemScriptHex: "dead",
	}); err != nil {
		t.Fatalf("cache escrow info update failed: %v", err)
	}
	if pc.activeEscrowInfo == nil || pc.activeEscrowInfo.FundingTxid != "abcd" {
		t.Fatalf("expected cached info in memory: %+v", pc.activeEscrowInfo)
	}
}
