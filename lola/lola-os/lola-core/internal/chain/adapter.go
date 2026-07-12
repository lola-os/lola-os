// Package chain defines the ChainAdapter interface that every blockchain
// integration (EVM, Solana, and any future chain) must implement. This is
// the seam that lets lola-core support "any chain" without the rest of the
// engine (replay, budget, registry, RPC server) knowing chain-specific
// details.
package chain

import (
	"context"
	"math/big"
)

// TxRequest describes a generic write operation, independent of chain
// family. Adapters translate this into their native transaction format.
type TxRequest struct {
	From            string
	To              string
	ValueWei        *big.Int // native unit, smallest denomination (wei, lamports)
	Data            []byte   // ABI-encoded call data (EVM) or instruction data (Solana)
	Nonce           *uint64  // nil = let the adapter/nonce manager decide
	GasLimit        uint64   // EVM only; ignored elsewhere
	MaxFeePerGas    *big.Int // EVM only (EIP-1559); ignored elsewhere
	MaxPriorityFee  *big.Int // EVM only (EIP-1559); ignored elsewhere
	IdempotencyKey  string
}

// TxReceipt is the chain-agnostic result of a broadcast transaction.
type TxReceipt struct {
	Hash        string
	Status      string // "pending", "confirmed", "failed"
	BlockNumber uint64
	GasUsed     uint64
	EffectiveGasPrice *big.Int
	Error       string
}

// ContractCallRequest describes a read-only ("call") or state-changing
// ("send") contract invocation.
type ContractCallRequest struct {
	ContractAddress string
	Method          string
	Args            []interface{}
	ABI             string // JSON ABI, if known; empty triggers ABI fetch
	From            string // for sends; ignored for reads
	ValueWei        *big.Int
}

// Balance represents a native or token balance query result.
type Balance struct {
	Address  string
	Token    string // empty = native asset
	RawValue *big.Int
	Decimals int
	Symbol   string
}

// ChainAdapter is the interface every blockchain integration implements.
// Implementations must be safe for concurrent use.
type ChainAdapter interface {
	// Name returns a short identifier, e.g. "ethereum", "polygon", "solana".
	Name() string

	// AddressFromKey derives the public account address that corresponds to
	// the given private key material, in the chain's native address format
	// (a 0x-hex address for EVM, a base58 pubkey for Solana). This lets
	// callers sign with a key by name without also having to state the
	// sender address.
	AddressFromKey(privateKeyHex string) (string, error)

	// Kind returns the chain family: "evm" or "solana".
	Kind() string

	// Ping verifies RPC connectivity and returns the latest block height
	// (or slot, for Solana) — used by `lola doctor`.
	Ping(ctx context.Context) (uint64, error)

	// NativeBalance returns the native asset balance for an address.
	NativeBalance(ctx context.Context, address string) (Balance, error)

	// TokenBalance returns an ERC20/SPL token balance for an address.
	TokenBalance(ctx context.Context, address, tokenAddress string) (Balance, error)

	// EstimateGas estimates the gas/fee cost of a transaction request.
	// Returns native-unit cost (wei, lamports).
	EstimateGas(ctx context.Context, req TxRequest) (*big.Int, error)

	// PendingNonce returns the next nonce/sequence the chain expects for
	// address (satisfies nonce.ChainClient).
	PendingNonce(ctx context.Context, address string) (uint64, error)

	// SendTransaction signs (using the provided private key material) and
	// broadcasts a transaction, returning its hash immediately
	// (status "pending").
	SendTransaction(ctx context.Context, req TxRequest, privateKeyHex string) (TxReceipt, error)

	// WaitForReceipt polls until the transaction is confirmed/failed or
	// ctx is cancelled.
	WaitForReceipt(ctx context.Context, hash string) (TxReceipt, error)

	// CallContract performs a read-only contract call.
	CallContract(ctx context.Context, req ContractCallRequest) (interface{}, error)

	// ExecuteContract performs a state-changing contract call (builds,
	// signs, and broadcasts a transaction invoking the given method).
	ExecuteContract(ctx context.Context, req ContractCallRequest, privateKeyHex string) (TxReceipt, error)

	// FetchABI retrieves a contract's ABI from a configured source
	// (Etherscan/Sourcify for EVM, IDL registry for Solana) for use in
	// pre-flight validation. Returns an error if no ABI could be found.
	FetchABI(ctx context.Context, contractAddress string) (string, error)
}

// Registry-style lookup map, populated by main() at startup from config.
type Set map[string]ChainAdapter

// Get returns the adapter for name, or an error if unconfigured.
func (s Set) Get(name string) (ChainAdapter, error) {
	a, ok := s[name]
	if !ok {
		return nil, &UnknownChainError{Name: name}
	}
	return a, nil
}

// UnknownChainError indicates a request referenced an unconfigured chain.
type UnknownChainError struct{ Name string }

func (e *UnknownChainError) Error() string {
	return "chain: unknown or unconfigured chain \"" + e.Name + "\""
}
