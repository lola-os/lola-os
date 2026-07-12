// Package lola is the LOLA OS Go SDK. Unlike the Python and TypeScript
// SDKs, it does not spawn a subprocess: it imports lola-core's pkg/sdk
// package directly and calls it in-process, since Go programs can link
// against lola-core's compiled engine code natively.
//
//	client, err := lola.NewClient(ctx, lola.ClientOptions{
//	    VaultPassphrase: os.Getenv("LOLA_VAULT_PASSPHRASE"),
//	})
//	balance, err := client.GetBalance(ctx, "ethereum", "0x...")
package lola

import (
	"context"
	"fmt"

	coresdk "github.com/lola-os/lola-core/pkg/sdk"
)

// Re-exported types so callers never need to import lola-core directly.
type (
	Balance             = coresdk.Balance
	TxReceipt           = coresdk.TxReceipt
	TxRequest           = coresdk.TxRequest
	ContractCallRequest = coresdk.ContractCallRequest
	BudgetState         = coresdk.BudgetState
	Overrides           = coresdk.Overrides
	ABIMismatchError    = coresdk.ABIMismatchError
	BudgetExceededError = coresdk.BudgetExceededError
)

var ErrReadOnly = coresdk.ErrReadOnly

// ClientOptions configures a new Client. If Config is nil, the standard
// ~/.lola/config.yaml + environment variable resolution (the same one
// lola-core's CLI uses) applies.
type ClientOptions struct {
	VaultPassphrase string
	// Config, if non-nil, is used as-is instead of loading from disk —
	// useful for tests or fully programmatic setups.
	Config *coresdk.Config
}

// Client is the Go SDK's main entry point, wrapping an in-process
// lola-core engine instance.
type Client struct {
	engine *coresdk.Engine
}

// NewClient loads configuration (from ~/.lola/config.yaml + env vars,
// unless opts.Config is set), opens the vault and registry, dials every
// configured chain, and returns a ready-to-use Client.
func NewClient(ctx context.Context, opts ClientOptions) (*Client, error) {
	var cfg coresdk.Config
	if opts.Config != nil {
		cfg = *opts.Config
	} else {
		loaded, err := coresdk.LoadConfig()
		if err != nil {
			return nil, fmt.Errorf("lola: loading config: %w", err)
		}
		cfg = loaded
	}

	engine, err := coresdk.New(ctx, coresdk.Options{Config: cfg, VaultPassphrase: opts.VaultPassphrase})
	if err != nil {
		return nil, err
	}
	return &Client{engine: engine}, nil
}

// Close releases all resources (vault key material, registry handle,
// budget breaker background loop).
func (c *Client) Close() error {
	return c.engine.Close()
}

// GetBalance returns the native asset balance for address on chainName.
// Pass overrides (or nil) to scope a one-off chain/RPC/budget override to
// this call only.
func (c *Client) GetBalance(ctx context.Context, chainName, address string, overrides *Overrides) (Balance, error) {
	return c.engine.GetBalance(ctx, chainName, address, resolveOverrides(ctx, overrides))
}

// GetTokenBalance returns an ERC20/SPL token balance.
func (c *Client) GetTokenBalance(ctx context.Context, chainName, address, tokenAddress string, overrides *Overrides) (Balance, error) {
	return c.engine.GetTokenBalance(ctx, chainName, address, tokenAddress, resolveOverrides(ctx, overrides))
}

// CallContract performs a read-only contract call. If req.ABI is empty,
// lola-core attempts to fetch a verified ABI automatically.
func (c *Client) CallContract(ctx context.Context, chainName string, req ContractCallRequest, overrides *Overrides) (interface{}, error) {
	return c.engine.CallContract(ctx, chainName, req, resolveOverrides(ctx, overrides))
}

// ExecuteContract builds, signs (using the named vault key), and
// broadcasts a state-changing contract call. May block on a
// human-in-the-loop approval prompt depending on config.
func (c *Client) ExecuteContract(ctx context.Context, chainName string, req ContractCallRequest, keyName, idempotencyKey string, overrides *Overrides) (TxReceipt, error) {
	return c.engine.ExecuteContract(ctx, chainName, req, keyName, idempotencyKey, resolveOverrides(ctx, overrides))
}

// SendTransaction sends a native asset transfer, signed with the named
// vault key.
func (c *Client) SendTransaction(ctx context.Context, chainName string, req TxRequest, keyName string, overrides *Overrides) (TxReceipt, error) {
	return c.engine.SendTransaction(ctx, chainName, req, keyName, resolveOverrides(ctx, overrides))
}

// TransferToken transfers an ERC20/SPL token by calling the token
// contract's standard `transfer(address,uint256)` method.
func (c *Client) TransferToken(ctx context.Context, chainName, tokenAddress, from, to string, amountRaw interface{}, keyName, idempotencyKey string, overrides *Overrides) (TxReceipt, error) {
	const erc20TransferABI = `[{"constant":false,"inputs":[{"name":"_to","type":"address"},{"name":"_value","type":"uint256"}],"name":"transfer","outputs":[{"name":"","type":"bool"}],"type":"function"}]`
	req := ContractCallRequest{
		ContractAddress: tokenAddress,
		Method:          "transfer",
		Args:            []interface{}{to, amountRaw},
		ABI:             erc20TransferABI,
		From:            from,
	}
	return c.engine.ExecuteContract(ctx, chainName, req, keyName, idempotencyKey, resolveOverrides(ctx, overrides))
}

// FetchExternalAPI performs a rate-limited, retrying GET request and
// JSON-decodes the response into out (pass a pointer).
func (c *Client) FetchExternalAPI(ctx context.Context, url string, out interface{}) error {
	return c.engine.FetchExternalAPI(ctx, url, out)
}

// BudgetStatus returns a snapshot of the current session's gas/USD spend
// and rate-limit window.
func (c *Client) BudgetStatus() BudgetState {
	return c.engine.BudgetStatus()
}

// VaultKeyNames lists the names of keys stored in the vault (never their
// values).
func (c *Client) VaultKeyNames() []string {
	return c.engine.VaultKeyNames()
}
