// Package rpc wires the jsonrpc.Server's method dispatch table to LOLA
// OS's actual engine components: chain adapters, the vault, the budget
// breaker, the nonce manager, the idempotency cache, and pre-flight ABI
// validation. This is the "API" the Python/Go/TypeScript SDKs talk to.
package rpc

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	abi "github.com/lola-os/lola-core/internal/abi"
	"github.com/lola-os/lola-core/internal/budget"
	"github.com/lola-os/lola-core/internal/chain"
	"github.com/lola-os/lola-core/internal/hitl"
	"github.com/lola-os/lola-core/internal/idempotency"
	"github.com/lola-os/lola-core/internal/jsonrpc"
	"github.com/lola-os/lola-core/internal/logging"
	"github.com/lola-os/lola-core/internal/nonce"
	"github.com/lola-os/lola-core/internal/oracle"
	"github.com/lola-os/lola-core/internal/registry"
	"github.com/lola-os/lola-core/internal/vault"
)

// Deps bundles every component the RPC layer needs.
type Deps struct {
	Chains      chain.Set
	Vault       *vault.Vault
	Breaker     *budget.Breaker
	Nonces      *nonce.Manager
	Idem        *idempotency.Cache
	Registry    *registry.Registry
	Oracle      *oracle.Gateway
	Logger      *logging.Logger
	Approver    hitl.Approver
	HITLOn      bool
	HITLTimeout time.Duration
	ReadOnly    bool
}

// Register attaches every LOLA OS RPC method to server.
func Register(server *jsonrpc.Server, d *Deps) {
	server.Handle("ping", d.handlePing)
	server.Handle("get_balance", d.handleGetBalance)
	server.Handle("get_token_balance", d.handleGetTokenBalance)
	server.Handle("call_contract", d.handleCallContract)
	server.Handle("send_transaction", d.handleSendTransaction)
	server.Handle("execute_contract", d.handleExecuteContract)
	server.Handle("transfer_token", d.handleTransferToken)
	server.Handle("get_price", d.handleGetPrice)
	server.Handle("fetch_external_api", d.handleFetchExternalAPI)
	server.Handle("budget_status", d.handleBudgetStatus)
	server.Handle("vault_list", d.handleVaultList)
}

func (d *Deps) chainAdapter(name string) (chain.ChainAdapter, *jsonrpc.Error) {
	a, err := d.Chains.Get(name)
	if err != nil {
		return nil, jsonrpc.NewError(jsonrpc.CodeUnknownChain, err.Error(), nil)
	}
	return a, nil
}

func parseParams(raw json.RawMessage, out interface{}) error {
	if len(raw) == 0 {
		return nil
	}
	return json.Unmarshal(raw, out)
}

// --- ping --------------------------------------------------------------

type pingParams struct {
	Chain string `json:"chain"`
}

func (d *Deps) handlePing(ctx context.Context, raw json.RawMessage) (interface{}, error) {
	var p pingParams
	if err := parseParams(raw, &p); err != nil {
		return nil, jsonrpc.NewError(jsonrpc.CodeInvalidParams, err.Error(), nil)
	}
	a, jerr := d.chainAdapter(p.Chain)
	if jerr != nil {
		return nil, jerr
	}
	block, err := a.Ping(ctx)
	if err != nil {
		return nil, jsonrpc.NewError(jsonrpc.CodeRPCConnection, err.Error(), nil)
	}
	return map[string]interface{}{"chain": p.Chain, "latest_block": block}, nil
}

// --- get_balance ---------------------------------------------------------

type getBalanceParams struct {
	Chain   string `json:"chain"`
	Address string `json:"address"`
}

func (d *Deps) handleGetBalance(ctx context.Context, raw json.RawMessage) (interface{}, error) {
	var p getBalanceParams
	if err := parseParams(raw, &p); err != nil {
		return nil, jsonrpc.NewError(jsonrpc.CodeInvalidParams, err.Error(), nil)
	}
	a, jerr := d.chainAdapter(p.Chain)
	if jerr != nil {
		return nil, jerr
	}
	bal, err := a.NativeBalance(ctx, p.Address)
	if err != nil {
		return nil, jsonrpc.NewError(jsonrpc.CodeRPCConnection, err.Error(), nil)
	}
	return balanceToJSON(bal), nil
}

type getTokenBalanceParams struct {
	Chain        string `json:"chain"`
	Address      string `json:"address"`
	TokenAddress string `json:"token_address"`
}

