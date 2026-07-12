package main

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/lola-os/lola-core/internal/config"
	"github.com/lola-os/lola-core/internal/registry"
)

var metricsFormat string

var metricsCmd = &cobra.Command{
	Use:   "metrics",
	Short: "Show operational metrics (requests, gas spent, success/failure counts)",
	Long: `Exposes counters accumulated by lola-core: total requests, gas spent,
rate-limit hits, and success/failure counts. Output as JSON lines (default)
or Prometheus exposition format with --format prometheus.`,
	RunE: runMetrics,
}

func init() {
	metricsCmd.Flags().StringVar(&metricsFormat, "format", "json", "output format: json or prometheus")
}

func runMetrics(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}
	reg, err := registry.Open(cfg.Registry.DBPath)
	if err != nil {
		return fmt.Errorf("opening registry: %w", err)
	}
	defer reg.Close()

	counters, err := reg.AllCounters()
	if err != nil {
		return fmt.Errorf("reading counters: %w", err)
	}

	txs, err := reg.ListTransactions(0)
	if err != nil {
		return fmt.Errorf("reading transactions: %w", err)
	}
	var confirmed, failed, pending int
	for _, t := range txs {
		switch t.Status {
		case registry.TxStatusConfirmed:
			confirmed++
		case registry.TxStatusFailed:
			failed++
		default:
			pending++
		}
	}

	if metricsFormat == "prometheus" {
		fmt.Println("# HELP lola_transactions_total Total transactions recorded by status")
		fmt.Println("# TYPE lola_transactions_total counter")
		fmt.Printf("lola_transactions_total{status=\"confirmed\"} %d\n", confirmed)
		fmt.Printf("lola_transactions_total{status=\"failed\"} %d\n", failed)
		fmt.Printf("lola_transactions_total{status=\"pending\"} %d\n", pending)
		fmt.Println("# HELP lola_counter Generic named counters tracked by lola-core")
		fmt.Println("# TYPE lola_counter counter")
		for name, v := range counters {
			fmt.Printf("lola_counter{name=%q} %d\n", name, v)
		}
		return nil
	}

	out := map[string]interface{}{
		"transactions": map[string]int{"confirmed": confirmed, "failed": failed, "pending": pending, "total": len(txs)},
		"counters":     counters,
	}
	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}
