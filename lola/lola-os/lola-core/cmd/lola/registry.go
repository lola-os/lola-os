package main

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/lola-os/lola-core/internal/config"
	"github.com/lola-os/lola-core/internal/registry"
)

var registryListLimit int

var registryCmd = &cobra.Command{
	Use:   "registry",
	Short: "Manage LOLA OS's local transaction registry",
}

var registryListCmd = &cobra.Command{
	Use:   "list",
	Short: "Show recent operations (tx hash, chain, method, timestamp, status)",
	RunE: func(cmd *cobra.Command, args []string) error {
		reg, err := openRegistry()
		if err != nil {
			return err
		}
		defer reg.Close()

		txs, err := reg.ListTransactions(registryListLimit)
		if err != nil {
			return err
		}
		if len(txs) == 0 {
			fmt.Println("No transactions recorded yet.")
			return nil
		}
		fmt.Printf("%-12s %-10s %-20s %-10s %-20s\n", "CHAIN", "STATUS", "METHOD", "GAS", "TIMESTAMP")
		fmt.Println(color.HiBlackString("────────────────────────────────────────────────────────────────────────"))
		for _, t := range txs {
			fmt.Printf("%-12s %-10s %-20s %-10d %-20s\n", t.Chain, statusColor(t.Status), truncate(t.Method, 20), t.Gas, t.Timestamp.Format("2006-01-02 15:04:05"))
			fmt.Printf("  %s\n", color.HiBlackString(t.Hash))
		}
		return nil
	},
}

var registryShowCmd = &cobra.Command{
	Use:   "show <tx_hash>",
	Short: "Show full details for a transaction by hash",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		reg, err := openRegistry()
		if err != nil {
			return err
		}
		defer reg.Close()

		t, err := reg.GetTransaction(args[0])
		if err != nil {
			return err
		}
		if t == nil {
			return fmt.Errorf("no transaction found with hash %s", args[0])
		}
		fmt.Printf("Hash:       %s\n", t.Hash)
		fmt.Printf("Chain:      %s\n", t.Chain)
		fmt.Printf("From:       %s\n", t.From)
		fmt.Printf("To:         %s\n", t.To)
		fmt.Printf("Value:      %s\n", t.Value)
		fmt.Printf("Gas:        %d\n", t.Gas)
		fmt.Printf("Gas price:  %s\n", t.GasPrice)
		fmt.Printf("Status:     %s\n", statusColor(t.Status))
		fmt.Printf("Method:     %s\n", t.Method)
		fmt.Printf("Plan ID:    %s\n", t.PlanID)
		fmt.Printf("Timestamp:  %s\n", t.Timestamp)
		if t.Error != "" {
			fmt.Printf("Error:      %s\n", color.RedString(t.Error))
		}
		return nil
	},
}

var registryClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Reset the local registry (truncates all tables)",
	RunE: func(cmd *cobra.Command, args []string) error {
		reg, err := openRegistry()
		if err != nil {
			return err
		}
		defer reg.Close()
		if err := reg.ClearAll(); err != nil {
			return err
		}
		fmt.Println(color.GreenString("✅ Registry cleared."))
		return nil
	},
}

func init() {
	registryListCmd.Flags().IntVar(&registryListLimit, "limit", 50, "maximum number of transactions to show")
	registryCmd.AddCommand(registryListCmd, registryShowCmd, registryClearCmd)
}

func openRegistry() (*registry.Registry, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}
	return registry.Open(cfg.Registry.DBPath)
}

func statusColor(s registry.TxStatus) string {
	switch s {
	case registry.TxStatusConfirmed:
		return color.GreenString(string(s))
	case registry.TxStatusFailed, registry.TxStatusDropped:
		return color.RedString(string(s))
	default:
		return color.YellowString(string(s))
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}
