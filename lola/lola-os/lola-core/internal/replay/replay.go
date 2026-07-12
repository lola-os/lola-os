// Package replay implements LOLA OS's structured execution replay engine
// (`lola replay <plan.json>`). It ingests a JSON plan describing a sequence
// of operations, executes them in order against a chain.ChainAdapter,
// supports ${variable} interpolation so later steps can reference earlier
// outputs, and records the full run in the SQLite registry.
//
// Plan JSON schema (see lola-ui docs "Replay" page for the full reference):
//
//	{
//	  "description": "string",
//	  "operations": [
//	    {
//	      "id": "step1",
//	      "type": "call_contract" | "send_transaction" | "transfer_token" |
//	              "execute_contract" | "swap_tokens" | "assert" | "wait",
//	      "chain": "ethereum",
//	      ... type-specific fields ...
//	    }
//	  ]
//	}
package replay

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/lola-os/lola-core/internal/chain"
	"github.com/lola-os/lola-core/internal/logging"
	"github.com/lola-os/lola-core/internal/registry"
)

// OpType enumerates supported operation kinds.
type OpType string

const (
	OpCallContract    OpType = "call_contract"
	OpSendTransaction OpType = "send_transaction"
	OpTransferToken   OpType = "transfer_token"
	OpExecuteContract OpType = "execute_contract"
	OpSwapTokens      OpType = "swap_tokens"
	OpAssert          OpType = "assert"
	OpWait            OpType = "wait"
)

// Operation is a single step in a plan. Fields are a superset across all
// operation types; which ones are read depends on Type.
type Operation struct {
	ID       string        `json:"id"`
	Type     OpType        `json:"type"`
	Chain    string        `json:"chain"`
	Contract string        `json:"contract,omitempty"`
	Method   string        `json:"method,omitempty"`
	Args     []interface{} `json:"args,omitempty"`
	ABI      string        `json:"abi,omitempty"`
	From     string        `json:"from,omitempty"`
	To       string        `json:"to,omitempty"`
	Token    string        `json:"token,omitempty"`
	Amount   string        `json:"amount,omitempty"` // decimal string, may contain ${vars}
	// assert
	Expression string      `json:"expression,omitempty"` // e.g. "${step1.output} > 0"
	Expected   interface{} `json:"expected,omitempty"`
	// wait
	Seconds int `json:"seconds,omitempty"`
	// idempotency
	IdempotencyKey string `json:"idempotency_key,omitempty"`
}

// Plan is the top-level structure of a plan.json file.
type Plan struct {
	Description string      `json:"description"`
	Operations  []Operation `json:"operations"`
}

// StepResult captures the outcome of a single executed operation.
type StepResult struct {
	ID      string      `json:"id"`
	Type    OpType      `json:"type"`
	Output  interface{} `json:"output,omitempty"`
	TxHash  string      `json:"tx_hash,omitempty"`
	Error   string      `json:"error,omitempty"`
	Skipped bool        `json:"skipped,omitempty"`
}

// Receipt is the full result of a plan execution, suitable for
// `--output receipt.json`.
type Receipt struct {
	PlanID      string       `json:"plan_id"`
	Description string       `json:"description"`
	DryRun      bool         `json:"dry_run"`
	Steps       []StepResult `json:"steps"`
	Success     bool         `json:"success"`
	StartedAt   time.Time    `json:"started_at"`
	FinishedAt  time.Time    `json:"finished_at"`
}

// Options configures a single replay run.
type Options struct {
	ForkURL            string // overrides the RPC URL for every chain reference (testing against a fork)
	DryRun             bool
	PrivateKeyResolver func(chainName, fromAddress string) (string, error)
}

// Engine executes plans against a set of configured chain adapters.
type Engine struct {
	chains chain.Set
	reg    *registry.Registry
	logger *logging.Logger
}

// New constructs a replay Engine.
func New(chains chain.Set, reg *registry.Registry, logger *logging.Logger) *Engine {
	if logger == nil {
		logger = logging.Default()
	}
	return &Engine{chains: chains, reg: reg, logger: logger}
}

