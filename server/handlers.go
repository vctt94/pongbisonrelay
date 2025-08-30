package server

import (
	"bytes"
	"context"
	crand "crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/companyzero/bisonrelay/clientrpc/types"
	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/decred/dcrd/chaincfg/chainhash"
	"github.com/decred/dcrd/chaincfg/v3"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/decred/dcrd/txscript/v4"

	"github.com/decred/dcrd/wire"
	"github.com/vctt94/pong-bisonrelay/ponggame"
	"github.com/vctt94/pong-bisonrelay/pongrpc/grpc/pong"
	"github.com/vctt94/pong-bisonrelay/server/serverdb"
)

// --- Referee service state and handlers ---

type refEscrowUTXO struct {
	TxID            string
	Vout            uint32
	Value           uint64
	RedeemScriptHex string
	PkScriptHex     string
	Owner           string // "A" or "B"
}

type refInputPreSig struct {
	InputID string
	RX      string
	SPrime  string
}

type refAdaptor struct {
	InputID string
	Gamma   string // 32B hex
	T       string // 33B hex
}

type refMatchState struct {
	mu sync.RWMutex

	// Stable identifiers
	MatchID string

	// Players' compressed pubkeys (hex, 33B)
	AComp string
	BComp string

	CSV      uint32
	XA       string
	XB       string
	Escrows  []refEscrowUTXO
	FeeAtoms uint64

	// Branch-scoped authoring state
	BrDraft map[pong.Branch]string // branch -> draft tx hex

	// Collected presigs per branch (inputID -> presig)
	PreSigsA map[string]refInputPreSig
	PreSigsB map[string]refInputPreSig

	// Legacy single-deposit (kept for compatibility with existing code)
	DepositPkScriptHex     string
	DepositRedeemScriptHex string
	RequiredAtoms          uint64 // per-player deposit amount

	// Per-player deposits (v0-min)
	ADepositPkScriptHex     string
	ADepositRedeemScriptHex string
	BDepositPkScriptHex     string
	BDepositRedeemScriptHex string
}

type branchStore struct {
	pre map[string]refInputPreSig
}

func (st *refMatchState) storeFor(br pong.Branch) branchStore {
	if br == pong.Branch_BRANCH_A {
		return branchStore{st.PreSigsA}
	}
	return branchStore{st.PreSigsB}
}

func (s *Server) handleReturnUnprocessedTips(ctx context.Context, clientID zkidentity.ShortID) error {
	// Tips model removed in POC; nothing to do.
	return nil
}

func (s *Server) handleFetchTotalUnprocessedTips(ctx context.Context, clientID zkidentity.ShortID) (int64, []*types.ReceivedTip, error) {
	// Tips model removed in POC; return zero.
	return 0, nil, nil
}

func (s *Server) handleGameLifecycle(ctx context.Context, players []*ponggame.Player, tips []*types.ReceivedTip) {
	game, err := s.gameManager.StartGame(ctx, players)
	if err != nil {
		s.log.Errorf("Failed to start game: %v", err)
		return
	}

	defer func() {
		// reset player status
		for _, player := range game.Players {
			player.ResetPlayer()
		}
		// remove game from gameManager after it ended
		delete(s.gameManager.Games, game.Id)
		s.log.Debugf("Game %s cleaned up", game.Id)
	}()

	game.Run()

	var wg sync.WaitGroup
	for _, player := range players {
		wg.Add(1)
		go func(player *ponggame.Player) {
			defer wg.Done()
			if player.NotifierStream != nil {
				err := player.NotifierStream.Send(&pong.NtfnStreamResponse{
					NotificationType: pong.NotificationType_GAME_START,
					Message:          "Game started with ID: " + game.Id,
					Started:          true,
					GameId:           game.Id,
				})
				if err != nil {
					s.log.Warnf("Failed to notify player %s: %v", player.ID, err)
				}
			}
			s.sendGameUpdates(ctx, player, game)
		}(player)
	}

	wg.Wait() // Wait for both players' streams to finish

	s.handleGameEnd(ctx, game, players, tips)
}

func (s *Server) handleGameEnd(ctx context.Context, game *ponggame.GameInstance, players []*ponggame.Player, tips []*types.ReceivedTip) {
	winner := game.Winner
	var winnerID string
	if winner != nil {
		winnerID = winner.String()
		s.log.Infof("Game ended. Winner: %s", winnerID)
	} else {
		s.log.Infof("Game ended in a draw.")
	}

	// Notify players of game outcome (no transfers)
	for _, player := range players {
		message := "Game ended in a draw."
		if player.ID == winner {
			message = "Congratulations, you won!"
		} else if winner != nil {
			message = "Sorry, you lost."
		}
		player.NotifierStream.Send(&pong.NtfnStreamResponse{
			NotificationType: pong.NotificationType_GAME_END,
			Message:          message,
			GameId:           game.Id,
		})
		// delete player from gameManager PlayerGameMap
		delete(s.gameManager.PlayerGameMap, *player.ID)
	}
}

