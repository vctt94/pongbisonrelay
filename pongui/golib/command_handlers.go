package golib

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/companyzero/bisonrelay/client/clientintf"
	"github.com/companyzero/bisonrelay/clientrpc/types"
	"github.com/companyzero/bisonrelay/lockfile"
	"github.com/companyzero/bisonrelay/rates"
	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/decred/slog"
	"github.com/vctt94/bisonbotkit/logging"
	"github.com/vctt94/pongbisonrelay"
	"github.com/vctt94/pongbisonrelay/client"
	"github.com/vctt94/pongbisonrelay/pongrpc/grpc/pong"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

const (
	appName = "bisonpong"
)

type clientCtx struct {
	ID     *localInfo
	c      *client.PongClient
	ctx    context.Context
	chat   types.ChatServiceClient
	cancel func()
	runMtx sync.Mutex
	runErr error

	log          slog.Logger
	certConfChan chan bool

	httpClient *http.Client
	rates      *rates.Rates

	// expirationDays are the expirtation days provided by the server when
	// connected
	expirationDays uint64

	serverState atomic.Value
}

var (
	cmtx sync.Mutex
	cs   map[uint32]*clientCtx
	lfs  map[string]*lockfile.LockFile = map[string]*lockfile.LockFile{}

	// The following are debug vars.
	sigUrgCount       atomic.Uint64
	isServerConnected atomic.Bool
)

func handleHello(name string) (string, error) {
	if name == "*bug" {
		return "", fmt.Errorf("name '%s' is an error", name)
	}
	return "hello " + name, nil
}

func isClientRunning(handle uint32) bool {
	cmtx.Lock()
	var res bool
	if cs != nil {
		res = cs[handle] != nil
	}
	cmtx.Unlock()
	return res
}