// LoadPlan reads and parses a plan.json file from disk.
func LoadPlan(path string) (Plan, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Plan{}, fmt.Errorf("replay: reading plan file %s: %w", path, err)
	}
	var p Plan
	if err := json.Unmarshal(data, &p); err != nil {
		return Plan{}, fmt.Errorf("replay: parsing plan file %s: %w", path, err)
	}
	if len(p.Operations) == 0 {
		return Plan{}, fmt.Errorf("replay: plan %s has no operations", path)
	}
	for i, op := range p.Operations {
		if op.ID == "" {
			return Plan{}, fmt.Errorf("replay: operation at index %d is missing an \"id\"", i)
		}
		if op.Type == "" {
			return Plan{}, fmt.Errorf("replay: operation %q is missing a \"type\"", op.ID)
		}
	}
	return p, nil
}

var varPattern = regexp.MustCompile(`\$\{([^}]+)\}`)

// interpolate replaces ${stepID.field} / ${stepID} references in s using
// previously recorded step outputs.
func interpolate(s string, results map[string]StepResult) string {
	return varPattern.ReplaceAllStringFunc(s, func(match string) string {
		key := varPattern.FindStringSubmatch(match)[1]
		parts := strings.SplitN(key, ".", 2)
		stepID := parts[0]
		res, ok := results[stepID]
		if !ok {
			return match // leave unresolved references intact for visibility
		}
		if len(parts) == 1 {
			return fmt.Sprintf("%v", res.Output)
		}
		field := parts[1]
		switch field {
		case "tx_hash":
			return res.TxHash
		case "output":
			return fmt.Sprintf("%v", res.Output)
		default:
			if m, ok := res.Output.(map[string]interface{}); ok {
				if v, ok := m[field]; ok {
					return fmt.Sprintf("%v", v)
				}
			}
			return match
		}
	})
}

func interpolateArgs(args []interface{}, results map[string]StepResult) []interface{} {
	out := make([]interface{}, len(args))
	for i, a := range args {
		if s, ok := a.(string); ok {
			out[i] = interpolate(s, results)
		} else {
			out[i] = a
		}
	}
	return out
}

// PlanIDForTx is a tiny helper so Run can decide whether a step type
// produces a transaction worth recording, without a long type switch
// duplicated in two places. Despite the name, it returns the *operation's*
// own ID (used as a recording marker), or "" if this op type never
// produces a transaction.
func (op Operation) PlanIDForTx() string {
	switch op.Type {
	case OpSendTransaction, OpTransferToken, OpExecuteContract, OpSwapTokens:
		return op.ID
	default:
		return ""
	}
}

// Run executes plan and returns a Receipt. The plan's full execution is
// recorded in the registry under a new plan ID.
func (e *Engine) Run(ctx context.Context, plan Plan, opts Options) (Receipt, error) {
	planID := uuid.NewString()
	if err := e.reg.CreatePlan(planID, plan.Description); err != nil {
		return Receipt{}, fmt.Errorf("replay: recording plan start: %w", err)
	}

	receipt := Receipt{
		PlanID:      planID,
		Description: plan.Description,
		DryRun:      opts.DryRun,
		StartedAt:   time.Now().UTC(),
	}

	results := map[string]StepResult{}
	success := true

	for _, op := range plan.Operations {
		e.logger.Info("executing replay step", map[string]interface{}{"id": op.ID, "type": string(op.Type)})
		res, err := e.runStep(ctx, op, results, opts)
		if err != nil {
			res.Error = err.Error()
			success = false
			e.logger.Error("replay step failed", map[string]interface{}{"id": op.ID, "error": err.Error()})
			results[op.ID] = res
			receipt.Steps = append(receipt.Steps, res)
			break // stop on first failure; partial receipts are still useful
		}
		results[op.ID] = res
		receipt.Steps = append(receipt.Steps, res)

		if op.PlanIDForTx() != "" && res.TxHash != "" {
			_ = e.reg.RecordTransaction(registry.Transaction{
				Hash: res.TxHash, Chain: op.Chain, Status: registry.TxStatusPending,
				Timestamp: time.Now().UTC(), PlanID: planID, Method: op.Method,
			})
		}
	}

	receipt.Success = success
	receipt.FinishedAt = time.Now().UTC()

	status := registry.PlanStatusCompleted
	if !success {
		status = registry.PlanStatusFailed
	}
	if err := e.reg.FinishPlan(planID, status, receipt); err != nil {
		e.logger.Error("failed to record plan completion", map[string]interface{}{"error": err.Error()})
	}

	if !success {
		return receipt, fmt.Errorf("replay: plan execution failed (see receipt for details)")
	}
	return receipt, nil
}

