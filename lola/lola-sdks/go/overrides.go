package lola

import "context"

type overridesKey struct{}

// WithOverrides returns a new context carrying ov, so that any Client
// call made with that context picks up the override automatically — the
// Go-idiomatic equivalent of Python's `with lola.override(...):` block or
// TypeScript's `configOverrides` option.
//
//	ctx := lola.WithOverrides(context.Background(), &lola.Overrides{Chain: "polygon"})
//	balance, err := client.GetBalance(ctx, "ethereum", addr, nil) // chain is overridden to "polygon"
//
// An explicit `overrides` argument passed directly to a Client method
// always takes precedence over one carried in the context, letting a
// single call override an enclosing Tool-level default.
func WithOverrides(ctx context.Context, ov *Overrides) context.Context {
	if ov == nil {
		return ctx
	}
	return context.WithValue(ctx, overridesKey{}, ov)
}

// OverridesFromContext extracts an *Overrides previously attached with
// WithOverrides, or nil if none is present.
func OverridesFromContext(ctx context.Context) *Overrides {
	v := ctx.Value(overridesKey{})
	if v == nil {
		return nil
	}
	ov, _ := v.(*Overrides)
	return ov
}

// resolveOverrides implements the precedence rule: an explicit override
// passed to a Client method wins; otherwise fall back to whatever is
// attached to ctx.
func resolveOverrides(ctx context.Context, explicit *Overrides) *Overrides {
	if explicit != nil {
		return explicit
	}
	return OverridesFromContext(ctx)
}
