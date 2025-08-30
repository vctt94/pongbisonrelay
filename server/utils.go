package server

import (
	"context"
	"fmt"
	"log"

	"github.com/decred/slog"
)

func GetDebugLevel(debugStr string) slog.Level {
	// Convert debugStr to slog.Level
	var debugLevel slog.Level
	switch debugStr {
	case "info":
		debugLevel = slog.LevelInfo
	case "warn":
		debugLevel = slog.LevelWarn
	case "error":
		debugLevel = slog.LevelError
	case "debug":
		debugLevel = slog.LevelDebug
	default:
		log.Fatalf("Unknown debug level: %s", debugStr)
	}

	return debugLevel
}

// withMatch runs fn with a live match state, restoring from DB if needed.
func (s *Server) withMatch(ctx context.Context, matchID string, fn func(st *refMatchState) error) error {
	s.RLock()
	st, ok := s.matches[matchID]
	s.RUnlock()
	if ok {
		return fn(st)
	}

	rec, err := s.db.FetchRefMatch(ctx, matchID)
	if err != nil || rec == nil {
		return fmt.Errorf("unknown match id %s", matchID)
	}

	st = &refMatchState{
		MatchID:                rec.MatchID,
		AComp:                  rec.AComp,
		BComp:                  rec.BComp,
		CSV:                    rec.CSV,
		XA:                     rec.XA,
		XB:                     rec.XB,
		PreSigsA:               make(map[string]refInputPreSig),
		PreSigsB:               make(map[string]refInputPreSig),
		DepositPkScriptHex:     rec.DepositPkScriptHex,
		DepositRedeemScriptHex: rec.DepositRedeemScriptHex,
		RequiredAtoms:          rec.RequiredAtoms,
	}

	s.Lock()
	s.matches[matchID] = st
	if s.watcher != nil && st.DepositPkScriptHex != "" {
		s.watcher.registerDeposit(st.DepositPkScriptHex, st.DepositRedeemScriptHex, st.MatchID)
	}
	s.Unlock()
	return fn(st)
}