func (e *Engine) runStep(ctx context.Context, op Operation, results map[string]StepResult, opts Options) (StepResult, error) {
	res := StepResult{ID: op.ID, Type: op.Type}

	var adapter chain.ChainAdapter
	var err error
	if op.Type != OpAssert && op.Type != OpWait {
		adapter, err = e.resolveAdapter(op.Chain, opts.ForkURL)
		if err != nil {
			return res, err
		}
	}

	switch op.Type {
	case OpWait:
		secs := op.Seconds
		if secs <= 0 {
			secs = 1
		}
		select {
		case <-ctx.Done():
			return res, ctx.Err()
		case <-time.After(time.Duration(secs) * time.Second):
		}
		res.Output = fmt.Sprintf("waited %ds", secs)
		return res, nil

	case OpAssert:
		expr := interpolate(op.Expression, results)
		ok, evalErr := evaluateAssertion(expr, op.Expected)
		if evalErr != nil {
			return res, fmt.Errorf("assert %q: %w", op.ID, evalErr)
		}
		if !ok {
			return res, fmt.Errorf("assert %q failed: expression %q did not hold", op.ID, expr)
		}
		res.Output = true
		return res, nil

	case OpCallContract:
		method := interpolate(op.Method, results)
		args := interpolateArgs(op.Args, results)
		req := chain.ContractCallRequest{
			ContractAddress: interpolate(op.Contract, results),
			Method:          method,
			Args:            args,
			ABI:             op.ABI,
		}
		out, err := adapter.CallContract(ctx, req)
		if err != nil {
			return res, err
		}
		res.Output = out
		return res, nil

	case OpSendTransaction:
		if opts.DryRun {
			res.Output = "dry-run: transaction not broadcast"
			return res, nil
		}
		privKey, err := opts.PrivateKeyResolver(op.Chain, interpolate(op.From, results))
		if err != nil {
			return res, fmt.Errorf("resolving signing key: %w", err)
		}
		valueWei, err := parseAmount(interpolate(op.Amount, results))
		if err != nil {
			return res, err
		}
		txReq := chain.TxRequest{
			From:           interpolate(op.From, results),
			To:             interpolate(op.To, results),
			ValueWei:       valueWei,
			IdempotencyKey: op.IdempotencyKey,
		}
		receipt, err := adapter.SendTransaction(ctx, txReq, privKey)
		if err != nil {
			return res, err
		}
		res.TxHash = receipt.Hash
		res.Output = receipt
		return res, nil

	case OpExecuteContract, OpSwapTokens:
		if opts.DryRun {
			res.Output = "dry-run: contract execution not broadcast"
			return res, nil
		}
		from := interpolate(op.From, results)
		privKey, err := opts.PrivateKeyResolver(op.Chain, from)
		if err != nil {
			return res, fmt.Errorf("resolving signing key: %w", err)
		}
		valueWei, _ := parseAmount(interpolate(op.Amount, results))
		req := chain.ContractCallRequest{
			ContractAddress: interpolate(op.Contract, results),
			Method:          interpolate(op.Method, results),
			Args:            interpolateArgs(op.Args, results),
			ABI:             op.ABI,
			From:            from,
			ValueWei:        valueWei,
		}
		receipt, err := adapter.ExecuteContract(ctx, req, privKey)
		if err != nil {
			return res, err
		}
		res.TxHash = receipt.Hash
		res.Output = receipt
		return res, nil

	case OpTransferToken:
		if opts.DryRun {
			res.Output = "dry-run: token transfer not broadcast"
			return res, nil
		}
		from := interpolate(op.From, results)
		privKey, err := opts.PrivateKeyResolver(op.Chain, from)
		if err != nil {
			return res, fmt.Errorf("resolving signing key: %w", err)
		}
		amount, err := parseAmount(interpolate(op.Amount, results))
		if err != nil {
			return res, err
		}
		const erc20TransferABI = `[{"constant":false,"inputs":[{"name":"_to","type":"address"},{"name":"_value","type":"uint256"}],"name":"transfer","outputs":[{"name":"","type":"bool"}],"type":"function"}]`
		req := chain.ContractCallRequest{
			ContractAddress: interpolate(op.Token, results),
			Method:          "transfer",
			Args:            []interface{}{interpolate(op.To, results), amount},
			ABI:             erc20TransferABI,
			From:            from,
		}
		receipt, err := adapter.ExecuteContract(ctx, req, privKey)
		if err != nil {
			return res, err
		}
		res.TxHash = receipt.Hash
		res.Output = receipt
		return res, nil

	default:
		return res, fmt.Errorf("replay: unknown operation type %q", op.Type)
	}
}

