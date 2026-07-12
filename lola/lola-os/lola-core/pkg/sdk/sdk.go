// Package sdk is the public, in-process entry point to the LOLA OS engine.
//
// The Python and TypeScript SDKs talk to lola-core over JSON-RPC by spawning
// it as a subprocess. Go programs, being able to link lola-core's code
// directly, don't need that hop: this package exposes the very same engine
// components the JSON-RPC server wires up (chain adapters, the encrypted
// vault, the budget circuit breaker, the nonce manager, the idempotency
// cache, the oracle gateway, and human-in-the-loop approval) as a typed
// Engine, so the Go SDK is a thin, allocation-cheap wrapper around it.
//
// Everything under lola-core/internal is invisible to code outside the
// module; this package is the single, deliberately-narrow seam that makes the
// engine reusable as a library.
package sdk

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"os"
	"time"

	abivalidate "github.com/lola-os/lola-core/internal/abi"
	"github.com/lola-os/lola-core/internal/budget"
	"github.com/lola-os/lola-core/internal/chain"
	"github.com/lola-os/lola-core/internal/chain/evm"
	"github.com/lola-os/lola-core/internal/chain/solana"
	"github.com/lola-os/lola-core/internal/config"
	"github.com/lola-os/lola-core/internal/hitl"
	"github.com/lola-os/lola-core/internal/idempotency"
	"github.com/lola-os/lola-core/internal/logging"
	"github.com/lola-os/lola-core/internal/nonce"
	"github.com/lola-os/lola-core/internal/oracle"
	"github.com/lola-os/lola-core/internal/registry"
	"github.com/lola-os/lola-core/internal/vault"
)

// --- Public types -----------------------------------------------------------

// Config is the resolved LOLA OS configuration (see config.yaml + LOLA_* env).
type Config = config.Config

// Overrides scopes a per-call chain/RPC/budget/mode change. Overrides may only
// make a call MORE restrictive (e.g. force read-only); they can never widen
// permissions granted by the global config.
type Overrides = config.Overrides

// BudgetState is a snapshot of the current session's spend accounting.
type BudgetState = budget.State

// ABIMismatchError is returned when call arguments don't match a contract ABI.
type ABIMismatchError = abivalidate.ABIMismatchError

// BudgetExceededError is returned when an operation would breach a budget limit.
type BudgetExceededError = budget.ErrBudgetExceeded

// ErrReadOnly is returned by write operations when the engine is in read-only
// mode (the default). Set `mode: live` in config.yaml (or LOLA_MODE=live) to
// enable writes.
var ErrReadOnly = errors.New("lola: engine is in read_only mode; write operations are disabled")

// Balance is a native or token balance result.
type Balance struct {
	Address  string
	Token    string // empty for the native asset
	RawValue *big.Int
	Decimals int
	Symbol   string
}

// TxReceipt is the chain-agnostic result of a broadcast transaction.
type TxReceipt struct {
	Hash        string
	Status      string // "pending", "confirmed", "failed"
	BlockNumber uint64
	GasUsed     uint64
}

// TxRequest describes a native-asset transfer.
type TxRequest struct {
	To       string
	ValueWei *big.Int
}

// ContractCallRequest describes a read-only or state-changing contract call.
type ContractCallRequest struct {
	ContractAddress string
	Method          string
	Args            []interface{}
	ABI             string // optional; empty triggers an ABI fetch from the explorer
	From            string // required for state-changing calls
	ValueWei        *big.Int
}

// Options configures a new Engine.
type Options struct {
	Config          Config
	VaultPassphrase string // required only if you perform write operations
}

// LoadConfig resolves configuration exactly as the CLI does: built-in defaults,
// overlaid by ~/.lola/config.yaml, overlaid by LOLA_* environment variables,
// with the built-in chain catalog filling in any omitted chain metadata.
func LoadConfig() (Config, error) {
	return config.Load()
}

// --- Engine -----------------------------------------------------------------

// Engine is the in-process LOLA OS core. It is safe for concurrent use.
type Engine struct {
	cfg      config.Config
	logger   *logging.Logger
	chains   chain.Set
	vault    *vault.Vault
	breaker  *budget.Breaker
	nonces   *nonce.Manager
	idem     *idempotency.Cache
	registry *registry.Registry
	oracle   *oracle.Gateway
	approver hitl.Approver

	readOnly    bool
	hitlOn      bool
	hitlTimeout time.Duration
}

