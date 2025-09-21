package main

import (
	"context"
	"encoding/hex"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/companyzero/bisonrelay/clientrpc/types"
	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/vctt94/bisonbotkit"
	"github.com/vctt94/bisonbotkit/logging"
	"github.com/vctt94/bisonbotkit/utils"
	"github.com/vctt94/pongbisonrelay/pongrpc/grpc/pong"
	"github.com/vctt94/pongbisonrelay/server"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
)

var (
	flagDataDir        = flag.String("datadir", "", "Directory for server data (certificates, keys, etc.)")
	flagIsF2P          = flag.Bool("isf2p", false, "Enable free-to-play mode")
	flagMinBetAmt      = flag.Float64("minbetamt", 0, "Minimum bet amount")
	flagRPCURL         = flag.String("rpcurl", "", "URL of the RPC server")
	flagGRPCHost       = flag.String("grpchost", "", "Host for gRPC server")
	flagGRPCPort       = flag.String("grpcport", "", "Port for gRPC server")
	flagServerCertPath = flag.String("servercert", "", "Path to server certificate")
	flagClientCertPath = flag.String("clientcert", "", "Path to client certificate")
	flagClientKeyPath  = flag.String("clientkey", "", "Path to client key")
	flagRPCUser        = flag.String("rpcuser", "", "RPC user")
	flagRPCPass        = flag.String("rpcpass", "", "RPC password")
	flagDebug          = flag.String("debug", "", "Debug level")

	// dcrd connectivity (optional)
	flagDcrdHost = flag.String("dcrdhost", "", "dcrd host:port (e.g. 127.0.0.1:19109)")
	flagDcrdCert = flag.String("dcrdcert", "", "Path to dcrd rpc.cert")
	flagDcrdUser = flag.String("dcrduser", "", "dcrd RPC user (default: rpcuser)")
	flagDcrdPass = flag.String("dcrdpass", "", "dcrd RPC password (default: rpcpass)")
)