func (d *Deps) handleGetTokenBalance(ctx context.Context, raw json.RawMessage) (interface{}, error) {
	var p getTokenBalanceParams
	if err := parseParams(raw, &p); err != nil {
		return nil, jsonrpc.NewError(jsonrpc.CodeInvalidParams, err.Error(), nil)
	}
	a, jerr := d.chainAdapter(p.Chain)
	if jerr != nil {
		return nil, jerr
	}
	bal, err := a.TokenBalance(ctx, p.Address, p.TokenAddress)
	if err != nil {
		return nil, jsonrpc.NewError(jsonrpc.CodeRPCConnection, err.Error(), nil)
	}
	return balanceToJSON(bal), nil
}

func balanceToJSON(b chain.Balance) map[string]interface{} {
	return map[string]interface{}{
		"address": b.Address, "token": b.Token, "raw_value": b.RawValue.String(),
		"decimals": b.Decimals, "symbol": b.Symbol,
	}
}

// --- call_contract ---------------------------------------------------------

type callContractParams struct {
	Chain    string        `json:"chain"`
	Contract string        `json:"contract"`
	Method   string        `json:"method"`
	Args     []interface{} `json:"args"`
	ABI      string        `json:"abi"`
}

func (d *Deps) handleCallContract(ctx context.Context, raw json.RawMessage) (interface{}, error) {
	var p callContractParams
	if err := parseParams(raw, &p); err != nil {
		return nil, jsonrpc.NewError(jsonrpc.CodeInvalidParams, err.Error(), nil)
	}
	a, jerr := d.chainAdapter(p.Chain)
	if jerr != nil {
		return nil, jerr
	}

	abiJSON := p.ABI
	if abiJSON == "" {
		fetched, err := a.FetchABI(ctx, p.Contract)
		if err != nil {
			return nil, jsonrpc.NewError(jsonrpc.CodeInvalidParams, "no ABI provided and fetch failed: "+err.Error(), nil)
		}
		abiJSON = fetched
	}
	if err := abi.Validate(abiJSON, p.Method, p.Args); err != nil {
		return nil, jsonrpc.NewError(jsonrpc.CodeABIMismatch, err.Error(), nil)
	}

	out, err := a.CallContract(ctx, chain.ContractCallRequest{
		ContractAddress: p.Contract, Method: p.Method, Args: p.Args, ABI: abiJSON,
	})
	if err != nil {
		return nil, jsonrpc.NewError(jsonrpc.CodeRPCConnection, err.Error(), nil)
	}
	return map[string]interface{}{"result": out}, nil
}

// --- send_transaction ---------------------------------------------------------

type sendTransactionParams struct {
	Chain          string `json:"chain"`
	From           string `json:"from"`
	To             string `json:"to"`
	ValueWei       string `json:"value_wei"`
	IdempotencyKey string `json:"idempotency_key"`
	KeyName        string `json:"key_name"` // name of the vault entry holding the private key
}

