// Package solana implements chain.ChainAdapter for the Solana network,
// using github.com/gagliardetto/solana-go for RPC access, transaction
// building, and signing.
package solana

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"

	solanago "github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/rpc"

	"github.com/lola-os/lola-core/internal/chain"
)

const lamportsPerSOL = 1_000_000_000

// Adapter implements chain.ChainAdapter for Solana.
type Adapter struct {
	name   string
	client *rpc.Client
	rpcURL string
}

// New constructs a Solana Adapter pointed at rpcURL.
func New(name, rpcURL string) *Adapter {
	return &Adapter{
		name:   name,
		client: rpc.New(rpcURL),
		rpcURL: rpcURL,
	}
}

func (a *Adapter) Name() string { return a.name }
func (a *Adapter) Kind() string { return "solana" }

// AddressFromKey derives the base58 public key for a base58 private key.
func (a *Adapter) AddressFromKey(privateKeyHex string) (string, error) {
	priv, err := solanago.PrivateKeyFromBase58(privateKeyHex)
	if err != nil {
		return "", fmt.Errorf("solana[%s]: parsing private key: %w", a.name, err)
	}
	return priv.PublicKey().String(), nil
}

func (a *Adapter) Ping(ctx context.Context) (uint64, error) {
	slot, err := a.client.GetSlot(ctx, rpc.CommitmentConfirmed)
	if err != nil {
		return 0, fmt.Errorf("solana[%s]: ping failed: %w", a.name, err)
	}
	return slot, nil
}

func (a *Adapter) NativeBalance(ctx context.Context, address string) (chain.Balance, error) {
	pub, err := solanago.PublicKeyFromBase58(address)
	if err != nil {
		return chain.Balance{}, fmt.Errorf("solana[%s]: invalid address %s: %w", a.name, address, err)
	}
	out, err := a.client.GetBalance(ctx, pub, rpc.CommitmentConfirmed)
	if err != nil {
		return chain.Balance{}, fmt.Errorf("solana[%s]: balance for %s: %w", a.name, address, err)
	}
	return chain.Balance{
		Address:  address,
		RawValue: bigIntFromUint64(out.Value),
		Decimals: 9,
		Symbol:   "SOL",
	}, nil
}

// TokenBalance returns the SPL token balance for address's largest token
// account holding tokenAddress (mint). For accounts with multiple token
// accounts for the same mint, callers needing exact per-account balances
// should use a dedicated RPC call with the specific token account address.
func (a *Adapter) TokenBalance(ctx context.Context, address, tokenAddress string) (chain.Balance, error) {
	owner, err := solanago.PublicKeyFromBase58(address)
	if err != nil {
		return chain.Balance{}, fmt.Errorf("solana[%s]: invalid owner address: %w", a.name, err)
	}
	mint, err := solanago.PublicKeyFromBase58(tokenAddress)
	if err != nil {
		return chain.Balance{}, fmt.Errorf("solana[%s]: invalid mint address: %w", a.name, err)
	}

	out, err := a.client.GetTokenAccountsByOwner(
		ctx,
		owner,
		&rpc.GetTokenAccountsConfig{Mint: &mint},
		&rpc.GetTokenAccountsOpts{Encoding: solanago.EncodingJSONParsed},
	)
	if err != nil {
		return chain.Balance{}, fmt.Errorf("solana[%s]: fetching token accounts: %w", a.name, err)
	}
	if len(out.Value) == 0 {
		return chain.Balance{Address: address, Token: tokenAddress, RawValue: bigIntFromUint64(0), Decimals: 0}, nil
	}

	// Sum balances across all token accounts for this mint (an owner may
	// have more than one account for the same mint).
	// With EncodingJSONParsed the node returns the SPL token account as
	// parsed JSON. We read it via the raw JSON payload rather than a
	// library-specific accessor, so we stay resilient to solana-go API
	// changes. Shape: { "parsed": { "info": { "tokenAmount": {
	// "amount": "<uint>", "decimals": <n> } } } }.
	var total uint64
	var decimals int
	for _, acc := range out.Value {
		raw := acc.Account.Data.GetRawJSON()
		if len(raw) == 0 {
			continue
		}
		var parsed struct {
			Parsed struct {
				Info struct {
					TokenAmount struct {
						Amount   string `json:"amount"`
						Decimals int    `json:"decimals"`
					} `json:"tokenAmount"`
				} `json:"info"`
			} `json:"parsed"`
		}
		if err := json.Unmarshal(raw, &parsed); err != nil {
			continue
		}
		decimals = parsed.Parsed.Info.TokenAmount.Decimals
		var amt uint64
		fmt.Sscanf(parsed.Parsed.Info.TokenAmount.Amount, "%d", &amt)
		total += amt
	}

	return chain.Balance{
		Address: address, Token: tokenAddress,
		RawValue: bigIntFromUint64(total), Decimals: decimals, Symbol: "",
	}, nil
}

func bigIntFromUint64(v uint64) *big.Int {
	return new(big.Int).SetUint64(v)
}

