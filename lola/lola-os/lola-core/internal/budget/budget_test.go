package budget

import (
	"testing"

	"github.com/lola-os/lola-core/internal/config"
	"github.com/lola-os/lola-core/internal/logging"
)

func newTestBreaker(t *testing.T, cfg config.BudgetConfig) *Breaker {
	t.Helper()
	b := New(cfg, logging.Default())
	t.Cleanup(b.Stop)
	return b
}

func TestBreaker_TripsOnGasLimit(t *testing.T) {
	b := newTestBreaker(t, config.BudgetConfig{
		MaxGasSpendPerSession: 1.0,
		Action:                config.BudgetActionPause,
	})

	if err := b.CheckWrite(0.5, 0); err != nil {
		t.Fatalf("expected first write under budget to pass, got %v", err)
	}
	b.Record(0.5, 0)

	if err := b.CheckWrite(0.6, 0); err == nil {
		t.Fatalf("expected write exceeding gas budget to be denied")
	}

	snap := b.Snapshot()
	if !snap.Paused {
		t.Fatalf("expected breaker to be paused after exceeding gas budget")
	}
}

func TestBreaker_TripsOnUSDLimit(t *testing.T) {
	b := newTestBreaker(t, config.BudgetConfig{
		MaxUSDSpendPerSession: 10.0,
		Action:                config.BudgetActionPause,
	})

	if err := b.CheckWrite(0, 9.0); err != nil {
		t.Fatalf("expected write under USD budget to pass: %v", err)
	}
	b.Record(0, 9.0)

	if err := b.CheckWrite(0, 5.0); err == nil {
		t.Fatalf("expected write exceeding USD budget to be denied")
	}
}

func TestBreaker_NotifyActionDoesNotBlock(t *testing.T) {
	b := newTestBreaker(t, config.BudgetConfig{
		MaxGasSpendPerSession: 1.0,
		Action:                config.BudgetActionNotify,
	})
	b.Record(2.0, 0) // already over budget

	if err := b.CheckWrite(0.1, 0); err != nil {
		t.Fatalf("notify action should not block writes, got error: %v", err)
	}
}

func TestBreaker_DenyActionBlocksImmediately(t *testing.T) {
	b := newTestBreaker(t, config.BudgetConfig{
		MaxGasSpendPerSession: 1.0,
		Action:                config.BudgetActionDeny,
	})

	if err := b.CheckWrite(2.0, 0); err == nil {
		t.Fatalf("deny action should reject a write that would exceed budget")
	}
	// Deny should not leave the breaker permanently paused — only the
	// single offending write is rejected.
	if b.Snapshot().Paused {
		t.Fatalf("deny action should not pause the breaker for future, smaller writes")
	}
}

func TestBreaker_ResumeAndResetSession(t *testing.T) {
	b := newTestBreaker(t, config.BudgetConfig{
		MaxGasSpendPerSession: 1.0,
		Action:                config.BudgetActionPause,
	})
	b.Record(2.0, 0)
	if !b.Snapshot().Paused {
		t.Fatalf("expected breaker to be paused")
	}

	b.Resume()
	if b.Snapshot().Paused {
		t.Fatalf("expected Resume to clear paused state")
	}
	// Resume doesn't reset counters, so a subsequent over-budget write
	// check should still fail.
	if err := b.CheckWrite(0.1, 0); err == nil {
		t.Fatalf("expected CheckWrite to still fail after Resume without ResetSession")
	}

	b.ResetSession()
	snap := b.Snapshot()
	if snap.GasSpent != 0 || snap.Paused {
		t.Fatalf("expected ResetSession to zero counters and clear pause, got %+v", snap)
	}
}

func TestBreaker_RateLimit(t *testing.T) {
	b := newTestBreaker(t, config.BudgetConfig{
		MaxRequestsPerMinute: 3,
		Action:               config.BudgetActionDeny,
	})
	for i := 0; i < 3; i++ {
		if err := b.CheckRate(); err != nil {
			t.Fatalf("request %d should be within rate limit, got %v", i, err)
		}
	}
	if err := b.CheckRate(); err == nil {
		t.Fatalf("4th request should exceed the per-minute rate limit")
	}
}