func handleInitClient(handle uint32, args initClient) (*localInfo, error) {
	cmtx.Lock()
	defer cmtx.Unlock()
	if cs == nil {
		cs = make(map[uint32]*clientCtx)
	}
	// If an existing client exists for this handle, decide whether to reuse or re-init.
	if existing := cs[handle]; existing != nil {
		needReinit := false
		reqCID := strings.TrimSpace(args.ClientID)
		if reqCID == "" {
			// Pre-login init request: keep existing client (prelogin or full) as-is.
			return existing.ID, nil
		}
		// Parse requested ID.
		var reqID zkidentity.ShortID
		if err := reqID.FromString(reqCID); err != nil {
			return nil, fmt.Errorf("invalid client_id format: %v", err)
		}
		// If existing ID is zero or differs from requested, reinit.
		if existing.ID == nil || existing.ID.ID == (clientintf.UserID{}) || existing.ID.ID != reqID {
			needReinit = true
		}
		if !needReinit {
			// Same authenticated client already loaded.
			return existing.ID, nil
		}
		// Stop and remove existing client before reinitializing.
		if existing.cancel != nil {
			existing.cancel()
		}
		delete(cs, handle)
	}

	// Ensure the data directory exists first
	if strings.TrimSpace(args.DataDir) == "" {
		return nil, fmt.Errorf("missing data_dir")
	}
	if err := os.MkdirAll(args.DataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory %s: %v", args.DataDir, err)
	}

	// Ensure the logs subdirectory exists
	logFile := strings.TrimSpace(args.LogFile)
	if logFile == "" {
		logFile = filepath.Join(args.DataDir, "logs", appName+".log")
	}
	logsDir := filepath.Dir(logFile)
	if !strings.HasPrefix(logsDir, args.DataDir) {
		logFile = filepath.Join(args.DataDir, "logs", appName+".log")
		logsDir = filepath.Dir(logFile)
	}
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create logs directory %s: %v", logsDir, err)
	}

	logBackend, err := logging.NewLogBackend(logging.LogConfig{
		LogFile:        logFile,
		DebugLevel:     args.DebugLevel,
		MaxLogFiles:    10,
		MaxBufferLines: 1000,
	})
	if err != nil {
		return nil, err
	}
	log := logBackend.Logger("bisonpong")

	// If no wallet-authenticated clientID is provided, create a minimal client
	// context that can serve local/non-auth features (e.g., historic escrows,
	// archiving session keys), without connecting to any server.
	var li *localInfo
	var haveAuth bool
	var id zkidentity.ShortID
	if strings.TrimSpace(args.ClientID) != "" && id.FromString(args.ClientID) == nil {
		haveAuth = true
		li = &localInfo{ID: id, Nick: "anon"}
	} else {
		// Minimal pre-login mode: leave ID zeroed out, no server info.
		li = &localInfo{Nick: "prelogin"}
	}

	fmt.Println(appName)
	// Build consolidated AppConfig for the pong client (without BR auth)
	// Load network from config file (defaults to testnet if not set)
	appCfg, err := client.LoadAppConfig(args.DataDir, appName)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	g, _ := errgroup.WithContext(ctx)

	// Set up NotificationManager to emit UI notifications and forward to Flutter.
	nmgr := client.NewNotificationManager()
	// Enable common UI notifications and shorten emit interval for responsiveness.
	nmgr.UpdateUIConfig(client.UINotificationsConfig{
		GameStarted:           true,
		WRCreated:             true,
		MaxLength:             255,
		CancelEmissionChannel: ctx.Done(),
	})

	var pc *client.PongClient
	// Full client with connectivity.
	pc, err = client.NewPongClient(args.ClientID, appCfg)
	if err != nil {
		cancel()
		return nil, err
	}
	// Ensure the client ID is set correctly (safeguard in case of any issues)
	if haveAuth && strings.TrimSpace(args.ClientID) != "" {
		if pc.ID() != strings.TrimSpace(args.ClientID) {
			pc.SetID(strings.TrimSpace(args.ClientID))
			log.Infof("Updated client ID to authenticated ID: %s", args.ClientID)
		}
	}
	li.ServerVersion = pc.ServerVersion()
	li.ServerIsF2P = pc.ServerIsF2P()

	cctx := &clientCtx{
		ID:     li,
		ctx:    ctx,
		c:      pc,
		cancel: cancel,
		log:    log,
	}
	cs[handle] = cctx

	if haveAuth {
		// Start the notification stream to receive server notifications
		if err := pc.RefStartNtfnStream(ctx); err != nil {
			cancel()
			cmtx.Lock()
			delete(cs, handle)
			cmtx.Unlock()
			return nil, fmt.Errorf("failed to start notification stream: %w", err)
		}
	}

	// Forward only UI notifications to Flutter via NTUINotification.
	nmgr.Register(client.OnUINotification(func(n client.UINotification) {
		// Forward a simplified payload that matches the Dart struct.
		payload := map[string]interface{}{
			"type":  string(n.Type),
			"text":  n.Text,
			"count": n.Count,
			// Use FromNick as a human-readable source; ensure string type.
			"from": n.FromNick,
		}
		notify(NTUINotification, payload, nil)
	}))

	// Forward structured state-change events as simplified UINotification payloads.
	go func() {
		// Per-second counter of forwarded game frames.
		lastLog := time.Now()
		var fwd int
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-pc.UpdatesCh():
				if ntfn, ok := msg.(*pong.NtfnStreamResponse); ok {
					switch ntfn.NotificationType {
					case pong.NotificationType_CONNECTION_STATE:
						extras := map[string]interface{}{
							"connected": ntfn.Connected,
						}
						fromJSON, _ := json.Marshal(extras)
						notify(NTUINotification, map[string]interface{}{
							"type":  "connection_state",
							"text":  "",
							"count": 0,
							"from":  string(fromJSON),
						}, nil)
						continue
					case pong.NotificationType_SERVER_CONFIG:
						extras := map[string]interface{}{"is_f2p": ntfn.ServerIsF2P}
						fromJSON, _ := json.Marshal(extras)
						notify(NTUINotification, map[string]interface{}{
							"type":  "server_config",
							"text":  ntfn.Message,
							"count": 0,
							"from":  string(fromJSON),
						}, nil)
					case pong.NotificationType_BET_AMOUNT_UPDATE:
						extras := map[string]interface{}{
							"player_id": ntfn.PlayerId,
							"bet_amt":   ntfn.BetAmt,
							"confs":     ntfn.Confs,
						}
						if strings.TrimSpace(ntfn.Message) != "" {
							var meta map[string]interface{}
							if err := json.Unmarshal([]byte(ntfn.Message), &meta); err == nil {
								for k, v := range meta {
									extras[k] = v
								}
							}
						}
						fromJSON, _ := json.Marshal(extras)
						payload := map[string]interface{}{
							"type":  "bet_update",
							"text":  "",
							"count": 0,
							"from":  string(fromJSON),
						}
						notify(NTUINotification, payload, nil)

					case pong.NotificationType_ON_WR_CREATED:
						// Convert proto WR to UI shape
						var wr *waitingRoom
						if ntfn.Wr != nil {
							players := make([]*player, len(ntfn.Wr.Players))
							for i, p := range ntfn.Wr.Players {
								pp, _ := playerFromServer(p)
								players[i] = pp
							}
							wr = &waitingRoom{ID: ntfn.Wr.Id, HostID: ntfn.Wr.HostId, BetAmt: ntfn.Wr.BetAmt, Players: players}
						}
						extras := map[string]interface{}{"waiting_room": wr}
						fromJSON, _ := json.Marshal(extras)
						notify(NTUINotification, map[string]interface{}{
							"type":  "wr_created",
							"text":  "",
							"count": 0,
							"from":  string(fromJSON),
						}, nil)

					case pong.NotificationType_ON_WR_REMOVED:
						extras := map[string]interface{}{"room_id": ntfn.RoomId}
						fromJSON, _ := json.Marshal(extras)
						notify(NTUINotification, map[string]interface{}{
							"type":  "wr_removed",
							"text":  ntfn.Message,
							"count": 0,
							"from":  string(fromJSON),
						}, nil)

					case pong.NotificationType_PLAYER_JOINED_WR:
						var wr *waitingRoom
						if ntfn.Wr != nil {
							players := make([]*player, len(ntfn.Wr.Players))
							for i, p := range ntfn.Wr.Players {
								pp, _ := playerFromServer(p)
								players[i] = pp
							}
							wr = &waitingRoom{ID: ntfn.Wr.Id, HostID: ntfn.Wr.HostId, BetAmt: ntfn.Wr.BetAmt, Players: players}
						}
						extras := map[string]interface{}{"waiting_room": wr, "player_id": ntfn.PlayerId}
						fromJSON, _ := json.Marshal(extras)
						notify(NTUINotification, map[string]interface{}{
							"type":  "player_joined_wr",
							"text":  ntfn.Message,
							"count": 0,
							"from":  string(fromJSON),
						}, nil)

					case pong.NotificationType_OPPONENT_DISCONNECTED:
						var wr *waitingRoom
						if ntfn.Wr != nil {
							players := make([]*player, len(ntfn.Wr.Players))
							for i, p := range ntfn.Wr.Players {
								pp, _ := playerFromServer(p)
								players[i] = pp
							}
							wr = &waitingRoom{ID: ntfn.Wr.Id, HostID: ntfn.Wr.HostId, BetAmt: ntfn.Wr.BetAmt, Players: players}
						}
						extras := map[string]interface{}{"waiting_room": wr, "player_id": ntfn.PlayerId}
						fromJSON, _ := json.Marshal(extras)
						notify(NTUINotification, map[string]interface{}{
							"type":  "player_left_wr",
							"text":  ntfn.Message,
							"count": 0,
							"from":  string(fromJSON),
						}, nil)

					case pong.NotificationType_GAME_START:
						extras := map[string]interface{}{"game_id": ntfn.GameId}
						fromJSON, _ := json.Marshal(extras)
						notify(NTUINotification, map[string]interface{}{
							"type":  "game_started",
							"text":  "",
							"count": 0,
							"from":  string(fromJSON),
						}, nil)

					case pong.NotificationType_GAME_READY_TO_PLAY:
						extras := map[string]interface{}{"game_id": ntfn.GameId}
						fromJSON, _ := json.Marshal(extras)
						notify(NTUINotification, map[string]interface{}{
							"type":  "game_ready_to_play",
							"text":  "Game is ready! Signal when you're ready to play.",
							"count": 0,
							"from":  string(fromJSON),
						}, nil)

					case pong.NotificationType_COUNTDOWN_UPDATE:
						extras := map[string]interface{}{"game_id": ntfn.GameId, "message": ntfn.Message}
						fromJSON, _ := json.Marshal(extras)
						notify(NTUINotification, map[string]interface{}{
							"type":  "countdown_update",
							"text":  ntfn.Message,
							"count": 0,
							"from":  string(fromJSON),
						}, nil)

					case pong.NotificationType_READY_TIMEOUT_HINT:
						// Structured ready-timeout hint with seconds provided by server.
						extras := map[string]interface{}{
							"game_id": ntfn.GameId,
							"message": ntfn.Message,
							"seconds": ntfn.ReadyTimeoutSeconds,
						}
						fromJSON, _ := json.Marshal(extras)
						notify(NTUINotification, map[string]interface{}{
							"type":  "ready_timeout",
							"text":  ntfn.Message,
							"count": int(ntfn.ReadyTimeoutSeconds),
							"from":  string(fromJSON),
						}, nil)

					case pong.NotificationType_ON_PLAYER_READY:
						// Include updated waiting room snapshot when available so UIs
						// can reflect both players' ready state.
						var wr *waitingRoom
						if ntfn.Wr != nil {
							players := make([]*player, len(ntfn.Wr.Players))
							for i, p := range ntfn.Wr.Players {
								pp, _ := playerFromServer(p)
								players[i] = pp
							}
							wr = &waitingRoom{
								ID:      ntfn.Wr.Id,
								HostID:  ntfn.Wr.HostId,
								BetAmt:  ntfn.Wr.BetAmt,
								Players: players,
							}
						}
						extras := map[string]interface{}{
							"waiting_room": wr,
							"player_id":    ntfn.PlayerId,
							"ready":        ntfn.Ready,
						}
						fromJSON, _ := json.Marshal(extras)
						notify(NTUINotification, map[string]interface{}{
							"type":  "player_ready",
							"text":  "",
							"count": 0,
							"from":  string(fromJSON),
						}, nil)

					case pong.NotificationType_GAME_END:
						extras := map[string]interface{}{"game_id": ntfn.GameId}
						fromJSON, _ := json.Marshal(extras)
						notify(NTUINotification, map[string]interface{}{
							"type":  "game_end",
							"text":  ntfn.Message,
							"count": 0,
							"from":  string(fromJSON),
						}, nil)

					case pong.NotificationType_MESSAGE:
						notify(NTUINotification, map[string]interface{}{
							"type": "message",
							"text": ntfn.Message,
						}, nil)

					}
				} else if gub, ok := msg.(*pong.GameUpdateBytes); ok {
					// Emit binary frame for high-frequency updates only.
					notify(NTGameFrame, gub.Data, nil)
					fwd++
					if time.Since(lastLog) >= time.Second {
						fwd = 0
						lastLog = time.Now()
					}
				}
			case err := <-pc.ErrorsCh():
				if err != nil {
					log.Errorf("PongClient error: %v", err)
				}
			}
		}
	}()

	go func() {
		// Handle client closure and errors
		if err := g.Wait(); err != nil {
			fmt.Printf("err: %+v\n\n", err)
			cctx.runMtx.Lock()
			cctx.runErr = err
			cctx.runMtx.Unlock()

			// Clean up the client if it stops running
			cmtx.Lock()
			delete(cs, handle)
			cmtx.Unlock()

			// Notify the system that the client stopped
			notify(NTClientStopped, nil, err)
		}
	}()

	return li, nil
}