func (d *Deps) handleSendTransaction(ctx context.Context, raw json.RawMessage) (interface{}, error) {
	var p sendTransactionParams
	if err := parseParams(raw, &p); err != nil {
		return nil, jsonrpc.NewError(jsonrpc.CodeInvalidParams, err.Error(), nil)
	}

	if cached, found, _ := lookupIdempotent(d, p.IdempotencyKey); found {
		return cached, nil
	}

	if d.ReadOnly {
		return nil, jsonrpc.NewError(jsonrpc.CodeInvalidParams, "lola-core is in read_only mode; write operations are disabled", nil)
	}

	a, jerr := d.chainAdapter(p.Chain)
	if jerr != nil {
		return nil, jerr
	}

	value, ok := new(big.Int).SetString(p.ValueWei, 10)
	if !ok {
		return nil, jsonrpc.NewError(jsonrpc.CodeInvalidParams, "invalid value_wei: "+p.ValueWei, nil)
	}

	if err := d.Breaker.CheckRate(); err != nil {
		return nil, jsonrpc.NewError(jsonrpc.CodeBudgetExceeded, err.Error(), nil)
	}

	estGas, err := a.EstimateGas(ctx, chain.TxRequest{From: p.From, To: p.To, ValueWei: value})
	if err != nil {
		return nil, jsonrpc.NewError(jsonrpc.CodeRPCConnection, "estimating gas: "+err.Error(), nil)
	}
	estGasFloat, _ := new(big.Float).SetInt(estGas).Float64()
	estGasFloat /= 1e18

	if err := d.Breaker.CheckWrite(estGasFloat, 0); err != nil {
		return nil, jsonrpc.NewError(jsonrpc.CodeBudgetExceeded, err.Error(), nil)
	}

	if d.HITLOn {
		decision, err := d.Approver.RequestApproval(ctx, hitl.Request{
			ID: fmt.Sprintf("send-%d", time.Now().UnixNano()), Chain: p.Chain,
			From: p.From, To: p.To, ValueHuman: p.ValueWei + " wei",
			Description: "Native asset transfer",
		}, d.HITLTimeout)
		if err != nil || decision != hitl.DecisionApprove {
			return nil, jsonrpc.NewError(jsonrpc.CodeApprovalDenied, "transaction was not approved by operator", nil)
		}
	}

	privKey, err := d.Vault.Get(p.KeyName)
	if err != nil {
		return nil, jsonrpc.NewError(jsonrpc.CodeInvalidParams, "resolving signing key: "+err.Error(), nil)
	}

	nextNonce, err := d.Nonces.Next(ctx, p.Chain, p.From, &chainClientAdapter{a})
	if err != nil {
		return nil, jsonrpc.NewError(jsonrpc.CodeInternalError, err.Error(), nil)
	}

	receipt, err := a.SendTransaction(ctx, chain.TxRequest{From: p.From, To: p.To, ValueWei: value, Nonce: &nextNonce}, privKey)
	if err != nil {
		_ = d.Nonces.Release(p.Chain, p.From, nextNonce)
		return nil, jsonrpc.NewError(jsonrpc.CodeInternalError, err.Error(), nil)
	}

	_ = d.Registry.RecordTransaction(registry.Transaction{
		Hash: receipt.Hash, Chain: p.Chain, From: p.From, To: p.To,
		Value: p.ValueWei, Status: registry.TxStatusPending, Timestamp: time.Now().UTC(),
	})
	d.Breaker.Record(estGasFloat, 0)

	result := map[string]interface{}{"tx_hash": receipt.Hash, "status": receipt.Status}
	_ = storeIdempotent(d, p.IdempotencyKey, result)
	return result, nil
}

// --- execute_contract ---------------------------------------------------------

type executeContractParams struct {
	Chain          string        `json:"chain"`
	From           string        `json:"from"`
	Contract       string        `json:"contract"`
	Method         string        `json:"method"`
	Args           []interface{} `json:"args"`
	ABI            string        `json:"abi"`
	ValueWei       string        `json:"value_wei"`
	IdempotencyKey string        `json:"idempotency_key"`
	KeyName        string        `json:"key_name"`
}

func (d *Deps) handleExecuteContract(ctx context.Context, raw json.RawMessage) (interface{}, error) {
	var p executeContractParams
	if err := parseParams(raw, &p); err != nil {
		return nil, jsonrpc.NewError(jsonrpc.CodeInvalidParams, err.Error(), nil)
	}

	if cached, found, _ := lookupIdempotent(d, p.IdempotencyKey); found {
		return cached, nil
	}
	if d.ReadOnly {
		return nil, jsonrpc.NewError(jsonrpc.CodeInvalidParams, "lola-core is in read_only mode; write operations are disabled", nil)
	}

	a, jerr := d.chainAdapter(p.Chain)
	if jerr != nil {
		return nil, jerr
	}

	abiJSON := p.ABI
	if abiJSON == "" {
		fetched, err := a.FetchABI(ctx, p.Contract)
		if err != nil {
			return nil, jsonrpc.NewError(jsonrpc.CodeInvalidParams, "no ABI provided and fetch failed: "+err.Error(), nil)
		}
		abiJSON = fetched
	}
	if err := abi.Validate(abiJSON, p.Method, p.Args); err != nil {
		return nil, jsonrpc.NewError(jsonrpc.CodeABIMismatch, err.Error(), nil)
	}

	if err := d.Breaker.CheckRate(); err != nil {
		return nil, jsonrpc.NewError(jsonrpc.CodeBudgetExceeded, err.Error(), nil)
	}
	if err := d.Breaker.CheckWrite(0, 0); err != nil {
		return nil, jsonrpc.NewError(jsonrpc.CodeBudgetExceeded, err.Error(), nil)
	}

	if d.HITLOn {
		decision, err := d.Approver.RequestApproval(ctx, hitl.Request{
			ID: fmt.Sprintf("exec-%d", time.Now().UnixNano()), Chain: p.Chain,
			From: p.From, Contract: p.Contract, Method: p.Method,
			Description: "Contract execution",
		}, d.HITLTimeout)
		if err != nil || decision != hitl.DecisionApprove {
			return nil, jsonrpc.NewError(jsonrpc.CodeApprovalDenied, "execution was not approved by operator", nil)
		}
	}

	privKey, err := d.Vault.Get(p.KeyName)
	if err != nil {
		return nil, jsonrpc.NewError(jsonrpc.CodeInvalidParams, "resolving signing key: "+err.Error(), nil)
	}

	var value *big.Int
	if p.ValueWei != "" {
		value, _ = new(big.Int).SetString(p.ValueWei, 10)
	}

	receipt, err := a.ExecuteContract(ctx, chain.ContractCallRequest{
		ContractAddress: p.Contract, Method: p.Method, Args: p.Args, ABI: abiJSON, From: p.From, ValueWei: value,
	}, privKey)
	if err != nil {
		return nil, jsonrpc.NewError(jsonrpc.CodeInternalError, err.Error(), nil)
	}

	_ = d.Registry.RecordTransaction(registry.Transaction{
		Hash: receipt.Hash, Chain: p.Chain, From: p.From, To: p.Contract, Method: p.Method,
		Status: registry.TxStatusPending, Timestamp: time.Now().UTC(),
	})
	d.Breaker.Record(0, 0)

	result := map[string]interface{}{"tx_hash": receipt.Hash, "status": receipt.Status}
	_ = storeIdempotent(d, p.IdempotencyKey, result)
	return result, nil
}

