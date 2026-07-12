// Command lola is the single-binary entrypoint for LOLA OS: it hosts the
// JSON-RPC engine (used by the SDKs) and the operational CLI (doctor,
// registry, metrics, replay, vault).
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "lola",
	Short: "LOLA OS — the bridge between AI agents and blockchains",
	Long: `LOLA OS connects AI agents to blockchains, oracles, and APIs.

Run "lola serve" to start the JSON-RPC engine (used by the Python, Go, and
TypeScript SDKs), or use the operational commands below directly.`,
}

func main() {
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(chainsCmd)
	rootCmd.AddCommand(doctorCmd)
	rootCmd.AddCommand(registryCmd)
	rootCmd.AddCommand(metricsCmd)
	rootCmd.AddCommand(replayCmd)
	rootCmd.AddCommand(vaultCmd)
	rootCmd.AddCommand(versionCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// version is set at build time via -ldflags "-X main.version=v1.0.0".
var version = "dev"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the lola-core version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("lola-core " + version)
	},
}
