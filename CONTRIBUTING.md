# Contributing to LOLA OS

Thanks for your interest in improving LOLA OS. This is a monorepo of
loosely-coupled subprojects — pick the one you're changing and follow its own
README for setup and tests.

## Ground rules

- **No hardcoded chain or protocol opinions.** Swaps and contract calls are thin
  passthroughs: the caller supplies the router, method, and ABI. Don't add
  routing, aggregation, or protocol-specific logic to the core.
- **Keep the operation set consistent across SDKs.** Every SDK exposes the same
  operations, named per-language convention (`snake_case` Python, `camelCase`
  TypeScript, `PascalCase` Go). A change to one usually implies the others.
- **Preserve the error contract.** Engine JSON-RPC errors carry LOLA-specific
  codes that SDKs translate into typed exceptions — keep both sides in sync.
- **Security-sensitive defaults are intentional.** The HITL WebSocket binds
  loopback-only; private keys only leave the encrypted vault in memory during a
  signing operation. Don't loosen these.

## Development

| Subproject | Setup + tests |
|---|---|
| `lola/lola-os/lola-core` | `go build ./... && go test ./...` |
| `lola/lola-sdks/python`  | `pip install -e . && pytest` |
| `lola/lola-sdks/typescript` | `npm install && npm run build && npm test` |
| `lola/lola-sdks/go`      | `go mod tidy && go build ./... && go test ./...` |
| `lola/lola-ui`           | `npm install && npm run dev` |

## Pull requests

1. Fork and create a topic branch.
2. Keep changes focused; add or update tests for behavior changes.
3. Make sure the affected subproject builds and its tests pass.
4. Open a PR with a clear description of the what and the why.

## License

By contributing, you agree that your contributions will be licensed under the
[Apache-2.0](./LICENSE) license.
