package replay

import (
	"context"
	"math/big"
	"path/filepath"
	"testing"

	"github.com/lola-os/lola-core/internal/chain"
	"github.com/lola-os/lola-core/internal/logging"
	"github.com/lola-os/lola-core/internal/registry"
)

// fakeAdapter is a minimal in-memory chain.ChainAdapter used to test the
// replay engine without any real network access.
type fakeAdapter struct {
	name        string
	callResults map[string]interface{} // method -> canned return value
	txCounter   int
}

func (f *fakeAdapter) Name() string { return f.name }
func (f *fakeAdapter) Kind() string { return "evm" }

func (f *fakeAdapter) AddressFromKey(privateKeyHex string) (string, error) {
	return "0x0000000000000000000000000000000000000000", nil
}
func (f *fakeAdapter) Ping(ctx context.Context) (uint64, error) { return 1, nil }
func (f *fakeAdapter) NativeBalance(ctx context.Context, address string) (chain.Balance, error) {
	return chain.Balance{Address: address, RawValue: big.NewInt(0)}, nil
}
func (f *fakeAdapter) TokenBalance(ctx context.Context, address, token string) (chain.Balance, error) {
	return chain.Balance{Address: address, RawValue: big.NewInt(0)}, nil
}
func (f *fakeAdapter) EstimateGas(ctx context.Context, req chain.TxRequest) (*big.Int, error) {
	return big.NewInt(21000), nil
}
func (f *fakeAdapter) PendingNonce(ctx context.Context, address string) (uint64, error) { return 0, nil }
func (f *fakeAdapter) SendTransaction(ctx context.Context, req chain.TxRequest, key string) (chain.TxReceipt, error) {
	f.txCounter++
	return chain.TxReceipt{Hash: "0xhash" + string(rune('0'+f.txCounter)), Status: "pending"}, nil
}
func (f *fakeAdapter) WaitForReceipt(ctx context.Context, hash string) (chain.TxReceipt, error) {
	return chain.TxReceipt{Hash: hash, Status: "confirmed"}, nil
}
func (f *fakeAdapter) CallContract(ctx context.Context, req chain.ContractCallRequest) (interface{}, error) {
	if v, ok := f.callResults[req.Method]; ok {
		return v, nil
	}
	return nil, nil
}
func (f *fakeAdapter) ExecuteContract(ctx context.Context, req chain.ContractCallRequest, key string) (chain.TxReceipt, error) {
	f.txCounter++
	return chain.TxReceipt{Hash: "0xexec" + string(rune('0'+f.txCounter)), Status: "pending"}, nil
}
func (f *fakeAdapter) FetchABI(ctx context.Context, addr string) (string, error) {
	return "[]", nil
}

func newTestEngine(t *testing.T, adapters chain.Set) (*Engine, *registry.Registry) {
	t.Helper()
	reg, err := registry.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("opening registry: %v", err)
	}
	t.Cleanup(func() { reg.Close() })
	return New(adapters, reg, logging.Default()), reg
}

func TestReplay_MultiStepPlanWithInterpolationAndAssertion(t *testing.T) {
	fake := &fakeAdapter{name: "ethereum", callResults: map[string]interface{}{
		"totalSupply": big.NewInt(1000000),
	}}
	engine, _ := newTestEngine(t, chain.Set{"ethereum": fake})

	plan := Plan{
		Description: "read supply then assert it's positive",
		Operations: []Operation{
			{ID: "read_supply", Type: OpCallContract, Chain: "ethereum", Contract: "0xToken", Method: "totalSupply", ABI: "[]"},
			{ID: "check_supply", Type: OpAssert, Expression: "${read_supply.output} > 0"},
			{ID: "pause", Type: OpWait, Seconds: 0},
		},
	}

	receipt, err := engine.Run(context.Background(), plan, Options{DryRun: true})
	if err != nil {
		t.Fatalf("expected plan to succeed, got error: %v", err)
	}
	if !receipt.Success {
		t.Fatalf("expected receipt.Success = true")
	}
	if len(receipt.Steps) != 3 {
		t.Fatalf("expected 3 step results, got %d", len(receipt.Steps))
	}
	if receipt.Steps[1].Error != "" {
		t.Fatalf("assertion step should have passed, got error: %s", receipt.Steps[1].Error)
	}
}

