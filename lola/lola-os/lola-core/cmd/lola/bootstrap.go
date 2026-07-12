package main

import (
	"context"
	"fmt"

	"github.com/lola-os/lola-core/internal/chain"
	"github.com/lola-os/lola-core/internal/chain/evm"
	"github.com/lola-os/lola-core/internal/chain/solana"
	"github.com/lola-os/lola-core/internal/config"
	"github.com/lola-os/lola-core/internal/logging"
)

// buildChainSet dials every configured chain and returns a ready chain.Set.
// Chains that fail to dial are skipped with a warning rather than aborting
// the whole process, so a misconfigured/unreachable testnet doesn't block
// use of the chains that do work.
func buildChainSet(ctx context.Context, cfg config.Config, logger *logging.Logger) chain.Set {
	set := chain.Set{}
	for name, cc := range cfg.Chains {
		switch cc.Kind {
		case "solana":
			set[name] = solana.New(name, cc.RPCURL)
		default: // "evm"
			a, err := evm.New(ctx, evmConfig(name, cc))
			if err != nil {
				logger.Warn("skipping chain: failed to connect", map[string]interface{}{"chain": name, "error": err.Error()})
				continue
			}
			set[name] = a
		}
	}
	return set
}

// requireChain is a small helper for commands that need a single named
// chain adapter built fresh (e.g. `lola doctor`, `lola replay --fork-url`).
func buildSingleChain(ctx context.Context, name string, cc config.ChainConfig) (chain.ChainAdapter, error) {
	switch cc.Kind {
	case "solana":
		return solana.New(name, cc.RPCURL), nil
	default:
		a, err := evm.New(ctx, evmConfig(name, cc))
		if err != nil {
			return nil, fmt.Errorf("connecting to %s: %w", name, err)
		}
		return a, nil
	}
}

// evmConfig adapts a config.ChainConfig into the evm adapter's Config.
func evmConfig(name string, cc config.ChainConfig) evm.Config {
	return evm.Config{
		Name:           name,
		RPCURL:         cc.RPCURL,
		ChainID:        cc.ChainID,
		NativeSymbol:   cc.NativeSymbol,
		NativeDecimals: cc.NativeDecimals,
		ExplorerAPI:    cc.Explorer,
		ExplorerKey:    cc.ExplorerKey,
	}
}