func (s *Server) handleWaitingRoomRemoved(wr *pong.WaitingRoom) {
	s.log.Infof("Waiting room %s removed", wr.Id)

	// Notify all users about the waiting room removal
	for _, user := range s.users {
		user.NotifierStream.Send(&pong.NtfnStreamResponse{
			NotificationType: pong.NotificationType_ON_WR_REMOVED,
			Message:          fmt.Sprintf("Waiting room %s was removed", wr.Id),
			RoomId:           wr.Id,
		})
	}
}

func (s *Server) ensureRefereeKey() {
	s.refereeKeyOnce.Do(func() {
		b := make([]byte, 32)
		_, _ = crand.Read(b)
		priv := secp256k1.PrivKeyFromBytes(b)
		s.refereePrivKeyHex = hex.EncodeToString(priv.Serialize())
		s.refereePubCompressed = hex.EncodeToString(priv.PubKey().SerializeCompressed())
	})
}

// buildEscrowRedeemScript builds the redeemScript for an escrow owned by ownerComp (33b compressed pubkey hex).
func buildEscrowRedeemScript(state *refMatchState, ownerCompHex string) ([]byte, error) {
	const STSchnorrSecp256k1 int64 = 2

	xa, err := hex.DecodeString(state.XA)
	if err != nil || len(xa) != 33 {
		return nil, fmt.Errorf("XA must be 33-byte compressed pubkey hex: %w", err)
	}
	xb, err := hex.DecodeString(state.XB)
	if err != nil || len(xb) != 33 {
		return nil, fmt.Errorf("XB must be 33-byte compressed pubkey hex: %w", err)
	}
	owner, err := hex.DecodeString(ownerCompHex)
	if err != nil || len(owner) != 33 {
		return nil, fmt.Errorf("owner must be 33-byte compressed pubkey hex: %w", err)
	}
	if state.CSV <= 0 {
		return nil, fmt.Errorf("CSV must be > 0")
	}

	b := txscript.NewScriptBuilder()
	b.AddOp(txscript.OP_IF).
		AddData(xa).AddInt64(STSchnorrSecp256k1).AddOp(txscript.OP_CHECKSIGALT).
		AddOp(txscript.OP_ELSE).
		AddOp(txscript.OP_IF).
		AddData(xb).AddInt64(STSchnorrSecp256k1).AddOp(txscript.OP_CHECKSIGALT).
		AddOp(txscript.OP_ELSE).
		AddInt64(int64(state.CSV)).AddOp(txscript.OP_CHECKSEQUENCEVERIFY).AddOp(txscript.OP_DROP).
		AddData(owner).AddInt64(STSchnorrSecp256k1).AddOp(txscript.OP_CHECKSIGALT).
		AddOp(txscript.OP_ENDIF).
		AddOp(txscript.OP_ENDIF)

	redeem, err := b.Script()
	if err != nil {
		return nil, fmt.Errorf("build redeem script: %w", err)
	}
	return redeem, nil
}

// Legacy SubmitPreSig RPC removed in favor of SettlementStream.

func (s *Server) RevealAdaptors(ctx context.Context, req *pong.RevealAdaptorsRequest) (*pong.RevealAdaptorsResponse, error) {
	var out *pong.RevealAdaptorsResponse
	err := s.withMatch(ctx, req.MatchId, func(st *refMatchState) error {
		bdHex := ""
		if st.BrDraft != nil {
			bdHex = st.BrDraft[req.Branch]
		}
		if bdHex == "" {
			return fmt.Errorf("no draft for branch")
		}
		txb, err := hex.DecodeString(bdHex)
		if err != nil {
			return fmt.Errorf("bad draft")
		}
		var tx wire.MsgTx
		if err := tx.Deserialize(bytes.NewReader(txb)); err != nil {
			return fmt.Errorf("bad tx")
		}
		bstore := st.storeFor(req.Branch)
		entries := make([]*pong.RevealAdaptorEntry, 0, len(bstore.pre))
		for id := range bstore.pre {
			parts := strings.Split(id, ":")
			if len(parts) != 2 {
				return fmt.Errorf("bad input id: %s", id)
			}
			var h chainhash.Hash
			if err := chainhash.Decode(&h, parts[0]); err != nil {
				return fmt.Errorf("bad txid: %w", err)
			}
			idx := -1
			for i, ti := range tx.TxIn {
				if ti.PreviousOutPoint.Hash == h {
					idx = i
					break
				}
			}
			if idx < 0 {
				return fmt.Errorf("input not in draft: %s", id)
			}
			redeemA, _ := buildEscrowRedeemScript(st, st.AComp)
			redeemB, _ := buildEscrowRedeemScript(st, st.BComp)
			mA, _ := txscript.CalcSignatureHash(redeemA, txscript.SigHashAll, &tx, idx, nil)
			mB, _ := txscript.CalcSignatureHash(redeemB, txscript.SigHashAll, &tx, idx, nil)
			gammaHexA, _ := deriveAdaptorGamma(st.MatchID, id, int32(req.Branch), hex.EncodeToString(mA), s.refereePrivKeyHex)
			gammaHex := gammaHexA
			if gammaHex == "" {
				gammaHex, _ = deriveAdaptorGamma(st.MatchID, id, int32(req.Branch), hex.EncodeToString(mB), s.refereePrivKeyHex)
			}
			entries = append(entries, &pong.RevealAdaptorEntry{InputId: id, Gamma: gammaHex})
		}
		out = &pong.RevealAdaptorsResponse{Branch: req.Branch, Entries: entries}
		return nil
	})
	return out, err
}

