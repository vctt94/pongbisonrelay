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
	"github.com/vctt94/bisonbotkit/botclient"
	"github.com/vctt94/bisonbotkit/config"
	"github.com/vctt94/bisonbotkit/logging"
	"github.com/vctt94/pongbisonrelay/client"
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
	cfg, err := config.LoadClientConfig(args.DataDir, "pongui.conf")
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %v", err)
	}

	// Apply overrides from args
	if args.RPCWebsocketURL != "" {
		cfg.RPCURL = args.RPCWebsocketURL
	}
	if args.RPCCertPath != "" {
		cfg.ServerCertPath = args.RPCCertPath
	}
	if args.RPCCLientCertPath != "" {
		cfg.ClientCertPath = args.RPCCLientCertPath
	}
	if args.RPCCLientKeyPath != "" {
		cfg.ClientKeyPath = args.RPCCLientKeyPath
	}
	if args.RPCUser != "" {
		cfg.RPCUser = args.RPCUser
	}
	if args.RPCPass != "" {
		cfg.RPCPass = args.RPCPass
	}

	logBackend, err := logging.NewLogBackend(logging.LogConfig{
		LogFile:        filepath.Join(args.DataDir, "logs", "pongui.log"),
		DebugLevel:     cfg.Debug,
		MaxLogFiles:    10,
		MaxBufferLines: 1000,
	})
	if err != nil {
		return nil, err
	}
	log := logBackend.Logger("pongui")

	ctx, cancel := context.WithCancel(context.Background())
	g, gctx := errgroup.WithContext(ctx)

	// Use botclient instead of manual JSON-RPC client
	c, err := botclient.NewClient(cfg, logBackend)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create bot client: %v", err)
	}

	// Start the bot client
	g.Go(func() error { return c.RPCClient.Run(gctx) })

	// Initialize clientID using botclient
	var publicIdentity types.PublicIdentity
	err = c.Chat.UserPublicIdentity(gctx, &types.PublicIdentityReq{}, &publicIdentity)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to get user public identity: %v", err)
	}
	var id zkidentity.ShortID
	id.FromBytes(publicIdentity.Identity[:])
	localInfo := &localInfo{
		ID:   id,
		Nick: publicIdentity.Nick,
	}

	pc, err := client.NewPongClient(localInfo.ID.String(), &client.PongClientCfg{
		ServerAddr:   args.ServerAddr,
		ChatClient:   c.Chat,
		Log:          logBackend.Logger("client"),
		GRPCCertPath: args.GRPCCertPath,
	})
	if err != nil {
		cancel()
		return nil, err
	}

	cctx := &clientCtx{
		ID:     localInfo,
		ctx:    gctx,
		c:      pc,
		chat:   c.Chat,
		cancel: cancel,
		log:    log,
	}
	cs[handle] = cctx

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
		wrp, err := cc.c.GetWRPlayers()
		if err != nil {
			return nil, err
		}
		res := make([]*player, len(wrp))
		for i, p := range wrp {
			res[i], err = playerFromServer(p)
			if err != nil {
				return nil, err
			}
		}
		return res, nil
	case CTGetWaitingRooms:
		rooms, err := cc.c.GetWaitingRooms()
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
		id := string(bytes.Trim(cmd.Payload, "\""))
		res, err := cc.c.JoinWaitingRoom(id, "")
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

		res, err := cc.c.CreateWaitingRoom(req.ClientID, req.BetAmt, "")
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
		err := cc.c.LeaveWaitingRoom(id)
		return nil, err
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
