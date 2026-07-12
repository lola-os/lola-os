// Package registry provides LOLA OS's persistent SQLite store
// (~/.lola/lola.db by default). It is the single source of truth for:
//
//   - transactions: every broadcast tx, its status, and metadata
//   - nonces: the last-known nonce per (chain, address), used by the
//     nonce manager for crash recovery
//   - execution_plans: replay engine plan executions and their results
//
// It uses modernc.org/sqlite, a pure-Go SQLite driver, so lola-core has
// no cgo dependency and stays a single static binary.
package registry

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

// TxStatus enumerates transaction lifecycle states.
type TxStatus string

const (
	TxStatusPending   TxStatus = "pending"
	TxStatusConfirmed TxStatus = "confirmed"
	TxStatusFailed    TxStatus = "failed"
	TxStatusDropped   TxStatus = "dropped"
)

// Transaction is a single recorded on-chain operation.
type Transaction struct {
	Hash      string
	Chain     string
	From      string
	To        string
	Value     string // decimal string, to avoid float precision loss
	Gas       uint64
	GasPrice  string
	Status    TxStatus
	Timestamp time.Time
	PlanID    string // empty if not part of a replay plan
	Method    string // contract method name, if applicable
	Error     string // populated if Status == failed
}

// PlanStatus enumerates execution_plans lifecycle states.
type PlanStatus string

const (
	PlanStatusRunning   PlanStatus = "running"
	PlanStatusCompleted PlanStatus = "completed"
	PlanStatusFailed    PlanStatus = "failed"
)

// ExecutionPlan is a recorded `lola replay` run.
type ExecutionPlan struct {
	ID          string
	Description string
	Status      PlanStatus
	Result      string // JSON-encoded receipt
	StartedAt   time.Time
	FinishedAt  *time.Time
}

// Registry wraps a *sql.DB with LOLA-specific query helpers.
type Registry struct {
	db *sql.DB
}

// Open opens (creating if necessary) the SQLite database at path and
// ensures the schema is migrated to the latest version.
func Open(path string) (*Registry, error) {
	// _pragma busy_timeout reduces "database is locked" errors when the
	// nonce manager and registry are both writing concurrently.
	dsn := fmt.Sprintf("file:%s?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)", path)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("registry: opening %s: %w", path, err)
	}
	db.SetMaxOpenConns(1) // SQLite + WAL: single writer is simplest & correct
	r := &Registry{db: db}
	if err := r.migrate(); err != nil {
		return nil, err
	}
	return r, nil
}

// Close closes the underlying database handle.
func (r *Registry) Close() error {
	return r.db.Close()
}

// DB exposes the underlying *sql.DB for callers (e.g. nonce manager) that
// need transactional access spanning multiple tables.
func (r *Registry) DB() *sql.DB {
	return r.db
}

