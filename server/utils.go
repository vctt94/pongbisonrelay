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
		matchID:                rec.MatchID,
		aComp:                  rec.AComp,
		bComp:                  rec.BComp,
		xa:                     rec.XA,
		xb:                     rec.XB,
		csv:                    rec.CSV,
		preSigA:                make(map[string]refInputPreSig),
		preSigB:                make(map[string]refInputPreSig),
		depositPkScriptHex:     rec.DepositPkScriptHex,
		depositRedeemScriptHex: rec.DepositRedeemScriptHex,
		reqAtoms:               rec.RequiredAtoms,
	}

	s.Lock()
	s.matches[matchID] = st
	if s.watcher != nil && st.depositPkScriptHex != "" {
		s.watcher.registerDeposit(st.depositPkScriptHex, st.depositRedeemScriptHex, st.matchID)
	}
	s.Unlock()
	return fn(st)
}
