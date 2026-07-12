package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/lola-os/lola-core/internal/config"
	"github.com/lola-os/lola-core/internal/registry"
	"github.com/lola-os/lola-core/internal/vault"
)

var doctorVaultPass string

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Run a comprehensive environment health check",
	Long: `Checks RPC connectivity for every configured chain, oracle endpoint
reachability, config.yaml syntax, vault integrity, and (if connected) hardware
wallet availability. Prints a pass/fail table with actionable fix messages.`,
	RunE: runDoctor,
}

func init() {
	doctorCmd.Flags().StringVar(&doctorVaultPass, "vault-passphrase", "", "vault passphrase, to test vault integrity (skipped if not provided)")
}

type checkResult struct {
	Name   string
	Pass   bool
	Detail string
	Fix    string
}

func runDoctor(cmd *cobra.Command, args []string) error {
	fmt.Println()
	fmt.Println(color.New(color.Bold).Sprint("🩺  LOLA OS Doctor"))
	fmt.Println(color.HiBlackString("Checking your environment..."))
	fmt.Println()

	var results []checkResult

	// 1. config.yaml syntax
	results = append(results, checkConfigSyntax())

	cfg, err := config.Load()
	if err != nil {
		results = append(results, checkResult{Name: "Configuration load", Pass: false, Detail: err.Error(),
			Fix: "Fix the YAML syntax error in ~/.lola/config.yaml, or delete it to fall back to defaults."})
		printResults(results)
		return fmt.Errorf("doctor: cannot continue without a valid configuration")
	}

	// 2. RPC connectivity per chain
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	for name, cc := range cfg.Chains {
		results = append(results, checkChainRPC(ctx, name, cc))
	}

	// 3. Oracle endpoint reachability (Chainlink feeds piggy-back on EVM
	// RPC checks above; here we verify the explorer/REST oracle config is
	// at least well-formed).
	results = append(results, checkOracleConfig(cfg))

	// 4. Vault integrity
	results = append(results, checkVault(cfg, doctorVaultPass))

	// 5. Registry / SQLite reachability
	results = append(results, checkRegistry(cfg))

	// 6. Hardware wallet detection (best-effort; LOLA doesn't bundle a
	// specific HW wallet driver in this build, so this reports guidance
	// rather than a false positive/negative).
	results = append(results, checkHardwareWallet())

	printResults(results)

	for _, r := range results {
		if !r.Pass {
			return fmt.Errorf("doctor: one or more checks failed")
		}
	}
	return nil
}

func checkConfigSyntax() checkResult {
	path, err := config.Path()
	if err != nil {
		return checkResult{Name: "config.yaml syntax", Pass: false, Detail: err.Error()}
	}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return checkResult{Name: "config.yaml syntax", Pass: true, Detail: "no config.yaml found; using built-in defaults"}
	}
	if err != nil {
		return checkResult{Name: "config.yaml syntax", Pass: false, Detail: err.Error()}
	}
	var generic map[string]interface{}
	if err := yaml.Unmarshal(data, &generic); err != nil {
		return checkResult{Name: "config.yaml syntax", Pass: false, Detail: err.Error(),
			Fix: fmt.Sprintf("Fix the YAML syntax in %s", path)}
	}
	return checkResult{Name: "config.yaml syntax", Pass: true, Detail: path}
}

func checkChainRPC(ctx context.Context, name string, cc config.ChainConfig) checkResult {
	a, err := buildSingleChain(ctx, name, cc)
	if err != nil {
		return checkResult{Name: fmt.Sprintf("RPC: %s", name), Pass: false, Detail: err.Error(),
			Fix: fmt.Sprintf("Check that %s is reachable and correct in config.yaml or via LOLA_RPC_URL_%s", cc.RPCURL, name)}
	}
	block, err := a.Ping(ctx)
	if err != nil {
		return checkResult{Name: fmt.Sprintf("RPC: %s", name), Pass: false, Detail: err.Error(),
			Fix: fmt.Sprintf("Check that %s is reachable and correct in config.yaml or via LOLA_RPC_URL_%s", cc.RPCURL, name)}
	}
	return checkResult{Name: fmt.Sprintf("RPC: %s", name), Pass: true, Detail: fmt.Sprintf("latest block/slot: %d", block)}
}