func (a *Adapter) EstimateGas(ctx context.Context, req chain.TxRequest) (*big.Int, error) {
	// Solana fees are flat per-signature (currently 5000 lamports per
	// signature) rather than a gas-limit * price model. We return a
	// conservative estimate; callers needing the live fee should use
	// GetFeeForMessage against an assembled transaction.
	fee, err := a.client.GetRecentBlockhash(ctx, rpc.CommitmentConfirmed)
	if err != nil {
		return bigIntFromUint64(5000), nil // fall back to the known base fee
	}
	_ = fee
	return bigIntFromUint64(5000), nil
}

func (a *Adapter) PendingNonce(ctx context.Context, address string) (uint64, error) {
	// Solana does not use account nonces in the EVM sense by default;
	// ordering is handled via recent blockhashes. We return 0 so the
	// generic nonce.Manager treats this as "no persistent counter needed".
	// Durable nonce accounts (for offline signing) are a future extension
	// point and intentionally not implemented here.
	return 0, nil
}

func (a *Adapter) SendTransaction(ctx context.Context, req chain.TxRequest, privateKeyHex string) (chain.TxReceipt, error) {
	priv, err := solanago.PrivateKeyFromBase58(privateKeyHex)
	if err != nil {
		return chain.TxReceipt{}, fmt.Errorf("solana[%s]: parsing private key: %w", a.name, err)
	}

	recent, err := a.client.GetLatestBlockhash(ctx, rpc.CommitmentFinalized)
	if err != nil {
		return chain.TxReceipt{}, fmt.Errorf("solana[%s]: fetching recent blockhash: %w", a.name, err)
	}

	to, err := solanago.PublicKeyFromBase58(req.To)
	if err != nil {
		return chain.TxReceipt{}, fmt.Errorf("solana[%s]: invalid recipient: %w", a.name, err)
	}
	from := priv.PublicKey()

	var lamports uint64
	if req.ValueWei != nil {
		lamports = req.ValueWei.Uint64()
	}

	instr := system.NewTransferInstruction(lamports, from, to).Build()

	tx, err := solanago.NewTransaction(
		[]solanago.Instruction{instr},
		recent.Value.Blockhash,
		solanago.TransactionPayer(from),
	)
	if err != nil {
		return chain.TxReceipt{}, fmt.Errorf("solana[%s]: building transaction: %w", a.name, err)
	}

	_, err = tx.Sign(func(key solanago.PublicKey) *solanago.PrivateKey {
		if key.Equals(from) {
			return &priv
		}
		return nil
	})
	if err != nil {
		return chain.TxReceipt{}, fmt.Errorf("solana[%s]: signing transaction: %w", a.name, err)
	}

	sig, err := a.client.SendTransaction(ctx, tx)
	if err != nil {
		return chain.TxReceipt{}, fmt.Errorf("solana[%s]: broadcasting transaction: %w", a.name, err)
	}

	return chain.TxReceipt{Hash: sig.String(), Status: "pending"}, nil
}

func (a *Adapter) WaitForReceipt(ctx context.Context, hash string) (chain.TxReceipt, error) {
	sig, err := solanago.SignatureFromBase58(hash)
	if err != nil {
		return chain.TxReceipt{}, fmt.Errorf("solana[%s]: invalid signature %s: %w", a.name, hash, err)
	}
	result, err := a.client.GetSignatureStatuses(ctx, true, sig)
	if err != nil {
		return chain.TxReceipt{}, fmt.Errorf("solana[%s]: fetching signature status: %w", a.name, err)
	}
	if len(result.Value) == 0 || result.Value[0] == nil {
		return chain.TxReceipt{Hash: hash, Status: "pending"}, nil
	}
	status := result.Value[0]
	receipt := chain.TxReceipt{Hash: hash, Status: "confirmed"}
	if status.Err != nil {
		receipt.Status = "failed"
		receipt.Error = fmt.Sprintf("%v", status.Err)
	}
	if status.Slot > 0 {
		receipt.BlockNumber = status.Slot
	}
	return receipt, nil
}

// CallContract is not applicable to Solana in the generic ABI sense used
// by EVM (Solana programs use instruction-based calls with IDLs rather
// than a uniform ABI). Read-only "calls" should instead be performed via
// the higher-level account-fetching methods (NativeBalance, TokenBalance)
// or a program-specific extension. We return a clear error rather than a
// silent no-op so SDK callers get an actionable message.
func (a *Adapter) CallContract(ctx context.Context, req chain.ContractCallRequest) (interface{}, error) {
	return nil, fmt.Errorf("solana[%s]: generic CallContract is not supported; use program-specific instructions or NativeBalance/TokenBalance", a.name)
}

func (a *Adapter) ExecuteContract(ctx context.Context, req chain.ContractCallRequest, privateKeyHex string) (chain.TxReceipt, error) {
	return chain.TxReceipt{}, fmt.Errorf("solana[%s]: generic ExecuteContract is not supported; build a program-specific instruction and use SendTransaction", a.name)
}

func (a *Adapter) FetchABI(ctx context.Context, contractAddress string) (string, error) {
	return "", fmt.Errorf("solana[%s]: ABI/IDL fetching is not implemented; supply program instruction data directly", a.name)
}