func handleClientCmd(cc *clientCtx, cmd *cmd) (interface{}, error) {
	chat := cc.chat

	switch cmd.Type {
	case CTGetUserNick:
		resp := &types.UserNickResponse{}
		hexUid := string(cmd.Payload)
		err := chat.UserNick(cc.ctx, &types.UserNickRequest{
			HexUid: strings.Trim(hexUid, `"`),
		}, resp)
		if err != nil {
			return nil, err
		}
		return resp.Nick, nil
	case CTGetWRPlayers:
		// Not supported via client API; return empty for now
		return []*player{}, nil
	case CTGetWaitingRooms:
		rooms, err := cc.c.RefGetWaitingRooms(cc.ctx)
		if err != nil {
			return nil, err
		}
		res := make([]*waitingRoom, len(rooms))
		for i, r := range rooms {
			players := make([]*player, len(r.Players))
			for i, p := range r.Players {
				var id zkidentity.ShortID
				err := id.FromString(p.Uid)
				if err != nil {
					return nil, err
				}

				players[i], err = playerFromServer(p)
				if err != nil {
					return nil, err
				}
			}
			res[i] = &waitingRoom{
				ID:      r.Id,
				HostID:  r.HostId,
				BetAmt:  r.BetAmt,
				Players: players,
			}
		}
		return res, nil
	case CTJoinWaitingRoom:
		// Accept either raw string room_id or JSON with escrow_id
		var roomID string
		var req joinWaitingRoom
		if err := json.Unmarshal(cmd.Payload, &req); err == nil && req.RoomID != "" {
			roomID = req.RoomID
			res, err := cc.c.RefJoinWaitingRoom(roomID, req.EscrowId)
			if err != nil {
				return nil, err
			}
			return &waitingRoom{
				ID:     res.Wr.Id,
				HostID: res.Wr.HostId,
				BetAmt: res.Wr.BetAmt,
			}, nil
		}
		roomID = string(bytes.Trim(cmd.Payload, "\""))
		res, err := cc.c.RefJoinWaitingRoom(roomID, "")
		if err != nil {
			return nil, err
		}
		return &waitingRoom{
			ID:     res.Wr.Id,
			HostID: res.Wr.HostId,
			BetAmt: res.Wr.BetAmt,
		}, nil

	case CTCreateWaitingRoom:
		args := cmd.Payload

		var req createWaitingRoom
		err := json.Unmarshal(args, &req)
		if err != nil {
			return nil, fmt.Errorf("invalid create waiting room payload: %v", err)
		}

		// EscrowId is optional; empty string lets server auto-pick
		res, err := cc.c.RefCreateWaitingRoom(req.ClientID, req.BetAmt, req.EscrowId)
		if err != nil {
			return nil, fmt.Errorf("failed to create waiting room: %v", err)
		}

		players := make([]*player, len(res.Players))
		for i, p := range res.Players {
			players[i], err = playerFromServer(p)
			if err != nil {
				return nil, err
			}
		}
		return &waitingRoom{
			ID:      res.Id,
			HostID:  res.HostId,
			BetAmt:  res.BetAmt,
			Players: players,
		}, nil

	case CTStopClient:
		// Gracefully stop the running client for this handle:
		// - cancel the context so background goroutines exit
		// - close the underlying gRPC connection
		// - remove the clientCtx from the global map so a future
		//   CTInitClient call will fully reinitialize a fresh client.
		if cc.cancel != nil {
			cc.cancel()
		}
		if cc.c != nil {
			_ = cc.c.Close()
		}
		cmtx.Lock()
		for h, c := range cs {
			if c == cc {
				delete(cs, h)
				break
			}
		}
		cmtx.Unlock()
		return nil, nil

	case CTLeaveWaitingRoom:
		id := strings.Trim(string(cmd.Payload), `"`)
		fmt.Printf("Leaving waiting room: %s\n", id)
		err := cc.c.RefLeaveWaitingRoom(id)
		return nil, err

	// Settlement-related commands
	case CTGenerateSessionKey:
		priv, pub, err := cc.c.GenerateNewSettlementSessionKey()
		if err != nil {
			return nil, err
		}
		return map[string]string{"priv": priv, "pub": pub}, nil
	case CTOpenEscrow:
		var req openEscrowReq
		if err := json.Unmarshal(cmd.Payload, &req); err != nil {
			return nil, fmt.Errorf("bad open escrow payload: %v", err)
		}
		// Reject xpub inputs - require wallet authentication for P2PK address
		payout := strings.TrimSpace(req.Payout)
		if strings.HasPrefix(payout, "tpub") || strings.HasPrefix(payout, "dpub") {
			return nil, fmt.Errorf("xpub not allowed - please login with wallet to use authenticated P2PK address")
		}
		params, err := cc.c.GetChainParams()
		if err != nil {
			return nil, fmt.Errorf("failed to get chain params: %v", err)
		}
		payoutPK33, err := pongbisonrelay.PayoutPubkeyFromConfHex(req.Payout, params)
		if err != nil {
			return nil, fmt.Errorf("payout key parse failed (expected P2PK address from wallet login): %v", err)
		}

		res, err := cc.c.OpenEscrowWithSession(cc.ctx, payoutPK33, req.BetAtoms, req.CSVBlocks)
		if err != nil {
			return nil, err
		}
		pubHex, err := cc.c.CurrentSettlementPubKey()
		if err != nil {
			return nil, err
		}
		var redeemHex string
		if pubBytes, derr := hex.DecodeString(pubHex); derr == nil {
			if redeem, rerr := pongbisonrelay.BuildPerDepositorRedeemScript(pubBytes, req.CSVBlocks); rerr == nil {
				redeemHex = hex.EncodeToString(redeem)
			}
		}
		if redeemHex == "" {
			return nil, fmt.Errorf("failed to derive redeem script for escrow")
		}
		return map[string]any{
			"escrow_id":         res.EscrowId,
			"deposit_address":   res.DepositAddress,
			"pk_script_hex":     res.PkScriptHex,
			"redeem_script_hex": redeemHex,
			"csv_blocks":        req.CSVBlocks,
		}, nil

	case CTStartPreSign:
		var req preSignReq
		if err := json.Unmarshal(cmd.Payload, &req); err != nil {
			return nil, fmt.Errorf("bad presign payload: %v", err)
		}
		fmt.Printf("start presign: match_id=%q\n", req.MatchID)
		if err := cc.c.RefStartSettlementHandshake(cc.ctx, req.MatchID); err != nil {
			fmt.Printf("presign failed: %v\n", err)
			return nil, err
		}
		return map[string]string{"status": "ok"}, nil

	case CTArchiveSessionKey:
		var req struct {
			MatchID    string                 `json:"match_id"`
			EscrowInfo map[string]interface{} `json:"escrow_info,omitempty"`
		}
		if err := json.Unmarshal(cmd.Payload, &req); err != nil {
			return nil, fmt.Errorf("bad archive payload: %v", err)
		}
		if req.EscrowInfo == nil {
			return nil, fmt.Errorf("archive payload requires escrow_info with funding details")
		}

		// Convert map to EscrowInfo struct and validate required fields.
		escrowInfo := &client.EscrowInfo{}
		if id, ok := req.EscrowInfo["escrow_id"].(string); ok {
			escrowInfo.EscrowID = id
		}
		if txid, ok := req.EscrowInfo["funding_txid"].(string); ok {
			escrowInfo.FundingTxid = txid
		}
		hasVout := false
		if vout, ok := req.EscrowInfo["funding_vout"].(float64); ok {
			escrowInfo.FundingVout = uint32(vout)
			hasVout = true
		}
		hasAmount := false
		if amount, ok := req.EscrowInfo["funded_amount"].(float64); ok {
			escrowInfo.FundedAmount = uint64(amount)
			hasAmount = true
		}
		if redeem, ok := req.EscrowInfo["redeem_script_hex"].(string); ok {
			escrowInfo.RedeemScriptHex = redeem
		}
		if pk, ok := req.EscrowInfo["pk_script_hex"].(string); ok {
			escrowInfo.PKScriptHex = pk
		}
		hasCSV := false
		if csv, ok := req.EscrowInfo["csv_blocks"].(float64); ok {
			escrowInfo.CSVBlocks = uint32(csv)
			hasCSV = true
		}
		if archived, ok := req.EscrowInfo["archived_at"].(float64); ok {
			escrowInfo.ArchivedAt = int64(archived)
		}

		switch {
		case escrowInfo.EscrowID == "":
			return nil, fmt.Errorf("escrow_info missing escrow_id")
		case escrowInfo.FundingTxid == "":
			return nil, fmt.Errorf("escrow_info missing funding_txid")
		case !hasVout:
			return nil, fmt.Errorf("escrow_info missing funding_vout")
		case !hasAmount:
			return nil, fmt.Errorf("escrow_info missing funded_amount")
		case escrowInfo.RedeemScriptHex == "":
			return nil, fmt.Errorf("escrow_info missing redeem_script_hex")
		case escrowInfo.PKScriptHex == "":
			return nil, fmt.Errorf("escrow_info missing pk_script_hex")
		case !hasCSV:
			return nil, fmt.Errorf("escrow_info missing csv_blocks")
		}
		if escrowInfo.ArchivedAt == 0 {
			escrowInfo.ArchivedAt = time.Now().Unix()
		}

		if err := cc.c.ArchiveSettlementSessionKeyWithEscrow(req.MatchID, escrowInfo); err != nil {
			return nil, err
		}
		return map[string]string{"status": "archived"}, nil

	case CTCacheEscrowInfo:
		var payload map[string]interface{}
		if err := json.Unmarshal(cmd.Payload, &payload); err != nil {
			return nil, fmt.Errorf("bad cache escrow payload: %v", err)
		}
		info := &client.EscrowInfo{}
		if id, ok := payload["escrow_id"].(string); ok {
			info.EscrowID = id
		}
		if info.EscrowID == "" {
			return nil, fmt.Errorf("cache escrow payload missing escrow_id")
		}
		if addr, ok := payload["deposit_address"].(string); ok {
			info.DepositAddress = addr
		}
		if txid, ok := payload["funding_txid"].(string); ok {
			info.FundingTxid = txid
		}
		if v, exists := payload["funding_vout"]; exists {
			if vout, ok := v.(float64); ok {
				info.FundingVout = uint32(vout)
				info.FundingVoutSet = true
			}
		}
		if v, exists := payload["funded_amount"]; exists {
			if amount, ok := v.(float64); ok {
				info.FundedAmount = uint64(amount)
				info.FundedAmountSet = true
			}
		}
		if redeem, ok := payload["redeem_script_hex"].(string); ok {
			info.RedeemScriptHex = redeem
		}
		if pk, ok := payload["pk_script_hex"].(string); ok {
			info.PKScriptHex = pk
		}
		if v, exists := payload["csv_blocks"]; exists {
			if csv, ok := v.(float64); ok {
				info.CSVBlocks = uint32(csv)
				info.CSVBlocksSet = true
			}
		}
		if archived, ok := payload["archived_at"].(float64); ok {
			info.ArchivedAt = int64(archived)
		}
		if err := cc.c.CacheEscrowInfo(info); err != nil {
			return nil, err
		}
		return map[string]string{"status": "cached"}, nil

	case CTUpdateHistoricEscrow:
		var payload map[string]interface{}
		if err := json.Unmarshal(cmd.Payload, &payload); err != nil {
			return nil, fmt.Errorf("bad update historic escrow payload: %v", err)
		}
		info := &client.EscrowInfo{}
		if id, ok := payload["escrow_id"].(string); ok {
			info.EscrowID = strings.TrimSpace(id)
		}
		if info.EscrowID == "" {
			return nil, fmt.Errorf("update historic escrow payload missing escrow_id")
		}
		if txid, ok := payload["funding_txid"].(string); ok {
			info.FundingTxid = strings.TrimSpace(txid)
		}
		if v, exists := payload["funding_vout"]; exists {
			switch vv := v.(type) {
			case float64:
				info.FundingVout = uint32(vv)
				info.FundingVoutSet = true
			case int:
				info.FundingVout = uint32(vv)
				info.FundingVoutSet = true
			case int32:
				info.FundingVout = uint32(vv)
				info.FundingVoutSet = true
			case int64:
				info.FundingVout = uint32(vv)
				info.FundingVoutSet = true
			}
		}
		if amount, exists := payload["funded_amount"]; exists {
			switch av := amount.(type) {
			case float64:
				info.FundedAmount = uint64(av)
				info.FundedAmountSet = true
			case int:
				info.FundedAmount = uint64(av)
				info.FundedAmountSet = true
			case int32:
				info.FundedAmount = uint64(av)
				info.FundedAmountSet = true
			case int64:
				info.FundedAmount = uint64(av)
				info.FundedAmountSet = true
			}
		}
		if redeem, ok := payload["redeem_script_hex"].(string); ok {
			info.RedeemScriptHex = strings.TrimSpace(redeem)
		}
		if pk, ok := payload["pk_script_hex"].(string); ok {
			info.PKScriptHex = strings.TrimSpace(pk)
		}
		if csv, exists := payload["csv_blocks"]; exists {
			switch cv := csv.(type) {
			case float64:
				info.CSVBlocks = uint32(cv)
				info.CSVBlocksSet = true
			case int:
				info.CSVBlocks = uint32(cv)
				info.CSVBlocksSet = true
			case int32:
				info.CSVBlocks = uint32(cv)
				info.CSVBlocksSet = true
			case int64:
				info.CSVBlocks = uint32(cv)
				info.CSVBlocksSet = true
			}
		}
		if archived, exists := payload["archived_at"]; exists {
			switch av := archived.(type) {
			case float64:
				info.ArchivedAt = int64(av)
			case int:
				info.ArchivedAt = int64(av)
			case int32:
				info.ArchivedAt = int64(av)
			case int64:
				info.ArchivedAt = av
			}
		}
		if err := cc.c.UpdateHistoricEscrow(info); err != nil {
			return nil, err
		}
		return map[string]string{"status": "updated"}, nil

	case CTDeleteHistoricEscrow:
		var req deleteHistoricEscrowReq
		if err := json.Unmarshal(cmd.Payload, &req); err != nil {
			return nil, fmt.Errorf("bad delete historic escrow payload: %v", err)
		}
		escrowID := strings.TrimSpace(req.EscrowID)
		if escrowID == "" {
			return nil, fmt.Errorf("delete historic escrow payload missing escrow_id")
		}
		if err := cc.c.DeleteHistoricEscrow(escrowID); err != nil {
			return nil, fmt.Errorf("failed to delete historic escrow: %w", err)
		}
		cc.log.Infof("CTDeleteHistoricEscrow: deleted escrow %s", escrowID)
		return map[string]string{"status": "deleted"}, nil

	case CTListHistoricEscrows:
		escrows, err := cc.c.LoadHistoricEscrows()
		if err != nil {
			return nil, fmt.Errorf("failed to load historic escrows: %v", err)
		}

		// Convert to a format the UI can use
		result := make([]map[string]interface{}, 0)
		for _, escrow := range escrows {
			result = append(result, map[string]interface{}{
				"escrow_id":         escrow.EscrowID,
				"funding_txid":      escrow.FundingTxid,
				"funding_vout":      escrow.FundingVout,
				"funded_amount":     escrow.FundedAmount,
				"redeem_script_hex": escrow.RedeemScriptHex,
				"pk_script_hex":     escrow.PKScriptHex,
				"csv_blocks":        escrow.CSVBlocks,
				"archived_at":       escrow.ArchivedAt,
			})
		}

		cc.log.Infof("CTListHistoricEscrows: returning %d escrows", len(result))

		return map[string]interface{}{
			"escrows": result,
		}, nil

	case CTCacheWalletAuthInfo:
		var payload map[string]interface{}
		if err := json.Unmarshal(cmd.Payload, &payload); err != nil {
			return nil, fmt.Errorf("bad cache wallet auth payload: %v", err)
		}
		walletAddr := ""
		payoutAddr := ""
		if addr, ok := payload["wallet_address"].(string); ok {
			walletAddr = strings.TrimSpace(addr)
		}
		if addr, ok := payload["payout_address_or_pubkey"].(string); ok {
			payoutAddr = strings.TrimSpace(addr)
		}
		if err := cc.c.CacheWalletAuthInfo(walletAddr, payoutAddr); err != nil {
			return nil, err
		}
		return map[string]string{"status": "cached"}, nil

	case CTGetWalletAuthInfo:
		walletAddr, payoutAddr := cc.c.GetWalletAuthInfo()
		return map[string]string{
			"wallet_address":           walletAddr,
			"payout_address_or_pubkey": payoutAddr,
		}, nil

	case CTGetActiveEscrowInfo:
		info := cc.c.GetActiveEscrowInfo()
		if info == nil {
			return map[string]interface{}{}, nil
		}
		result := map[string]interface{}{
			"escrow_id": info.EscrowID,
		}
		if info.DepositAddress != "" {
			result["deposit_address"] = info.DepositAddress
		}
		if info.FundingTxid != "" {
			result["funding_txid"] = info.FundingTxid
		}
		if info.FundingVoutSet || info.FundingVout != 0 {
			result["funding_vout"] = info.FundingVout
		}
		if info.FundedAmountSet || info.FundedAmount != 0 {
			result["funded_amount"] = info.FundedAmount
		}
		if info.RedeemScriptHex != "" {
			result["redeem_script_hex"] = info.RedeemScriptHex
		}
		if info.PKScriptHex != "" {
			result["pk_script_hex"] = info.PKScriptHex
		}
		if info.CSVBlocksSet || info.CSVBlocks != 0 {
			result["csv_blocks"] = info.CSVBlocks
		}
		if info.Status != "" {
			result["status"] = info.Status
		}
		if info.ArchivedAt != 0 {
			result["archived_at"] = info.ArchivedAt
		}
		return result, nil

	case CTGetCurrentWaitingRoom:
		roomID := ""
		if cc.c != nil {
			roomID = cc.c.CurrentWaitingRoomID()
		}
		return map[string]interface{}{"room_id": roomID}, nil

	case CTValidateRefundSession:
		var req struct {
			EscrowID string `json:"escrow_id"`
		}
		if err := json.Unmarshal(cmd.Payload, &req); err != nil {
			return nil, fmt.Errorf("bad validate refund payload: %v", err)
		}
		ok, reason := cc.c.ValidateHistoricRefundSession(req.EscrowID)
		return map[string]any{"ok": ok, "reason": reason}, nil

	// Player action commands
	case CTSendInput:
		var req struct {
			Input string `json:"input"`
		}
		if err := json.Unmarshal(cmd.Payload, &req); err != nil {
			return nil, fmt.Errorf("bad send input payload: %v", err)
		}
		if err := cc.c.RefSendInput(req.Input); err != nil {
			return nil, err
		}
		return map[string]string{"status": "ok"}, nil

	case CTSignalReadyToPlay:
		var req struct {
			GameID string `json:"game_id"`
		}
		if err := json.Unmarshal(cmd.Payload, &req); err != nil {
			return nil, fmt.Errorf("bad signal ready payload: %v", err)
		}
		if err := cc.c.RefSignalReadyToPlay(req.GameID); err != nil {
			return nil, err
		}
		return map[string]bool{"success": true}, nil

	case CTUnreadyGameStream:
		if err := cc.c.RefUnreadyGameStream(); err != nil {
			return nil, err
		}
		return map[string]string{"status": "ok"}, nil

	case CTStartGameStream:
		if err := cc.c.RefStartGameStream(); err != nil {
			return nil, err
		}
		return map[string]string{"status": "ok"}, nil

	case CTRefundEscrow:
		var req refundEscrowReq
		if err := json.Unmarshal(cmd.Payload, &req); err != nil {
			return nil, fmt.Errorf("bad refund escrow payload: %v", err)
		}

		// Get escrow details
		details, err := cc.c.GetEscrowDetails(req.EscrowID)
		if err != nil {
			return &refundEscrowRes{
				CanRefund: false,
				Reason:    fmt.Sprintf("failed to get escrow details: %v", err),
			}, nil
		}

		// Get settlement private key for this escrow
		privHex, err := cc.c.GetSettlementPrivKeyForEscrow(req.EscrowID)
		if err != nil {
			return &refundEscrowRes{
				CanRefund: false,
				Reason:    fmt.Sprintf("failed to get settlement private key: %v", err),
			}, nil
		}

		// Get chain parameters
		params, err := cc.c.GetChainParams()
		if err != nil {
			return &refundEscrowRes{
				CanRefund: false,
				Reason:    fmt.Sprintf("failed to get chain params: %v", err),
			}, nil
		}
		if params == nil {
			return &refundEscrowRes{
				CanRefund: false,
				Reason:    "chain params is nil",
			}, nil
		}
		// Log the network being used for debugging
		cc.log.Infof("RefundEscrow: using params.Name=%s for address=%s", params.Name, req.DestAddr)

		// Use CSV blocks from request if provided, otherwise use stored value
		csvBlocks := req.CSVBlocks
		if csvBlocks == 0 {
			csvBlocks = details.CSVBlocks
		}

		// Use fee from request if provided, otherwise default to 20000
		feeAtoms := req.FeeAtoms
		if feeAtoms == 0 {
			feeAtoms = 20000
		}

		// Use UTXO value from request if provided, otherwise use stored value
		utxoValue := details.FundedAmount
		if req.UtxoValue > 0 {
			utxoValue = req.UtxoValue
			cc.log.Infof("RefundEscrow: using provided utxo_value=%d instead of stored funded_amount=%d", req.UtxoValue, details.FundedAmount)
		}

		// Build the refund transaction
		refundTxHex, err := client.BuildCSVRefundTx(
			privHex,
			details.FundingTxHash,
			details.FundingVout,
			utxoValue,
			details.RedeemScriptHex,
			req.DestAddr,
			feeAtoms,
			csvBlocks,
			params,
		)
		if err != nil {
			return &refundEscrowRes{
				CanRefund: false,
				Reason:    fmt.Sprintf("failed to build refund transaction: %v", err),
			}, nil
		}

		// Return success response
		return &refundEscrowRes{
			RefundTxHex: refundTxHex,
			UTXOTxid:    details.FundingTxHash,
			UTXOVout:    details.FundingVout,
			UTXOValue:   utxoValue,
			RedeemHex:   details.RedeemScriptHex,
			CSVBlocks:   csvBlocks,
			CanRefund:   true,
		}, nil
	}
	return nil, nil
}

