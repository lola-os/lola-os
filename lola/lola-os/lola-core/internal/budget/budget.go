// Package budget implements LOLA OS's circuit breaker: a background-safe
// guard that tracks gas spent, USD-equivalent spend (via oracle price
// conversion), and request rate within the current session, and blocks
// further write operations once a configured limit is exceeded.
//
// The breaker does not run as a literal background goroutine polling a
// remote source; rather, every write operation must call Record() after
// it executes (or Check() before it executes), so accounting is always
// exact and driven by real activity rather than estimation. A lightweight
// goroutine is used only to roll over the per-minute request-rate window.
package budget

import (
	"fmt"
	"sync"
	"time"

	"github.com/lola-os/lola-core/internal/config"
	"github.com/lola-os/lola-core/internal/logging"
)

// ErrBudgetExceeded is returned by Check when a limit has been exceeded and
// the configured action is "deny" or "pause".
type ErrBudgetExceeded struct {
	Limit   string // "gas", "usd", or "rate"
	Limit_  float64
	Current float64
}

func (e *ErrBudgetExceeded) Error() string {
	return fmt.Sprintf("budget exceeded: %s limit is %.4f, current usage is %.4f", e.Limit, e.Limit_, e.Current)
}

// State is a point-in-time snapshot of the breaker's session accounting.
type State struct {
	GasSpent      float64
	USDSpent      float64
	RequestsThisMinute int
	Paused        bool
	PausedReason  string
}

// Breaker tracks spend within a single lola-core process lifetime ("session").
// It is safe for concurrent use.
type Breaker struct {
	mu sync.Mutex

	cfg    config.BudgetConfig
	logger *logging.Logger

	gasSpent float64
	usdSpent float64

	windowStart   time.Time
	requestsInWin int

	paused       bool
	pausedReason string

	stopCh chan struct{}
}

// New constructs a Breaker from the given budget configuration.
func New(cfg config.BudgetConfig, logger *logging.Logger) *Breaker {
	if logger == nil {
		logger = logging.Default()
	}
	b := &Breaker{
		cfg:         cfg,
		logger:      logger,
		windowStart: time.Now(),
		stopCh:      make(chan struct{}),
	}
	go b.rateWindowLoop()
	return b
}

// Stop terminates the background rate-window roller. Call on shutdown.
func (b *Breaker) Stop() {
	close(b.stopCh)
}

func (b *Breaker) rateWindowLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			b.mu.Lock()
			b.requestsInWin = 0
			b.windowStart = time.Now()
			b.mu.Unlock()
		case <-b.stopCh:
			return
		}
	}
}

// CheckRate records one request against the per-minute rate limit and
// returns an error if the limit has been exceeded and the action requires
// blocking.
func (b *Breaker) CheckRate() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.requestsInWin++
	if b.cfg.MaxRequestsPerMinute > 0 && b.requestsInWin > b.cfg.MaxRequestsPerMinute {
		err := &ErrBudgetExceeded{Limit: "rate", Limit_: float64(b.cfg.MaxRequestsPerMinute), Current: float64(b.requestsInWin)}
		return b.handleExceeded(err)
	}
	return nil
}

// CheckWrite verifies, *before* a write operation executes, that
// proceeding would not push spend over budget, and that the breaker is not
// already paused. estimatedGas and estimatedUSD are the operation's
// projected cost (e.g. from gas estimation + oracle price).
func (b *Breaker) CheckWrite(estimatedGas, estimatedUSD float64) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.paused {
		return fmt.Errorf("budget: write operations are paused (%s)", b.pausedReason)
	}

	if b.cfg.MaxGasSpendPerSession > 0 && b.gasSpent+estimatedGas > b.cfg.MaxGasSpendPerSession {
		err := &ErrBudgetExceeded{Limit: "gas", Limit_: b.cfg.MaxGasSpendPerSession, Current: b.gasSpent + estimatedGas}
		return b.handleExceeded(err)
	}
	if b.cfg.MaxUSDSpendPerSession > 0 && b.usdSpent+estimatedUSD > b.cfg.MaxUSDSpendPerSession {
		err := &ErrBudgetExceeded{Limit: "usd", Limit_: b.cfg.MaxUSDSpendPerSession, Current: b.usdSpent + estimatedUSD}
		return b.handleExceeded(err)
	}
	return nil
}

// Record commits the actual gas/USD cost of a completed write operation to
// the session totals. Call this *after* a transaction is confirmed (or
// known to have consumed gas), using actual — not estimated — values.
func (b *Breaker) Record(actualGas, actualUSD float64) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.gasSpent += actualGas
	b.usdSpent += actualUSD

	// Auto-pausing on recorded spend is only correct under the "pause"
	// action. Under "notify" the breaker must keep writes flowing (it only
	// logs), and under "deny" each offending write is rejected individually
	// by CheckWrite rather than latching the whole breaker paused.
	if b.cfg.Action != config.BudgetActionPause {
		return
	}
	if b.cfg.MaxGasSpendPerSession > 0 && b.gasSpent >= b.cfg.MaxGasSpendPerSession {
		b.trip(fmt.Sprintf("gas budget of %.4f reached (spent %.4f)", b.cfg.MaxGasSpendPerSession, b.gasSpent))
	}
	if b.cfg.MaxUSDSpendPerSession > 0 && b.usdSpent >= b.cfg.MaxUSDSpendPerSession {
		b.trip(fmt.Sprintf("USD budget of %.2f reached (spent %.2f)", b.cfg.MaxUSDSpendPerSession, b.usdSpent))
	}
}

// handleExceeded applies the configured BudgetAction once a limit check
// fails. Must be called with b.mu held.
func (b *Breaker) handleExceeded(err *ErrBudgetExceeded) error {
	switch b.cfg.Action {
	case config.BudgetActionNotify:
		b.logger.Warn("budget limit exceeded (notify-only, write proceeding)", map[string]interface{}{
			"limit": err.Limit, "max": err.Limit_, "current": err.Current,
		})
		return nil
	case config.BudgetActionDeny:
		b.logger.Error("budget limit exceeded, denying this operation", map[string]interface{}{
			"limit": err.Limit, "max": err.Limit_, "current": err.Current,
		})
		return err
	default: // pause
		b.trip(err.Error())
		return err
	}
}

// trip flips the breaker into the paused state and emits a CRITICAL log.
// Must be called with b.mu held.
func (b *Breaker) trip(reason string) {
	if b.paused {
		return
	}
	b.paused = true
	b.pausedReason = reason
	b.logger.Critical("circuit breaker tripped — all write operations paused", map[string]interface{}{
		"reason": reason,
	})
}

// Resume clears the paused state, e.g. after an operator reviews session
// spend and decides to continue. This does NOT reset the underlying
// counters — call ResetSession for that.
func (b *Breaker) Resume() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.paused = false
	b.pausedReason = ""
}

// ResetSession zeroes all session spend counters. Typically called at the
// start of a new `lola-core` process, or explicitly via the JSON-RPC API.
func (b *Breaker) ResetSession() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.gasSpent = 0
	b.usdSpent = 0
	b.paused = false
	b.pausedReason = ""
}

// Snapshot returns the current breaker state for diagnostics/metrics.
func (b *Breaker) Snapshot() State {
	b.mu.Lock()
	defer b.mu.Unlock()
	return State{
		GasSpent:           b.gasSpent,
		USDSpent:            b.usdSpent,
		RequestsThisMinute: b.requestsInWin,
		Paused:             b.paused,
		PausedReason:       b.pausedReason,
	}
}