// SettlementStream
func (s *Server) SettlementStream(stream pong.PongReferee_SettlementStreamServer) error {
	ctx := stream.Context()
	var matchID string
	for {
		in, err := stream.Recv()
		if err != nil {
			return err
		}
		if in.GetMatchId() != "" {
			matchID = in.GetMatchId()
		}
		if h := in.GetHello(); h != nil {
			var reqAtoms uint64
			var pkHex string
			s.RLock()
			st := s.matches[matchID]
			s.RUnlock()
			if st != nil {
				reqAtoms = st.RequiredAtoms
				pkHex = st.DepositPkScriptHex
			}
			_ = stream.Send(&pong.ServerMsg{MatchId: matchID, Kind: &pong.ServerMsg_Role{Role: &pong.AssignRole{Role: pong.AssignRole_A, RequiredAtoms: reqAtoms, DepositPkscriptHex: pkHex}}})

			aHex, bHex, inputs, err := s.buildBothBranchesForStream(ctx, matchID)
			if err != nil {
				_ = stream.Send(&pong.ServerMsg{MatchId: matchID, Kind: &pong.ServerMsg_Info{Info: &pong.Info{Text: "error building drafts"}}})
				continue
			}
			nps := &pong.NeedPreSigs{BranchesToPresign: []pong.Branch{pong.Branch_BRANCH_A, pong.Branch_BRANCH_B}, DraftAwinsTxHex: aHex, DraftBwinsTxHex: bHex, Inputs: inputs}
			_ = stream.Send(&pong.ServerMsg{MatchId: matchID, Kind: &pong.ServerMsg_Req{Req: nps}})
			continue
		}
		// The client computes R' locally in computePreSig(xPrivHex, mHex, TCompHex string),
		//  extracts r' = x(R'), and sends only (r', s').
		if pb := in.GetPresigs(); pb != nil {
			s.withMatch(ctx, matchID, func(st *refMatchState) error {
				bs := st.storeFor(pb.Branch)
				for _, sig := range pb.Presigs {
					rx := hex.EncodeToString(sig.Rprime32)
					sp := hex.EncodeToString(sig.Sprime32)
					st.mu.Lock()
					bs.pre[sig.InputId] = refInputPreSig{InputID: sig.InputId, RX: rx, SPrime: sp}
					st.mu.Unlock()
				}
				return nil
			})
			_ = stream.Send(&pong.ServerMsg{MatchId: matchID, Kind: &pong.ServerMsg_Info{Info: &pong.Info{Text: "presigs received"}}})
			continue
		}
		if ack := in.GetAck(); ack != nil {
			_ = ack
			continue
		}
	}
}

