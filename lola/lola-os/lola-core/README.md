# lola-core

The Go engine at the heart of LOLA OS. A single binary that connects AI
agents to blockchains, oracles, and APIs — with an encrypted vault,
human-in-the-loop approval, a persistent SQLite registry, nonce
management, idempotency, a budget circuit breaker, pre-flight ABI
validation, and a structured execution replay engine.

The Python and TypeScript SDKs in `lola-sdks/` talk to this binary over
JSON-RPC; the Go SDK links `pkg/sdk` directly and runs the same engine
in-process. `go build ./...` and `go test ./...` both pass on Go 1.22+.

## Prerequisites

- Go 1.22 or newer
- No cgo / C toolchain required — the SQLite driver (`modernc.org/sqlite`)
  is pure Go, so `lola-core` builds to a single static binary on macOS,
  Linux, and Windows, for both amd64 and arm64.

## Building

```bash
cd lola-core
go mod tidy      # resolves dependency versions and writes go.sum
go build -o bin/lola ./cmd/lola
```

Cross-compiling for another platform:

```bash
GOOS=linux GOARCH=arm64 go build -o bin/lola-linux-arm64 ./cmd/lola
GOOS=darwin GOARCH=arm64 go build -o bin/lola-darwin-arm64 ./cmd/lola
GOOS=windows GOARCH=amd64 go build -o bin/lola-windows-amd64.exe ./cmd/lola
```

## Running the tests

```bash
go test ./...
```

The suite is network-independent (budget breaker, nonce manager
concurrency, ABI validation and argument coercion, replay engine logic,
vault, registry, config/catalog resolution) and runs with no external
services. Live-testnet integration tests aren't bundled — wire real RPC
calls into `_test.go` files following the patterns in `internal/chain/evm`
and `internal/chain/solana` if you want them.

## Quick start

```bash
# 1. Create your encrypted vault and store a signing key
./bin/lola vault init
./bin/lola vault add deployer

# 2. Check your environment
./bin/lola doctor --vault-passphrase "your-passphrase"

# 3. Start the JSON-RPC engine (this is what the SDKs spawn automatically)
LOLA_VAULT_PASSPHRASE="your-passphrase" ./bin/lola serve

# 4. In another terminal, run a structured plan
./bin/lola replay ../lola-infra/examples/plan.json --dry-run
```

## CLI reference

| Command | Purpose |
|---|---|
| `lola serve` | Start the JSON-RPC engine (stdio and/or `--tcp host:port`) |
| `lola chains` | List every supported chain (`--json`) and which are enabled |
| `lola doctor` | Health-check RPC connectivity, config, vault, registry |
| `lola replay <plan.json>` | Execute a structured plan (`--dry-run`, `--fork-url`, `--output`) |
| `lola registry list/show/clear` | Inspect or reset the local transaction history |
| `lola metrics` | Print operational counters as JSON or Prometheus text |
| `lola vault init/add/list/remove` | Manage the encrypted key vault |

## Supported chains

`lola-core` has a built-in catalog of 40+ chains — every major EVM L1/L2
(Ethereum, Polygon, Arbitrum, Optimism, Base, BNB Chain, Avalanche,
Gnosis, Linea, Scroll, zkSync Era, Polygon zkEVM, Mantle, Blast, Celo,
Moonbeam, Cronos, and more), the common testnets (Sepolia, Holesky, Base
Sepolia, Amoy, Fuji…), and Solana (mainnet/devnet/testnet). See
`internal/chain/catalog.go`, or run `lola chains`.

Enabling a chain is as short as its name in `~/.lola/config.yaml` — the
chain id, native symbol, public RPC, and explorer are filled in from the
catalog automatically. Anything you set overrides the catalog default, so
you keep full control (custom RPC, private networks, etc.).

## Configuration

See `config.example.yaml`. Copy it to `~/.lola/config.yaml` and edit, or
run with no config file at all to use the built-in defaults. Every key can
also be set via `LOLA_*` environment variables (e.g. `LOLA_MODE=live`,
`LOLA_RPC_URL_ethereum=…`, `LOLA_VAULT_PASSPHRASE=…`) — see
`internal/config/config.go` for the full list. Precedence, highest first:
runtime context overrides → environment variables → `config.yaml` →
built-in defaults.

## Package layout

```
cmd/lola/              CLI entrypoint and subcommands (cobra)
pkg/sdk/               Public in-process engine — the seam the Go SDK links against
internal/config/       YAML + env config loading, catalog fill, context overrides
internal/chain/        ChainAdapter interface, the chain catalog, + evm/ and solana/ adapters
internal/vault/        AES-256-GCM + scrypt encrypted key storage
internal/registry/     SQLite persistence (transactions, nonces, plans)
internal/nonce/        Atomic, persisted nonce manager
internal/idempotency/  24h TTL idempotent-result cache
internal/budget/       Gas/USD/rate circuit breaker
internal/oracle/       Chainlink price feeds + generic REST gateway
internal/abi/          Pre-flight ABI validation + JSON→ABI argument coercion
internal/replay/       Structured execution replay engine
internal/hitl/         Console approval UI + localhost WebSocket server
internal/jsonrpc/      Minimal JSON-RPC 2.0 server (stdio + TCP)
internal/rpc/          Wires JSON-RPC methods to the engine components
internal/logging/      Rich/JSON structured logger
```

`pkg/sdk` is the one public package outside `internal/`: it exposes the
engine (chains, vault, budget, nonce, idempotency, oracle, HITL) as a typed
`Engine` so Go programs can embed lola-core as a library. Everything else
stays internal to the module.

## Security notes

- Private keys are only ever held in memory for the duration of a signing
  operation; at rest they live in the AES-256-GCM-encrypted vault file.
- The HITL WebSocket server refuses to bind to anything but a loopback
  address, since the protocol has no authentication by design (it's meant
  for a UI running on the same machine).
- `lola-core` never panics on bad input; RPC handlers return structured
  JSON-RPC errors with LOLA-specific codes (see `internal/jsonrpc/jsonrpc.go`)
  that SDKs are expected to translate into typed exceptions
  (`BudgetExceededError`, `ABIMismatchError`, `RPCConnectionError`).

## Known extension points (intentionally not implemented)

- Hardware wallet signing (Ledger/Trezor) — `lola doctor` reports this
  honestly as "not bundled" rather than faking detection.
- Solana program-specific instruction building beyond native SOL
  transfers and SPL balance reads — `CallContract`/`ExecuteContract` on the
  Solana adapter return a clear error explaining the gap.
- Chainlink price reads in the RPC layer (`get_price`) need a connected
  `ethclient` passed through — wire this in `cmd/lola/serve.go` for your
  specific deployment chain.