func (e *Engine) resolveAdapter(chainName, forkURL string) (chain.ChainAdapter, error) {
	a, err := e.chains.Get(chainName)
	if err != nil {
		return nil, fmt.Errorf("replay: resolving chain %q: %w", chainName, err)
	}
	// True fork execution (pointing at a local anvil/hardhat fork) is
	// handled by the CLI layer constructing a fork-specific adapter set
	// before calling Run — see cmd/lola/replay.go — rather than mutating
	// shared adapters here.
	_ = forkURL
	return a, nil
}

func parseAmount(s string) (*big.Int, error) {
	if s == "" {
		return big.NewInt(0), nil
	}
	v, ok := new(big.Int).SetString(s, 10)
	if !ok {
		f, ok2 := new(big.Float).SetString(s)
		if !ok2 {
			return nil, fmt.Errorf("invalid amount %q", s)
		}
		v, _ = f.Int(nil)
	}
	return v, nil
}

// evaluateAssertion supports a minimal comparison grammar sufficient for
// plan assertions: "<value> == <expected>", "!=", ">", ">=", "<", "<=", or,
// with no operator, direct equality against the `expected` field.
func evaluateAssertion(expr string, expected interface{}) (bool, error) {
	ops := []string{">=", "<=", "==", "!=", ">", "<"}
	for _, op := range ops {
		if idx := strings.Index(expr, op); idx >= 0 {
			lhs := strings.TrimSpace(expr[:idx])
			rhs := strings.TrimSpace(expr[idx+len(op):])
			lf, lerr := parseFloatLoose(lhs)
			rf, rerr := parseFloatLoose(rhs)
			if lerr != nil || rerr != nil {
				switch op {
				case "==":
					return lhs == rhs, nil
				case "!=":
					return lhs != rhs, nil
				default:
					return false, fmt.Errorf("cannot numerically compare %q and %q", lhs, rhs)
				}
			}
			switch op {
			case ">=":
				return lf >= rf, nil
			case "<=":
				return lf <= rf, nil
			case "==":
				return lf == rf, nil
			case "!=":
				return lf != rf, nil
			case ">":
				return lf > rf, nil
			case "<":
				return lf < rf, nil
			}
		}
	}
	return fmt.Sprintf("%v", expected) == strings.TrimSpace(expr), nil
}

func parseFloatLoose(s string) (float64, error) {
	s = strings.TrimSpace(s)
	var f float64
	_, err := fmt.Sscanf(s, "%g", &f)
	if err != nil {
		return 0, err
	}
	return f, nil
}