func handleCreateLockFile(rootDir string) error {
	filePath := filepath.Join(rootDir, clientintf.LockFileName)

	cmtx.Lock()
	defer cmtx.Unlock()

	lf := lfs[filePath]
	if lf != nil {
		// Already running on this DB from this process.
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	lf, err := lockfile.Create(ctx, filePath)
	cancel()
	if err != nil {
		return fmt.Errorf("unable to create lockfile %q: %v", filePath, err)
	}
	lfs[filePath] = lf
	return nil
}

func handleCloseLockFile(rootDir string) error {
	filePath := filepath.Join(rootDir, clientintf.LockFileName)

	cmtx.Lock()
	lf := lfs[filePath]
	delete(lfs, filePath)
	cmtx.Unlock()

	if lf == nil {
		return fmt.Errorf("nil lockfile")
	}
	return lf.Close()
}

// --- Wallet-auth handlers (no running client required) ---

// handleRequestNonceForHandle prefers the running client's server config for the given handle.
func handleRequestNonce(handle uint32, args requestNonceArgs) (interface{}, error) {
	// Prefer the active client's server address and cert path if a client exists.
	cmtx.Lock()
	var cctx *clientCtx
	if cs != nil {
		cctx = cs[handle]
	}
	cmtx.Unlock()

	serverAddr := strings.TrimSpace(args.ServerAddr)
	grpcCertPath := strings.TrimSpace(args.GRPCCertPath)
	if cctx != nil && cctx.c != nil {
		if sa := strings.TrimSpace(cctx.c.ServerAddr()); sa != "" {
			serverAddr = sa
		}
		if cp := strings.TrimSpace(cctx.c.GRPCCertPath()); cp != "" {
			grpcCertPath = cp
		}
	}
	if serverAddr == "" {
		return nil, fmt.Errorf("missing server_addr")
	}

	// Build TLS creds using consolidated cert path.
	pc := &client.PongConf{GRPCCertPath: grpcCertPath}
	creds, err := client.LoadTLSCreds(pc)
	if err != nil {
		return nil, fmt.Errorf("load TLS creds: %w", err)
	}
	conn, err := grpc.NewClient(serverAddr, grpc.WithTransportCredentials(creds))
	if err != nil {
		return nil, fmt.Errorf("dial server: %w", err)
	}
	defer conn.Close()

	c := pong.NewPongAuthClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := c.RequestNonce(ctx, &pong.RequestNonceRequest{})
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"nonce":        res.GetNonce(),
		"ttl_sec":      res.GetTtlSec(),
		"address_hint": res.GetAddressHint(),
	}, nil
}

