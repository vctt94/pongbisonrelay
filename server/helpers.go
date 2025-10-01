package server

import (
	"fmt"

	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/vctt94/pongbisonrelay/ponggame"
	"github.com/vctt94/pongbisonrelay/pongrpc/grpc/pong"
)

// notify sends a simple text notification if the stream is live.
// (Optional: de-dupe by caching the last message per player to avoid spam.)
func (s *Server) notify(p *ponggame.Player, resp *pong.NtfnStreamResponse) error {
	if p == nil {
		return fmt.Errorf("player nil")
	}
	if p.ID == nil {
		return fmt.Errorf("player id nil")
	}
	if p.NotifierStream == nil {
		return fmt.Errorf("player stream nil")
	}
	if resp == nil {
		return fmt.Errorf("nil response")
	}
	// Serialize send per-player to avoid concurrent Stream.Send races.
	if err := p.SendNotif(resp); err != nil {
		s.log.Warnf("notify: failed to deliver to %s: %v", p.ID.String(), err)
		return err
	}
	return nil
}

func (s *Server) notifyallusers(resp *pong.NtfnStreamResponse) {
	// Notify all users.
	s.usersMu.RLock()
	usersSnap := make([]*ponggame.Player, 0, len(s.users))
	for _, u := range s.users {
		usersSnap = append(usersSnap, u)
	}
	s.usersMu.RUnlock()

	for _, user := range usersSnap {
		if user.NotifierStream == nil {
			s.log.Errorf("user %s without NotifierStream", user.ID)
			continue
		}
		_ = s.notify(user, resp)
	}
}

// ensureBoundFunding either binds the canonical funding input if not yet bound
// (requiring exactly one UTXO), or verifies the previously-bound input still
// exists and that no extra deposits were made.
func (s *Server) ensureBoundFunding(es *escrowSession) error {
	es.mu.Lock()
	defer es.mu.Unlock()

	// Must have a watcher snapshot that says "funded".
	if !es.latest.OK || es.latest.UTXOCount == 0 {
		return fmt.Errorf("escrow not yet funded")
	}

	// Normalize current UTXO list.
	var current []*pong.EscrowUTXO
	for _, u := range es.lastUTXOs {
		if u != nil && u.Txid != "" {
			current = append(current, u)
		}
	}
	// If we haven't cached details yet, we can't bind.
	if len(current) == 0 {
		return fmt.Errorf("escrow UTXO details not available yet; wait for index")
	}

	if es.boundInputID == "" {
		// Not yet bound: require an exact-value funding UTXO matching betAtoms.
		var matches []*pong.EscrowUTXO
		for _, u := range current {
			if u.Value == es.betAtoms {
				matches = append(matches, u)
			}
		}
		if len(matches) == 0 {
			return fmt.Errorf("no funding UTXO with exact amount: want %d atoms", es.betAtoms)
		}
		// Fail if there are multiple deposits (of any value) — require exactly one UTXO total.
		if len(current) != 1 || es.latest.UTXOCount != 1 || len(matches) != 1 {
			return fmt.Errorf("unexpected deposits present (total=%d, exact-matches=%d); require a single exact %d atoms deposit", len(current), len(matches), es.betAtoms)
		}
		u := matches[0]
		boundID := fmt.Sprintf("%s:%d", u.Txid, u.Vout)
		es.boundInputID = boundID
		es.boundInput = u
		s.notify(es.player, &pong.NtfnStreamResponse{NotificationType: pong.NotificationType_MESSAGE, Message: fmt.Sprintf("Escrow bound to %s (exact %d atoms).", boundID, es.betAtoms)})
		return nil
	}

	// Already bound: verify the exact input is still present.
	var found *pong.EscrowUTXO
	for _, u := range current {
		if fmt.Sprintf("%s:%d", u.Txid, u.Vout) == es.boundInputID {
			found = u
			break
		}
	}
	if found == nil {
		// The canonical input disappeared (reorg/spent?) -> invalidate.
		return fmt.Errorf("bound funding UTXO %s not present", es.boundInputID)
	}
	es.boundInput = found
	// Enforce: only the bound input must exist and must match the exact amount.
	if len(current) != 1 || es.latest.UTXOCount != 1 {
		return fmt.Errorf("unexpected additional deposits (%d); only the bound %s with %d atoms is allowed", len(current), es.boundInputID, es.betAtoms)
	}
	if es.boundInput != nil && es.boundInput.Value != es.betAtoms {
		return fmt.Errorf("bound funding amount mismatch: have %d want %d", es.boundInput.Value, es.betAtoms)
	}

	return nil
}

// escrowForRoomPlayer returns the escrow session bound to (wrID, ownerUID)
// via the roomEscrows mapping. It does not attempt to pick a "newest" escrow.
func (s *Server) escrowForRoomPlayer(ownerUID zkidentity.ShortID, wrID string) *escrowSession {
	s.roomEscrowsMu.RLock()
	defer s.roomEscrowsMu.RUnlock()
	m := s.roomEscrows[ownerUID]
	if m == nil {
		return nil
	}
	escrowID := m[wrID]
	if escrowID == "" {
		return nil
	}
	return s.escrows[escrowID]
}

// clearPreSigns removes all presign artifacts from the escrow session.
// This should be called when a player leaves to prevent memory leaks and stale data.
func (es *escrowSession) clearPreSigns() {
	es.mu.Lock()
	defer es.mu.Unlock()

	if es.preSign != nil {
		es.preSign = make(map[string]*PreSignCtx)
	}
}
