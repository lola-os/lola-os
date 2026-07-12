// Package config loads LOLA OS configuration from ~/.lola/config.yaml,
// overlays environment variables, and applies sensible defaults.
//
// Precedence (highest wins): runtime context overrides > environment
// variables > config.yaml > built-in defaults.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/lola-os/lola-core/internal/chain"
)

// BudgetAction describes what happens when a budget limit is exceeded.
type BudgetAction string

const (
	BudgetActionPause  BudgetAction = "pause"
	BudgetActionNotify BudgetAction = "notify"
	BudgetActionDeny   BudgetAction = "deny"
)

// BudgetConfig configures the circuit breaker in internal/budget.
type BudgetConfig struct {
	MaxGasSpendPerSession float64      `yaml:"max_gas_spend_per_session"` // in native units (e.g. ETH)
	MaxUSDSpendPerSession float64      `yaml:"max_usd_spend_per_session"`
	MaxRequestsPerMinute  int          `yaml:"max_requests_per_minute"`
	Action                BudgetAction `yaml:"action"`
}

// ChainConfig holds per-chain RPC and adapter settings. Only `name` is ever
// required from the user: for any chain in the built-in catalog, the missing
// fields (kind, rpc_url, chain_id, native symbol/decimals, explorer) are
// filled in automatically at load time. Anything set here overrides the
// catalog default, so you always keep full control.
type ChainConfig struct {
	Name           string `yaml:"name"`
	Kind           string `yaml:"kind"` // "evm" or "solana"
	RPCURL         string `yaml:"rpc_url"`
	ChainID        int64  `yaml:"chain_id,omitempty"`
	NativeSymbol   string `yaml:"native_symbol,omitempty"`
	NativeDecimals int    `yaml:"native_decimals,omitempty"`
	Explorer       string `yaml:"explorer_api_url,omitempty"`
	ExplorerKey    string `yaml:"explorer_api_key,omitempty"`
}

// VaultConfig configures the encrypted key vault location and KDF params.
type VaultConfig struct {
	Path    string `yaml:"path"`
	ScryptN int    `yaml:"scrypt_n"`
	ScryptR int    `yaml:"scrypt_r"`
	ScryptP int    `yaml:"scrypt_p"`
}

// HITLConfig configures human-in-the-loop approval behavior.
type HITLConfig struct {
	Enabled        bool   `yaml:"enabled"`
	TimeoutSeconds int    `yaml:"timeout_seconds"`
	WebSocketAddr  string `yaml:"websocket_addr"` // e.g. "127.0.0.1:8765"
	WebSocketOn    bool   `yaml:"websocket_enabled"`
}

// LoggingConfig configures the rich/structured logger.
type LoggingConfig struct {
	Level  string `yaml:"level"`  // debug, info, warn, error
	Format string `yaml:"format"` // "rich" or "json"
}

// RegistryConfig configures the SQLite persistence layer.
type RegistryConfig struct {
	DBPath string `yaml:"db_path"`
}

// OracleConfig configures Chainlink + generic REST oracle access.
type OracleConfig struct {
	ChainlinkFeeds map[string]string `yaml:"chainlink_feeds"` // symbol -> contract address
	RESTTimeoutMS  int               `yaml:"rest_timeout_ms"`
	RESTMaxRetries int               `yaml:"rest_max_retries"`
}

// Config is the root LOLA OS configuration object.
type Config struct {
	Mode     string                 `yaml:"mode"` // "read_only" or "live"
	Chains   map[string]ChainConfig `yaml:"chains"`
	Vault    VaultConfig            `yaml:"vault"`
	HITL     HITLConfig             `yaml:"hitl"`
	Logging  LoggingConfig          `yaml:"logging"`
	Registry RegistryConfig         `yaml:"registry"`
	Budget   BudgetConfig           `yaml:"budget"`
	Oracle   OracleConfig           `yaml:"oracle"`
}

// HomeDir returns ~/.lola, creating it if necessary.
func HomeDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolving home directory: %w", err)
	}
	dir := filepath.Join(home, ".lola")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", fmt.Errorf("creating ~/.lola: %w", err)
	}
	return dir, nil
}

// chainConfigFromCatalog builds a ChainConfig from built-in catalog metadata.
func chainConfigFromCatalog(info chain.Info) ChainConfig {
	return ChainConfig{
		Name:           info.Name,
		Kind:           info.Kind,
		RPCURL:         info.DefaultRPC,
		ChainID:        info.ChainID,
		NativeSymbol:   info.NativeSymbol,
		NativeDecimals: info.NativeDecimals,
		Explorer:       info.ExplorerAPI,
	}
}

