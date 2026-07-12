// Package oracle implements LOLA OS's oracle gateway: reading Chainlink
// price feeds directly from their on-chain aggregator contracts, and
// making generic REST API calls with bounded retries and basic rate
// limiting for arbitrary off-chain data sources.
package oracle

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// aggregatorV3ABI covers Chainlink's standard AggregatorV3Interface,
// sufficient for reading the latest price and decimals.
const aggregatorV3ABI = `[
	{"inputs":[],"name":"latestRoundData","outputs":[
		{"internalType":"uint80","name":"roundId","type":"uint80"},
		{"internalType":"int256","name":"answer","type":"int256"},
		{"internalType":"uint256","name":"startedAt","type":"uint256"},
		{"internalType":"uint256","name":"updatedAt","type":"uint256"},
		{"internalType":"uint80","name":"answeredInRound","type":"uint80"}
	],"stateMutability":"view","type":"function"},
	{"inputs":[],"name":"decimals","outputs":[{"internalType":"uint8","name":"","type":"uint8"}],"stateMutability":"view","type":"function"}
]`

// PriceResult is the outcome of a Chainlink feed read.
type PriceResult struct {
	Pair      string
	Price     float64
	Decimals  int
	UpdatedAt time.Time
}

// Gateway provides price feed and generic REST access.
type Gateway struct {
	feeds      map[string]string // "ETH/USD" -> contract address
	httpClient *http.Client
	maxRetries int

	rl *rateLimiter
}

// New constructs a Gateway. evmClients maps chain name to an already-dialed
// ethclient (Chainlink feeds are read via standard EVM calls).
func New(feeds map[string]string, restTimeoutMS, restMaxRetries int) *Gateway {
	return &Gateway{
		feeds:      feeds,
		httpClient: &http.Client{Timeout: time.Duration(restTimeoutMS) * time.Millisecond},
		maxRetries: restMaxRetries,
		rl:         newRateLimiter(30, time.Minute), // generic default; callers can wrap with their own limits
	}
}

// GetPrice reads the latest Chainlink price for pair (e.g. "ETH/USD") using
// client, an already-connected EVM client for the chain hosting the feed.
func (g *Gateway) GetPrice(ctx context.Context, client *ethclient.Client, pair string) (PriceResult, error) {
	addr, ok := g.feeds[pair]
	if !ok {
		return PriceResult{}, fmt.Errorf("oracle: no Chainlink feed configured for pair %q", pair)
	}
	parsed, err := abi.JSON(strings.NewReader(aggregatorV3ABI))
	if err != nil {
		return PriceResult{}, fmt.Errorf("oracle: parsing aggregator ABI: %w", err)
	}
	contract := common.HexToAddress(addr)

	decData, err := parsed.Pack("decimals")
	if err != nil {
		return PriceResult{}, fmt.Errorf("oracle: packing decimals call: %w", err)
	}
	decOut, err := client.CallContract(ctx, callMsg(contract, decData), nil)
	if err != nil {
		return PriceResult{}, fmt.Errorf("oracle: calling decimals on %s: %w", addr, err)
	}
	var decimals uint8
	if err := parsed.UnpackIntoInterface(&decimals, "decimals", decOut); err != nil {
		return PriceResult{}, fmt.Errorf("oracle: unpacking decimals: %w", err)
	}

	roundData, err := parsed.Pack("latestRoundData")
	if err != nil {
		return PriceResult{}, fmt.Errorf("oracle: packing latestRoundData call: %w", err)
	}
	roundOut, err := client.CallContract(ctx, callMsg(contract, roundData), nil)
	if err != nil {
		return PriceResult{}, fmt.Errorf("oracle: calling latestRoundData on %s: %w", addr, err)
	}
	values, err := parsed.Unpack("latestRoundData", roundOut)
	if err != nil {
		return PriceResult{}, fmt.Errorf("oracle: unpacking latestRoundData: %w", err)
	}
	answer, ok := values[1].(*big.Int)
	if !ok {
		return PriceResult{}, fmt.Errorf("oracle: unexpected answer type from feed %s", addr)
	}
	updatedAtBig, _ := values[3].(*big.Int)

	divisor := new(big.Float).SetFloat64(pow10(int(decimals)))
	priceFloat := new(big.Float).Quo(new(big.Float).SetInt(answer), divisor)
	price, _ := priceFloat.Float64()

	var updatedAt time.Time
	if updatedAtBig != nil {
		updatedAt = time.Unix(updatedAtBig.Int64(), 0).UTC()
	}

	return PriceResult{Pair: pair, Price: price, Decimals: int(decimals), UpdatedAt: updatedAt}, nil
}

func pow10(n int) float64 {
	v := 1.0
	for i := 0; i < n; i++ {
		v *= 10
	}
	return v
}

func callMsg(to common.Address, data []byte) ethereum.CallMsg {
	return ethereum.CallMsg{To: &to, Data: data}
}

// FetchJSON performs a GET request against url, retrying with exponential
// backoff up to g.maxRetries times on network errors or 5xx responses, and
// decodes the JSON response body into out. It honors a simple token-bucket
// rate limiter shared across all generic REST calls from this Gateway.
func (g *Gateway) FetchJSON(ctx context.Context, url string, out interface{}) error {
	if err := g.rl.Wait(ctx); err != nil {
		return fmt.Errorf("oracle: rate limit wait: %w", err)
	}

	var lastErr error
	backoff := 250 * time.Millisecond
	for attempt := 0; attempt <= g.maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}
			backoff *= 2
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return fmt.Errorf("oracle: building request: %w", err)
		}
		resp, err := g.httpClient.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		body, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode >= 500 {
			lastErr = fmt.Errorf("server error: status %d", resp.StatusCode)
			continue
		}
		if resp.StatusCode >= 400 {
			return fmt.Errorf("oracle: request to %s failed with status %d: %s", url, resp.StatusCode, string(body))
		}
		if readErr != nil {
			lastErr = readErr
			continue
		}
		if out != nil {
			if err := json.Unmarshal(body, out); err != nil {
				return fmt.Errorf("oracle: decoding response from %s: %w", url, err)
			}
		}
		return nil
	}
	return fmt.Errorf("oracle: request to %s failed after %d attempts: %w", url, g.maxRetries+1, lastErr)
}

// --- simple token-bucket rate limiter ---------------------------------------

type rateLimiter struct {
	mu        sync.Mutex
	tokens    int
	max       int
	window    time.Duration
	lastReset time.Time
}

func newRateLimiter(maxPerWindow int, window time.Duration) *rateLimiter {
	return &rateLimiter{tokens: maxPerWindow, max: maxPerWindow, window: window, lastReset: time.Now()}
}

func (r *rateLimiter) Wait(ctx context.Context) error {
	for {
		r.mu.Lock()
		if time.Since(r.lastReset) > r.window {
			r.tokens = r.max
			r.lastReset = time.Now()
		}
		if r.tokens > 0 {
			r.tokens--
			r.mu.Unlock()
			return nil
		}
		wait := r.window - time.Since(r.lastReset)
		r.mu.Unlock()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(wait):
		}
	}
}
