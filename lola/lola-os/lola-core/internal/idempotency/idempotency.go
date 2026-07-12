// Package idempotency provides a cache of write-operation results keyed by
// a caller-supplied idempotency key, so that retried requests (e.g. an AI
// agent that resends a tool call after a timeout) don't cause duplicate
// on-chain broadcasts. Results are persisted to SQLite with a 24h TTL.
package idempotency

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/lola-os/lola-core/internal/registry"
)

// DefaultTTL is the standard idempotency window per the blueprint.
const DefaultTTL = 24 * time.Hour

// Cache wraps a *registry.Registry to provide idempotent result storage.
type Cache struct {
	reg *registry.Registry
	ttl time.Duration
}

// New constructs a Cache with the default 24h TTL.
func New(reg *registry.Registry) *Cache {
	return &Cache{reg: reg, ttl: DefaultTTL}
}

// WithTTL returns a copy of the cache using a custom TTL (mainly for tests).
func (c *Cache) WithTTL(ttl time.Duration) *Cache {
	return &Cache{reg: c.reg, ttl: ttl}
}

// Lookup returns a previously stored result for key, unmarshaled into out,
// and true, if a non-expired entry exists. Otherwise it returns false.
func (c *Cache) Lookup(key string, out interface{}) (bool, error) {
	if key == "" {
		return false, nil // empty key = idempotency not requested
	}
	raw, found, err := c.reg.GetIdempotentResult(key, c.ttl)
	if err != nil {
		return false, fmt.Errorf("idempotency: lookup failed: %w", err)
	}
	if !found {
		return false, nil
	}
	if out != nil {
		if err := json.Unmarshal([]byte(raw), out); err != nil {
			return false, fmt.Errorf("idempotency: decoding cached result: %w", err)
		}
	}
	return true, nil
}

// Store saves result under key for later replay within the TTL window.
// A no-op if key is empty.
func (c *Cache) Store(key string, result interface{}) error {
	if key == "" {
		return nil
	}
	data, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("idempotency: encoding result: %w", err)
	}
	if err := c.reg.SetIdempotentResult(key, string(data)); err != nil {
		return fmt.Errorf("idempotency: storing result: %w", err)
	}
	return nil
}

// Prune deletes expired entries. Intended to be called periodically by a
// background loop (see internal/budget for the existing ticker pattern).
func (c *Cache) Prune() error {
	return c.reg.PruneIdempotencyKeys(c.ttl)
}
