package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/lola-os/lola-core/internal/budget"
	"github.com/lola-os/lola-core/internal/config"
	"github.com/lola-os/lola-core/internal/hitl"
	"github.com/lola-os/lola-core/internal/idempotency"
	"github.com/lola-os/lola-core/internal/jsonrpc"
	"github.com/lola-os/lola-core/internal/logging"
	"github.com/lola-os/lola-core/internal/nonce"
	"github.com/lola-os/lola-core/internal/oracle"
	"github.com/lola-os/lola-core/internal/registry"
	"github.com/lola-os/lola-core/internal/rpc"
	"github.com/lola-os/lola-core/internal/vault"
)

var (
	serveTCPAddr    string
	serveVaultPass  string
	serveNoStdio    bool
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the LOLA OS JSON-RPC engine",
	Long: `Starts lola-core's JSON-RPC server. By default it serves over
stdin/stdout (the transport used by the Python and TypeScript SDKs when
they spawn lola-core as a subprocess). Pass --tcp to also (or instead)
listen on a local TCP socket.`,
	RunE: runServe,
}

func init() {
	serveCmd.Flags().StringVar(&serveTCPAddr, "tcp", "", "also listen on this TCP address, e.g. 127.0.0.1:8899")
	serveCmd.Flags().BoolVar(&serveNoStdio, "no-stdio", false, "disable the stdio JSON-RPC transport (use only --tcp)")
	serveCmd.Flags().StringVar(&serveVaultPass, "vault-passphrase", "", "vault passphrase (falls back to LOLA_VAULT_PASSPHRASE env var, then an interactive prompt)")
}

func runServe(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}
	logger := logging.New(os.Stderr, logging.ParseLevel(cfg.Logging.Level), cfg.Logging.Format)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh
		logger.Info("shutting down", nil)
		cancel()
	}()

	reg, err := registry.Open(cfg.Registry.DBPath)
	if err != nil {
		return fmt.Errorf("opening registry: %w", err)
	}
	defer reg.Close()

	passphrase := serveVaultPass
	if passphrase == "" {
		passphrase = os.Getenv("LOLA_VAULT_PASSPHRASE")
	}
	if passphrase == "" {
		return fmt.Errorf("a vault passphrase is required: pass --vault-passphrase or set LOLA_VAULT_PASSPHRASE")
	}
	v, err := vault.OpenOrCreate(cfg.Vault.Path, passphrase)
	if err != nil {
		return fmt.Errorf("opening vault: %w", err)
	}
	defer v.Close()

	chains := buildChainSet(ctx, cfg, logger)
	if len(chains) == 0 {
		logger.Warn("no chains connected successfully; check your config.yaml RPC URLs", nil)
	}

	breaker := budget.New(cfg.Budget, logger)
	defer breaker.Stop()

	nonceMgr := nonce.New(reg)
	idemCache := idempotency.New(reg)
	oracleGw := oracle.New(cfg.Oracle.ChainlinkFeeds, cfg.Oracle.RESTTimeoutMS, cfg.Oracle.RESTMaxRetries)

	var approver hitl.Approver
	if cfg.HITL.WebSocketOn {
		wsServer, err := hitl.NewWebSocketServer(cfg.HITL.WebSocketAddr)
		if err != nil {
			return fmt.Errorf("starting HITL websocket server: %w", err)
		}
		if err := wsServer.Start(); err != nil {
			return fmt.Errorf("starting HITL websocket server: %w", err)
		}
		logger.Info("HITL websocket server listening", map[string]interface{}{"addr": cfg.HITL.WebSocketAddr})
		approver = wsServer
	} else {
		approver = hitl.NewConsoleApprover()
	}

	deps := &rpc.Deps{
		Chains: chains, Vault: v, Breaker: breaker, Nonces: nonceMgr,
		Idem: idemCache, Registry: reg, Oracle: oracleGw, Logger: logger,
		Approver: approver, HITLOn: cfg.HITL.Enabled,
		HITLTimeout: time.Duration(cfg.HITL.TimeoutSeconds) * time.Second,
		ReadOnly:    cfg.Mode == "read_only",
	}

	server := jsonrpc.NewServer()
	rpc.Register(server, deps)

	logger.Info("lola-core ready", map[string]interface{}{"mode": cfg.Mode, "chains": len(chains)})

	errCh := make(chan error, 2)
	if serveTCPAddr != "" {
		go func() { errCh <- server.ServeTCP(ctx, serveTCPAddr) }()
		logger.Info("JSON-RPC TCP listener started", map[string]interface{}{"addr": serveTCPAddr})
	}
	if !serveNoStdio {
		go func() { errCh <- server.ServeStdio(ctx, os.Stdin, os.Stdout) }()
	}

	select {
	case <-ctx.Done():
		return nil
	case err := <-errCh:
		return err
	}
}