// buildBothBranchesForStream builds both branch drafts and NeedPreSigs inputs for the stream.
func (s *Server) buildBothBranchesForStream(ctx context.Context, matchID string) (string, string, []*pong.NeedPreSigs_PerInput, error) {
	s.RLock()
	state, ok := s.matches[matchID]
	s.RUnlock()
	if !ok {
		return "", "", nil, fmt.Errorf("unknown match id")
	}
	s.ensureRefereeKey()
	var escrows []refEscrowUTXO
	if state.ADepositPkScriptHex != "" {
		if us, _, ok := s.watcher.queryDeposit(state.ADepositPkScriptHex); ok && len(us) > 0 {
			e := us[0]
			escrows = append(escrows, refEscrowUTXO{TxID: e.Txid, Vout: e.Vout, Value: e.Value, RedeemScriptHex: e.RedeemScriptHex, PkScriptHex: e.PkScriptHex, Owner: "A"})
		}
	}
	if state.BDepositPkScriptHex != "" {
		if us, _, ok := s.watcher.queryDeposit(state.BDepositPkScriptHex); ok && len(us) > 0 {
			e := us[0]
			escrows = append(escrows, refEscrowUTXO{TxID: e.Txid, Vout: e.Vout, Value: e.Value, RedeemScriptHex: e.RedeemScriptHex, PkScriptHex: e.PkScriptHex, Owner: "B"})
		}
	}
	if len(escrows) != 2 {
		return "", "", nil, fmt.Errorf("need 2 deposits")
	}
	build := func(branch pong.Branch, payoutX string) (string, map[string]struct{ m, t string }, error) {
		tx := wire.NewMsgTx()
		totalIn := int64(0)
		for _, e := range escrows {
			var h chainhash.Hash
			_ = chainhash.Decode(&h, e.TxID)
			op := wire.NewOutPoint(&h, e.Vout, 0)
			tx.AddTxIn(wire.NewTxIn(op, int64(e.Value), nil))
			totalIn += int64(e.Value)
		}
		fee := state.FeeAtoms
		if totalIn < int64(fee) {
			return "", nil, fmt.Errorf("fee exceeds total inputs")
		}
		pay := totalIn - int64(fee)
		x, _ := hex.DecodeString(payoutX)
		pkAlt, _ := txscript.NewScriptBuilder().AddData(x).AddInt64(2).AddOp(txscript.OP_CHECKSIGALT).Script()
		tx.AddTxOut(&wire.TxOut{Value: pay, Version: 0, PkScript: pkAlt})
		var buf bytes.Buffer
		_ = tx.Serialize(&buf)
		txHex := hex.EncodeToString(buf.Bytes())
		perInput := make(map[string]struct{ m, t string })
		for idx, e := range escrows {
			ownerComp := state.AComp
			if e.Owner == "B" {
				ownerComp = state.BComp
			}
			redeem, err := buildEscrowRedeemScript(state, ownerComp)
			if err != nil {
				return "", nil, fmt.Errorf("buildEscrowRedeemScript: %w", err)
			}
			m, err := txscript.CalcSignatureHash(redeem, txscript.SigHashAll, tx, idx, nil)
			if err != nil || len(m) != 32 {
				return "", nil, fmt.Errorf("sighash")
			}
			id := fmt.Sprintf("%s:%d", e.TxID, e.Vout)
			_, TCompHex := deriveAdaptorGamma(state.MatchID, id, int32(branch), hex.EncodeToString(m), s.refereePrivKeyHex)
			perInput[id] = struct{ m, t string }{m: hex.EncodeToString(m), t: TCompHex}
		}
		return txHex, perInput, nil
	}
	aHex, aInputs, err := build(pong.Branch_BRANCH_A, state.XA)
	if err != nil {
		return "", "", nil, err
	}
	bHex, bInputs, err := build(pong.Branch_BRANCH_B, state.XB)
	if err != nil {
		return "", "", nil, err
	}
	state.mu.Lock()
	state.Escrows = escrows
	if state.BrDraft == nil {
		state.BrDraft = make(map[pong.Branch]string)
	}
	state.BrDraft[pong.Branch_BRANCH_A] = aHex
	state.BrDraft[pong.Branch_BRANCH_B] = bHex
	state.mu.Unlock()
	var out []*pong.NeedPreSigs_PerInput
	for _, e := range escrows {
		id := fmt.Sprintf("%s:%d", e.TxID, e.Vout)
		redeemOwner := state.AComp
		if e.Owner == "B" {
			redeemOwner = state.BComp
		}
		redeem, err := buildEscrowRedeemScript(state, redeemOwner)
		if err != nil {
			return "", "", nil, fmt.Errorf("buildEscrowRedeemScript: %w", err)
		}
		tA, _ := hex.DecodeString(aInputs[id].t)
		tB, _ := hex.DecodeString(bInputs[id].t)
		out = append(out, &pong.NeedPreSigs_PerInput{InputId: id, RedeemScriptHex: hex.EncodeToString(redeem), MAwinsHex: aInputs[id].m, TAwinsCompressed: tA, MBwinsHex: bInputs[id].m, TBwinsCompressed: tB})
	}
	return aHex, bHex, out, nil
}

