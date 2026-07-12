# lola-sdks

Three language SDKs for LOLA OS, sharing one design: add a decorator (or,
in Go, a higher-order function) to any function and it can talk to
blockchains, oracles, and APIs through the local `lola-core` engine. No
hosted backend, no API keys, no billing — everything runs on your own
machine.

```
lola-sdks/
├── python/      pip install lola-os      — @lola_tool decorator
├── go/          import lola-go            — Tool() higher-order function, direct library import
├── typescript/  npm install lola-os       — lolaTool() / @LolaTool
└── scripts/     build-all.sh, test-all.sh
```

## Status

All three SDKs build and pass their test suites:

| SDK | Build | Tests |
|---|---|---|
| **Python** | `pip install -e .` | 44 passing (`pytest`) |
| **TypeScript** | `npm run build` (tsc) | 21 passing (`npm test`) |
| **Go** | `go build ./...` | passing (`go test ./...`) |

The Go SDK links `lola-core/pkg/sdk` directly, so build `lola-core` first
(`cd ../lola-os/lola-core && go build ./...`) — the SDK's `go.mod` already
points a `replace` directive at it. The Python and TypeScript SDKs spawn
`lola-core` as a subprocess and talk JSON-RPC to it; point them at a built
binary via `LOLA_CORE_BINARY` or each SDK's config (see per-SDK READMEs).

## Why the Go SDK links `lola-core/pkg/sdk`

Go enforces that packages under any module's `internal/` directory are
invisible to code outside that module — and `lola-core`'s engine logic
all lives under `internal/`. That's the right structure for `lola-core`
itself, but it means the Go SDK (a separate module) cannot import
`lola-core/internal/chain`, `internal/vault`, etc. directly, no matter how
the import path is written — the compiler refuses it.

The solution: `lola-core/pkg/sdk` is a narrow, public package inside the
`lola-core` module that wires up the same engine components the JSON-RPC
server uses and exposes them as a typed `Engine`. The Go SDK is a thin
`Client` wrapper around that `Engine`, so it imports lola-core directly as
a library (no subprocess) — the one seam Go's visibility rules allow.

## Quick start per SDK

**Python:**
```bash
cd python && pip install -e .
python3 -c "from lola_os import lola_tool, get_balance; print('ok')"
```

**TypeScript:**
```bash
cd typescript && npm install && npm run build && npm test
```

**Go** (after building `lola-core`):
```bash
cd go && go mod tidy && go build ./... && go test ./...
```

Each SDK's own README has full quickstart, API reference, and
context-override examples for its language's idioms (Python
`contextvars` + `with` blocks, Go `context.Context` propagation, and
TypeScript `AsyncLocalStorage`).

## Design consistency across the three SDKs

All three expose the same operation set, named consistently for each
language's conventions (`snake_case` in Python, `camelCase` in
TypeScript, `PascalCase` methods in Go):

- Native + token balance reads
- Read-only contract calls (with automatic ABI fetch + pre-flight validation)
- State-changing contract execution, native transfers, token transfers
- DEX swaps (via direct router contract calls — no swap routing logic is
  hardcoded; you supply the router's method/ABI, same as any other
  contract call)
- Chainlink oracle price reads, generic rate-limited REST fetches
- A client-side batch/multi-operation helper
- Structured log streaming
- **Context overrides**: every SDK lets you scope a chain/RPC/budget
  override to a single call or block without touching global config —
  implemented natively per language rather than bolted on: Python
  contextvars, Go context.Context values, TypeScript AsyncLocalStorage.

## What's deliberately not included

- No swap-routing or DEX-aggregation logic — `swapTokens`/`swap_tokens`/
  the Go equivalent are thin wrappers around a direct contract call you
  supply, consistent with the blueprint's "we don't hardcode any chain or
  protocol opinions" philosophy.
- No bundled `lola-core` binaries — each SDK's `bin/README.txt` explains
  how to build one and point the SDK at it.
- No WebAssembly browser build for the TypeScript SDK in this pass (the
  Node `child_process`-based transport is what's implemented; see that
  SDK's README for the gap).
