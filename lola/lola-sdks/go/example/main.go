//go:build example

// Package main demonstrates the Go SDK. Build with `go run -tags example
// ./example` after building lola-core (see the root README). This file is
// excluded from normal builds via the `example` build tag so importing
// the lola module doesn't require a runnable main().
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/lola-os/lola-go"
)

func main() {
	ctx := context.Background()

	client, err := lola.NewClient(ctx, lola.ClientOptions{
		VaultPassphrase: os.Getenv("LOLA_VAULT_PASSPHRASE"),
	})
	if err != nil {
		log.Fatalf("connecting to lola-core: %v", err)
	}
	defer client.Close()

	// A plain function, wrapped as a named tool. Any agent framework that
	// accepts a `func(context.Context) (interface{}, error)` (or similar)
	// can register checkBalance directly.
	checkBalance := lola.Tool(lola.ToolOptions{Name: "check_balance"}, func(ctx context.Context) (interface{}, error) {
		return client.GetBalance(ctx, "ethereum", "0x0000000000000000000000000000000000000000", nil)
	})

	result, err := checkBalance(ctx)
	if err != nil {
		log.Fatalf("check_balance failed: %v", err)
	}
	fmt.Printf("balance: %+v\n", result)

	// Scoping a context override to a single call: check the same
	// balance on Polygon instead, without touching global config.
	polygonCtx := lola.WithOverrides(ctx, &lola.Overrides{Chain: "polygon"})
	polygonResult, err := checkBalance(polygonCtx)
	if err != nil {
		log.Fatalf("check_balance (polygon) failed: %v", err)
	}
	fmt.Printf("polygon balance: %+v\n", polygonResult)
}