// v0-min: AllocateMatch builds per-player deposit scripts/addresses with fixed CSV and stores minimal state.
func (s *Server) AllocateMatch(ctx context.Context, req *pong.AllocateMatchRequest) (*pong.AllocateMatchResponse, error) {
	s.ensureRefereeKey()
	if req.AC == "" || req.BC == "" || req.BetAtoms == 0 {
		return nil, fmt.Errorf("invalid request")
	}
	matchID := fmt.Sprintf("%d", time.Now().UnixNano())
	state := &refMatchState{
		MatchID:       matchID,
		AComp:         req.AC,
		BComp:         req.BC,
		CSV:           64,
		XA:            req.AC,
		XB:            req.BC,
		PreSigsA:      make(map[string]refInputPreSig),
		PreSigsB:      make(map[string]refInputPreSig),
		RequiredAtoms: req.BetAtoms,
		FeeAtoms:      s.pocFeeAtoms,
	}
	// Build redeem scripts for each player using shared XA/XB and distinct P_i
	redeema, err := buildEscrowRedeemScript(state, state.AComp)
	if err != nil {
		return nil, fmt.Errorf("buildEscrowRedeemScript: %w", err)
	}
	redeemb, err := buildEscrowRedeemScript(state, state.BComp)
	if err != nil {
		return nil, fmt.Errorf("buildEscrowRedeemScript: %w", err)
	}
	s.Lock()
	s.matches[matchID] = state
	// Store per-player deposit scripts
	if err := setDepositFromRedeemLocked(state, hex.EncodeToString(redeema)); err == nil {
		state.ADepositRedeemScriptHex = state.DepositRedeemScriptHex
		state.ADepositPkScriptHex = state.DepositPkScriptHex
	}
	if err := setDepositFromRedeemLocked(state, hex.EncodeToString(redeemb)); err == nil {
		state.BDepositRedeemScriptHex = state.DepositRedeemScriptHex
		state.BDepositPkScriptHex = state.DepositPkScriptHex
	}
	s.Unlock()
	// Addresses
	addrA, _ := addrFromPkScript(state.ADepositPkScriptHex, chaincfg.TestNet3Params())
	addrB, _ := addrFromPkScript(state.BDepositPkScriptHex, chaincfg.TestNet3Params())
	return &pong.AllocateMatchResponse{
		MatchId:          matchID,
		Csv:              64,
		BetAtoms:         req.BetAtoms,
		FeeAtoms:         s.pocFeeAtoms,
		ARedeemScriptHex: state.ADepositRedeemScriptHex,
		APkScriptHex:     state.ADepositPkScriptHex,
		ADepositAddress:  addrA,
		BRedeemScriptHex: state.BDepositRedeemScriptHex,
		BPkScriptHex:     state.BDepositPkScriptHex,
		BDepositAddress:  addrB,
	}, nil
}

// v0-min: WaitFunding streams until both A and B deposits reach >=1 conf.
func (s *Server) WaitFunding(req *pong.WaitFundingRequest, stream pong.PongReferee_WaitFundingServer) error {
	ctx := stream.Context()
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			s.RLock()
			state, ok := s.matches[req.MatchId]
			s.RUnlock()
			if !ok {
				return fmt.Errorf("unknown match id")
			}
			if s.dcrd == nil || s.watcher == nil {
				return fmt.Errorf("chain access not configured")
			}
			// Ensure watcher has both scripts registered
			if state.ADepositPkScriptHex != "" {
				s.watcher.registerDeposit(state.ADepositPkScriptHex, state.ADepositRedeemScriptHex, state.MatchID+"/A")
			}
			if state.BDepositPkScriptHex != "" {
				s.watcher.registerDeposit(state.BDepositPkScriptHex, state.BDepositRedeemScriptHex, state.MatchID+"/B")
			}
			utxosA, confA, _ := s.watcher.queryDeposit(state.ADepositPkScriptHex)
			utxosB, confB, _ := s.watcher.queryDeposit(state.BDepositPkScriptHex)
			var ua, ub *pong.EscrowUTXO
			if len(utxosA) > 0 {
				ua = utxosA[0]
			}
			if len(utxosB) > 0 {
				ub = utxosB[0]
			}
			confirmed := confA >= 1 && confB >= 1 && ua != nil && ub != nil
			if err := stream.Send(&pong.WaitFundingResponse{Confirmed: confirmed, ConfsA: confA, ConfsB: confB, UtxoA: ua, UtxoB: ub}); err != nil {
				return err
			}
			if confirmed {
				return nil
			}
		}
	}
}