// Handle GetClientConfig using running client's app config if available.
func handleGetClientConfigForHandle(handle uint32, dataDir string) (interface{}, error) {
	cmtx.Lock()
	var cctx *clientCtx
	if cs != nil {
		cctx = cs[handle]
	}
	cmtx.Unlock()
	if cctx != nil && cctx.c != nil {
		pc := cctx.c.AppConfig()
		if pc != nil {
			return map[string]any{
				"server_addr":      pc.ServerAddr,
				"grpc_cert_path":   pc.GRPCCertPath,
				"debug":            pc.Debug,
				"show_perfoverlay": pc.ShowPerfOverlay,
				"data_dir":         pc.DataDir,
			}, nil
		}
	}
	// Fallback to default behavior.
	return handleGetClientConfig(dataDir)
}

// handleVerifyLoginForHandle uses the running client's server config when available
// so VerifyLogin is performed against the same server used for gameplay and escrow.
func handleVerifyLogin(handle uint32, args verifyLoginArgs) (interface{}, error) {
	// Prefer the active client's server address and cert path if a client exists.
	cmtx.Lock()
	var cctx *clientCtx
	if cs != nil {
		cctx = cs[handle]
	}
	cmtx.Unlock()

	serverAddr := strings.TrimSpace(args.ServerAddr)
	grpcCertPath := strings.TrimSpace(args.GRPCCertPath)
	if cctx != nil && cctx.c != nil {
		if sa := strings.TrimSpace(cctx.c.ServerAddr()); sa != "" {
			serverAddr = sa
		}
		if cp := strings.TrimSpace(cctx.c.GRPCCertPath()); cp != "" {
			grpcCertPath = cp
		}
	}
	if serverAddr == "" {
		return nil, fmt.Errorf("missing server or cert path")
	}
	if strings.TrimSpace(args.Address) == "" || strings.TrimSpace(args.Nonce) == "" || strings.TrimSpace(args.Signature) == "" {
		return nil, fmt.Errorf("missing address, nonce or signature")
	}

	// Build TLS creds using consolidated cert path.
	pc := &client.PongConf{GRPCCertPath: grpcCertPath}
	creds, err := client.LoadTLSCreds(pc)
	if err != nil {
		return nil, fmt.Errorf("load TLS creds: %w", err)
	}
	conn, err := grpc.NewClient(serverAddr, grpc.WithTransportCredentials(creds))
	if err != nil {
		return nil, fmt.Errorf("dial server: %w", err)
	}
	defer conn.Close()

	c := pong.NewPongAuthClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	res, err := c.VerifyLogin(ctx, &pong.VerifyLoginRequest{
		Address:   args.Address,
		Nonce:     args.Nonce,
		Signature: args.Signature,
	})
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"ok":          res.GetOk(),
		"token":       res.GetToken(),
		"client_id":   res.GetClientId(),
		"comp_pubkey": res.GetCompPubkey(),
		"p2pk_addr":   res.GetP2PkAddr(),
	}, nil
}

