package lola

import (
	"context"
)

// ToolFunc is the shape every LOLA tool function must have: it takes a
// context (for cancellation/timeouts) and returns a result plus an error.
// Use generics-free `interface{}` for the result so tools can return
// whatever shape fits their use case (a struct, a map, a primitive).
type ToolFunc func(ctx context.Context) (interface{}, error)

// ToolOptions configures a single Tool() wrapping.
type ToolOptions struct {
	Name        string
	Description string
	Overrides   *Overrides
}

// Tool wraps fn as a named LOLA tool, applying ContextOverrides (if set)
// for the duration of fn's execution. This mirrors the Python SDK's
// `@lola_tool(config_overrides=...)` and the TypeScript SDK's
// `@lolaTool({ configOverrides: ... })`.
//
// Usage:
//
//	checkBalance := lola.Tool(lola.ToolOptions{Name: "check_balance"}, func(ctx context.Context) (interface{}, error) {
//	    return client.GetBalance(ctx, "ethereum", "0x...", nil)
//	})
//	result, err := checkBalance(context.Background())
//
// Because Go has no decorator syntax, `Tool` is a higher-order function:
// it takes your function and returns a wrapped version with the same
// signature, ready to be registered with whatever agent framework you're
// using (the returned ToolFunc satisfies most frameworks' "any callable
// returning (result, error)" tool interface directly).
func Tool(opts ToolOptions, fn ToolFunc) ToolFunc {
	return func(ctx context.Context) (interface{}, error) {
		if opts.Overrides != nil {
			ctx = WithOverrides(ctx, opts.Overrides)
		}
		return fn(ctx)
	}
}

// Described is a small helper for agent frameworks that want a
// name/description pair alongside the callable, without forcing every
// framework integration to know about ToolOptions.
type Described struct {
	Name        string
	Description string
	Fn          ToolFunc
}

// Describe wraps a Tool()-produced ToolFunc with its metadata for
// frameworks that expect a (name, description, callable) triple.
func Describe(opts ToolOptions, fn ToolFunc) Described {
	return Described{Name: opts.Name, Description: opts.Description, Fn: Tool(opts, fn)}
}
