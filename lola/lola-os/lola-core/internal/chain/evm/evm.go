// Package evm implements chain.ChainAdapter for EVM-compatible chains
// (Ethereum, Polygon, and any other chain reachable via a standard
// JSON-RPC endpoint), using go-ethereum's ethclient and abi packages.
package evm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"time"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	abivalidate "github.com/lola-os/lola-core/internal/abi"
	"github.com/lola-os/lola-core/internal/chain"
)

// erc20ABI covers the handful of standard ERC20 read methods LOLA needs
// for balance lookups without requiring the caller to supply a full ABI.
const erc20ABI = `[
	{"constant":true,"inputs":[{"name":"_owner","type":"address"}],"name":"balanceOf","outputs":[{"name":"balance","type":"uint256"}],"type":"function"},
	{"constant":true,"inputs":[],"name":"decimals","outputs":[{"name":"","type":"uint8"}],"type":"function"},
	{"constant":true,"inputs":[],"name":"symbol","outputs":[{"name":"","type":"string"}],"type":"function"}
]`

// Config carries everything the EVM adapter needs to serve one chain. It maps
// directly from config.ChainConfig / the built-in chain catalog.
type Config struct {
	Name           string
	RPCURL         string
	ChainID        int64
	NativeSymbol   string
	NativeDecimals int
	ExplorerAPI    string
	ExplorerKey    string
}

// Adapter implements chain.ChainAdapter for a single EVM-compatible chain.
type Adapter struct {
	name           string
	chainID        *big.Int
	client         *ethclient.Client
	rpcURL         string
	nativeSymbol   string
	nativeDecimals int
	explorerAPI    string
	explorerKey    string
	httpClient     *http.Client
}

// New dials cfg.RPCURL and returns a ready-to-use Adapter. The native-asset
// symbol/decimals come from configuration (fed by the chain catalog), so
// balances on any EVM chain — ETH, MATIC, BNB, AVAX, and so on — are labelled
// correctly without chain-specific code here.
func New(ctx context.Context, cfg Config) (*Adapter, error) {
	client, err := ethclient.DialContext(ctx, cfg.RPCURL)
	if err != nil {
		return nil, fmt.Errorf("evm[%s]: dialing %s: %w", cfg.Name, cfg.RPCURL, err)
	}
	symbol := cfg.NativeSymbol
	if symbol == "" {
		symbol = "ETH"
	}
	decimals := cfg.NativeDecimals
	if decimals == 0 {
		decimals = 18
	}
	a := &Adapter{
		name:           cfg.Name,
		chainID:        big.NewInt(cfg.ChainID),
		client:         client,
		rpcURL:         cfg.RPCURL,
		nativeSymbol:   symbol,
		nativeDecimals: decimals,
		explorerAPI:    cfg.ExplorerAPI,
		explorerKey:    cfg.ExplorerKey,
		httpClient:     &http.Client{Timeout: 10 * time.Second},
	}
	return a, nil
}

func (a *Adapter) Name() string { return a.name }
func (a *Adapter) Kind() string { return "evm" }

// EVMClient exposes the underlying go-ethereum client so callers that need a
// raw connection (e.g. the oracle's Chainlink aggregator reads) can reuse the
// one this adapter already dialed, rather than opening a second connection.
func (a *Adapter) EVMClient() *ethclient.Client { return a.client }

// AddressFromKey derives the 0x-hex account address for an ECDSA private key.
func (a *Adapter) AddressFromKey(privateKeyHex string) (string, error) {
	key, err := crypto.HexToECDSA(strings.TrimPrefix(privateKeyHex, "0x"))
	if err != nil {
		return "", fmt.Errorf("evm[%s]: parsing private key: %w", a.name, err)
	}
	return crypto.PubkeyToAddress(key.PublicKey).Hex(), nil
}

func (a *Adapter) Ping(ctx context.Context) (uint64, error) {
	n, err := a.client.BlockNumber(ctx)
	if err != nil {
		return 0, fmt.Errorf("evm[%s]: ping failed: %w", a.name, err)
	}
	return n, nil
}

func (a *Adapter) NativeBalance(ctx context.Context, address string) (chain.Balance, error) {
	addr := common.HexToAddress(address)
	bal, err := a.client.BalanceAt(ctx, addr, nil)
	if err != nil {
		return chain.Balance{}, fmt.Errorf("evm[%s]: native balance for %s: %w", a.name, address, err)
	}
	return chain.Balance{Address: address, RawValue: bal, Decimals: a.nativeDecimals, Symbol: a.nativeSymbol}, nil
}