// --- transfer_token ---------------------------------------------------------

type transferTokenParams struct {
	Chain          string `json:"chain"`
	From           string `json:"from"`
	To             string `json:"to"`
	Token          string `json:"token"`
	AmountRaw      string `json:"amount_raw"` // smallest-unit decimal string
	IdempotencyKey string `json:"idempotency_key"`
	KeyName        string `json:"key_name"`
}

const erc20TransferABI = `[{"constant":false,"inputs":[{"name":"_to","type":"address"},{"name":"_value","type":"uint256"}],"name":"transfer","outputs":[{"name":"","type":"bool"}],"type":"function"}]`

func (d *Deps) handleTransferToken(ctx context.Context, raw json.RawMessage) (interface{}, error) {
	var p transferTokenParams
	if err := parseParams(raw, &p); err != nil {
		return nil, jsonrpc.NewError(jsonrpc.CodeInvalidParams, err.Error(), nil)
	}
	if cached, found, _ := lookupIdempotent(d, p.IdempotencyKey); found {
		return cached, nil
	}
	if d.ReadOnly {
		return nil, jsonrpc.NewError(jsonrpc.CodeInvalidParams, "lola-core is in read_only mode; write operations are disabled", nil)
	}

	a, jerr := d.chainAdapter(p.Chain)
	if jerr != nil {
		return nil, jerr
	}
	amount, ok := new(big.Int).SetString(p.AmountRaw, 10)
	if !ok {
		return nil, jsonrpc.NewError(jsonrpc.CodeInvalidParams, "invalid amount_raw: "+p.AmountRaw, nil)
	}

	if err := d.Breaker.CheckRate(); err != nil {
		return nil, jsonrpc.NewError(jsonrpc.CodeBudgetExceeded, err.Error(), nil)
	}
	if err := d.Breaker.CheckWrite(0, 0); err != nil {
		return nil, jsonrpc.NewError(jsonrpc.CodeBudgetExceeded, err.Error(), nil)
	}

	if d.HITLOn {
		decision, err := d.Approver.RequestApproval(ctx, hitl.Request{
			ID: fmt.Sprintf("transfer-%d", time.Now().UnixNano()), Chain: p.Chain,
			From: p.From, To: p.To, Contract: p.Token, ValueHuman: p.AmountRaw + " (raw units)",
			Description: "Token transfer",
		}, d.HITLTimeout)
		if err != nil || decision != hitl.DecisionApprove {
			return nil, jsonrpc.NewError(jsonrpc.CodeApprovalDenied, "transfer was not approved by operator", nil)
		}
	}

	privKey, err := d.Vault.Get(p.KeyName)
	if err != nil {
		return nil, jsonrpc.NewError(jsonrpc.CodeInvalidParams, "resolving signing key: "+err.Error(), nil)
	}

	receipt, err := a.ExecuteContract(ctx, chain.ContractCallRequest{
		ContractAddress: p.Token, Method: "transfer", Args: []interface{}{p.To, amount}, ABI: erc20TransferABI, From: p.From,
	}, privKey)
	if err != nil {
		return nil, jsonrpc.NewError(jsonrpc.CodeInternalError, err.Error(), nil)
	}

	_ = d.Registry.RecordTransaction(registry.Transaction{
		Hash: receipt.Hash, Chain: p.Chain, From: p.From, To: p.To, Method: "transfer",
		Status: registry.TxStatusPending, Timestamp: time.Now().UTC(),
	})
	d.Breaker.Record(0, 0)

	result := map[string]interface{}{"tx_hash": receipt.Hash, "status": receipt.Status}
	_ = storeIdempotent(d, p.IdempotencyKey, result)
	return result, nil
}

