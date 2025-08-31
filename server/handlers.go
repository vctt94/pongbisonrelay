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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

	// match id identifiers
	matchID string

	// Players' 33-byte compressed pubkeys (33 bytes)
	aComp []byte
	bComp []byte
	xa    []byte
	xb    []byte

	csv      uint32
	reqAtoms uint64
	feeAtoms uint64

	// Funding/escrow
	escrows []refEscrowUTXO

	brDraft map[pong.Branch]string    // branch -> tx hex
	preSigA map[string]refInputPreSig // inputID -> presig (branch A)
	preSigB map[string]refInputPreSig // inputID -> presig (branch B)

	// hex fields kept for watcher/DB edges
	depositPkScriptHex      string
	depositRedeemScriptHex  string
	aDepositPkScriptHex     string
	aDepositRedeemScriptHex string
	bDepositPkScriptHex     string
	bDepositRedeemScriptHex string
}

// Safe getters (return copies so callers can't mutate internal slices).
func (s *refMatchState) AComp() []byte {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]byte(nil), s.aComp...)
}
func (s *refMatchState) BComp() []byte {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]byte(nil), s.bComp...)
}
func (s *refMatchState) XA() []byte {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]byte(nil), s.xa...)
}
func (s *refMatchState) XB() []byte {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]byte(nil), s.xb...)
}
func (s *refMatchState) CSV() uint32           { s.mu.RLock(); defer s.mu.RUnlock(); return s.csv }
func (s *refMatchState) RequiredAtoms() uint64 { s.mu.RLock(); defer s.mu.RUnlock(); return s.reqAtoms }
func (s *refMatchState) FeeAtoms() uint64      { s.mu.RLock(); defer s.mu.RUnlock(); return s.feeAtoms }

// Deposit script getters (byte copies; decode errors return nil)
func (s *refMatchState) DepositPkScript() []byte {
	s.mu.RLock()
	defer s.mu.RUnlock()
	b, err := hex.DecodeString(s.depositPkScriptHex)
	if err != nil {
		return nil
	}
	return append([]byte(nil), b...)
}
func (s *refMatchState) DepositRedeemScript() []byte {
	s.mu.RLock()
	defer s.mu.RUnlock()
	b, err := hex.DecodeString(s.depositRedeemScriptHex)
	if err != nil {
		return nil
	}
	return append([]byte(nil), b...)
}
func (s *refMatchState) ADepositPkScript() []byte {
	s.mu.RLock()
	defer s.mu.RUnlock()
	b, err := hex.DecodeString(s.aDepositPkScriptHex)
	if err != nil {
		return nil
	}
	return append([]byte(nil), b...)
}
func (s *refMatchState) BDepositPkScript() []byte {
	s.mu.RLock()
	defer s.mu.RUnlock()
	b, err := hex.DecodeString(s.bDepositPkScriptHex)
	if err != nil {
		return nil
	}
	return append([]byte(nil), b...)
}

