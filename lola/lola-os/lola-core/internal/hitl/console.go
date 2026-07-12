// Package hitl implements LOLA OS's human-in-the-loop approval flows: a
// rich console prompt (this file) and a localhost-only WebSocket server
// (see websocket.go) that broadcasts the same approval requests for
// custom UIs to consume.
package hitl

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
)

// Decision is the outcome of a human approval request.
type Decision string

const (
	DecisionApprove Decision = "approve"
	DecisionDeny    Decision = "deny"
	DecisionSkip    Decision = "skip"
	DecisionTimeout Decision = "timeout"
)

// Request describes a write operation awaiting human approval.
type Request struct {
	ID          string
	Chain       string
	Method      string
	Contract    string
	From        string
	To          string
	ValueHuman  string // human-readable amount, e.g. "0.5 ETH"
	GasEstimate string
	Description string
}

// Approver is anything capable of resolving a Request to a Decision.
// ConsoleApprover and WebSocketServer both implement this.
type Approver interface {
	RequestApproval(ctx context.Context, req Request, timeout time.Duration) (Decision, error)
}

// ConsoleApprover prompts the operator via the terminal, per the LOLA
// branding guide: rich boxes, icons, and clear colour-coded states.
type ConsoleApprover struct {
	in  *bufio.Reader
	out *os.File
}

// NewConsoleApprover constructs a ConsoleApprover reading from stdin and
// writing prompts to stderr (so stdout stays clean for JSON-RPC output).
func NewConsoleApprover() *ConsoleApprover {
	return &ConsoleApprover{in: bufio.NewReader(os.Stdin), out: os.Stderr}
}

// RequestApproval renders req and blocks (up to timeout) for an operator
// response of [a]pprove, [d]eny, or [s]kip.
func (c *ConsoleApprover) RequestApproval(ctx context.Context, req Request, timeout time.Duration) (Decision, error) {
	c.render(req)

	answerCh := make(chan string, 1)
	go func() {
		line, _ := c.in.ReadString('\n')
		answerCh <- strings.ToLower(strings.TrimSpace(line))
	}()

	select {
	case <-ctx.Done():
		return DecisionTimeout, ctx.Err()
	case <-time.After(timeout):
		fmt.Fprintln(c.out, color.YellowString("\n⚠️  Approval timed out — treating as deny for safety."))
		return DecisionTimeout, nil
	case answer := <-answerCh:
		switch answer {
		case "a", "approve", "y", "yes":
			fmt.Fprintln(c.out, color.GreenString("✅ Approved."))
			return DecisionApprove, nil
		case "s", "skip":
			fmt.Fprintln(c.out, color.HiBlackString("🔄 Skipped."))
			return DecisionSkip, nil
		default:
			fmt.Fprintln(c.out, color.RedString("❌ Denied."))
			return DecisionDeny, nil
		}
	}
}

func (c *ConsoleApprover) render(req Request) {
	bold := color.New(color.Bold).SprintFunc()
	border := color.HiBlackString(strings.Repeat("─", 56))

	fmt.Fprintln(c.out)
	fmt.Fprintln(c.out, border)
	fmt.Fprintln(c.out, bold("⚠️  APPROVAL REQUIRED"))
	fmt.Fprintln(c.out, border)
	if req.Description != "" {
		fmt.Fprintf(c.out, "  %s\n", req.Description)
	}
	fmt.Fprintf(c.out, "  Chain:     %s\n", req.Chain)
	if req.Method != "" {
		fmt.Fprintf(c.out, "  Method:    %s\n", req.Method)
	}
	if req.Contract != "" {
		fmt.Fprintf(c.out, "  Contract:  %s\n", req.Contract)
	}
	if req.From != "" {
		fmt.Fprintf(c.out, "  From:      %s\n", req.From)
	}
	if req.To != "" {
		fmt.Fprintf(c.out, "  To:        %s\n", req.To)
	}
	if req.ValueHuman != "" {
		fmt.Fprintf(c.out, "  Value:     %s\n", req.ValueHuman)
	}
	if req.GasEstimate != "" {
		fmt.Fprintf(c.out, "  Est. gas:  %s\n", req.GasEstimate)
	}
	fmt.Fprintln(c.out, border)
	fmt.Fprint(c.out, bold("  [a]pprove / [d]eny / [s]kip: "))
}
