# LOLA OS — Developer Guide

Your private, end-to-end guide to the system: how it fits together, how to
build and run every part, how multichain works, and where the seams and
extension points are. Public-facing docs live in each subproject's README and
in `lola-ui`; this file is the map for you, the maintainer.

---

## 1. The one-paragraph mental model

A developer adds one decorator (`@lola_tool` in Python, `lolaTool()` in
TypeScript, `Tool()` in Go) to a function. That function can now read and
write blockchains, read oracle prices, and call REST APIs. All the hard parts
— key management, signing, nonces, gas/budget limits, approval, ABI checks —
are handled by a single local Go engine, **lola-core**. Nothing is hosted;
there are no API keys and no billing. Everything runs on the user's machine.

```
   your agent code (Python / TS / Go)
            │  @lola_tool / lolaTool() / Tool()
            ▼
   ┌──────────────────────┐
   │      lola-core        │   one Go binary — the engine
   │  vault · budget · HITL │
   │  nonce · idempotency   │
   │  ABI validation        │
   │  chain adapters        │──▶ EVM chains (40+) & Solana
   │  oracle gateway        │──▶ Chainlink feeds, REST APIs
   └──────────────────────┘
```

- **Python & TypeScript SDKs** spawn `lola-core` as a subprocess and talk
  **JSON-RPC 2.0** to it over stdio (or TCP).
- **Go SDK** links `lola-core/pkg/sdk` directly and runs the engine
  **in-process** — no subprocess.

---

## 2. Repository layout

```
lola/                         ← you are here (the monorepo root; has the Makefile)
├── Makefile                  one entry point for everything (make help)
├── guide.md                  this file
├── lola-os/
│   ├── lola-core/            the Go engine (build/test/CLI)
│   │   ├── cmd/lola/         CLI: serve, chains, doctor, replay, registry, metrics, vault
│   │   ├── pkg/sdk/          PUBLIC in-process engine — the Go SDK links this
│   │   └── internal/         chain (+catalog), vault, budget, nonce, idempotency,
│   │                          abi, oracle, replay, hitl, jsonrpc, rpc, registry, config, logging
│   └── lola-infra/examples/  example plan.json for `lola replay`
├── lola-sdks/
│   ├── python/               pip install lola-os        (@lola_tool)
│   ├── typescript/           npm install lola-os        (lolaTool)
│   └── go/                    import lola-go             (Tool)
├── lola-ui/                  Next.js 14 docs + landing site (MDX)
└── lola-infra/               Dockerfile, docker-compose, dashboard, scripts

../lolaos-frontend/           standalone Vite + React 19 + three.js marketing site
```

Two front ends exist and are separate: **lola-ui** (Next.js docs) and
**lolaos-frontend** (Vite marketing site, the flashier one). Don't conflate
them.

---

## 3. Build, test, run — the fast path

From this directory:

```bash
make            # list every target
make build      # build lola-core + all three SDKs
make test       # run every test suite
make doctor     # health-check config, RPC, vault, registry
make chains     # list the 40+ supported chains
make serve LOLA_VAULT_PASSPHRASE=yourpass   # start the engine on stdio + :8899
```

Per-part commands (all also available as make targets):

| Part | Build | Test |
|---|---|---|
| lola-core | `cd lola-os/lola-core && go build ./...` | `go test ./...` |
| Python SDK | `cd lola-sdks/python && pip install -e .` | `pytest` (44 tests) |
| TypeScript SDK | `cd lola-sdks/typescript && npm run build` | `npm test` (21 tests) |
| Go SDK | `cd lola-sdks/go && go build ./...` | `go test ./...` |
| Docs site | `cd lola-ui && npm run build` | — |
| Marketing site | `cd ../lolaos-frontend && npm run build` | — |
| Docker | `make infra-up` | `make infra-test` |

**Order matters for Go:** build `lola-core` before the Go SDK — the SDK's
`go.mod` has a `replace` pointing at `../../lola-os/lola-core`.

---

## 4. Configuration & the control-vs-simplicity model

The design goal: abstract the hard parts away, but never take the wheel out of
the user's hands. Two knobs:

- **`~/.lola/config.yaml`** — everything non-secret. Chains, budget limits,
  HITL behavior, logging, vault/registry paths. Copy
  `lola-os/lola-core/config.example.yaml` to start. If absent, sensible
  defaults apply.
- **`LOLA_*` environment variables / `.env`** — secrets and overrides. Vault
  passphrase (`LOLA_VAULT_PASSPHRASE`), per-chain RPC URLs
  (`LOLA_RPC_URL_<chain>`), mode (`LOLA_MODE=live`), budget caps, etc.

Precedence (highest wins): **runtime context overrides → env vars →
config.yaml → built-in defaults.** See `internal/config/config.go`.

Safety default: `mode: read_only`. Writes are refused until you set
`mode: live` (or `LOLA_MODE=live`). SDK context-overrides can only *tighten*
to read-only, never loosen — a call can't escape a globally read-only engine.

---

## 5. Multichain — how it works and how to extend it

The heart of "works with any chain" is the **catalog** in
`internal/chain/catalog.go`: a table of 40+ chains (every major EVM L1/L2, the
common testnets, and Solana) with each chain's id, native symbol/decimals, a
public RPC, and an Etherscan-compatible explorer endpoint.

