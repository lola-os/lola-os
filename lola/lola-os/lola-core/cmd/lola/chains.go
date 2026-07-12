package main

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/spf13/cobra"

	"github.com/lola-os/lola-core/internal/chain"
	"github.com/lola-os/lola-core/internal/config"
)

var chainsJSON bool

// chainsCmd lists every chain LOLA OS knows about out of the box, marking
// which are enabled in the current configuration. This is the fast way to
// discover what you can connect to without reading source or docs.
var chainsCmd = &cobra.Command{
	Use:   "chains",
	Short: "List supported blockchains (the built-in catalog) and which are enabled",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, _ := config.Load()
		enabled := map[string]bool{}
		for name := range cfg.Chains {
			enabled[name] = true
		}

		names := chain.CatalogNames()
		sort.Strings(names)

		if chainsJSON {
			type row struct {
				Name         string `json:"name"`
				Display      string `json:"display"`
				Kind         string `json:"kind"`
				ChainID      int64  `json:"chain_id"`
				NativeSymbol string `json:"native_symbol"`
				Testnet      bool   `json:"testnet"`
				Enabled      bool   `json:"enabled"`
			}
			out := make([]row, 0, len(names))
			for _, n := range names {
				info, _ := chain.Lookup(n)
				out = append(out, row{info.Name, info.Display, info.Kind, info.ChainID, info.NativeSymbol, info.Testnet, enabled[n]})
			}
			b, _ := json.MarshalIndent(out, "", "  ")
			fmt.Println(string(b))
			return nil
		}

		fmt.Printf("\n  LOLA OS supports %d chains out of the box (● = enabled now).\n", len(names))
		fmt.Print("  Enable any of them by name in ~/.lola/config.yaml.\n\n")
		fmt.Printf("  %-2s %-18s %-24s %-8s %-8s %s\n", "", "NAME", "NETWORK", "KIND", "SYMBOL", "CHAIN ID")
		for _, n := range names {
			info, _ := chain.Lookup(n)
			marker := " "
			if enabled[n] {
				marker = "●"
			}
			id := ""
			if info.ChainID != 0 {
				id = fmt.Sprintf("%d", info.ChainID)
			}
			net := info.Display
			if info.Testnet {
				net += " (testnet)"
			}
			fmt.Printf("  %-2s %-18s %-24s %-8s %-8s %s\n", marker, info.Name, net, info.Kind, info.NativeSymbol, id)
		}
		fmt.Println()
		return nil
	},
}

func init() {
	chainsCmd.Flags().BoolVar(&chainsJSON, "json", false, "output the catalog as JSON")
}