// --- Config management (no running client required) ---

type saveClientConfigArgs struct {
	ServerAddr      string `json:"server_addr"`
	GRPCCertPath    string `json:"grpc_cert_path"`
	Network         string `json:"network"`
	Debug           string `json:"debug"`
	ShowPerfOverlay bool   `json:"show_perfoverlay"`
	DataDir         string `json:"data_dir"`
}

func handleGetClientConfig(dataDir string) (interface{}, error) {
	// Load current config strictly from provided data dir (no fallback).
	dir := strings.TrimSpace(dataDir)
	if dir == "" {
		return map[string]any{
			"server_addr":      "",
			"grpc_cert_path":   "",
			"network":          "mainnet",
			"debug":            "info",
			"show_perfoverlay": false,
			"data_dir":         "",
		}, nil
	}
	appCfg, err := client.LoadAppConfig(dir, appName)
	if err != nil {
		// On first run, config might not exist yet. Return default values
		// so the UI can show the config page without errors.
		defaultDataDir := dir
		return map[string]any{
			"server_addr":      "",
			"grpc_cert_path":   "",
			"network":          "mainnet",
			"debug":            "info",
			"show_perfoverlay": false,
			"data_dir":         defaultDataDir,
		}, nil
	}
	pc := appCfg.PongConf
	return map[string]any{
		"server_addr":      pc.ServerAddr,
		"grpc_cert_path":   pc.GRPCCertPath,
		"debug":            pc.Debug,
		"show_perfoverlay": pc.ShowPerfOverlay,
		"data_dir":         pc.DataDir,
	}, nil
}