func (a *Adapter) TokenBalance(ctx context.Context, address, tokenAddress string) (chain.Balance, error) {
	parsed, err := abi.JSON(strings.NewReader(erc20ABI))
	if err != nil {
		return chain.Balance{}, fmt.Errorf("evm[%s]: parsing ERC20 ABI: %w", a.name, err)
	}
	tokenAddr := common.HexToAddress(tokenAddress)
	owner := common.HexToAddress(address)

	balData, err := parsed.Pack("balanceOf", owner)
	if err != nil {
		return chain.Balance{}, fmt.Errorf("evm[%s]: packing balanceOf: %w", a.name, err)
	}
	balOut, err := a.client.CallContract(ctx, gethCallMsg(tokenAddr, balData), nil)
	if err != nil {
		return chain.Balance{}, fmt.Errorf("evm[%s]: calling balanceOf: %w", a.name, err)
	}
	var balance *big.Int
	if err := parsed.UnpackIntoInterface(&balance, "balanceOf", balOut); err != nil {
		return chain.Balance{}, fmt.Errorf("evm[%s]: unpacking balanceOf: %w", a.name, err)
	}

	decimals := uint8(18)
	if decData, err := parsed.Pack("decimals"); err == nil {
		if decOut, err := a.client.CallContract(ctx, gethCallMsg(tokenAddr, decData), nil); err == nil {
			_ = parsed.UnpackIntoInterface(&decimals, "decimals", decOut)
		}
	}
	symbol := ""
	if symData, err := parsed.Pack("symbol"); err == nil {
		if symOut, err := a.client.CallContract(ctx, gethCallMsg(tokenAddr, symData), nil); err == nil {
			_ = parsed.UnpackIntoInterface(&symbol, "symbol", symOut)
		}
	}

	return chain.Balance{
		Address: address, Token: tokenAddress, RawValue: balance,
		Decimals: int(decimals), Symbol: symbol,
	}, nil
}

func gethCallMsg(to common.Address, data []byte) ethereum.CallMsg {
	return ethereum.CallMsg{To: &to, Data: data}
}

func (a *Adapter) PendingNonce(ctx context.Context, address string) (uint64, error) {
	n, err := a.client.PendingNonceAt(ctx, common.HexToAddress(address))
	if err != nil {
		return 0, fmt.Errorf("evm[%s]: pending nonce for %s: %w", a.name, address, err)
	}
	return n, nil
}

func (a *Adapter) EstimateGas(ctx context.Context, req chain.TxRequest) (*big.Int, error) {
	to := common.HexToAddress(req.To)
	gasLimit, err := a.client.EstimateGas(ctx, ethereum.CallMsg{From: common.HexToAddress(req.From), To: &to, Value: req.ValueWei, Data: req.Data})
	if err != nil {
		return nil, fmt.Errorf("evm[%s]: estimating gas: %w", a.name, err)
	}
	gasPrice, err := a.client.SuggestGasPrice(ctx)
	if err != nil {
		return nil, fmt.Errorf("evm[%s]: suggesting gas price: %w", a.name, err)
	}
	cost := new(big.Int).Mul(big.NewInt(int64(gasLimit)), gasPrice)
	return cost, nil
}

func (a *Adapter) SendTransaction(ctx context.Context, req chain.TxRequest, privateKeyHex string) (chain.TxReceipt, error) {
	key, err := crypto.HexToECDSA(strings.TrimPrefix(privateKeyHex, "0x"))
	if err != nil {
		return chain.TxReceipt{}, fmt.Errorf("evm[%s]: parsing private key: %w", a.name, err)
	}

	var nonce uint64
	if req.Nonce != nil {
		nonce = *req.Nonce
	} else {
		nonce, err = a.PendingNonce(ctx, req.From)
		if err != nil {
			return chain.TxReceipt{}, err
		}
	}

	gasTipCap := req.MaxPriorityFee
	gasFeeCap := req.MaxFeePerGas
	if gasTipCap == nil || gasFeeCap == nil {
		suggested, err := a.client.SuggestGasPrice(ctx)
		if err != nil {
			return chain.TxReceipt{}, fmt.Errorf("evm[%s]: suggesting gas price: %w", a.name, err)
		}
		gasTipCap = suggested
		gasFeeCap = new(big.Int).Mul(suggested, big.NewInt(2))
	}

	gasLimit := req.GasLimit
	if gasLimit == 0 {
		to := common.HexToAddress(req.To)
		gasLimit, err = a.client.EstimateGas(ctx, ethereum.CallMsg{From: common.HexToAddress(req.From), To: &to, Value: req.ValueWei, Data: req.Data})
		if err != nil {
			return chain.TxReceipt{}, fmt.Errorf("evm[%s]: estimating gas: %w", a.name, err)
		}
	}

	value := req.ValueWei
	if value == nil {
		value = big.NewInt(0)
	}
	to := common.HexToAddress(req.To)

	tx := gethtypes.NewTx(&gethtypes.DynamicFeeTx{
		ChainID:   a.chainID,
		Nonce:     nonce,
		GasTipCap: gasTipCap,
		GasFeeCap: gasFeeCap,
		Gas:       gasLimit,
		To:        &to,
		Value:     value,
		Data:      req.Data,
	})

	signer := gethtypes.NewLondonSigner(a.chainID)
	signedTx, err := gethtypes.SignTx(tx, signer, key)
	if err != nil {
		return chain.TxReceipt{}, fmt.Errorf("evm[%s]: signing transaction: %w", a.name, err)
	}

	if err := a.client.SendTransaction(ctx, signedTx); err != nil {
		return chain.TxReceipt{}, fmt.Errorf("evm[%s]: broadcasting transaction: %w", a.name, err)
	}

	return chain.TxReceipt{Hash: signedTx.Hash().Hex(), Status: "pending"}, nil
}