// New builds an Engine: it opens the registry and (if a passphrase is given)
// the encrypted vault, dials every configured chain, and starts the budget
// breaker. Call Close when done.
func New(ctx context.Context, opts Options) (*Engine, error) {
	cfg := opts.Config
	logger := logging.New(os.Stderr, logging.ParseLevel(cfg.Logging.Level), cfg.Logging.Format)

	reg, err := registry.Open(cfg.Registry.DBPath)
	if err != nil {
		return nil, fmt.Errorf("lola: opening registry: %w", err)
	}

	var v *vault.Vault
	if opts.VaultPassphrase != "" {
		v, err = vault.OpenOrCreate(cfg.Vault.Path, opts.VaultPassphrase)
		if err != nil {
			reg.Close()
			return nil, fmt.Errorf("lola: opening vault: %w", err)
		}
	}

	e := &Engine{
		cfg:         cfg,
		logger:      logger,
		chains:      buildChains(ctx, cfg, logger),
		vault:       v,
		breaker:     budget.New(cfg.Budget, logger),
		nonces:      nonce.New(reg),
		idem:        idempotency.New(reg),
		registry:    reg,
		oracle:      oracle.New(cfg.Oracle.ChainlinkFeeds, cfg.Oracle.RESTTimeoutMS, cfg.Oracle.RESTMaxRetries),
		approver:    hitl.NewConsoleApprover(),
		readOnly:    cfg.Mode == "read_only",
		hitlOn:      cfg.HITL.Enabled,
		hitlTimeout: time.Duration(cfg.HITL.TimeoutSeconds) * time.Second,
	}
	return e, nil
}

// Close releases the vault key material, registry handle, and budget breaker
// background loop.
func (e *Engine) Close() error {
	e.breaker.Stop()
	if e.vault != nil {
		e.vault.Close()
	}
	return e.registry.Close()
}

// buildChains dials every configured chain, skipping (with a warning) any that
// fail so one bad endpoint never blocks the rest.
func buildChains(ctx context.Context, cfg config.Config, logger *logging.Logger) chain.Set {
	set := chain.Set{}
	for name, cc := range cfg.Chains {
		a, err := buildAdapter(ctx, name, cc)
		if err != nil {
			logger.Warn("skipping chain: failed to connect", map[string]interface{}{"chain": name, "error": err.Error()})
			continue
		}
		set[name] = a
	}
	return set
}

func buildAdapter(ctx context.Context, name string, cc config.ChainConfig) (chain.ChainAdapter, error) {
	if cc.Kind == "solana" {
		return solana.New(name, cc.RPCURL), nil
	}
	return evm.New(ctx, evm.Config{
		Name:           name,
		RPCURL:         cc.RPCURL,
		ChainID:        cc.ChainID,
		NativeSymbol:   cc.NativeSymbol,
		NativeDecimals: cc.NativeDecimals,
		ExplorerAPI:    cc.Explorer,
		ExplorerKey:    cc.ExplorerKey,
	})
}

// resolveAdapter picks the chain adapter for a call, honoring a chain and/or
// RPC-URL override. A chain named only in an override — or pointed at a
// one-off RPC — is dialed on demand, with catalog metadata filling any blanks.
func (e *Engine) resolveAdapter(ctx context.Context, chainName string, ov *Overrides) (chain.ChainAdapter, error) {
	name := chainName
	if ov != nil && ov.Chain != "" {
		name = ov.Chain
	}

	if ov != nil && ov.RPCURL != "" {
		cc := e.chainConfig(name)
		cc.RPCURL = ov.RPCURL
		return buildAdapter(ctx, name, cc)
	}
	if a, err := e.chains.Get(name); err == nil {
		return a, nil
	}
	// Not pre-dialed: build on demand from config/catalog so any of the
	// catalog's chains works without being enabled up front.
	return buildAdapter(ctx, name, e.chainConfig(name))
}

// chainConfig returns the effective config for a chain: the one from loaded
// config if present, else derived from the built-in catalog, else an EVM
// default (for a bare custom RPC).
func (e *Engine) chainConfig(name string) config.ChainConfig {
	if cc, ok := e.cfg.Chains[name]; ok {
		return cc
	}
	if info, ok := chain.Lookup(name); ok {
		return config.ChainConfig{
			Name: info.Name, Kind: info.Kind, RPCURL: info.DefaultRPC, ChainID: info.ChainID,
			NativeSymbol: info.NativeSymbol, NativeDecimals: info.NativeDecimals, Explorer: info.ExplorerAPI,
		}
	}
	return config.ChainConfig{Name: name, Kind: "evm", NativeSymbol: "ETH", NativeDecimals: 18}
}

// effectiveReadOnly lets an override tighten to read-only but never loosen it.
func (e *Engine) effectiveReadOnly(ov *Overrides) bool {
	if e.readOnly {
		return true
	}
	return ov != nil && ov.Mode == "read_only"
}

func (e *Engine) hitlTimeoutFor(ov *Overrides) time.Duration {
	if ov != nil && ov.HITLTimeoutSecs != nil {
		return time.Duration(*ov.HITLTimeoutSecs) * time.Second
	}
	return e.hitlTimeout
}