func handleSaveClientConfig(handle uint32, args saveClientConfigArgs) (interface{}, error) {
	// Load existing to get datadir and defaults.
	loadDir := strings.TrimSpace(args.DataDir)
	if loadDir == "" {
		return nil, fmt.Errorf("missing data_dir")
	}
	appCfg, err := client.LoadAppConfig(loadDir, appName)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}
	pc := appCfg.PongConf

	// If a sandboxed data dir was provided by the UI, honor it.
	if strings.TrimSpace(args.DataDir) != "" {
		pc.DataDir = strings.TrimSpace(args.DataDir)
	}

	// Apply provided fields (empty strings are ignored for strings).
	if strings.TrimSpace(args.ServerAddr) != "" {
		pc.ServerAddr = strings.TrimSpace(args.ServerAddr)
	}
	if strings.TrimSpace(args.GRPCCertPath) != "" {
		pc.GRPCCertPath = strings.TrimSpace(args.GRPCCertPath)
	}
	if strings.TrimSpace(args.Debug) != "" {
		pc.Debug = strings.TrimSpace(args.Debug)
	}
	pc.ShowPerfOverlay = args.ShowPerfOverlay

	confPath := filepath.Join(pc.DataDir, appName+".conf")
	if err := client.WriteClientConfigFile(pc, confPath); err != nil {
		return nil, fmt.Errorf("write config: %w", err)
	}

	// Refresh the cached configuration for the requesting client (if running) so
	// subsequent reads observe the newly saved values.
	cmtx.Lock()
	cctx := cs[handle]
	cmtx.Unlock()

	if cctx != nil && cctx.c != nil {
		current := cctx.c.AppConfig()
		if current != nil && filepath.Clean(current.DataDir) == filepath.Clean(pc.DataDir) {
			cctx.c.UpdateAppConfig(pc)
		}
	}

	return map[string]string{"status": "ok"}, nil
}