// Setters (validate & copy)
func (s *refMatchState) SetXA(b []byte) error {
	if len(b) != 33 {
		return fmt.Errorf("XA must be 33 bytes")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.xa = append([]byte(nil), b...)
	return nil
}
func (s *refMatchState) SetXB(b []byte) error {
	if len(b) != 33 {
		return fmt.Errorf("XB must be 33 bytes")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.xb = append([]byte(nil), b...)
	return nil
}
func (s *refMatchState) SetAComp(b []byte) error {
	if len(b) != 33 {
		return fmt.Errorf("AComp must be 33 bytes")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.aComp = append([]byte(nil), b...)
	return nil
}
func (s *refMatchState) SetBComp(b []byte) error {
	if len(b) != 33 {
		return fmt.Errorf("BComp must be 33 bytes")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.bComp = append([]byte(nil), b...)
	return nil
}

// Draft / presig helpers
func (s *refMatchState) SetDraft(br pong.Branch, txHex string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.brDraft == nil {
		s.brDraft = make(map[pong.Branch]string)
	}
	s.brDraft[br] = txHex
}
func (s *refMatchState) Draft(br pong.Branch) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	h, ok := s.brDraft[br]
	return h, ok
}
func (s *refMatchState) AddPreSig(br pong.Branch, inID string, ps refInputPreSig) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if br == pong.Branch_BRANCH_A {
		if s.preSigA == nil {
			s.preSigA = make(map[string]refInputPreSig)
		}
		s.preSigA[inID] = ps
		return
	}
	if s.preSigB == nil {
		s.preSigB = make(map[string]refInputPreSig)
	}
	s.preSigB[inID] = ps
}
func (s *refMatchState) PreSigs(br pong.Branch) map[string]refInputPreSig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.preSigA
	if br == pong.Branch_BRANCH_B {
		src = s.preSigB
	}
	out := make(map[string]refInputPreSig, len(src))
	for k, v := range src {
		out[k] = v
	}
	return out
}

type branchStore struct {
	pre map[string]refInputPreSig
}

func (st *refMatchState) storeFor(br pong.Branch) branchStore {
	if br == pong.Branch_BRANCH_A {
		return branchStore{st.preSigA}
	}
	return branchStore{st.preSigB}
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
// buildEscrowRedeemScript builds the redeemScript using XA/XB/owner compressed pubkeys (33 bytes).
func buildEscrowRedeemScript(state *refMatchState, ownerComp []byte) ([]byte, error) {
	const STSchnorrSecp256k1 int64 = 2

	if len(state.xa) != 33 {
		return nil, fmt.Errorf("XA must be 33-byte compressed pubkey")
	}
	if len(state.xb) != 33 {
		return nil, fmt.Errorf("XB must be 33-byte compressed pubkey")
	}
	if len(ownerComp) != 33 {
		return nil, fmt.Errorf("owner must be 33-byte compressed pubkey")
	}
	if state.csv == 0 {
		return nil, fmt.Errorf("CSV must be > 0")
	}

	b := txscript.NewScriptBuilder()
	b.AddOp(txscript.OP_IF).
		AddData(state.xa).AddInt64(STSchnorrSecp256k1).AddOp(txscript.OP_CHECKSIGALT).
		AddOp(txscript.OP_ELSE).
		AddOp(txscript.OP_IF).
		AddData(state.xb).AddInt64(STSchnorrSecp256k1).AddOp(txscript.OP_CHECKSIGALT).
		AddOp(txscript.OP_ELSE).
		AddInt64(int64(state.csv)).AddOp(txscript.OP_CHECKSEQUENCEVERIFY).AddOp(txscript.OP_DROP).
		AddData(ownerComp).AddInt64(STSchnorrSecp256k1).AddOp(txscript.OP_CHECKSIGALT).
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
		bdHex, _ := st.Draft(req.Branch)
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
		entries := make([]*pong.RevealAdaptorEntry, 0)
		for id := range st.PreSigs(req.Branch) {
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
			redeemA, _ := buildEscrowRedeemScript(st, st.AComp())
			redeemB, _ := buildEscrowRedeemScript(st, st.BComp())
			mA, _ := txscript.CalcSignatureHash(redeemA, txscript.SigHashAll, &tx, idx, nil)
			mB, _ := txscript.CalcSignatureHash(redeemB, txscript.SigHashAll, &tx, idx, nil)
			gammaHexA, _ := deriveAdaptorGamma(st.matchID, id, int32(req.Branch), hex.EncodeToString(mA), s.refereePrivKeyHex)
			gammaHex := gammaHexA
			if gammaHex == "" {
				gammaHex, _ = deriveAdaptorGamma(st.matchID, id, int32(req.Branch), hex.EncodeToString(mB), s.refereePrivKeyHex)
			}
			entries = append(entries, &pong.RevealAdaptorEntry{InputId: id, Gamma: gammaHex})
		}
		out = &pong.RevealAdaptorsResponse{Branch: req.Branch, Entries: entries}
		return nil
	})
	return out, err
}

