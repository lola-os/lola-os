package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/lola-os/lola-core/internal/config"
	"github.com/lola-os/lola-core/internal/logging"
	"github.com/lola-os/lola-core/internal/registry"
	"github.com/lola-os/lola-core/internal/replay"
	"github.com/lola-os/lola-core/internal/vault"
)

var (
	replayForkURL        string
	replayDryRun         bool
	replayOutput         string
	replayVaultPass      string
	replayKeyNameForFrom string
)

var replayCmd = &cobra.Command{
	Use:   "replay <plan.json>",
	Short: "Execute a structured execution plan",
	Long: `Ingests a JSON plan describing a sequence of operations
(call_contract, send_transaction, transfer_token, execute_contract,
swap_tokens, assert, wait), executes them in order, and records the run
in the local registry. See the "Replay" page in the docs for the full
plan.json schema.`,
	Args: cobra.ExactArgs(1),
	RunE: runReplay,
}

func init() {
	replayCmd.Flags().StringVar(&replayForkURL, "fork-url", "", "execute against this RPC URL instead of the configured one (e.g. a local fork)")
	replayCmd.Flags().BoolVar(&replayDryRun, "dry-run", false, "simulate only; do not broadcast any transactions")
	replayCmd.Flags().StringVar(&replayOutput, "output", "", "write the execution receipt to this file as JSON")
	replayCmd.Flags().StringVar(&replayVaultPass, "vault-passphrase", "", "vault passphrase for resolving signing keys (skipped in --dry-run)")
	replayCmd.Flags().StringVar(&replayKeyNameForFrom, "key-name", "default", "vault entry name to use as the signing key for every `from` address in the plan")
}

func runReplay(cmd *cobra.Command, args []string) error {
	planPath := args[0]
	plan, err := replay.LoadPlan(planPath)
	if err != nil {
		return err
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}
	logger := logging.New(os.Stderr, logging.ParseLevel(cfg.Logging.Level), cfg.Logging.Format)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	if replayForkURL != "" {
		// Apply the fork URL to every configured chain referenced by the
		// plan so execution happens against the fork rather than the
		// configured production RPC.
		for name, cc := range cfg.Chains {
			cc.RPCURL = replayForkURL
			cfg.Chains[name] = cc
		}
		logger.Info("running replay against fork", map[string]interface{}{"fork_url": replayForkURL})
	}

	chains := buildChainSet(ctx, cfg, logger)

	reg, err := registry.Open(cfg.Registry.DBPath)
	if err != nil {
		return fmt.Errorf("opening registry: %w", err)
	}
	defer reg.Close()

	var keyResolver func(chainName, from string) (string, error)
	if !replayDryRun {
		passphrase := replayVaultPass
		if passphrase == "" {
			passphrase = os.Getenv("LOLA_VAULT_PASSPHRASE")
		}
		if passphrase == "" {
			return fmt.Errorf("a vault passphrase is required for non-dry-run replay: pass --vault-passphrase or set LOLA_VAULT_PASSPHRASE")
		}
		v, err := vault.Open(cfg.Vault.Path, passphrase)
		if err != nil {
			return fmt.Errorf("opening vault: %w", err)
		}
		defer v.Close()
		keyResolver = func(chainName, from string) (string, error) {
			return v.Get(replayKeyNameForFrom)
		}
	}

	engine := replay.New(chains, reg, logger)
	receipt, runErr := engine.Run(ctx, plan, replay.Options{
		ForkURL: replayForkURL, DryRun: replayDryRun, PrivateKeyResolver: keyResolver,
	})

	printReceipt(receipt)

	if replayOutput != "" {
		data, err := json.MarshalIndent(receipt, "", "  ")
		if err != nil {
			return fmt.Errorf("encoding receipt: %w", err)
		}
		if err := os.WriteFile(replayOutput, data, 0o644); err != nil {
			return fmt.Errorf("writing receipt to %s: %w", replayOutput, err)
		}
		fmt.Printf("Receipt written to %s\n", replayOutput)
	}

	return runErr
}

func printReceipt(r replay.Receipt) {
	fmt.Println()
	fmt.Println(color.New(color.Bold).Sprintf("Plan: %s", r.Description))
	fmt.Println(color.HiBlackString("Plan ID: " + r.PlanID))
	if r.DryRun {
		fmt.Println(color.YellowString("(dry run — no transactions broadcast)"))
	}
	fmt.Println()
	for _, s := range r.Steps {
		icon := color.GreenString("✅")
		if s.Error != "" {
			icon = color.RedString("❌")
		}
		fmt.Printf("  %s  %-12s %-18s", icon, s.ID, s.Type)
		if s.TxHash != "" {
			fmt.Printf(" tx=%s", s.TxHash)
		}
		if s.Error != "" {
			fmt.Printf(" error=%s", s.Error)
		}
		fmt.Println()
	}
	fmt.Println()
	if r.Success {
		fmt.Println(color.GreenString("✅ Plan completed successfully."))
	} else {
		fmt.Println(color.RedString("❌ Plan failed."))
	}
	fmt.Println()
}
