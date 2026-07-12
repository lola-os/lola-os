# Contributing to lola-sdks

These three SDKs (`python/`, `go/`, `typescript/`) intentionally mirror
each other's feature set and behavior. When adding or changing
functionality, please keep them in sync.

## Adding a new operation

If you add a new RPC method to `lola-core` (e.g. a new convenience
operation), add the corresponding wrapper to **all three** SDKs in the
same pull request:

1. **Python**: add a sync + `_async` pair in `python/lola_os/functions.py`,
   export both from `python/lola_os/__init__.py`, and add tests in
   `python/tests/test_functions.py`.
2. **Go**: add a method on `lola-core/pkg/sdk.Engine` first (since that's
   the only public seam into `lola-core`'s internals), then a matching
   `Client` method in `go/client.go`.
3. **TypeScript**: add a function in `typescript/src/functions.ts`,
   export it from `typescript/src/index.ts`, and add it to the
   `multiOperation` dispatch table if it fits the batch-operation shape.

Keep parameter order and naming conventions idiomatic per language
(`snake_case` Python, `camelCase` TypeScript, `PascalCase` Go methods),
but keep the underlying RPC method name and wire-format field names
(`snake_case`, matching `lola-core/internal/jsonrpc`) identical across all
three — that consistency is what lets the docs describe one JSON-RPC API
surface instead of three.

## Context overrides

Every override field added to `lola-core/internal/config.Overrides` should
get a matching field in:

- Python: `lola_os/context.py`'s `Overrides` dataclass
- Go: it's already a direct alias of `config.Overrides` via
  `pkg/sdk.Overrides` — no extra mapping needed
- TypeScript: `src/context.ts`'s `Overrides` interface, plus a case in
  `overridesToRpcParams`

## Testing

- Python: `pytest` (or the manual runner pattern in this repo's history,
  if `pytest` isn't installable in your environment — see
  `python/README.md`)
- Go: `go test ./...` (requires `lola-core` to compile first, since this
  SDK imports `lola-core/pkg/sdk`)
- TypeScript: `npm test` (compiles via `tsc`, runs via the custom
  `run-tests.js` runner — Jest/Mocha/Vitest will also work against the
  same `test/*.test.ts` files if you'd rather use one of those)

Mock the `LolaCore`/`Client`/`Engine` boundary in unit tests rather than
spawning a real `lola-core` process — see each SDK's `tests/test_functions.py`
/ `test/functions.test.ts` for the established pattern. Save real
subprocess/binary integration tests for a dedicated integration suite
(not yet present in this repo — a good first contribution!).

## Style

- Python: type hints everywhere, docstrings on every public function,
  `from __future__ import annotations` at the top of every module.
- Go: every exported identifier gets a doc comment; errors are wrapped
  with `fmt.Errorf("package: doing thing: %w", err)` for traceability.
- TypeScript: explicit return types on every exported function; avoid
  `any` except at genuine JS-interop boundaries (and prefer `unknown` +
  a type guard there instead).

## Versioning

All three SDKs are versioned together (currently `1.0.0`) and should ship
together — a Python SDK release that depends on a `lola-core` RPC method
the TypeScript SDK doesn't yet support is a regression in consistency,
even if neither SDK is technically broken.
