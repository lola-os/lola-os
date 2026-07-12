package nonce

import (
	"context"
	"path/filepath"
	"sync"
	"testing"

	"github.com/lola-os/lola-core/internal/registry"
)

// fakeChainClient implements ChainClient with a fixed PendingNonce, used
// to bootstrap the manager in tests without a real RPC connection.
type fakeChainClient struct {
	pending uint64
}

func (f *fakeChainClient) PendingNonce(ctx context.Context, address string) (uint64, error) {
	return f.pending, nil
}

func newTestRegistry(t *testing.T) *registry.Registry {
	t.Helper()
	dir := t.TempDir()
	reg, err := registry.Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("opening test registry: %v", err)
	}
	t.Cleanup(func() { reg.Close() })
	return reg
}

func TestManager_BootstrapsFromChainOnFirstUse(t *testing.T) {
	reg := newTestRegistry(t)
	mgr := New(reg)
	client := &fakeChainClient{pending: 42}

	n, err := mgr.Next(context.Background(), "ethereum", "0xabc", client)
	if err != nil {
		t.Fatalf("Next returned error: %v", err)
	}
	if n != 42 {
		t.Fatalf("expected bootstrap nonce 42, got %d", n)
	}
}

func TestManager_IncrementsMonotonically(t *testing.T) {
	reg := newTestRegistry(t)
	mgr := New(reg)
	client := &fakeChainClient{pending: 0}

	first, err := mgr.Next(context.Background(), "ethereum", "0xabc", client)
	if err != nil {
		t.Fatal(err)
	}
	second, err := mgr.Next(context.Background(), "ethereum", "0xabc", client)
	if err != nil {
		t.Fatal(err)
	}
	if second != first+1 {
		t.Fatalf("expected second nonce to be first+1, got first=%d second=%d", first, second)
	}
}

// TestManager_ConcurrentAllocationsAreUnique is the specifically-requested
// concurrency test: many goroutines requesting nonces simultaneously for
// the same (chain, address) must never receive duplicate values.
func TestManager_ConcurrentAllocationsAreUnique(t *testing.T) {
	reg := newTestRegistry(t)
	mgr := New(reg)
	client := &fakeChainClient{pending: 0}

	const n = 50
	results := make([]uint64, n)
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			nonce, err := mgr.Next(context.Background(), "ethereum", "0xconcurrent", client)
			if err != nil {
				t.Errorf("goroutine %d: Next failed: %v", idx, err)
				return
			}
			results[idx] = nonce
		}(i)
	}
	wg.Wait()

	seen := map[uint64]bool{}
	for _, r := range results {
		if seen[r] {
			t.Fatalf("duplicate nonce allocated: %d", r)
		}
		seen[r] = true
	}
	if len(seen) != n {
		t.Fatalf("expected %d unique nonces, got %d", n, len(seen))
	}
}

func TestManager_ReleaseRollsBackUnusedNonce(t *testing.T) {
	reg := newTestRegistry(t)
	mgr := New(reg)
	client := &fakeChainClient{pending: 0}

	n, err := mgr.Next(context.Background(), "ethereum", "0xabc", client)
	if err != nil {
		t.Fatal(err)
	}
	if err := mgr.Release("ethereum", "0xabc", n); err != nil {
		t.Fatalf("Release failed: %v", err)
	}

	again, err := mgr.Next(context.Background(), "ethereum", "0xabc", client)
	if err != nil {
		t.Fatal(err)
	}
	if again != n {
		t.Fatalf("expected Release to allow re-issuing nonce %d, got %d", n, again)
	}
}

func TestManager_ReleaseIsNoOpIfNonceAlreadyAdvanced(t *testing.T) {
	reg := newTestRegistry(t)
	mgr := New(reg)
	client := &fakeChainClient{pending: 0}

	first, _ := mgr.Next(context.Background(), "ethereum", "0xabc", client)
	second, _ := mgr.Next(context.Background(), "ethereum", "0xabc", client)

	// Releasing the older nonce after a newer one has been issued must
	// not corrupt state.
	if err := mgr.Release("ethereum", "0xabc", first); err != nil {
		t.Fatalf("Release returned error: %v", err)
	}
	third, err := mgr.Next(context.Background(), "ethereum", "0xabc", client)
	if err != nil {
		t.Fatal(err)
	}
	if third != second+1 {
		t.Fatalf("expected nonce sequence to continue from %d, got %d", second+1, third)
	}
}