// --- oracle ---------------------------------------------------------

type getPriceParams struct {
	Chain string `json:"chain"`
	Pair  string `json:"pair"`
}

// evmClientHolder is implemented by the EVM chain adapter, exposing the
// underlying go-ethereum client so oracle reads (Chainlink aggregator calls)
// can run against the same connection the adapter already dialed.
type evmClientHolder interface {
	EVMClient() *ethclient.Client
}

func (d *Deps) handleGetPrice(ctx context.Context, raw json.RawMessage) (interface{}, error) {
	var p getPriceParams
	if err := parseParams(raw, &p); err != nil {
		return nil, jsonrpc.NewError(jsonrpc.CodeInvalidParams, err.Error(), nil)
	}
	a, jerr := d.chainAdapter(p.Chain)
	if jerr != nil {
		return nil, jerr
	}
	holder, ok := a.(evmClientHolder)
	if !ok || a.Kind() != "evm" {
		return nil, jsonrpc.NewError(jsonrpc.CodeInvalidParams,
			fmt.Sprintf("get_price is only supported on EVM chains (Chainlink feeds are read on-chain); %q is a %s chain", p.Chain, a.Kind()), nil)
	}
	res, err := d.Oracle.GetPrice(ctx, holder.EVMClient(), p.Pair)
	if err != nil {
		return nil, jsonrpc.NewError(jsonrpc.CodeRPCConnection, err.Error(), nil)
	}
	return map[string]interface{}{
		"pair":       res.Pair,
		"price":      res.Price,
		"decimals":   res.Decimals,
		"updated_at": res.UpdatedAt.UTC().Format(time.RFC3339),
	}, nil
}

type fetchExternalAPIParams struct {
	URL string `json:"url"`
}

func (d *Deps) handleFetchExternalAPI(ctx context.Context, raw json.RawMessage) (interface{}, error) {
	var p fetchExternalAPIParams
	if err := parseParams(raw, &p); err != nil {
		return nil, jsonrpc.NewError(jsonrpc.CodeInvalidParams, err.Error(), nil)
	}
	var out interface{}
	if err := d.Oracle.FetchJSON(ctx, p.URL, &out); err != nil {
		return nil, jsonrpc.NewError(jsonrpc.CodeRPCConnection, err.Error(), nil)
	}
	return out, nil
}

// --- budget / vault introspection ---------------------------------------------------------

func (d *Deps) handleBudgetStatus(ctx context.Context, raw json.RawMessage) (interface{}, error) {
	s := d.Breaker.Snapshot()
	return map[string]interface{}{
		"gas_spent": s.GasSpent, "usd_spent": s.USDSpent,
		"requests_this_minute": s.RequestsThisMinute, "paused": s.Paused, "paused_reason": s.PausedReason,
	}, nil
}

func (d *Deps) handleVaultList(ctx context.Context, raw json.RawMessage) (interface{}, error) {
	return map[string]interface{}{"entries": d.Vault.List()}, nil
}

// --- helpers ---------------------------------------------------------

func lookupIdempotent(d *Deps, key string) (interface{}, bool, error) {
	if key == "" {
		return nil, false, nil
	}
	var out map[string]interface{}
	found, err := d.Idem.Lookup(key, &out)
	return out, found, err
}

func storeIdempotent(d *Deps, key string, result interface{}) error {
	return d.Idem.Store(key, result)
}

// chainClientAdapter adapts a chain.ChainAdapter to nonce.ChainClient.
type chainClientAdapter struct {
	a chain.ChainAdapter
}

func (c *chainClientAdapter) PendingNonce(ctx context.Context, address string) (uint64, error) {
	return c.a.PendingNonce(ctx, address)
}