// buildBranchForStream builds a single-branch draft and NeedPreSigs inputs for the stream.
func (s *Server) buildBranchForStream(ctx context.Context, matchID string, br pong.Branch) (string, []*pong.NeedPreSigs_PerInput, error) {
	s.RLock()
	state, ok := s.matches[matchID]
	s.RUnlock()
	if !ok {
		return "", nil, fmt.Errorf("unknown match id")
	}
	s.ensureRefereeKey()
	var escrows []refEscrowUTXO
	if state.aDepositPkScriptHex != "" {
		if us, _, ok := s.watcher.queryDeposit(state.aDepositPkScriptHex); ok && len(us) > 0 {
			e := us[0]
			escrows = append(escrows, refEscrowUTXO{TxID: e.Txid, Vout: e.Vout, Value: e.Value, RedeemScriptHex: e.RedeemScriptHex, PkScriptHex: e.PkScriptHex, Owner: "A"})
		}
	}
	if state.bDepositPkScriptHex != "" {
		if us, _, ok := s.watcher.queryDeposit(state.bDepositPkScriptHex); ok && len(us) > 0 {
			e := us[0]
			escrows = append(escrows, refEscrowUTXO{TxID: e.Txid, Vout: e.Vout, Value: e.Value, RedeemScriptHex: e.RedeemScriptHex, PkScriptHex: e.PkScriptHex, Owner: "B"})
		}
	}
	if len(escrows) != 2 {
		return "", nil, fmt.Errorf("need 2 deposits")
	}

	// Build the branch draft paying to XA (A) or XB (B)
	build := func(payoutX []byte) (string, error) {
		tx := wire.NewMsgTx()
		totalIn := int64(0)
		for _, e := range escrows {
			var h chainhash.Hash
			_ = chainhash.Decode(&h, e.TxID)
			op := wire.NewOutPoint(&h, e.Vout, 0)
			tx.AddTxIn(wire.NewTxIn(op, int64(e.Value), nil))
			totalIn += int64(e.Value)
		}
		fee := state.FeeAtoms()
		if totalIn < int64(fee) {
			return "", fmt.Errorf("fee exceeds total inputs")
		}
		pay := totalIn - int64(fee)
		pkAlt, _ := txscript.NewScriptBuilder().AddData(payoutX).AddInt64(2).AddOp(txscript.OP_CHECKSIGALT).Script()
		tx.AddTxOut(&wire.TxOut{Value: pay, Version: 0, PkScript: pkAlt})
		var buf bytes.Buffer
		_ = tx.Serialize(&buf)
		return hex.EncodeToString(buf.Bytes()), nil
	}

	var payout []byte
	if br == pong.Branch_BRANCH_A {
		payout = state.XA()
	} else {
		payout = state.XB()
	}
	txHex, err := build(payout)
	if err != nil {
		return "", nil, err
	}
	state.SetDraft(br, txHex)

	// Per-input values (m, T) for this branch only
	out := make([]*pong.NeedPreSigs_PerInput, 0, len(escrows))
	for idx, e := range escrows {
		redeemOwner := state.AComp()
		if e.Owner == "B" {
			redeemOwner = state.BComp()
		}
		redeem, err := buildEscrowRedeemScript(state, redeemOwner)
		if err != nil {
			return "", nil, fmt.Errorf("buildEscrowRedeemScript: %w", err)
		}
		// Sighash for this input
		var tx wire.MsgTx
		if b, err := hex.DecodeString(txHex); err == nil {
			if err := tx.Deserialize(bytes.NewReader(b)); err != nil {
				return "", nil, fmt.Errorf("bad tx hex")
			}
		} else {
			return "", nil, fmt.Errorf("bad tx hex")
		}
		m, err := txscript.CalcSignatureHash(redeem, txscript.SigHashAll, &tx, idx, nil)
		if err != nil || len(m) != 32 {
			return "", nil, fmt.Errorf("sighash")
		}
		id := fmt.Sprintf("%s:%d", e.TxID, e.Vout)
		_, TCompHex := deriveAdaptorGamma(state.matchID, id, int32(br), hex.EncodeToString(m), s.refereePrivKeyHex)
		entry := &pong.NeedPreSigs_PerInput{InputId: id, RedeemScriptHex: hex.EncodeToString(redeem)}
		entry.MHex = hex.EncodeToString(m)
		entry.TCompressed, _ = hex.DecodeString(TCompHex)
		out = append(out, entry)
	}
	return txHex, out, nil
}