func TestReplay_FailingAssertionStopsExecution(t *testing.T) {
	fake := &fakeAdapter{name: "ethereum", callResults: map[string]interface{}{
		"totalSupply": big.NewInt(0),
	}}
	engine, _ := newTestEngine(t, chain.Set{"ethereum": fake})

	plan := Plan{
		Description: "assertion should fail and halt the plan",
		Operations: []Operation{
			{ID: "read_supply", Type: OpCallContract, Chain: "ethereum", Contract: "0xToken", Method: "totalSupply", ABI: "[]"},
			{ID: "check_supply", Type: OpAssert, Expression: "${read_supply.output} > 0"},
			{ID: "never_runs", Type: OpWait, Seconds: 0},
		},
	}

	receipt, err := engine.Run(context.Background(), plan, Options{DryRun: true})
	if err == nil {
		t.Fatalf("expected plan to fail due to false assertion")
	}
	if receipt.Success {
		t.Fatalf("expected receipt.Success = false")
	}
	if len(receipt.Steps) != 2 {
		t.Fatalf("expected execution to stop after the failing assertion, got %d steps", len(receipt.Steps))
	}
}

func TestReplay_DryRunDoesNotBroadcast(t *testing.T) {
	fake := &fakeAdapter{name: "ethereum"}
	engine, _ := newTestEngine(t, chain.Set{"ethereum": fake})

	plan := Plan{
		Operations: []Operation{
			{ID: "send", Type: OpSendTransaction, Chain: "ethereum", From: "0xfrom", To: "0xto", Amount: "100"},
		},
	}

	receipt, err := engine.Run(context.Background(), plan, Options{DryRun: true})
	if err != nil {
		t.Fatalf("dry run should not error: %v", err)
	}
	if fake.txCounter != 0 {
		t.Fatalf("expected dry-run to avoid calling SendTransaction, but txCounter=%d", fake.txCounter)
	}
	if receipt.Steps[0].TxHash != "" {
		t.Fatalf("dry-run step should not have a tx hash")
	}
}

func TestReplay_VariableInterpolationChainsStepOutputs(t *testing.T) {
	fake := &fakeAdapter{name: "ethereum"}
	engine, _ := newTestEngine(t, chain.Set{"ethereum": fake})

	resolver := func(chainName, from string) (string, error) { return "fakekey", nil }

	plan := Plan{
		Operations: []Operation{
			{ID: "tx1", Type: OpSendTransaction, Chain: "ethereum", From: "0xfrom", To: "0xto", Amount: "1"},
			// References tx1's tx_hash; we just assert it resolved to a
			// non-empty, non-literal value (i.e. interpolation happened).
			{ID: "check", Type: OpAssert, Expression: "${tx1.tx_hash} != "},
		},
	}

	receipt, err := engine.Run(context.Background(), plan, Options{PrivateKeyResolver: resolver})
	if err != nil {
		t.Fatalf("expected plan to succeed: %v (steps: %+v)", err, receipt.Steps)
	}
	if receipt.Steps[0].TxHash == "" {
		t.Fatalf("expected tx1 to produce a tx hash")
	}
}

func TestReplay_RecordsPlanInRegistry(t *testing.T) {
	fake := &fakeAdapter{name: "ethereum"}
	engine, reg := newTestEngine(t, chain.Set{"ethereum": fake})

	plan := Plan{Description: "trivial", Operations: []Operation{
		{ID: "w", Type: OpWait, Seconds: 0},
	}}

	receipt, err := engine.Run(context.Background(), plan, Options{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	stored, err := reg.GetPlan(receipt.PlanID)
	if err != nil {
		t.Fatalf("fetching stored plan: %v", err)
	}
	if stored == nil {
		t.Fatalf("expected plan to be recorded in registry")
	}
	if stored.Status != registry.PlanStatusCompleted {
		t.Fatalf("expected plan status 'completed', got %s", stored.Status)
	}
}