// AllocateEscrow returns a deposit P2SH address and scripts for a player
func (s *Server) AllocateEscrow(ctx context.Context, req *pong.AllocateEscrowRequest) (*pong.AllocateEscrowResponse, error) {
	s.ensureRefereeKey()
	// v0-min: restore allocation per player to avoid duplicate matches
	var matchID string
	if req.PlayerId != "" {
		if mid, ok := s.refAllocByPlayer[req.PlayerId]; ok {
			matchID = mid
		} else {
			// Try persistent mapping
			if mid, err := s.db.FetchRefAlloc(ctx, req.PlayerId); err == nil && mid != "" {
				matchID = mid
				s.Lock()
				s.refAllocByPlayer[req.PlayerId] = mid
				s.Unlock()
			}
		}
	}
	// If we have a match id, try to restore and return it
	if matchID != "" {
		s.RLock()
		state, ok := s.matches[matchID]
		s.RUnlock()
		if !ok {
			if rec, err := s.db.FetchRefMatch(ctx, matchID); err == nil && rec != nil {
				rs := &refMatchState{
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
				// Mirror deposit into A-deposit fields for funding queries
				rs.ADepositPkScriptHex = rs.DepositPkScriptHex
				rs.ADepositRedeemScriptHex = rs.DepositRedeemScriptHex
				s.Lock()
				s.matches[matchID] = rs
				if s.dcrd != nil && s.watcher != nil && rs.DepositPkScriptHex != "" {
					s.watcher.registerDeposit(rs.DepositPkScriptHex, rs.DepositRedeemScriptHex, rs.MatchID)
				}
				s.Unlock()
				state = rs
				ok = true
			}
		}
		if ok {
			resp := &pong.AllocateEscrowResponse{
				MatchId:         matchID,
				Role:            pong.Branch_BRANCH_A,
				XA:              state.XA,
				XB:              state.XB,
				Csv:             state.CSV,
				DepositAddress:  "",
				RedeemScriptHex: state.DepositRedeemScriptHex,
				PkScriptHex:     state.DepositPkScriptHex,
				RequiredAtoms:   state.RequiredAtoms,
			}
			if state.DepositPkScriptHex != "" {
				if addr, err := addrFromPkScript(state.DepositPkScriptHex, chaincfg.TestNet3Params()); err == nil {
					resp.DepositAddress = addr
				} else {
					s.log.Warnf("AllocateEscrow restore: cannot extract address: %v", err)
				}
			}
			return resp, nil
		}
		return nil, fmt.Errorf("AllocateEscrow: mapping exists for player %s but match %s not restorable", req.PlayerId, matchID)
	}
	// For now, create a new minimal state per player
	matchID = fmt.Sprintf("%d", time.Now().UnixNano())
	if req.PlayerId == "" || req.AC == "" || req.BetAtoms == 0 || req.Csv == 0 {
		return nil, fmt.Errorf("invalid allocate escrow request")
	}

	s.log.Infof("AllocateEscrow: player=%s bet_atoms=%d csv=%d -> creating match", req.PlayerId, req.BetAtoms, req.Csv)
	// Minimal: allocate as role A for the first player requesting; production would pair players and set roles
	// Ensure distinct branch keys for PoC: A-branch uses player's key; B-branch uses placeholder server key until pairing is implemented.
	xA := req.AC
	xB := s.refereePubCompressed

	state := &refMatchState{
		MatchID:  matchID,
		AComp:    req.AC,
		BComp:    s.refereePubCompressed, // placeholder until paired
		CSV:      req.Csv,
		XA:       xA,
		XB:       xB,
		PreSigsA: make(map[string]refInputPreSig),
		PreSigsB: make(map[string]refInputPreSig),
	}
	s.Lock()
	s.matches[matchID] = state
	s.Unlock()

	// Build redeemScript for role A and persist via helper under lock
	redeem, err := buildEscrowRedeemScript(state, req.AC)
	if err != nil {
		return nil, fmt.Errorf("buildEscrowRedeemScript: %w", err)
	}
	s.Lock()
	if err := setDepositFromRedeemLocked(state, hex.EncodeToString(redeem)); err != nil {
		s.Unlock()
		return nil, err
	}
	// Mirror legacy single-deposit into A-deposit fields so WaitFunding can track it.
	state.ADepositRedeemScriptHex = state.DepositRedeemScriptHex
	state.ADepositPkScriptHex = state.DepositPkScriptHex
	state.RequiredAtoms = req.BetAtoms
	if s.dcrd != nil && s.watcher != nil {
		s.watcher.registerDeposit(state.DepositPkScriptHex, state.DepositRedeemScriptHex, state.MatchID)
	}
	// Remember mapping so subsequent AllocateEscrow from same player reuses match
	if req.PlayerId != "" {
		if s.refAllocByPlayer == nil {
			s.refAllocByPlayer = make(map[string]string)
		}
		s.refAllocByPlayer[req.PlayerId] = matchID
		_ = s.db.SaveRefAlloc(ctx, req.PlayerId, matchID)
	}
	s.Unlock()

	// Persist minimal state and mapping AFTER deposit fields are set
	_ = s.db.SaveRefMatch(ctx, &serverdb.RefMatchRecord{
		MatchID:                state.MatchID,
		AComp:                  state.AComp,
		BComp:                  state.BComp,
		CSV:                    state.CSV,
		XA:                     state.XA,
		XB:                     state.XB,
		DepositPkScriptHex:     state.DepositPkScriptHex,
		DepositRedeemScriptHex: state.DepositRedeemScriptHex,
		RequiredAtoms:          state.RequiredAtoms,
	})
	// v0-min: no SaveRefAlloc mapping by player

	// Derive address strictly from pkScript using stdscript
	depositAddress := ""
	if state.DepositPkScriptHex != "" {
		if a, err := addrFromPkScript(state.DepositPkScriptHex, chaincfg.TestNet3Params()); err == nil {
			depositAddress = a
		} else {
			s.log.Warnf("AllocateEscrow: cannot extract address from pkScript: %v", err)
		}
	}

	// Optional runtime guard: ensure pkScript length looks correct and log inconsistencies
	if depositAddress != "" && state.DepositPkScriptHex != "" {
		pk := strings.ToLower(state.DepositPkScriptHex)
		if len(pk) != 46 { // 23 bytes * 2 hex
			s.log.Warnf("pkScript unexpected length for match %s: %d", matchID, len(pk))
		}
	}

	pkLen := 0
	if state.DepositPkScriptHex != "" {
		if pkb, err := hex.DecodeString(state.DepositPkScriptHex); err == nil {
			pkLen = len(pkb)
		}
	}
	s.log.Infof("AllocateEscrow: match_id=%s deposit_address=%s redeem_len=%d pk_len=%d", matchID, depositAddress, len(redeem), pkLen)
	// Diagnostics: also log pkScript used to derive the address
	if state.DepositPkScriptHex != "" {
		s.log.Infof("AllocateEscrow: pkScript=%s addr=%s", state.DepositPkScriptHex, depositAddress)
	}

	resp := &pong.AllocateEscrowResponse{
		MatchId:         matchID,
		Role:            pong.Branch_BRANCH_A,
		XA:              xA,
		XB:              xB,
		Csv:             req.Csv,
		DepositAddress:  depositAddress,
		RedeemScriptHex: state.DepositRedeemScriptHex,
		PkScriptHex:     state.DepositPkScriptHex,
		RequiredAtoms:   req.BetAtoms,
	}
	return resp, nil
}

func poll(ctx context.Context, every time.Duration, f func() (done bool, err error)) error {
	t := time.NewTicker(every)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-t.C:
			done, err := f()
			if err != nil {
				return err
			}
			if done {
				return nil
			}
		}
	}
}

