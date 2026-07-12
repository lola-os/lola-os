package sdk

import (
	"context"
	"fmt"
	"math/big"
	"time"

	abivalidate "github.com/lola-os/lola-core/internal/abi"
	"github.com/lola-os/lola-core/internal/chain"
	"github.com/lola-os/lola-core/internal/hitl"
	"github.com/lola-os/lola-core/internal/registry"
)

// GetBalance returns the native-asset balance for address on chainName.
func (e *Engine) GetBalance(ctx context.Context, chainName, address string, ov *Overrides) (Balance, error) {
	a, err := e.resolveAdapter(ctx, chainName, ov)
	if err != nil {
		return Balance{}, err
	}
	bal, err := a.NativeBalance(ctx, address)
	if err != nil {
		return Balance{}, err
	}
	return toBalance(bal), nil
}

// GetTokenBalance returns an ERC20/SPL token balance for address on chainName.
func (e *Engine) GetTokenBalance(ctx context.Context, chainName, address, tokenAddress string, ov *Overrides) (Balance, error) {
	a, err := e.resolveAdapter(ctx, chainName, ov)
	if err != nil {
		return Balance{}, err
	}
	bal, err := a.TokenBalance(ctx, address, tokenAddress)
	if err != nil {
		return Balance{}, err
	}
	return toBalance(bal), nil
}

// CallContract performs a read-only contract call. If req.ABI is empty, the
// adapter attempts to fetch a verified ABI from the configured explorer.
func (e *Engine) CallContract(ctx context.Context, chainName string, req ContractCallRequest, ov *Overrides) (interface{}, error) {
	a, err := e.resolveAdapter(ctx, chainName, ov)
	if err != nil {
		return nil, err
	}
	abiJSON := req.ABI
	if abiJSON == "" {
		fetched, ferr := a.FetchABI(ctx, req.ContractAddress)
		if ferr != nil {
			return nil, fmt.Errorf("lola: no ABI provided and fetch failed: %w", ferr)
		}
		abiJSON = fetched
	}
	if err := abivalidate.Validate(abiJSON, req.Method, req.Args); err != nil {
		return nil, err
	}
	return a.CallContract(ctx, chain.ContractCallRequest{
		ContractAddress: req.ContractAddress, Method: req.Method, Args: req.Args, ABI: abiJSON,
	})
}

// SendTransaction sends a native-asset transfer signed with the named vault
// key. It enforces read-only mode, the rate/spend budget, optional HITL
// approval, nonce management, and idempotency — the same guarantees the
// JSON-RPC engine provides.
func (e *Engine) SendTransaction(ctx context.Context, chainName string, req TxRequest, keyName string, ov *Overrides) (TxReceipt, error) {
	if e.effectiveReadOnly(ov) {
		return TxReceipt{}, ErrReadOnly
	}
	a, err := e.resolveAdapter(ctx, chainName, ov)
	if err != nil {
		return TxReceipt{}, err
	}
	privKey, err := e.signingKey(keyName)
	if err != nil {
		return TxReceipt{}, err
	}
	from, err := a.AddressFromKey(privKey)
	if err != nil {
		return TxReceipt{}, err
	}
	value := req.ValueWei
	if value == nil {
		value = big.NewInt(0)
	}

	if err := e.breaker.CheckRate(); err != nil {
		return TxReceipt{}, err
	}
	estGas, err := a.EstimateGas(ctx, chain.TxRequest{From: from, To: req.To, ValueWei: value})
	if err != nil {
		return TxReceipt{}, fmt.Errorf("lola: estimating gas: %w", err)
	}
	estGasFloat := weiToFloat(estGas)
	if err := e.breaker.CheckWrite(estGasFloat, 0); err != nil {
		return TxReceipt{}, err
	}

	if e.hitlOn {
		if err := e.approve(ctx, ov, hitl.Request{
			ID: fmt.Sprintf("send-%d", nowNano()), Chain: chainName, From: from, To: req.To,
			ValueHuman: value.String() + " wei", Description: "Native asset transfer",
		}); err != nil {
			return TxReceipt{}, err
		}
	}

	nextNonce, err := e.nonces.Next(ctx, chainName, from, a)
	if err != nil {
		return TxReceipt{}, err
	}
	receipt, err := a.SendTransaction(ctx, chain.TxRequest{From: from, To: req.To, ValueWei: value, Nonce: &nextNonce}, privKey)
	if err != nil {
		_ = e.nonces.Release(chainName, from, nextNonce)
		return TxReceipt{}, err
	}

	_ = e.registry.RecordTransaction(registry.Transaction{
		Hash: receipt.Hash, Chain: chainName, From: from, To: req.To,
		Value: value.String(), Status: registry.TxStatusPending, Timestamp: time.Now().UTC(),
	})
	e.breaker.Record(estGasFloat, 0)
	return toReceipt(receipt), nil
}

