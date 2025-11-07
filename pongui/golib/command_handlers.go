package golib

import (
	"bytes"
	"context"
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
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials"
    "golang.org/x/sync/errgroup"
)

const (
	appName = "pongui"
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
	if cs[handle] != nil {
		return cs[handle].ID, nil
	}

	// Ensure the data directory exists first
	if err := os.MkdirAll(args.DataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory %s: %v", args.DataDir, err)
	}

	// Ensure the logs subdirectory exists
	logsDir := filepath.Dir(args.LogFile)
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create logs directory %s: %v", logsDir, err)
	}

	// Load configuration using botclient config
	// cfg, err := config.LoadClientConfig(args.DataDir, "pongui.conf")
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to load config: %v", err)
	// }

	// // Apply overrides from args when available
	// if args.RPCWebsocketURL != "" {
	// 	cfg.RPCURL = args.RPCWebsocketURL
	// }
	// if args.RPCCertPath != "" {
	// 	cfg.BRClientCert = args.RPCCertPath
	// }
	// if args.RPCCLientCertPath != "" {
	// 	cfg.BRClientRPCCert = args.RPCCLientCertPath
	// }
	// if args.RPCCLientKeyPath != "" {
	// 	cfg.BRClientRPCKey = args.RPCCLientKeyPath
	// }
	// if args.RPCUser != "" {
	// 	cfg.RPCUser = args.RPCUser
	// }
	// if args.RPCPass != "" {
	// 	cfg.RPCPass = args.RPCPass
	// }
	// if args.DebugLevel != "" {
	// 	cfg.Debug = args.DebugLevel
	// }

	// // Validate required BR RPC fields.
	// var missing []string
	// if strings.TrimSpace(cfg.RPCURL) == "" {
	// 	missing = append(missing, "brrpcurl")
	// }
	// if strings.TrimSpace(cfg.BRClientCert) == "" {
	// 	missing = append(missing, "brclientcert")
	// }
	// if strings.TrimSpace(cfg.BRClientRPCCert) == "" {
	// 	missing = append(missing, "brclientrpccert")
	// }
	// if strings.TrimSpace(cfg.BRClientRPCKey) == "" {
	// 	missing = append(missing, "brclientrpckey")
	// }
	// if strings.TrimSpace(cfg.RPCUser) == "" {
	// 	missing = append(missing, "rpcuser")
	// }
	// if strings.TrimSpace(cfg.RPCPass) == "" {
	// 	missing = append(missing, "rpcpass")
	// }
	// if len(missing) > 0 {
	// 	return nil, fmt.Errorf("missing required fields in client config: %s", strings.Join(missing, ", "))
	// }

	logBackend, err := logging.NewLogBackend(logging.LogConfig{
		LogFile:        filepath.Join(args.DataDir, "logs", "pongui.log"),
		DebugLevel:     args.DebugLevel,
		MaxLogFiles:    10,
		MaxBufferLines: 1000,
	})
	if err != nil {
		return nil, err
	}
	log := logBackend.Logger("pongui")

	ctx, cancel := context.WithCancel(context.Background())
	g, _ := errgroup.WithContext(ctx)

	// Start a BR RPC client
	// c, err := botclient.NewClient(cfg)
	// if err != nil {
	// 	cancel()
	// 	return nil, fmt.Errorf("failed to create bot client: %v", err)
	// }

	// Start the bot client
	// g.Go(func() error { return c.RPCClient.Run(gctx) })

	// Require wallet-authenticated clientID (no random ID generation).
	if strings.TrimSpace(args.ClientID) == "" {
		cancel()
		return nil, fmt.Errorf("client_id is required - wallet authentication must be completed before initializing client")
	}
	var id zkidentity.ShortID
	if err := id.FromString(args.ClientID); err != nil {
		cancel()
		return nil, fmt.Errorf("invalid client_id format: %v", err)
	}
	localInfo := &localInfo{ID: id, Nick: "anon"}

	// Build consolidated AppConfig for the pong client (without BR auth)
	appCfg := &client.AppConfig{
		DataDir:      args.DataDir,
		BR:           nil,
		ServerAddr:   args.ServerAddr,
		GRPCCertPath: args.GRPCCertPath,
	}
    // Set up NotificationManager to emit UI notifications and forward to Flutter.
    nmgr := client.NewNotificationManager()
    // Enable common UI notifications and shorten emit interval for responsiveness.
    nmgr.UpdateUIConfig(client.UINotificationsConfig{
        GameStarted:          true,
        WRCreated:            true,
        MaxLength:            255,
        CancelEmissionChannel: ctx.Done(),
    })

    pc, err := client.NewPongClient(args.ClientID, &client.PongClientCfg{
        AppCfg:        appCfg,
        Log:           logBackend.Logger("client"),
        Notifications: nmgr,
    })
	if err != nil {
		cancel()
		return nil, err
	}

	cctx := &clientCtx{
		ID:     localInfo,
		ctx:    ctx,
		c:      pc,
		cancel: cancel,
		log:    log,
	}
	cs[handle] = cctx

	// Start the notification stream to receive server notifications
	if err := pc.RefStartNtfnStream(ctx); err != nil {
		cancel()
		cmtx.Lock()
		delete(cs, handle)
		cmtx.Unlock()
		return nil, fmt.Errorf("failed to start notification stream: %w", err)
	}

    // Forward only UI notifications to Flutter via NTUINotification.
    nmgr.Register(client.OnUINotification(func(n client.UINotification) {
        // Forward a simplified payload that matches the Dart struct.
        payload := map[string]interface{}{
            "type":  string(n.Type),
            "text":  n.Text,
            "count": n.Count,
            // Use FromNick as a human-readable source; ensure string type.
            "from":  n.FromNick,
        }
        notify(NTUINotification, payload, nil)
    }))

    // Forward structured state-change events as simplified UINotification payloads.
    go func() {
        for {
            select {
            case <-ctx.Done():
                return
            case msg := <-pc.UpdatesCh():
                if ntfn, ok := msg.(*pong.NtfnStreamResponse); ok {
                    switch ntfn.NotificationType {
                    case pong.NotificationType_BET_AMOUNT_UPDATE:
                        extras := map[string]interface{}{
                            "player_id": ntfn.PlayerId,
                            "bet_amt":   ntfn.BetAmt,
                            "confs":     ntfn.Confs,
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

                    case pong.NotificationType_ON_PLAYER_READY:
                        extras := map[string]interface{}{"player_id": ntfn.PlayerId, "ready": ntfn.Ready}
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

                    }
                } else if gub, ok := msg.(*pong.GameUpdateBytes); ok {
                    // Emit binary frame for high-frequency updates only.
                    notify(NTGameFrame, gub.Data, nil)
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

	return localInfo, nil
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
		cc.cancel()
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
		// Accept only 33/65B hex pubkey or Decred P2PK address from wallet auth
		payoutPK33, err := pongbisonrelay.PayoutPubkeyFromConfHex(req.Payout)
		if err != nil {
			return nil, fmt.Errorf("payout key parse failed (expected P2PK address from wallet login): %v", err)
		}

		res, err := cc.c.OpenEscrowWithSession(cc.ctx, payoutPK33, req.BetAtoms, req.CSVBlocks)
		if err != nil {
			return nil, err
		}
		return map[string]any{
			"escrow_id":       res.EscrowId,
			"deposit_address": res.DepositAddress,
			"pk_script_hex":   res.PkScriptHex,
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
			MatchID string `json:"match_id"`
		}
		if err := json.Unmarshal(cmd.Payload, &req); err != nil {
			return nil, fmt.Errorf("bad archive payload: %v", err)
		}
		if err := cc.c.ArchiveSettlementSessionKey(req.MatchID); err != nil {
			return nil, err
		}
		return map[string]string{"status": "archived"}, nil

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

func handleRequestNonce(args requestNonceArgs) (interface{}, error) {
    if strings.TrimSpace(args.ServerAddr) == "" {
        return nil, fmt.Errorf("missing server_addr")
    }
    if strings.TrimSpace(args.GRPCCertPath) == "" {
        return nil, fmt.Errorf("missing grpc_cert_path")
    }

    creds, err := credentials.NewClientTLSFromFile(args.GRPCCertPath, "")
    if err != nil {
        return nil, fmt.Errorf("load TLS cert: %w", err)
    }
    conn, err := grpc.NewClient(args.ServerAddr, grpc.WithTransportCredentials(creds))
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
        "nonce":         res.GetNonce(),
        "ttl_sec":       res.GetTtlSec(),
        "address_hint":  res.GetAddressHint(),
    }, nil
}

func handleVerifyLogin(args verifyLoginArgs) (interface{}, error) {
    if strings.TrimSpace(args.ServerAddr) == "" || strings.TrimSpace(args.GRPCCertPath) == "" {
        return nil, fmt.Errorf("missing server or cert path")
    }
    if strings.TrimSpace(args.Address) == "" || strings.TrimSpace(args.Nonce) == "" || strings.TrimSpace(args.Signature) == "" {
        return nil, fmt.Errorf("missing address, nonce or signature")
    }

    creds, err := credentials.NewClientTLSFromFile(args.GRPCCertPath, "")
    if err != nil {
        return nil, fmt.Errorf("load TLS cert: %w", err)
    }
    conn, err := grpc.NewClient(args.ServerAddr, grpc.WithTransportCredentials(creds))
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
        "ok":         res.GetOk(),
        "token":      res.GetToken(),
        "client_id":  res.GetClientId(),
        "comp_pubkey": res.GetCompPubkey(),
        "p2pk_addr":  res.GetP2PkAddr(),
    }, nil
}