func (a *Adapter) WaitForReceipt(ctx context.Context, hash string) (chain.TxReceipt, error) {
	txHash := common.HexToHash(hash)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return chain.TxReceipt{}, ctx.Err()
		case <-ticker.C:
			receipt, err := a.client.TransactionReceipt(ctx, txHash)
			if err != nil {
				continue // not yet mined
			}
			status := "confirmed"
			if receipt.Status == 0 {
				status = "failed"
			}
			return chain.TxReceipt{
				Hash:              hash,
				Status:            status,
				BlockNumber:       receipt.BlockNumber.Uint64(),
				GasUsed:           receipt.GasUsed,
				EffectiveGasPrice: receipt.EffectiveGasPrice,
			}, nil
		}
	}
}

func (a *Adapter) CallContract(ctx context.Context, req chain.ContractCallRequest) (interface{}, error) {
	parsedABI, callData, method, err := a.encodeCall(req)
	if err != nil {
		return nil, err
	}
	contractAddr := common.HexToAddress(req.ContractAddress)
	out, err := a.client.CallContract(ctx, gethCallMsg(contractAddr, callData), nil)
	if err != nil {
		return nil, fmt.Errorf("evm[%s]: calling %s.%s: %w", a.name, req.ContractAddress, req.Method, err)
	}
	values, err := parsedABI.Unpack(method.Name, out)
	if err != nil {
		return nil, fmt.Errorf("evm[%s]: unpacking result of %s: %w", a.name, req.Method, err)
	}
	if len(values) == 1 {
		return values[0], nil
	}
	return values, nil
}

func (a *Adapter) ExecuteContract(ctx context.Context, req chain.ContractCallRequest, privateKeyHex string) (chain.TxReceipt, error) {
	_, callData, _, err := a.encodeCall(req)
	if err != nil {
		return chain.TxReceipt{}, err
	}
	txReq := chain.TxRequest{
		From:     req.From,
		To:       req.ContractAddress,
		ValueWei: req.ValueWei,
		Data:     callData,
	}
	return a.SendTransaction(ctx, txReq, privateKeyHex)
}

func (a *Adapter) encodeCall(req chain.ContractCallRequest) (abi.ABI, []byte, *abi.Method, error) {
	abiJSON := req.ABI
	if abiJSON == "" {
		fetched, err := a.FetchABI(context.Background(), req.ContractAddress)
		if err != nil {
			return abi.ABI{}, nil, nil, fmt.Errorf("evm[%s]: no ABI provided and fetch failed: %w", a.name, err)
		}
		abiJSON = fetched
	}
	parsed, err := abi.JSON(strings.NewReader(abiJSON))
	if err != nil {
		return abi.ABI{}, nil, nil, fmt.Errorf("evm[%s]: parsing ABI: %w", a.name, err)
	}
	method, ok := parsed.Methods[req.Method]
	if !ok {
		return abi.ABI{}, nil, nil, fmt.Errorf("evm[%s]: method %q not found in ABI", a.name, req.Method)
	}
	// Arguments arrive from the SDKs as plain JSON values (hex strings,
	// numbers, slices). Coerce them into the concrete Go types the ABI
	// encoder requires before packing.
	coerced, err := abivalidate.CoerceArgs(method.Inputs, req.Args)
	if err != nil {
		return abi.ABI{}, nil, nil, fmt.Errorf("evm[%s]: coercing args for %s: %w", a.name, req.Method, err)
	}
	data, err := parsed.Pack(req.Method, coerced...)
	if err != nil {
		return abi.ABI{}, nil, nil, fmt.Errorf("evm[%s]: encoding args for %s: %w", a.name, req.Method, err)
	}
	return parsed, data, &method, nil
}

// FetchABI retrieves a verified contract ABI from the configured explorer
// API (Etherscan-compatible "getabi" endpoint). If no explorer is
// configured, it returns an error advising the caller to supply an ABI
// directly.
func (a *Adapter) FetchABI(ctx context.Context, contractAddress string) (string, error) {
	if a.explorerAPI == "" {
		return "", fmt.Errorf("evm[%s]: no explorer API configured; pass `abi` explicitly for %s", a.name, contractAddress)
	}
	url := fmt.Sprintf("%s?module=contract&action=getabi&address=%s&apikey=%s", a.explorerAPI, contractAddress, a.explorerKey)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("evm[%s]: building ABI request: %w", a.name, err)
	}
	resp, err := a.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("evm[%s]: fetching ABI: %w", a.name, err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("evm[%s]: reading ABI response: %w", a.name, err)
	}
	var parsed struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Result  string `json:"result"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", fmt.Errorf("evm[%s]: parsing explorer response: %w", a.name, err)
	}
	if parsed.Status != "1" {
		return "", fmt.Errorf("evm[%s]: explorer error: %s", a.name, parsed.Message)
	}
	return parsed.Result, nil
}
