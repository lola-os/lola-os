# lola-go (Go SDK)

[![Go Reference](https://pkg.go.dev/badge/github.com/lola-os/lola-go.svg)](https://pkg.go.dev/github.com/lola-os/lola-go)
![Go Report Card](https://goreportcard.com/badge/github.com/lola-os/lola-go?style=flat-square)
![License](https://img.shields.io/badge/license-Apache_2.0-555555?labelColor=252525&style=flat-square)

The Go SDK imports `lola-core` directly as a library — there's no
subprocess, no JSON-RPC serialization overhead, just Go function calls
into the same engine code the CLI uses.

```go
client, err := lola.NewClient(ctx, lola.ClientOptions{
    VaultPassphrase: os.Getenv("LOLA_VAULT_PASSPHRASE"),
})
balance, err := client.GetBalance(ctx, "ethereum", "0x...", nil)
```

## How this works (and why it's structured this way)

Go enforces that any package under a module's `internal/` directory is
invisible outside that module's own tree — which is exactly how
`lola-core` is organized. So this SDK can't import
`lola-core/internal/...` packages directly; Go's compiler refuses it
regardless of file paths.

The fix: `lola-core/pkg/sdk` is a small, public package inside the
`lola-core` module that wires up the same engine components the JSON-RPC
server uses (chain adapters, vault, registry, budget breaker, nonce
manager, idempotency cache) and exposes them as a typed `Engine`. This Go
SDK (`lola-sdks/go`) is a thin, ergonomic `Client` wrapper around that
`Engine` — context-override propagation, a `Tool()` helper, and friendlier
method signatures, but no protocol translation in between.

## Status

`go build ./...` and `go test ./...` pass on Go 1.22+. The `go.mod` uses a
`replace` directive pointing at `../../lola-os/lola-core`, matching the
monorepo layout — adjust the path if your checkout differs. Build
`lola-core` first (the SDK links its `pkg/sdk` package), then this SDK.

## Building

```bash
# from the monorepo root (the lola/ directory)
cd lola-os/lola-core && go mod tidy && go build ./... && cd -
cd lola-sdks/go && go mod tidy
go build ./...
go vet ./...
go test ./...
```

## Context overrides

Go has no decorator syntax and no implicit thread-local state, so
overrides are propagated the idiomatic Go way: through `context.Context`.

```go
ctx := lola.WithOverrides(context.Background(), &lola.Overrides{
    Chain: "polygon",
    BudgetMaxUSD: ptrFloat64(5.0),
})
balance, err := client.GetBalance(ctx, "ethereum", addr, nil) // chain is overridden to "polygon"
```

An override passed directly as a Client method's last argument always
wins over one carried in the context, so a single call can override an
enclosing default:

```go
// ctx carries a Polygon override from an outer Tool, but this one call
// explicitly asks for Solana instead:
balance, err := client.GetBalance(ctx, "ethereum", addr, &lola.Overrides{Chain: "solana"})
```

## The `Tool()` helper

```go
checkBalance := lola.Tool(lola.ToolOptions{Name: "check_balance"}, func(ctx context.Context) (interface{}, error) {
    return client.GetBalance(ctx, "ethereum", "0x...", nil)
})

result, err := checkBalance(context.Background())
```

`Tool()` returns a plain `func(context.Context) (interface{}, error)` —
compatible with most Go agent frameworks' tool-registration signature
without any adapter code. Scope an override to every call of one specific
tool with `ToolOptions.Overrides`.

See `example/main.go` for a complete runnable example (build with
`go run -tags example ./example` once lola-core is built).

## API reference

| Method | Description |
|---|---|
| `NewClient(ctx, ClientOptions)` | Opens the vault/registry, dials configured chains |
| `GetBalance(ctx, chain, address, *Overrides)` | Native balance |
| `GetTokenBalance(ctx, chain, address, token, *Overrides)` | ERC20/SPL balance |
| `CallContract(ctx, chain, ContractCallRequest, *Overrides)` | Read-only call |
| `ExecuteContract(ctx, chain, ContractCallRequest, keyName, idempotencyKey, *Overrides)` | State-changing call |
| `SendTransaction(ctx, chain, TxRequest, keyName, *Overrides)` | Native transfer |
| `TransferToken(ctx, chain, token, from, to, amountRaw, keyName, idempotencyKey, *Overrides)` | Token transfer |
| `FetchExternalAPI(ctx, url, &out)` | Rate-limited, retrying REST GET |
| `BudgetStatus()` | Current session spend snapshot |
| `VaultKeyNames()` | List stored key names (never values) |
| `Close()` | Releases vault/registry resources |

## Tests

`overrides_test.go` covers context-propagation and precedence logic and
has no external dependencies. `client.go`'s actual engine calls are only
exercisable once `lola-core` compiles, since the whole `lola` package
(including the test binary) must compile against `pkg/sdk`.

---

Built and maintained by **[0xSemantic](https://github.com/0xSemantic)** —
developer and visionary behind LOLA OS. Licensed under Apache-2.0.