// defaultChains returns the out-of-the-box set of enabled chains, built from
// the catalog so their metadata (chain id, native symbol, explorer) is always
// consistent with what lola-core knows.
func defaultChains() map[string]ChainConfig {
	chains := make(map[string]ChainConfig, len(chain.DefaultEnabled))
	for _, name := range chain.DefaultEnabled {
		if info, ok := chain.Lookup(name); ok {
			chains[name] = chainConfigFromCatalog(info)
		}
	}
	return chains
}

// Default returns a Config populated with sensible, safe-by-default values.
func Default() Config {
	home, _ := HomeDir()
	return Config{
		Mode:   "read_only",
		Chains: defaultChains(),
		Vault: VaultConfig{
			Path:    filepath.Join(home, "vault.enc"),
			ScryptN: 1 << 15,
			ScryptR: 8,
			ScryptP: 1,
		},
		HITL: HITLConfig{
			Enabled:        true,
			TimeoutSeconds: 120,
			WebSocketAddr:  "127.0.0.1:8765",
			WebSocketOn:    false,
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "rich",
		},
		Registry: RegistryConfig{
			DBPath: filepath.Join(home, "lola.db"),
		},
		Budget: BudgetConfig{
			MaxGasSpendPerSession: 0.5,
			MaxUSDSpendPerSession: 50.0,
			MaxRequestsPerMinute:  60,
			Action:                BudgetActionPause,
		},
		Oracle: OracleConfig{
			ChainlinkFeeds: map[string]string{
				"ETH/USD": "0x5f4eC3Df9cbd43714FE2740f5E3616155c5b8419",
			},
			RESTTimeoutMS:  5000,
			RESTMaxRetries: 3,
		},
	}
}

// Load reads ~/.lola/config.yaml if present, merges it over the defaults,
// then applies environment variable overrides, and returns the result.
func Load() (Config, error) {
	cfg := Default()

	home, err := HomeDir()
	if err != nil {
		return cfg, err
	}
	path := filepath.Join(home, "config.yaml")

	if data, err := os.ReadFile(path); err == nil {
		var fileCfg Config
		if err := yaml.Unmarshal(data, &fileCfg); err != nil {
			return cfg, fmt.Errorf("parsing %s: %w", path, err)
		}
		cfg = mergeConfig(cfg, fileCfg)
	} else if !os.IsNotExist(err) {
		return cfg, fmt.Errorf("reading %s: %w", path, err)
	}

	applyEnvOverrides(&cfg)
	normalizeChains(&cfg)
	return cfg, nil
}

// normalizeChains fills in any missing per-chain fields from the built-in
// catalog, so a user can enable a supported chain with as little as its name.
// User-supplied values always win; the catalog only supplies blanks. The map
// key is used as the chain name when the entry omits its own `name`.
func normalizeChains(cfg *Config) {
	if cfg.Chains == nil {
		return
	}
	for key, cc := range cfg.Chains {
		if cc.Name == "" {
			cc.Name = key
		}
		if info, ok := chain.Lookup(cc.Name); ok {
			if cc.Kind == "" {
				cc.Kind = info.Kind
			}
			if cc.RPCURL == "" {
				cc.RPCURL = info.DefaultRPC
			}
			if cc.ChainID == 0 {
				cc.ChainID = info.ChainID
			}
			if cc.NativeSymbol == "" {
				cc.NativeSymbol = info.NativeSymbol
			}
			if cc.NativeDecimals == 0 {
				cc.NativeDecimals = info.NativeDecimals
			}
			if cc.Explorer == "" {
				cc.Explorer = info.ExplorerAPI
			}
		}
		// A chain with no kind and no catalog match defaults to EVM, the
		// most common case for a custom/private network.
		if cc.Kind == "" {
			cc.Kind = "evm"
		}
		if cc.NativeSymbol == "" {
			cc.NativeSymbol = "ETH"
		}
		if cc.NativeDecimals == 0 {
			cc.NativeDecimals = 18
		}
		cfg.Chains[key] = cc
	}
}

// Save writes the config to ~/.lola/config.yaml.
func Save(cfg Config) error {
	home, err := HomeDir()
	if err != nil {
		return err
	}
	path := filepath.Join(home, "config.yaml")
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}
	return os.WriteFile(path, data, 0o600)
}

// Path returns the path to config.yaml without loading it.
func Path() (string, error) {
	home, err := HomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "config.yaml"), nil
}