- **Enable a chain** with just its name in `config.yaml` — the rest is filled
  in from the catalog (`normalizeChains` in `config.go`). Override any field to
  take control (your own RPC, a custom symbol, a private network's chain id).
- **Default-enabled set:** `chain.DefaultEnabled` (the popular mainnets + key
  testnets + Solana). Dialing is lazy, so enabling many chains is cheap.
- **`lola chains`** lists everything and marks what's enabled (`--json` for
  scripting).

**To add a new chain to the catalog:** append one `Info{...}` line to
`Catalog` in `catalog.go`. That's it — config, `doctor`, `chains`, and all
three SDKs pick it up automatically. EVM chains need only a correct
`ChainID`/`DefaultRPC`; Solana-family chains set `Kind: "solana"`.

**Adding a whole new chain *family*** (e.g. a non-EVM, non-Solana chain):
implement the `chain.ChainAdapter` interface (`internal/chain/adapter.go`) in a
new `internal/chain/<family>` package, and wire it into `buildAdapter` in both
`cmd/lola/bootstrap.go` and `pkg/sdk/sdk.go`. The rest of the engine (budget,
nonce, registry, replay, RPC) is chain-agnostic and needs no changes.

---

## 6. The engine internals (what each package does)

| Package | Responsibility |
|---|---|
| `internal/chain` | `ChainAdapter` interface + the chain catalog |
| `internal/chain/evm` | go-ethereum-backed adapter for all EVM chains |
| `internal/chain/solana` | solana-go-backed adapter (native SOL + SPL reads) |
| `internal/vault` | AES-256-GCM + scrypt encrypted key store |
| `internal/budget` | gas/USD/rate circuit breaker (`pause`/`notify`/`deny`) |
| `internal/nonce` | atomic, persisted per-(chain,address) nonce manager |
| `internal/idempotency` | 24h TTL result cache so retries don't double-send |
| `internal/abi` | pre-flight ABI validation **and** JSON→ABI argument coercion |
| `internal/oracle` | Chainlink price reads + rate-limited retrying REST gateway |
| `internal/replay` | executes structured `plan.json` (call/assert/send/wait) |
| `internal/hitl` | console + loopback-WebSocket approval |
| `internal/jsonrpc` | minimal JSON-RPC 2.0 server (stdio + TCP) |
| `internal/rpc` | maps JSON-RPC methods to the engine components |
| `internal/config` | config loading, catalog fill, context overrides |
| `pkg/sdk` | **public** typed `Engine` wrapping all of the above |

Two important, subtle pieces that were made correct:

1. **Argument coercion (`internal/abi`).** SDK arguments arrive as JSON
   (strings, numbers). go-ethereum's `abi.Pack` needs exact Go types
   (`common.Address`, `*big.Int`, `[N]byte`, correctly-typed slices).
   `CoerceArgs` bridges the two; it's used by both the validator and the real
   EVM call path (`evm.encodeCall`), so a call that validates will also pack.
2. **`AddressFromKey` on the adapter.** Lets the SDK derive the sender address
   from a vault key by name, so callers don't have to state their own address.

---

## 7. Security model (keep these invariants)

- Private keys live only in the AES-256-GCM vault at rest, and only in memory
  for the duration of a signing operation.
- The **HITL WebSocket server binds loopback only** (`internal/hitl`), on
  purpose — it has no auth. Never change it to bind a routable address. The
  infra dashboard reaches it via an nginx reverse-proxy inside a shared network
  namespace (see `lola-infra/docker-compose.yml`).
- No hardcoded chain/protocol opinions: swaps and contract calls are thin
  passthroughs — the caller supplies router/method/ABI. Don't add routing.
- Writes are off (`read_only`) by default and gated by the budget breaker and
  (optionally) HITL approval.

---

## 8. Infra / deployment

- **Docker:** `make infra-up` builds a distroless image (multi-stage: Go build
  → static binary) and starts the engine's JSON-RPC on `:8899`.
  `make infra-dashboard` adds the optional HITL approval UI on `:8080`.
  `make infra-test` runs the smoke test (doctor, registry, replay, metrics).
  Build context is the monorepo root; the Dockerfile copies from
  `lola-os/lola-core/`.
- **No Docker:** `lola-infra/scripts/deploy-systemd.sh` installs the binary as
  a systemd service that reads the passphrase from an env file (never from
  process listings or shell history).

---

## 9. Known extension points (intentionally open)

- Hardware-wallet signing (Ledger/Trezor) — `doctor` reports it as not bundled
  rather than faking detection; implement a `HardwareSigner` seam.
- Solana program-specific instruction building beyond native SOL transfers and
  SPL balance reads — `CallContract`/`ExecuteContract` on the Solana adapter
  return a clear "not supported" error by design.
- Per-call USD budget overrides in the in-process Go engine currently fall back
  to the global breaker; chain/RPC/read-only overrides are fully honored.
- Live-testnet integration tests aren't bundled (the unit suite is
  network-independent). Add them under the existing `_test.go` patterns.

---

## 10. Release checklist

1. `make build && make test` — engine + all SDKs green.
2. `make doctor` — RPC reachability for the chains you ship enabled.
3. `make ui-build && make web-build` — both front ends compile.
4. `make infra-up && make infra-test` — container smoke test (needs registry
   access to pull the Go/distroless base images).
5. Grep for stray secrets and ensure `mode: read_only` is the shipped default.
6. Bump versions: `lola-core` `-ldflags -X main.version`, each SDK's manifest.