func (s *Server) setAutoReady(st *refMatchState, playerID string) {
	var id zkidentity.ShortID
	if err := id.FromString(playerID); err != nil {
		return
	}
	pl := s.gameManager.PlayerSessions.GetPlayer(id)
	if pl == nil || pl.Ready {
		return
	}
	pl.Ready = true

	// Notify waiting-room peers if present
	if pl.WR != nil {
		if pwr, err := pl.WR.Marshal(); err == nil {
			for _, p := range pl.WR.Players {
				if p.NotifierStream != nil {
					_ = p.NotifierStream.Send(&pong.NtfnStreamResponse{
						NotificationType: pong.NotificationType_ON_PLAYER_READY,
						Message:          fmt.Sprintf("Player %s is ready (funding confirmed)", pl.Nick),
						PlayerId:         playerID,
						Wr:               pwr,
						Ready:            true,
					})
				}
			}
		}
	}

	// If the player is already in a game instance, also notify game peers
	if g := s.gameManager.GetPlayerGame(id); g != nil {
		peers := append([]*ponggame.Player(nil), g.Players...)
		for _, p := range peers {
			if p.NotifierStream != nil {
				_ = p.NotifierStream.Send(&pong.NtfnStreamResponse{
					NotificationType: pong.NotificationType_ON_PLAYER_READY,
					Message:          fmt.Sprintf("Player %s is ready (funding confirmed)", pl.Nick),
					PlayerId:         playerID,
					GameId:           g.Id,
					Ready:            true,
				})
			}
		}
		g.Lock()
		if g.PlayersReady == nil {
			g.PlayersReady = make(map[string]bool)
		}
		g.PlayersReady[playerID] = true
		g.Unlock()
	}
}

// queryFunding returns (utxos, minConf, err)
func (s *Server) queryFunding(st *refMatchState) ([]*pong.EscrowUTXO, uint32, error) {
	// Prefer watcher with per-player deposits (A/B). Fall back to escrows list if already set.
	if st.ADepositPkScriptHex != "" || st.BDepositPkScriptHex != "" {
		if s.watcher == nil {
			return nil, 0, fmt.Errorf("watcher not initialized")
		}
		if st.ADepositPkScriptHex != "" {
			s.watcher.registerDeposit(st.ADepositPkScriptHex, st.ADepositRedeemScriptHex, st.MatchID+"/A")
		}
		if st.BDepositPkScriptHex != "" {
			s.watcher.registerDeposit(st.BDepositPkScriptHex, st.BDepositRedeemScriptHex, st.MatchID+"/B")
		}
		var all []*pong.EscrowUTXO
		min := ^uint32(0)
		if us, c, ok := s.watcher.queryDeposit(st.ADepositPkScriptHex); ok {
			all = append(all, us...)
			if c < min {
				min = c
			}
		}
		if us, c, ok := s.watcher.queryDeposit(st.BDepositPkScriptHex); ok {
			all = append(all, us...)
			if c < min {
				min = c
			}
		}
		if min == ^uint32(0) {
			min = 0
		}
		return all, min, nil
	}

	// Escrows path (already resolved to txids)
	if s.dcrd == nil {
		return nil, 0, fmt.Errorf("dcrd RPC not configured")
	}
	if len(st.Escrows) == 0 {
		return nil, 0, nil
	}
	utxos := make([]*pong.EscrowUTXO, 0, len(st.Escrows))
	min := ^uint32(0)
	for _, e := range st.Escrows {
		utxos = append(utxos, &pong.EscrowUTXO{Txid: e.TxID, Vout: e.Vout, Value: e.Value, RedeemScriptHex: e.RedeemScriptHex, PkScriptHex: e.PkScriptHex, Owner: e.Owner})
		var h chainhash.Hash
		if err := chainhash.Decode(&h, e.TxID); err != nil {
			return nil, 0, err
		}
		vtx, err := s.dcrd.GetRawTransactionVerbose(context.Background(), &h)
		if err != nil {
			return nil, 0, err
		}
		c := uint32(vtx.Confirmations)
		if c < min {
			min = c
		}
	}
	if min == ^uint32(0) {
		min = 0
	}
	return utxos, min, nil
}