func mergeConfig(base, override Config) Config {
	if override.Mode != "" {
		base.Mode = override.Mode
	}
	if len(override.Chains) > 0 {
		if base.Chains == nil {
			base.Chains = map[string]ChainConfig{}
		}
		for k, v := range override.Chains {
			base.Chains[k] = v
		}
	}
	if override.Vault.Path != "" {
		base.Vault = override.Vault
	}
	if override.HITL.WebSocketAddr != "" || override.HITL.TimeoutSeconds != 0 {
		base.HITL = override.HITL
	}
	if override.Logging.Level != "" {
		base.Logging = override.Logging
	}
	if override.Registry.DBPath != "" {
		base.Registry = override.Registry
	}
	var zero BudgetConfig
	if override.Budget != zero {
		base.Budget = override.Budget
	}
	if len(override.Oracle.ChainlinkFeeds) > 0 {
		base.Oracle = override.Oracle
	}
	return base
}

// applyEnvOverrides overlays LOLA_* environment variables onto cfg.
// Supported variables are documented in the security/configuration guide.
func applyEnvOverrides(cfg *Config) {
	if v := os.Getenv("LOLA_MODE"); v != "" {
		cfg.Mode = v
	}
	if v := os.Getenv("LOLA_LOG_LEVEL"); v != "" {
		cfg.Logging.Level = v
	}
	if v := os.Getenv("LOLA_LOG_FORMAT"); v != "" {
		cfg.Logging.Format = v
	}
	if v := os.Getenv("LOLA_DB_PATH"); v != "" {
		cfg.Registry.DBPath = v
	}
	if v := os.Getenv("LOLA_VAULT_PATH"); v != "" {
		cfg.Vault.Path = v
	}
	if v := os.Getenv("LOLA_HITL_TIMEOUT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.HITL.TimeoutSeconds = n
		}
	}
	if v := os.Getenv("LOLA_BUDGET_MAX_USD"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			cfg.Budget.MaxUSDSpendPerSession = f
		}
	}
	if v := os.Getenv("LOLA_BUDGET_MAX_GAS"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			cfg.Budget.MaxGasSpendPerSession = f
		}
	}
	if v := os.Getenv("LOLA_BUDGET_ACTION"); v != "" {
		cfg.Budget.Action = BudgetAction(strings.ToLower(v))
	}
	// Per-chain RPC overrides: LOLA_RPC_URL_<CHAINNAME>=https://...
	for _, env := range os.Environ() {
		const prefix = "LOLA_RPC_URL_"
		if !strings.HasPrefix(env, prefix) {
			continue
		}
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		}
		chainKey := strings.ToLower(strings.TrimPrefix(parts[0], prefix))
		if c, ok := cfg.Chains[chainKey]; ok {
			c.RPCURL = parts[1]
			cfg.Chains[chainKey] = c
		}
	}
}

// Overrides backs the SDKs' "context overrides" feature (e.g. Python's
// `with lola.override(...)`), letting a single request carry a temporary
// config patch without mutating global state.
type Overrides struct {
	Chain           string
	RPCURL          string
	Mode            string
	BudgetMaxGas    *float64
	BudgetMaxUSD    *float64
	HITLTimeoutSecs *int
}

// WithOverrides returns a copy of c with the given overrides applied.
func (c Config) WithOverrides(o Overrides) Config {
	out := c
	chains := make(map[string]ChainConfig, len(c.Chains))
	for k, v := range c.Chains {
		chains[k] = v
	}
	out.Chains = chains

	if o.Mode != "" {
		out.Mode = o.Mode
	}
	if o.Chain != "" {
		ch, ok := chains[o.Chain]
		if !ok {
			// A brand-new chain referenced only in an override: seed it from
			// the catalog so its kind/id/symbol are correct even if the
			// caller only supplied a name (and optionally an RPC URL).
			if info, found := chain.Lookup(o.Chain); found {
				ch = chainConfigFromCatalog(info)
			} else {
				ch = ChainConfig{Name: o.Chain, Kind: "evm", NativeSymbol: "ETH", NativeDecimals: 18}
			}
		}
		if o.RPCURL != "" {
			ch.RPCURL = o.RPCURL
		}
		chains[o.Chain] = ch
	}
	if o.BudgetMaxGas != nil {
		out.Budget.MaxGasSpendPerSession = *o.BudgetMaxGas
	}
	if o.BudgetMaxUSD != nil {
		out.Budget.MaxUSDSpendPerSession = *o.BudgetMaxUSD
	}
	if o.HITLTimeoutSecs != nil {
		out.HITL.TimeoutSeconds = *o.HITLTimeoutSecs
	}
	return out
}