func realMain() error {
	flag.Parse()

	var appdata string
	if *flagDataDir != "" {
		appdata = utils.CleanAndExpandPath(*flagDataDir)
	} else {
		appdata = utils.AppDataDir("pongbot", false)
	}
	cfg, err := LoadPongBotConfig(appdata, "pongbot.conf")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	if cfg == nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	// Create channels for tip events
	tipChan := make(chan types.ReceivedTip)
	tipProgressChan := make(chan types.TipProgressEvent)

	// Assign channels to the bot config
	cfg.BotConfig.TipReceivedChan = tipChan
	cfg.BotConfig.TipProgressChan = tipProgressChan

	logBackend, err := logging.NewLogBackend(logging.LogConfig{
		LogFile:        filepath.Join(appdata, "logs", "pongbot.log"),
		DebugLevel:     cfg.Debug,
		MaxLogFiles:    10,
		MaxBufferLines: 1000,
	})
	if err != nil {
		return fmt.Errorf("NewLogBackend failed: %w", err)
	}

	log := logBackend.Logger("Bot")
	cfg.TipLog = logBackend.Logger("tipprogress")
	cfg.PMLog = logBackend.Logger("pm")
	cfg.TipReceivedLog = logBackend.Logger("tipreceived")

	if *flagIsF2P {
		cfg.IsF2P = *flagIsF2P
	}
	if *flagMinBetAmt != 0 {
		cfg.MinBetAmt = *flagMinBetAmt
	}
	if *flagRPCURL != "" {
		cfg.RPCURL = *flagRPCURL
	}
	if *flagGRPCHost != "" {
		cfg.GRPCHost = *flagGRPCHost
	}
	if *flagGRPCPort != "" {
		cfg.GRPCPort = *flagGRPCPort
	}
	if *flagServerCertPath != "" {
		cfg.ServerCertPath = utils.CleanAndExpandPath(*flagServerCertPath)
	}
	if *flagClientCertPath != "" {
		cfg.ClientCertPath = utils.CleanAndExpandPath(*flagClientCertPath)
	}
	if *flagClientKeyPath != "" {
		cfg.ClientKeyPath = utils.CleanAndExpandPath(*flagClientKeyPath)
	}
	if *flagRPCUser != "" {
		cfg.RPCUser = *flagRPCUser
	}
	if *flagRPCPass != "" {
		cfg.RPCPass = *flagRPCPass
	}
	if *flagDebug != "" {
		cfg.Debug = *flagDebug
	}

	// dcrd flags override config when provided (keep prior behavior)
	if *flagDcrdHost != "" {
		cfg.DcrdHost = *flagDcrdHost
	}
	if *flagDcrdCert != "" {
		cfg.DcrdCert = utils.CleanAndExpandPath(*flagDcrdCert)
	}
	if *flagDcrdUser != "" {
		cfg.DcrdUser = *flagDcrdUser
	}
	if *flagDcrdPass != "" {
		cfg.DcrdPass = *flagDcrdPass
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := os.MkdirAll(cfg.DataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %v", err)
	}

	g, gctx := errgroup.WithContext(ctx)

	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		return fmt.Errorf("failed to listen on gRPC port: %v", err)
	}

	bot, err := bisonbotkit.NewBot(cfg.BotConfig)
	if err != nil {
		return fmt.Errorf("failed to create JSON-RPC client: %w", err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Infof("Received shutdown signal")
		cancel()
	}()

	req := &types.PublicIdentityReq{}
	var publicIdentity types.PublicIdentity
	if err := bot.UserPublicIdentity(ctx, req, &publicIdentity); err != nil {
		return fmt.Errorf("failed to get public identity: %w", err)
	}

	clientID := hex.EncodeToString(publicIdentity.Identity[:])
	var zkShortID zkidentity.ShortID
	copy(zkShortID[:], clientID)

	srv, err := server.NewServer(&zkShortID, server.ServerConfig{
		ServerDir:       cfg.DataDir,
		IsF2P:           cfg.IsF2P,
		MinBetAmt:       cfg.MinBetAmt,
		LogBackend:      logBackend,
		DcrdHostPort:    cfg.DcrdHost,
		DcrdRPCCertPath: cfg.DcrdCert,
		DcrdRPCUser:     cfg.DcrdUser,
		DcrdRPCPass:     cfg.DcrdPass,
		AdaptorSecret:   cfg.AdaptorSecret,
	})
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}

	g.Go(func() error { return srv.Run(gctx) })

	certPath := filepath.Join(cfg.DataDir, "server.cert")
	keyPath := filepath.Join(cfg.DataDir, "server.key")
	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		return fmt.Errorf("failed to load TLS credentials: %w", err)
	}
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		return fmt.Errorf("failed to load TLS credentials: %w", err)
	}

	creds, err := credentials.NewServerTLSFromFile(certPath, keyPath)
	if err != nil {
		return fmt.Errorf("failed to load TLS credentials: %w", err)
	}
	grpcServer := grpc.NewServer(
		grpc.Creds(creds),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime: 30 * time.Second, // If a client sends pings more often than this, the server will send a GOAWAY
		}),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:              30 * time.Second, // Send keepalive pings every X interval
			Timeout:           10 * time.Second, // Wait this long for an ACK before considering the connection dead.
			MaxConnectionIdle: 5 * time.Minute,  // If a connection is idle (no RPCs in flight) for this long, send a GOAWAY and close.
		}),
	)

	pong.RegisterPongGameServer(grpcServer, srv)
	pong.RegisterPongRefereeServer(grpcServer, srv)
	// Register waiting room service (required by clients calling GetWaitingRooms, etc.)
	pong.RegisterPongWaitingRoomServer(grpcServer, srv)

	g.Go(func() error {
		<-gctx.Done()
		log.Info("Stopping gRPC server...")
		grpcServer.GracefulStop()
		return nil
	})

	g.Go(func() error {
		log.Infof("server listening at %v", lis.Addr())
		if err := grpcServer.Serve(lis); err != nil && err != grpc.ErrServerStopped {
			return fmt.Errorf("failed to serve gRPC server: %v", err)
		}
		return nil
	})

	// Wait for shutdown signal
	<-gctx.Done()
	log.Info("Shutting down servers...")

	// Make sure to call bot.Close() during shutdown
	bot.Close()
	log.Info("Server shutdown complete")
	select {
	case <-ctx.Done():
		return nil
	default:
		// proceed
	}

	return g.Wait()
}

func main() {
	err := realMain()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