// ExecuteContract builds, signs, and broadcasts a state-changing contract
// call. Set idempotencyKey (non-empty) to make retries safe: a repeated call
// with the same key returns the original receipt instead of re-broadcasting.
func (e *Engine) ExecuteContract(ctx context.Context, chainName string, req ContractCallRequest, keyName, idempotencyKey string, ov *Overrides) (TxReceipt, error) {
	if idempotencyKey != "" {
		var cached TxReceipt
		if found, _ := e.idem.Lookup(idempotencyKey, &cached); found {
			return cached, nil
		}
	}
	if e.effectiveReadOnly(ov) {
		return TxReceipt{}, ErrReadOnly
	}
	a, err := e.resolveAdapter(ctx, chainName, ov)
	if err != nil {
		return TxReceipt{}, err
	}
	privKey, err := e.signingKey(keyName)
	if err != nil {
		return TxReceipt{}, err
	}
	from := req.From
	if from == "" {
		from, err = a.AddressFromKey(privKey)
		if err != nil {
			return TxReceipt{}, err
		}
	}

	abiJSON := req.ABI
	if abiJSON == "" {
		fetched, ferr := a.FetchABI(ctx, req.ContractAddress)
		if ferr != nil {
			return TxReceipt{}, fmt.Errorf("lola: no ABI provided and fetch failed: %w", ferr)
		}
		abiJSON = fetched
	}
	if err := abivalidate.Validate(abiJSON, req.Method, req.Args); err != nil {
		return TxReceipt{}, err
	}

	if err := e.breaker.CheckRate(); err != nil {
		return TxReceipt{}, err
	}
	if err := e.breaker.CheckWrite(0, 0); err != nil {
		return TxReceipt{}, err
	}

	if e.hitlOn {
		if err := e.approve(ctx, ov, hitl.Request{
			ID: fmt.Sprintf("exec-%d", nowNano()), Chain: chainName, From: from,
			Contract: req.ContractAddress, Method: req.Method, Description: "Contract execution",
		}); err != nil {
			return TxReceipt{}, err
		}
	}

	receipt, err := a.ExecuteContract(ctx, chain.ContractCallRequest{
		ContractAddress: req.ContractAddress, Method: req.Method, Args: req.Args, ABI: abiJSON,
		From: from, ValueWei: req.ValueWei,
	}, privKey)
	if err != nil {
		return TxReceipt{}, err
	}

	_ = e.registry.RecordTransaction(registry.Transaction{
		Hash: receipt.Hash, Chain: chainName, From: from, To: req.ContractAddress, Method: req.Method,
		Status: registry.TxStatusPending, Timestamp: time.Now().UTC(),
	})
	e.breaker.Record(0, 0)

	out := toReceipt(receipt)
	if idempotencyKey != "" {
		_ = e.idem.Store(idempotencyKey, out)
	}
	return out, nil
}

// FetchExternalAPI performs a rate-limited, retrying GET and JSON-decodes the
// response into out (pass a pointer).
func (e *Engine) FetchExternalAPI(ctx context.Context, url string, out interface{}) error {
	return e.oracle.FetchJSON(ctx, url, out)
}

// BudgetStatus returns a snapshot of the current session's spend.
func (e *Engine) BudgetStatus() BudgetState {
	return e.breaker.Snapshot()
}

// VaultKeyNames lists the names of keys stored in the vault (never their
// values). Returns nil if the vault was not opened (no passphrase supplied).
func (e *Engine) VaultKeyNames() []string {
	if e.vault == nil {
		return nil
	}
	return e.vault.List()
}

// --- helpers ---------------------------------------------------------------

func (e *Engine) signingKey(name string) (string, error) {
	if e.vault == nil {
		return "", fmt.Errorf("lola: no vault open — construct the client with a VaultPassphrase to sign transactions")
	}
	key, err := e.vault.Get(name)
	if err != nil {
		return "", fmt.Errorf("lola: resolving signing key %q: %w", name, err)
	}
	return key, nil
}

func (e *Engine) approve(ctx context.Context, ov *Overrides, req hitl.Request) error {
	decision, err := e.approver.RequestApproval(ctx, req, e.hitlTimeoutFor(ov))
	if err != nil || decision != hitl.DecisionApprove {
		return fmt.Errorf("lola: operation was not approved by the operator")
	}
	return nil
}

func toBalance(b chain.Balance) Balance {
	return Balance{Address: b.Address, Token: b.Token, RawValue: b.RawValue, Decimals: b.Decimals, Symbol: b.Symbol}
}

func toReceipt(r chain.TxReceipt) TxReceipt {
	return TxReceipt{Hash: r.Hash, Status: r.Status, BlockNumber: r.BlockNumber, GasUsed: r.GasUsed}
}

func weiToFloat(wei *big.Int) float64 {
	f, _ := new(big.Float).Quo(new(big.Float).SetInt(wei), big.NewFloat(1e18)).Float64()
	return f
}

func nowNano() int64 { return time.Now().UnixNano() }
