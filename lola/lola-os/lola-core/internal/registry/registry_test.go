package registry

import (
	"path/filepath"
	"testing"
	"time"
)

func newTestRegistry(t *testing.T) *Registry {
	t.Helper()
	reg, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("opening registry: %v", err)
	}
	t.Cleanup(func() { reg.Close() })
	return reg
}

func TestRegistry_TransactionRoundTrip(t *testing.T) {
	reg := newTestRegistry(t)
	tx := Transaction{
		Hash: "0xabc", Chain: "ethereum", From: "0xfrom", To: "0xto",
		Value: "1000", Gas: 21000, GasPrice: "5", Status: TxStatusPending, Timestamp: time.Now().UTC(),
	}
	if err := reg.RecordTransaction(tx); err != nil {
		t.Fatalf("RecordTransaction failed: %v", err)
	}

	got, err := reg.GetTransaction("0xabc")
	if err != nil {
		t.Fatalf("GetTransaction failed: %v", err)
	}
	if got == nil || got.Hash != "0xabc" || got.Status != TxStatusPending {
		t.Fatalf("unexpected result: %+v", got)
	}

	tx.Status = TxStatusConfirmed
	if err := reg.RecordTransaction(tx); err != nil {
		t.Fatalf("upsert failed: %v", err)
	}
	got, _ = reg.GetTransaction("0xabc")
	if got.Status != TxStatusConfirmed {
		t.Fatalf("expected upsert to update status to confirmed, got %s", got.Status)
	}
}

func TestRegistry_NoncePersistence(t *testing.T) {
	reg := newTestRegistry(t)
	if _, exists, _ := reg.GetNonce("ethereum", "0xabc"); exists {
		t.Fatalf("expected no nonce initially")
	}
	if err := reg.SetNonce("ethereum", "0xabc", 5); err != nil {
		t.Fatalf("SetNonce failed: %v", err)
	}
	n, exists, err := reg.GetNonce("ethereum", "0xabc")
	if err != nil || !exists || n != 5 {
		t.Fatalf("expected nonce 5, got n=%d exists=%v err=%v", n, exists, err)
	}
}

func TestRegistry_ExecutionPlanLifecycle(t *testing.T) {
	reg := newTestRegistry(t)
	if err := reg.CreatePlan("plan-1", "test plan"); err != nil {
		t.Fatalf("CreatePlan failed: %v", err)
	}
	p, err := reg.GetPlan("plan-1")
	if err != nil || p == nil || p.Status != PlanStatusRunning {
		t.Fatalf("expected running plan, got %+v err=%v", p, err)
	}

	if err := reg.FinishPlan("plan-1", PlanStatusCompleted, map[string]string{"ok": "true"}); err != nil {
		t.Fatalf("FinishPlan failed: %v", err)
	}
	p, _ = reg.GetPlan("plan-1")
	if p.Status != PlanStatusCompleted || p.FinishedAt == nil {
		t.Fatalf("expected completed plan with finished_at set, got %+v", p)
	}
}

func TestRegistry_IdempotencyKeyTTL(t *testing.T) {
	reg := newTestRegistry(t)
	if err := reg.SetIdempotentResult("key1", `{"tx_hash":"0xabc"}`); err != nil {
		t.Fatalf("SetIdempotentResult failed: %v", err)
	}

	result, found, err := reg.GetIdempotentResult("key1", time.Hour)
	if err != nil || !found || result != `{"tx_hash":"0xabc"}` {
		t.Fatalf("expected cached result within TTL, got found=%v result=%q err=%v", found, result, err)
	}

	_, found, err = reg.GetIdempotentResult("key1", -time.Second) // already-expired window
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found {
		t.Fatalf("expected expired idempotency key to not be found")
	}
}

func TestRegistry_ClearAll(t *testing.T) {
	reg := newTestRegistry(t)
	_ = reg.RecordTransaction(Transaction{Hash: "0xabc", Chain: "ethereum", Status: TxStatusPending, Timestamp: time.Now()})
	_ = reg.SetNonce("ethereum", "0xabc", 1)
	_ = reg.CreatePlan("plan-1", "x")

	if err := reg.ClearAll(); err != nil {
		t.Fatalf("ClearAll failed: %v", err)
	}

	txs, _ := reg.ListTransactions(0)
	if len(txs) != 0 {
		t.Fatalf("expected transactions to be cleared, found %d", len(txs))
	}
	if _, exists, _ := reg.GetNonce("ethereum", "0xabc"); exists {
		t.Fatalf("expected nonces to be cleared")
	}
	if p, _ := reg.GetPlan("plan-1"); p != nil {
		t.Fatalf("expected execution_plans to be cleared")
	}
}
