# LOLA OS — `lola-core`

The Go engine at the heart of LOLA OS: a single binary that connects AI
agents to blockchains, oracles, and APIs — with an encrypted vault,
human-in-the-loop approval, a persistent SQLite registry, nonce
management, idempotency, a budget circuit breaker, pre-flight ABI
validation, and a structured execution replay engine.

This binary is what the Python, Go, and TypeScript SDKs in `lola-sdks/`
talk to — the Python and TypeScript SDKs over JSON-RPC, and the Go SDK by
linking `lola-core/pkg/sdk` directly in-process.

## What's in here

```
lola-os/
├── lola-core/        The Go engine — build, test, and CLI reference in lola-core/README.md
└── lola-infra/
    └── examples/
        └── plan.json  Example structured execution plan for `lola replay`
```

## Get started

```bash
cd lola-core
go mod tidy
go build -o bin/lola ./cmd/lola
go test ./...
```

Then:

```bash
./bin/lola chains                    # list the 40+ supported chains
./bin/lola vault init
./bin/lola vault add deployer
./bin/lola doctor --vault-passphrase "your-passphrase"
```

Full usage — every subcommand, config option, and the multichain catalog —
is in `lola-core/README.md`. From the monorepo root you can also just run
`make build && make test && make doctor`.

## Multichain out of the box

`lola-core` ships with a built-in catalog of every major EVM chain
(Ethereum, Polygon, Arbitrum, Optimism, Base, BNB Chain, Avalanche, Linea,
Scroll, zkSync, Blast, Mantle, and more), the common testnets, and Solana.
Enable any of them by name in `~/.lola/config.yaml` — the chain id, native
symbol, public RPC, and explorer are filled in automatically, and you can
override any of it (point a chain at your own Alchemy/Infura/QuickNode
endpoint, for instance). Run `lola chains` to list them all.
