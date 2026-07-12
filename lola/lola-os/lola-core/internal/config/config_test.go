package config

import "testing"

// TestDefaultChainsComeFromCatalog verifies the out-of-the-box config enables
// the expected chains and that their metadata is populated from the catalog.
func TestDefaultChainsComeFromCatalog(t *testing.T) {
	cfg := Default()
	eth, ok := cfg.Chains["ethereum"]
	if !ok {
		t.Fatal("expected 'ethereum' to be enabled by default")
	}
	if eth.ChainID != 1 || eth.NativeSymbol != "ETH" || eth.RPCURL == "" {
		t.Fatalf("ethereum default metadata not filled from catalog: %+v", eth)
	}
	if _, ok := cfg.Chains["solana"]; !ok {
		t.Fatal("expected 'solana' to be enabled by default")
	}
}

// TestNormalizeFillsCatalogMetadata confirms that a chain enabled with only a
// name (as a user would write in config.yaml) is completed from the catalog,
// while user-supplied overrides are preserved.
func TestNormalizeFillsCatalogMetadata(t *testing.T) {
	cfg := Config{
		Chains: map[string]ChainConfig{
			// User enables Arbitrum with nothing but the map key.
			"arbitrum": {},
			// User points Base at their own RPC but leaves the rest blank.
			"base": {RPCURL: "https://my-private-base.example"},
			// A fully custom private EVM chain not in the catalog.
			"myrollup": {ChainID: 999999, RPCURL: "http://localhost:8545"},
		},
	}
	normalizeChains(&cfg)

	arb := cfg.Chains["arbitrum"]
	if arb.Name != "arbitrum" || arb.Kind != "evm" || arb.ChainID != 42161 || arb.NativeSymbol != "ETH" || arb.RPCURL == "" {
		t.Fatalf("arbitrum not normalized from catalog: %+v", arb)
	}

	base := cfg.Chains["base"]
	if base.RPCURL != "https://my-private-base.example" {
		t.Fatalf("user RPC override for base was clobbered: %+v", base)
	}
	if base.ChainID != 8453 {
		t.Fatalf("base chain id not filled from catalog: %+v", base)
	}

	custom := cfg.Chains["myrollup"]
	if custom.Kind != "evm" || custom.NativeSymbol != "ETH" || custom.NativeDecimals != 18 {
		t.Fatalf("custom chain did not get EVM defaults: %+v", custom)
	}
	if custom.ChainID != 999999 {
		t.Fatalf("custom chain id was altered: %+v", custom)
	}
}