func (s *Server) FinalizeWinner(ctx context.Context, req *pong.FinalizeWinnerRequest) (*pong.FinalizeWinnerResponse, error) {
	var out *pong.FinalizeWinnerResponse
	err := s.withMatch(ctx, req.MatchId, func(st *refMatchState) error {
		bs := st.storeFor(req.Branch)
		if len(bs.pre) == 0 {
			return fmt.Errorf("no pre-sigs for branch")
		}

		st.mu.RLock()
		for id, ps := range bs.pre {
			spb, _ := hex.DecodeString(ps.SPrime)
			var s1 secp256k1.ModNScalar
			s1.SetByteSlice(spb)
			// In POC, we don't compute final s here; winner client finalizes with gamma and broadcasts.
			sb := s1.Bytes()
			s.log.Infof("FinalizeWinner (POC log): %s r'=%s s'=%s -> s_final = s' + γ (revealed to winner)", id, ps.RX, ps.SPrime)
			_ = sb
		}
		st.mu.RUnlock()
		out = &pong.FinalizeWinnerResponse{BroadcastTxid: ""}
		return nil
	})
	return out, err
}

// SignalReadyToPlay marks the caller ready and, if everyone is ready, triggers next phase.
func (s *Server) SignalReadyToPlay(ctx context.Context, req *pong.SignalReadyToPlayRequest) (*pong.SignalReadyToPlayResponse, error) {
	var clientID zkidentity.ShortID
	if err := clientID.FromString(req.ClientId); err != nil {
		return nil, fmt.Errorf("invalid client id: %w", err)
	}

	s.log.Debugf("Client %s signaling ready for game %s", req.ClientId, req.GameId)

	player := s.gameManager.PlayerSessions.GetPlayer(clientID)
	if player == nil {
		return nil, fmt.Errorf("player not found for client %s", clientID)
	}

	game := s.gameManager.GetPlayerGame(clientID)
	if game == nil || (req.GameId != "" && game.Id != req.GameId) {
		return nil, fmt.Errorf("game instance not found for client %s", clientID)
	}

	// Update ready state idempotently and snapshot for notifications without holding the lock.
	var (
		already     bool
		allReady    bool
		playersSnap []*ponggame.Player
		nick        = player.Nick
	)

	game.Lock()
	if game.PlayersReady == nil {
		game.PlayersReady = make(map[string]bool)
	}
	key := clientID.String()
	if game.PlayersReady[key] {
		already = true
	} else {
		game.PlayersReady[key] = true
	}

	// Count readiness.
	readyCount := 0
	total := len(game.Players)
	for _, p := range game.Players {
		if game.PlayersReady[p.ID.String()] {
			readyCount++
		}
	}
	allReady = (total > 0 && readyCount == total)

	// Snapshot players slice for notifying outside the lock.
	playersSnap = append(playersSnap, game.Players...)
	game.Unlock()

	// Notify: this player is ready (only once).
	if !already {
		s.broadcastReady(playersSnap, req.ClientId, game.Id, fmt.Sprintf("Player %s is ready to start the game", nick))
	}

	// // If everyone is ready, signal next phase (start game, fetch drafts/adaptors, etc.).
	if allReady {
		// s.broadcastAllReady(playersSnap, game.Id)
		// go s.onAllPlayersReady(game) // hook to kick the next phase safely out-of-band
	}

	return &pong.SignalReadyToPlayResponse{
		Success: true,
		Message: "Ready signal received",
	}, nil
}

func (s *Server) broadcastReady(players []*ponggame.Player, playerID, gameID, msg string) {
	for _, p := range players {
		if p.NotifierStream == nil {
			continue
		}
		_ = p.NotifierStream.Send(&pong.NtfnStreamResponse{
			NotificationType: pong.NotificationType_ON_PLAYER_READY,
			Message:          msg,
			PlayerId:         playerID,
			GameId:           gameID,
			Ready:            true,
		})
	}
}