// SettlementStream
func (s *Server) SettlementStream(stream pong.PongReferee_SettlementStreamServer) error {
	ctx := stream.Context()
	var matchID string
	var roleForStream pong.Branch
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
			if st == nil {
				return status.Errorf(codes.NotFound, "unknown match")
			}
			// Determine role from client's compressed pubkey
			if bytes.Equal(h.CompPubkey, st.AComp()) {
				roleForStream = pong.Branch_BRANCH_A
			} else if bytes.Equal(h.CompPubkey, st.BComp()) {
				roleForStream = pong.Branch_BRANCH_B
			} else {
				return status.Errorf(codes.PermissionDenied, "pubkey not part of match")
			}
			reqAtoms = st.RequiredAtoms()
			pkHex = st.depositPkScriptHex

			_ = stream.Send(&pong.ServerMsg{MatchId: matchID, Kind: &pong.ServerMsg_Role{Role: &pong.AssignRole{Role: pong.AssignRole_Role(roleForStream), RequiredAtoms: reqAtoms, DepositPkscriptHex: pkHex}}})

			txHex, inputs, err := s.buildBranchForStream(ctx, matchID, roleForStream)
			if err != nil {
				_ = stream.Send(&pong.ServerMsg{MatchId: matchID, Kind: &pong.ServerMsg_Info{Info: &pong.Info{Text: err.Error()}}})
				continue
			}
			nps := &pong.NeedPreSigs{Branch: roleForStream, DraftTxHex: txHex, Inputs: inputs}
			fmt.Printf("nps: %v\n", nps)
			_ = stream.Send(&pong.ServerMsg{MatchId: matchID, Kind: &pong.ServerMsg_Req{Req: nps}})
			continue
		}
		// The client computes R' locally in computePreSig(xPrivHex, mHex, TCompHex string),
		//  extracts r' = x(R'), and sends only (r', s').
		if pb := in.GetPresigs(); pb != nil {
			if pb.Branch != roleForStream {
				return status.Errorf(codes.PermissionDenied, "client cannot presign branch %v", pb.Branch)
			}
			s.withMatch(ctx, matchID, func(st *refMatchState) error {
				for _, sig := range pb.Presigs {
					rx := hex.EncodeToString(sig.Rprime32)
					sp := hex.EncodeToString(sig.Sprime32)
					st.AddPreSig(pb.Branch, sig.InputId, refInputPreSig{InputID: sig.InputId, RX: rx, SPrime: sp})
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

// v0-min: AllocateMatch builds per-player deposit scripts/addresses with fixed CSV and stores minimal state.
func (s *Server) AllocateMatch(ctx context.Context, req *pong.AllocateMatchRequest) (*pong.AllocateMatchResponse, error) {
	s.ensureRefereeKey()
	if req.AC == "" || req.BC == "" || req.BetAtoms == 0 {
		return nil, fmt.Errorf("invalid request")
	}
	matchID := fmt.Sprintf("%d", time.Now().UnixNano())
	aComp, _ := hex.DecodeString(req.AC)
	bComp, _ := hex.DecodeString(req.BC)
	state := &refMatchState{
		matchID:  matchID,
		aComp:    aComp,
		bComp:    bComp,
		csv:      64,
		xa:       aComp,
		xb:       bComp,
		preSigA:  make(map[string]refInputPreSig),
		preSigB:  make(map[string]refInputPreSig),
		reqAtoms: req.BetAtoms,
		feeAtoms: s.pocFeeAtoms,
	}
	// Build redeem scripts for each player using shared XA/XB and distinct P_i
	redeema, err := buildEscrowRedeemScript(state, state.AComp())
	if err != nil {
		return nil, fmt.Errorf("buildEscrowRedeemScript: %w", err)
	}
	redeemb, err := buildEscrowRedeemScript(state, state.BComp())
	if err != nil {
		return nil, fmt.Errorf("buildEscrowRedeemScript: %w", err)
	}
	s.Lock()
	s.matches[matchID] = state
	// Store per-player deposit scripts
	if err := setDepositFromRedeemLocked(state, hex.EncodeToString(redeema)); err == nil {
		state.aDepositRedeemScriptHex = state.depositRedeemScriptHex
		state.aDepositPkScriptHex = state.depositPkScriptHex
	}
	if err := setDepositFromRedeemLocked(state, hex.EncodeToString(redeemb)); err == nil {
		state.bDepositRedeemScriptHex = state.depositRedeemScriptHex
		state.bDepositPkScriptHex = state.depositPkScriptHex
	}
	s.Unlock()
	// Addresses
	addrA, _ := addrFromPkScript(state.aDepositPkScriptHex, chaincfg.TestNet3Params())
	addrB, _ := addrFromPkScript(state.bDepositPkScriptHex, chaincfg.TestNet3Params())
	return &pong.AllocateMatchResponse{
		MatchId:          matchID,
		Csv:              64,
		BetAtoms:         req.BetAtoms,
		FeeAtoms:         s.pocFeeAtoms,
		ARedeemScriptHex: state.aDepositRedeemScriptHex,
		APkScriptHex:     state.aDepositPkScriptHex,
		ADepositAddress:  addrA,
		BRedeemScriptHex: state.bDepositRedeemScriptHex,
		BPkScriptHex:     state.bDepositPkScriptHex,
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
			if state.aDepositPkScriptHex != "" {
				s.watcher.registerDeposit(state.aDepositPkScriptHex, state.aDepositRedeemScriptHex, state.matchID+"/A")
			}
			if state.bDepositPkScriptHex != "" {
				s.watcher.registerDeposit(state.bDepositPkScriptHex, state.bDepositRedeemScriptHex, state.matchID+"/B")
			}
			utxosA, confA, _ := s.watcher.queryDeposit(state.aDepositPkScriptHex)
			utxosB, confB, _ := s.watcher.queryDeposit(state.bDepositPkScriptHex)
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
					matchID:                rec.MatchID,
					aComp:                  append([]byte(nil), rec.AComp...),
					bComp:                  append([]byte(nil), rec.BComp...),
					csv:                    rec.CSV,
					xa:                     append([]byte(nil), rec.XA...),
					xb:                     append([]byte(nil), rec.XB...),
					preSigA:                make(map[string]refInputPreSig),
					preSigB:                make(map[string]refInputPreSig),
					depositPkScriptHex:     rec.DepositPkScriptHex,
					depositRedeemScriptHex: rec.DepositRedeemScriptHex,
					reqAtoms:               rec.RequiredAtoms,
				}
				// Mirror deposit into A-deposit fields for funding queries
				rs.aDepositPkScriptHex = rs.depositPkScriptHex
				rs.aDepositRedeemScriptHex = rs.depositRedeemScriptHex
				s.Lock()
				s.matches[matchID] = rs
				if s.dcrd != nil && s.watcher != nil && rs.depositPkScriptHex != "" {
					s.watcher.registerDeposit(rs.depositPkScriptHex, rs.depositRedeemScriptHex, rs.matchID)
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
				XA:              hex.EncodeToString(state.XA()),
				XB:              hex.EncodeToString(state.XB()),
				Csv:             state.CSV(),
				DepositAddress:  "",
				RedeemScriptHex: state.depositRedeemScriptHex,
				PkScriptHex:     state.depositPkScriptHex,
				RequiredAtoms:   state.RequiredAtoms(),
			}
			if state.depositPkScriptHex != "" {
				if addr, err := addrFromPkScript(state.depositPkScriptHex, chaincfg.TestNet3Params()); err == nil {
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
	xAb, _ := hex.DecodeString(xA)
	xBb, _ := hex.DecodeString(xB)
	aCb, _ := hex.DecodeString(req.AC)
	state := &refMatchState{
		matchID: matchID,
		aComp:   aCb,
		bComp:   xBb, // placeholder until paired
		csv:     req.Csv,
		xa:      xAb,
		xb:      xBb,
		preSigA: make(map[string]refInputPreSig),
		preSigB: make(map[string]refInputPreSig),
	}
	s.Lock()
	s.matches[matchID] = state
	s.Unlock()

	// Build redeemScript for role A and persist via helper under lock
	redeem, err := buildEscrowRedeemScript(state, aCb)
	if err != nil {
		return nil, fmt.Errorf("buildEscrowRedeemScript: %w", err)
	}
	s.Lock()
	if err := setDepositFromRedeemLocked(state, hex.EncodeToString(redeem)); err != nil {
		s.Unlock()
		return nil, err
	}
	// Mirror legacy single-deposit into A-deposit fields so WaitFunding can track it.
	state.aDepositRedeemScriptHex = state.depositRedeemScriptHex
	state.aDepositPkScriptHex = state.depositPkScriptHex
	state.reqAtoms = req.BetAtoms
	if s.dcrd != nil && s.watcher != nil {
		s.watcher.registerDeposit(state.depositPkScriptHex, state.depositRedeemScriptHex, state.matchID)
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
		MatchID:                state.matchID,
		AComp:                  append([]byte(nil), state.AComp()...),
		BComp:                  append([]byte(nil), state.BComp()...),
		XA:                     append([]byte(nil), state.XA()...),
		XB:                     append([]byte(nil), state.XB()...),
		CSV:                    state.CSV(),
		DepositPkScriptHex:     state.depositPkScriptHex,
		DepositRedeemScriptHex: state.depositRedeemScriptHex,
		RequiredAtoms:          state.RequiredAtoms(),
	})
	// v0-min: no SaveRefAlloc mapping by player

	// Derive address strictly from pkScript using stdscript
	depositAddress := ""
	if state.depositPkScriptHex != "" {
		if a, err := addrFromPkScript(state.depositPkScriptHex, chaincfg.TestNet3Params()); err == nil {
			depositAddress = a
		} else {
			s.log.Warnf("AllocateEscrow: cannot extract address from pkScript: %v", err)
		}
	}

	// Optional runtime guard: ensure pkScript length looks correct and log inconsistencies
	if depositAddress != "" && state.depositPkScriptHex != "" {
		pk := strings.ToLower(state.depositPkScriptHex)
		if len(pk) != 46 { // 23 bytes * 2 hex
			s.log.Warnf("pkScript unexpected length for match %s: %d", matchID, len(pk))
		}
	}

	pkLen := 0
	if state.depositPkScriptHex != "" {
		if pkb, err := hex.DecodeString(state.depositPkScriptHex); err == nil {
			pkLen = len(pkb)
		}
	}
	s.log.Infof("AllocateEscrow: match_id=%s deposit_address=%s redeem_len=%d pk_len=%d", matchID, depositAddress, len(redeem), pkLen)
	// Diagnostics: also log pkScript used to derive the address
	if state.depositPkScriptHex != "" {
		s.log.Infof("AllocateEscrow: pkScript=%s addr=%s", state.depositPkScriptHex, depositAddress)
	}

	resp := &pong.AllocateEscrowResponse{
		MatchId:         matchID,
		Role:            pong.Branch_BRANCH_A,
		XA:              xA,
		XB:              xB,
		Csv:             req.Csv,
		DepositAddress:  depositAddress,
		RedeemScriptHex: state.depositRedeemScriptHex,
		PkScriptHex:     state.depositPkScriptHex,
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
	if st.aDepositPkScriptHex != "" || st.bDepositPkScriptHex != "" {
		if s.watcher == nil {
			return nil, 0, fmt.Errorf("watcher not initialized")
		}
		if st.aDepositPkScriptHex != "" {
			s.watcher.registerDeposit(st.aDepositPkScriptHex, st.aDepositRedeemScriptHex, st.matchID+"/A")
		}
		if st.bDepositPkScriptHex != "" {
			s.watcher.registerDeposit(st.bDepositPkScriptHex, st.bDepositRedeemScriptHex, st.matchID+"/B")
		}
		var all []*pong.EscrowUTXO
		min := ^uint32(0)
		if us, c, ok := s.watcher.queryDeposit(st.aDepositPkScriptHex); ok {
			all = append(all, us...)
			if c < min {
				min = c
			}
		}
		if us, c, ok := s.watcher.queryDeposit(st.bDepositPkScriptHex); ok {
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
	if len(st.escrows) == 0 {
		return nil, 0, nil
	}
	utxos := make([]*pong.EscrowUTXO, 0, len(st.escrows))
	min := ^uint32(0)
	for _, e := range st.escrows {
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
		pre := st.PreSigs(req.Branch)
		if len(pre) == 0 {
			return fmt.Errorf("no pre-sigs for branch")
		}
		st.mu.RLock()
		for id, ps := range pre {
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

// presigsComplete returns true if collected presigs cover all escrow inputs for both branches.
func (s *Server) presigsComplete(st *refMatchState) (aDone, bDone bool) {
	st.mu.RLock()
	need := len(st.escrows)
	st.mu.RUnlock()
	if need == 0 {
		return false, false
	}
	return len(st.PreSigs(pong.Branch_BRANCH_A)) >= need, len(st.PreSigs(pong.Branch_BRANCH_B)) >= need
}
