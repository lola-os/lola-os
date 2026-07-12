// Package nonce implements LOLA OS's nonce manager: an atomic, in-memory
// counter per (chain, address) pair, persisted to the SQLite registry so
// that a crash or restart doesn't cause nonce reuse or gaps.
//
// Concurrency model: a single process-wide mutex per (chain, address) key
// serializes nonce allocation for that pair. This is intentionally simple
// and correct rather than lock-free; nonce allocation is not a hot path
// (it happens once per write transaction), so the cost of a mutex is
// negligible compared to network/RPC latency.
package nonce

import (
	"context"
	"fmt"
	"sync"

	"github.com/lola-os/lola-core/internal/registry"
)

// ChainClient is the minimal interface the nonce manager needs from a
// chain adapter to bootstrap/reconcile its counter against on-chain truth.
type ChainClient interface {
	// PendingNonce returns the next nonce the chain expects for address,
	// accounting for pending (not yet mined) transactions.
	PendingNonce(ctx context.Context, address string) (uint64, error)
}

// Manager allocates monotonically increasing nonces per (chain, address).
type Manager struct {
	reg *registry.Registry

	mu     sync.Mutex
	locks  map[string]*sync.Mutex // one lock per "chain|address" key
	locksM sync.Mutex             // protects the locks map itself
}

// New constructs a Manager backed by reg.
func New(reg *registry.Registry) *Manager {
	return &Manager{
		reg:   reg,
		locks: map[string]*sync.Mutex{},
	}
}

func key(chain, address string) string {
	return chain + "|" + address
}

func (m *Manager) lockFor(chain, address string) *sync.Mutex {
	k := key(chain, address)
	m.locksM.Lock()
	defer m.locksM.Unlock()
	l, ok := m.locks[k]
	if !ok {
		l = &sync.Mutex{}
		m.locks[k] = l
	}
	return l
}

// Next returns the next nonce to use for (chain, address) and persists the
// reservation immediately, so a concurrent caller (or a crash-recovery
// pass) never reuses it.
//
// If no nonce has ever been recorded for this pair, it falls back to
// client.PendingNonce to bootstrap from on-chain state.
func (m *Manager) Next(ctx context.Context, chain, address string, client ChainClient) (uint64, error) {
	l := m.lockFor(chain, address)
	l.Lock()
	defer l.Unlock()

	current, exists, err := m.reg.GetNonce(chain, address)
	if err != nil {
		return 0, fmt.Errorf("nonce: reading persisted nonce: %w", err)
	}

	var next uint64
	if !exists {
		// Bootstrap from chain state.
		pending, err := client.PendingNonce(ctx, address)
		if err != nil {
			return 0, fmt.Errorf("nonce: bootstrapping from chain: %w", err)
		}
		next = pending
	} else {
		next = current + 1
		// Defensive reconciliation: if our persisted nonce has drifted
		// behind chain state (e.g. transactions sent from another tool),
		// prefer the higher of the two to avoid an AlreadyKnown error.
		if pending, err := client.PendingNonce(ctx, address); err == nil && pending > next {
			next = pending
		}
	}

	if err := m.reg.SetNonce(chain, address, next); err != nil {
		return 0, fmt.Errorf("nonce: persisting reservation: %w", err)
	}
	return next, nil
}

// Release rolls back a previously reserved nonce, e.g. when a transaction
// build failed *before* broadcast and the nonce was never actually
// consumed on-chain. This avoids leaving permanent gaps from purely local
// failures. It is a best-effort operation: if another nonce has already
// been issued on top of it, Release is a no-op.
func (m *Manager) Release(chain, address string, nonce uint64) error {
	l := m.lockFor(chain, address)
	l.Lock()
	defer l.Unlock()

	current, exists, err := m.reg.GetNonce(chain, address)
	if err != nil {
		return fmt.Errorf("nonce: reading persisted nonce: %w", err)
	}
	if !exists || current != nonce {
		// Someone already moved past it; releasing now would corrupt state.
		return nil
	}
	if nonce == 0 {
		// There is no nonce below 0 to roll back to. Clearing the entry
		// makes the next allocation re-bootstrap from chain state, which
		// re-issues 0 (rather than incorrectly skipping to 1).
		return m.reg.DeleteNonce(chain, address)
	}
	return m.reg.SetNonce(chain, address, nonce-1)
}

// Reconcile forcibly resets the persisted nonce for (chain, address) to
// match the chain's current pending nonce. Used by `lola doctor` to repair
// drift, and by tests.
func (m *Manager) Reconcile(ctx context.Context, chain, address string, client ChainClient) (uint64, error) {
	l := m.lockFor(chain, address)
	l.Lock()
	defer l.Unlock()

	pending, err := client.PendingNonce(ctx, address)
	if err != nil {
		return 0, fmt.Errorf("nonce: reconciling: %w", err)
	}
	// We store "last used" semantics (Next() returns current+1), so we
	// persist pending-1 such that the next allocation yields `pending`.
	var toStore uint64
	if pending > 0 {
		toStore = pending - 1
	}
	if err := m.reg.SetNonce(chain, address, toStore); err != nil {
		return 0, err
	}
	return pending, nil
}