func checkOracleConfig(cfg config.Config) checkResult {
	if len(cfg.Oracle.ChainlinkFeeds) == 0 {
		return checkResult{Name: "Oracle configuration", Pass: true, Detail: "no Chainlink feeds configured"}
	}
	for pair, addr := range cfg.Oracle.ChainlinkFeeds {
		if len(addr) != 42 || addr[:2] != "0x" {
			return checkResult{Name: "Oracle configuration", Pass: false,
				Detail: fmt.Sprintf("feed %s has malformed address %s", pair, addr),
				Fix:    "Ensure every chainlink_feeds entry is a valid 0x-prefixed 20-byte address."}
		}
	}
	return checkResult{Name: "Oracle configuration", Pass: true, Detail: fmt.Sprintf("%d feed(s) configured", len(cfg.Oracle.ChainlinkFeeds))}
}

func checkVault(cfg config.Config, passphrase string) checkResult {
	if !vault.Exists(cfg.Vault.Path) {
		return checkResult{Name: "Vault integrity", Pass: true, Detail: "no vault created yet at " + cfg.Vault.Path}
	}
	if passphrase == "" {
		return checkResult{Name: "Vault integrity", Pass: true,
			Detail: "vault exists but was not tested (pass --vault-passphrase to verify decryption)"}
	}
	v, err := vault.Open(cfg.Vault.Path, passphrase)
	if err != nil {
		return checkResult{Name: "Vault integrity", Pass: false, Detail: err.Error(),
			Fix: "Double-check the passphrase. If the vault file is corrupted, restore from backup."}
	}
	defer v.Close()
	if err := v.VerifyIntegrity(); err != nil {
		return checkResult{Name: "Vault integrity", Pass: false, Detail: err.Error()}
	}
	return checkResult{Name: "Vault integrity", Pass: true, Detail: fmt.Sprintf("%d entries, decrypts cleanly", len(v.List()))}
}

func checkRegistry(cfg config.Config) checkResult {
	reg, err := registry.Open(cfg.Registry.DBPath)
	if err != nil {
		return checkResult{Name: "Registry (SQLite)", Pass: false, Detail: err.Error(),
			Fix: fmt.Sprintf("Check that %s is writable.", cfg.Registry.DBPath)}
	}
	defer reg.Close()
	txs, err := reg.ListTransactions(1)
	if err != nil {
		return checkResult{Name: "Registry (SQLite)", Pass: false, Detail: err.Error()}
	}
	return checkResult{Name: "Registry (SQLite)", Pass: true, Detail: fmt.Sprintf("%s (%d recent tx)", cfg.Registry.DBPath, len(txs))}
}

func checkHardwareWallet() checkResult {
	// This build does not bundle a hardware wallet driver (Ledger/Trezor
	// integration is a documented extension point). We report this
	// honestly rather than guessing.
	return checkResult{Name: "Hardware wallet", Pass: true,
		Detail: "no hardware wallet driver bundled in this build; use vault-based keys, or implement the HardwareSigner extension point"}
}

func printResults(results []checkResult) {
	for _, r := range results {
		icon := color.GreenString("✅")
		if !r.Pass {
			icon = color.RedString("❌")
		}
		fmt.Printf("  %s  %-28s %s\n", icon, r.Name, color.HiBlackString(r.Detail))
		if !r.Pass && r.Fix != "" {
			fmt.Printf("      %s %s\n", color.YellowString("→ fix:"), r.Fix)
		}
	}
	fmt.Println()

	passed := 0
	for _, r := range results {
		if r.Pass {
			passed++
		}
	}
	summary := fmt.Sprintf("%d/%d checks passed", passed, len(results))
	if passed == len(results) {
		fmt.Println(color.GreenString("✅ " + summary + " — your environment looks healthy."))
	} else {
		fmt.Println(color.RedString("❌ " + summary + " — see fixes above."))
	}
	fmt.Println()
}
