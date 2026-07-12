package lola

import (
	"context"
	"testing"
)

func TestWithOverrides_RoundTrip(t *testing.T) {
	ov := &Overrides{Chain: "polygon"}
	ctx := WithOverrides(context.Background(), ov)

	got := OverridesFromContext(ctx)
	if got == nil || got.Chain != "polygon" {
		t.Fatalf("expected to retrieve overrides with Chain=polygon, got %+v", got)
	}
}

func TestOverridesFromContext_NilWhenAbsent(t *testing.T) {
	if got := OverridesFromContext(context.Background()); got != nil {
		t.Fatalf("expected nil overrides on a bare context, got %+v", got)
	}
}

func TestWithOverrides_NilIsNoOp(t *testing.T) {
	ctx := WithOverrides(context.Background(), nil)
	if got := OverridesFromContext(ctx); got != nil {
		t.Fatalf("expected WithOverrides(nil) to not attach anything, got %+v", got)
	}
}

func TestResolveOverrides_ExplicitWinsOverContext(t *testing.T) {
	ctxOv := &Overrides{Chain: "ethereum"}
	ctx := WithOverrides(context.Background(), ctxOv)

	explicit := &Overrides{Chain: "solana"}
	resolved := resolveOverrides(ctx, explicit)
	if resolved.Chain != "solana" {
		t.Fatalf("expected explicit override to win, got chain=%s", resolved.Chain)
	}
}

func TestResolveOverrides_FallsBackToContext(t *testing.T) {
	ctxOv := &Overrides{Chain: "ethereum"}
	ctx := WithOverrides(context.Background(), ctxOv)

	resolved := resolveOverrides(ctx, nil)
	if resolved == nil || resolved.Chain != "ethereum" {
		t.Fatalf("expected to fall back to context override, got %+v", resolved)
	}
}

func TestTool_AppliesOverridesToContextSeenByFn(t *testing.T) {
	var seenChain string
	fn := func(ctx context.Context) (interface{}, error) {
		if ov := OverridesFromContext(ctx); ov != nil {
			seenChain = ov.Chain
		}
		return "ok", nil
	}

	wrapped := Tool(ToolOptions{Name: "test_tool", Overrides: &Overrides{Chain: "polygon"}}, fn)
	result, err := wrapped(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "ok" {
		t.Fatalf("unexpected result: %v", result)
	}
	if seenChain != "polygon" {
		t.Fatalf("expected wrapped fn to observe overrides.Chain=polygon, got %q", seenChain)
	}
}

func TestTool_NoOverridesLeavesContextUntouched(t *testing.T) {
	fn := func(ctx context.Context) (interface{}, error) {
		if OverridesFromContext(ctx) != nil {
			t.Errorf("expected no overrides on context when ToolOptions.Overrides is nil")
		}
		return nil, nil
	}
	wrapped := Tool(ToolOptions{Name: "plain_tool"}, fn)
	if _, err := wrapped(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDescribe_CarriesMetadata(t *testing.T) {
	fn := func(ctx context.Context) (interface{}, error) { return 42, nil }
	d := Describe(ToolOptions{Name: "answer", Description: "the answer"}, fn)
	if d.Name != "answer" || d.Description != "the answer" {
		t.Fatalf("unexpected metadata: %+v", d)
	}
	result, err := d.Fn(context.Background())
	if err != nil || result != 42 {
		t.Fatalf("unexpected result=%v err=%v", result, err)
	}
}