func (r *Registry) migrate() error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS transactions (
			hash       TEXT PRIMARY KEY,
			chain      TEXT NOT NULL,
			from_addr  TEXT NOT NULL,
			to_addr    TEXT NOT NULL,
			value      TEXT NOT NULL DEFAULT '0',
			gas        INTEGER NOT NULL DEFAULT 0,
			gas_price  TEXT NOT NULL DEFAULT '0',
			status     TEXT NOT NULL,
			timestamp  DATETIME NOT NULL,
			plan_id    TEXT NOT NULL DEFAULT '',
			method     TEXT NOT NULL DEFAULT '',
			error      TEXT NOT NULL DEFAULT ''
		)`,
		`CREATE INDEX IF NOT EXISTS idx_transactions_timestamp ON transactions(timestamp DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_transactions_plan_id ON transactions(plan_id)`,

		`CREATE TABLE IF NOT EXISTS nonces (
			chain   TEXT NOT NULL,
			address TEXT NOT NULL,
			nonce   INTEGER NOT NULL,
			updated_at DATETIME NOT NULL,
			PRIMARY KEY (chain, address)
		)`,

		`CREATE TABLE IF NOT EXISTS execution_plans (
			id          TEXT PRIMARY KEY,
			description TEXT NOT NULL DEFAULT '',
			status      TEXT NOT NULL,
			result      TEXT NOT NULL DEFAULT '',
			started_at  DATETIME NOT NULL,
			finished_at DATETIME
		)`,

		`CREATE TABLE IF NOT EXISTS idempotency_keys (
			key        TEXT PRIMARY KEY,
			result     TEXT NOT NULL,
			created_at DATETIME NOT NULL
		)`,

		`CREATE TABLE IF NOT EXISTS metrics_counters (
			name  TEXT PRIMARY KEY,
			value INTEGER NOT NULL DEFAULT 0
		)`,
	}
	for _, s := range stmts {
		if _, err := r.db.Exec(s); err != nil {
			return fmt.Errorf("registry: migration failed: %w\nstatement: %s", err, s)
		}
	}
	return nil
}

// --- Transactions ---------------------------------------------------------

// RecordTransaction inserts or updates (upserts) a transaction by hash.
func (r *Registry) RecordTransaction(tx Transaction) error {
	_, err := r.db.Exec(`
		INSERT INTO transactions (hash, chain, from_addr, to_addr, value, gas, gas_price, status, timestamp, plan_id, method, error)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(hash) DO UPDATE SET
			status = excluded.status,
			error = excluded.error,
			gas = excluded.gas,
			gas_price = excluded.gas_price
	`, tx.Hash, tx.Chain, tx.From, tx.To, tx.Value, tx.Gas, tx.GasPrice, string(tx.Status), tx.Timestamp, tx.PlanID, tx.Method, tx.Error)
	if err != nil {
		return fmt.Errorf("registry: recording transaction %s: %w", tx.Hash, err)
	}
	return nil
}

// GetTransaction fetches a transaction by hash.
func (r *Registry) GetTransaction(hash string) (*Transaction, error) {
	row := r.db.QueryRow(`
		SELECT hash, chain, from_addr, to_addr, value, gas, gas_price, status, timestamp, plan_id, method, error
		FROM transactions WHERE hash = ?`, hash)
	var t Transaction
	var status string
	if err := row.Scan(&t.Hash, &t.Chain, &t.From, &t.To, &t.Value, &t.Gas, &t.GasPrice, &status, &t.Timestamp, &t.PlanID, &t.Method, &t.Error); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("registry: fetching transaction %s: %w", hash, err)
	}
	t.Status = TxStatus(status)
	return &t, nil
}

// ListTransactions returns the most recent transactions, newest first,
// up to limit rows (0 means a sensible default of 50).
func (r *Registry) ListTransactions(limit int) ([]Transaction, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.db.Query(`
		SELECT hash, chain, from_addr, to_addr, value, gas, gas_price, status, timestamp, plan_id, method, error
		FROM transactions ORDER BY timestamp DESC LIMIT ?`, limit)
	if err != nil {
		return nil, fmt.Errorf("registry: listing transactions: %w", err)
	}
	defer rows.Close()

	var out []Transaction
	for rows.Next() {
		var t Transaction
		var status string
		if err := rows.Scan(&t.Hash, &t.Chain, &t.From, &t.To, &t.Value, &t.Gas, &t.GasPrice, &status, &t.Timestamp, &t.PlanID, &t.Method, &t.Error); err != nil {
			return nil, fmt.Errorf("registry: scanning transaction row: %w", err)
		}
		t.Status = TxStatus(status)
		out = append(out, t)
	}
	return out, rows.Err()
}

// ClearTransactions truncates the transactions table.
func (r *Registry) ClearTransactions() error {
	_, err := r.db.Exec(`DELETE FROM transactions`)
	return err
}

// --- Nonces ----------------------------------------------------------------

// GetNonce returns the last persisted nonce for (chain, address), and
// whether an entry existed.
func (r *Registry) GetNonce(chain, address string) (uint64, bool, error) {
	row := r.db.QueryRow(`SELECT nonce FROM nonces WHERE chain = ? AND address = ?`, chain, address)
	var n uint64
	if err := row.Scan(&n); err != nil {
		if err == sql.ErrNoRows {
			return 0, false, nil
		}
		return 0, false, fmt.Errorf("registry: fetching nonce: %w", err)
	}
	return n, true, nil
}

// SetNonce persists the nonce for (chain, address).
func (r *Registry) SetNonce(chain, address string, nonce uint64) error {
	_, err := r.db.Exec(`
		INSERT INTO nonces (chain, address, nonce, updated_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(chain, address) DO UPDATE SET nonce = excluded.nonce, updated_at = excluded.updated_at
	`, chain, address, nonce, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("registry: setting nonce: %w", err)
	}
	return nil
}

// DeleteNonce removes any persisted nonce for (chain, address), so the next
// allocation re-bootstraps from on-chain state. Used when releasing a
// reservation of nonce 0, where there is no lower value to roll back to.
func (r *Registry) DeleteNonce(chain, address string) error {
	if _, err := r.db.Exec(`DELETE FROM nonces WHERE chain = ? AND address = ?`, chain, address); err != nil {
		return fmt.Errorf("registry: deleting nonce: %w", err)
	}
	return nil
}

// --- Execution plans ---------------------------------------------------------

// CreatePlan inserts a new execution_plans row in "running" status.
func (r *Registry) CreatePlan(id, description string) error {
	_, err := r.db.Exec(`
		INSERT INTO execution_plans (id, description, status, result, started_at)
		VALUES (?, ?, ?, '', ?)
	`, id, description, string(PlanStatusRunning), time.Now().UTC())
	if err != nil {
		return fmt.Errorf("registry: creating plan %s: %w", id, err)
	}
	return nil
}

// FinishPlan marks a plan as completed or failed and stores its JSON result.
func (r *Registry) FinishPlan(id string, status PlanStatus, result interface{}) error {
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("registry: marshaling plan result: %w", err)
	}
	_, err = r.db.Exec(`
		UPDATE execution_plans SET status = ?, result = ?, finished_at = ? WHERE id = ?
	`, string(status), string(resultJSON), time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("registry: finishing plan %s: %w", id, err)
	}
	return nil
}

// GetPlan fetches an execution plan by ID.
func (r *Registry) GetPlan(id string) (*ExecutionPlan, error) {
	row := r.db.QueryRow(`
		SELECT id, description, status, result, started_at, finished_at
		FROM execution_plans WHERE id = ?`, id)
	var p ExecutionPlan
	var status string
	var finishedAt sql.NullTime
	if err := row.Scan(&p.ID, &p.Description, &status, &p.Result, &p.StartedAt, &finishedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("registry: fetching plan %s: %w", id, err)
	}
	p.Status = PlanStatus(status)
	if finishedAt.Valid {
		p.FinishedAt = &finishedAt.Time
	}
	return &p, nil
}

// --- Idempotency -------------------------------------------------------------

// GetIdempotentResult returns a cached result for key if present and not
// expired (ttl), else ("", false, nil).
func (r *Registry) GetIdempotentResult(key string, ttl time.Duration) (string, bool, error) {
	row := r.db.QueryRow(`SELECT result, created_at FROM idempotency_keys WHERE key = ?`, key)
	var result string
	var createdAt time.Time
	if err := row.Scan(&result, &createdAt); err != nil {
		if err == sql.ErrNoRows {
			return "", false, nil
		}
		return "", false, fmt.Errorf("registry: fetching idempotency key: %w", err)
	}
	if time.Since(createdAt) > ttl {
		return "", false, nil
	}
	return result, true, nil
}

// SetIdempotentResult stores result under key for later replay.
func (r *Registry) SetIdempotentResult(key, result string) error {
	_, err := r.db.Exec(`
		INSERT INTO idempotency_keys (key, result, created_at) VALUES (?, ?, ?)
		ON CONFLICT(key) DO UPDATE SET result = excluded.result, created_at = excluded.created_at
	`, key, result, time.Now().UTC())
	return err
}

// PruneIdempotencyKeys deletes entries older than ttl. Intended to be
// called periodically (e.g. by the budget/metrics background loop).
func (r *Registry) PruneIdempotencyKeys(ttl time.Duration) error {
	cutoff := time.Now().UTC().Add(-ttl)
	_, err := r.db.Exec(`DELETE FROM idempotency_keys WHERE created_at < ?`, cutoff)
	return err
}

// --- Metrics counters ---------------------------------------------------------

// IncrCounter atomically increments a named counter by delta and returns
// its new value.
func (r *Registry) IncrCounter(name string, delta int64) (int64, error) {
	_, err := r.db.Exec(`
		INSERT INTO metrics_counters (name, value) VALUES (?, ?)
		ON CONFLICT(name) DO UPDATE SET value = value + excluded.value
	`, name, delta)
	if err != nil {
		return 0, fmt.Errorf("registry: incrementing counter %s: %w", name, err)
	}
	row := r.db.QueryRow(`SELECT value FROM metrics_counters WHERE name = ?`, name)
	var v int64
	if err := row.Scan(&v); err != nil {
		return 0, fmt.Errorf("registry: reading counter %s: %w", name, err)
	}
	return v, nil
}

// AllCounters returns a snapshot of every metrics counter.
func (r *Registry) AllCounters() (map[string]int64, error) {
	rows, err := r.db.Query(`SELECT name, value FROM metrics_counters`)
	if err != nil {
		return nil, fmt.Errorf("registry: listing counters: %w", err)
	}
	defer rows.Close()
	out := map[string]int64{}
	for rows.Next() {
		var name string
		var v int64
		if err := rows.Scan(&name, &v); err != nil {
			return nil, err
		}
		out[name] = v
	}
	return out, rows.Err()
}

// ClearAll truncates every LOLA-managed table. Used by `lola registry clear`.
func (r *Registry) ClearAll() error {
	stmts := []string{
		`DELETE FROM transactions`,
		`DELETE FROM nonces`,
		`DELETE FROM execution_plans`,
		`DELETE FROM idempotency_keys`,
		`DELETE FROM metrics_counters`,
	}
	for _, s := range stmts {
		if _, err := r.db.Exec(s); err != nil {
			return fmt.Errorf("registry: clearing tables: %w", err)
		}
	}
	return nil
}
