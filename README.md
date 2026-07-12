<h1 align="center">LOLA OS</h1>

<p align="center">
  <b>Add one decorator to any function — and it can talk to blockchains, oracles, and REST APIs.</b><br>
  Python · Go · TypeScript. Runs entirely on your machine. No hosted backend, no API keys, no billing.
</p>

<p align="center">
  <a href="https://test.pypi.org/project/lola-os/"><img src="https://img.shields.io/badge/PyPI-lola--os-3775A9?logo=pypi&logoColor=white" alt="PyPI"></a>
  <a href="https://www.npmjs.com/package/lola-os"><img src="https://img.shields.io/badge/npm-lola--os-CB3837?logo=npm&logoColor=white" alt="npm"></a>
  <a href="https://pkg.go.dev/github.com/lola-os/lola-go"><img src="https://img.shields.io/badge/Go-lola--go-00ADD8?logo=go&logoColor=white" alt="Go"></a>
  <a href="./LICENSE"><img src="https://img.shields.io/badge/license-Apache--2.0-blue.svg" alt="License: Apache-2.0"></a>
  <img src="https://img.shields.io/badge/chains-40%2B-111111" alt="40+ chains">
  <img src="https://img.shields.io/badge/EVM%20%2B%20Solana-supported-8250df" alt="EVM + Solana">
  <img src="https://img.shields.io/badge/runs-100%25%20local-2ea44f" alt="Runs locally">
</p>

<p align="center">
  <a href="#why">Why</a> ·
  <a href="#quick-start">Quick start</a> ·
  <a href="#the-operation-set">Operations</a> ·
  <a href="#architecture">Architecture</a> ·
  <a href="#repository-map">Repo map</a> ·
  <a href="#security">Security</a> ·
  <a href="#contributing">Contributing</a>
</p>

---

## Why

Connecting an AI agent (or any program) to a blockchain usually means learning a
chain-specific SDK, wiring up RPC endpoints, managing keys, handling nonces,
estimating gas, validating ABIs, and building guardrails so a bad call can't
drain a wallet. LOLA OS collapses all of that into **one decorator**:

```python
from lola_os import lola_tool, get_balance

@lola_tool
def check(address: str):
    return get_balance("ethereum", address)
```

The same capability is exposed in Go and TypeScript with idiomatic naming. Under
the hood, a single native engine — [`lola-core`](https://github.com/lola-os/lola-core) —
does the real work: chain adapters, an encrypted key vault, nonce management,
idempotency, a spend/rate circuit breaker, pre-flight ABI validation, and
human-in-the-loop approval. Everything runs **locally**.

## Quick start

<table>
<tr><th>Python</th><th>TypeScript</th><th>Go</th></tr>
<tr>
<td>

```bash
pip install lola-os
```
```python
from lola_os import lola_tool, get_balance

@lola_tool
def check(addr: str):
    return get_balance("ethereum", addr)
```

</td>
<td>

```bash
npm install lola-os
```
```ts
import { lolaTool, getBalance } from "lola-os";

const check = lolaTool(
  (addr: string) => getBalance("ethereum", addr)
);
```

</td>
<td>

```bash
go get github.com/lola-os/lola-go
```
```go
import "github.com/lola-os/lola-go"

client, _ := lola.NewClient(ctx, lola.ClientOptions{})
bal, _ := client.GetBalance(ctx, "ethereum", addr, nil)
```

</td>
</tr>
</table>

> **Engine binary.** The Python and TypeScript SDKs drive `lola-core` as a
> subprocess over JSON-RPC; the Go SDK links it in-process. Grab a prebuilt
> binary from the [lola-core releases](https://github.com/lola-os/lola-core/releases)
> and put it on your `PATH` (or point `LOLA_CORE_BINARY` at it), or build it
> from source. See [`lola/lola-os/lola-core`](./lola/lola-os/lola-core).

## The operation set

Every SDK exposes the **same operations**, named per-language convention
(`snake_case` Python, `camelCase` TypeScript, `PascalCase` Go):

| Operation | What it does |
|---|---|
| Native / token balance | Read native and ERC-20/SPL balances |
| Read-only contract call | Auto-fetches the ABI and pre-flight-validates arguments |
| State-changing execution | Signs and submits transactions with nonce + budget guards |
| Token transfer | Native and token transfers |
| DEX swap | Thin passthrough over a router call you supply — no routing is hardcoded |
| Chainlink price read | Oracle price feeds |
| External REST fetch | Rate-limited HTTP gateway |
| Batch | Run several operations together |
| Log streaming | Structured logs from the engine |

**Context overrides** let you scope a chain / RPC / budget change to a single
call or block — implemented natively per language (`contextvars` in Python,
`context.Context` in Go, `AsyncLocalStorage` in TypeScript).

> **No hardcoded chain or protocol opinions.** Swaps and contract calls are thin
> passthroughs: you supply the router, method, and ABI. LOLA adds safety and
> ergonomics, not routing logic.

## Architecture

```
        ┌───────────────┐   ┌───────────────┐   ┌───────────────┐
        │  Python SDK   │   │TypeScript SDK │   │    Go SDK     │
        │  (lola_os)    │   │  (lola-os)    │   │  (lola-go)    │
        └──────┬────────┘   └──────┬────────┘   └──────┬────────┘
               │  JSON-RPC 2.0     │  JSON-RPC 2.0     │ in-process
               │  (subprocess)     │  (subprocess)     │ (pkg/sdk)
               └─────────┬─────────┴─────────┬─────────┘
                         ▼                   ▼
                    ┌──────────────────────────────┐
                    │          lola-core           │
                    │  chains (EVM + Solana) ·      │
                    │  vault · nonce · idempotency ·│
                    │  budget · ABI · oracle ·      │
                    │  replay · HITL · registry     │
                    └──────────────────────────────┘
```

- **Python & TypeScript** spawn `lola-core` and speak JSON-RPC 2.0 over stdio
  (or TCP via `serve --tcp`).
- **Go** imports `github.com/lola-os/lola-core/pkg/sdk` directly — same engine,
  no subprocess.
- Engine errors carry LOLA-specific codes that SDKs translate into typed
  exceptions (`BudgetExceededError`, `ABIMismatchError`, `RPCConnectionError`).

## Repository map

This is a monorepo of loosely-coupled subprojects. Work inside one at a time.

| Path | What it is |
|---|---|
| [`lola/lola-os/lola-core`](./lola/lola-os/lola-core) | The Go engine. Also published standalone at [`lola-os/lola-core`](https://github.com/lola-os/lola-core) for `go get` + binary releases. |
| [`lola/lola-sdks/python`](./lola/lola-sdks/python) | Python SDK → [`lola-os` on PyPI](https://pypi.org/project/lola-os/) |
| [`lola/lola-sdks/typescript`](./lola/lola-sdks/typescript) | TypeScript SDK → [`lola-os` on npm](https://www.npmjs.com/package/lola-os) |
| [`lola/lola-sdks/go`](./lola/lola-sdks/go) | Go SDK. Published standalone at [`lola-os/lola-go`](https://github.com/lola-os/lola-go). |
| [`lola/lola-ui`](./lola/lola-ui) | Next.js documentation + landing site (MDX, grayscale design system) |
| [`lola/lola-infra`](./lola/lola-infra) | Docker Compose harness, Dockerfile, deploy scripts, example `plan.json` |
| [`lolaos-frontend`](./lolaos-frontend) | Marketing site (Vite + React 19 + three.js) |
| [`documentation/do`](./documentation/do) | Product blueprint, branding, and roadmap |

## Supported chains

40+ chains out of the box — every major EVM L1/L2 (Ethereum, Polygon, Arbitrum,
Optimism, Base, BNB Chain, Avalanche, Gnosis, Linea, Scroll, zkSync Era, Mantle,
Blast, Celo, and more), common testnets (Sepolia, Holesky, Base Sepolia, Amoy,
Fuji…), and Solana (mainnet/devnet/testnet). Enable any by name in
`~/.lola/config.yaml`; run `lola chains` to list them all.

## Security

- **Private keys** live only in an AES-256-GCM + scrypt encrypted vault at rest,
  and in memory only for the duration of a signing operation.
- **The budget circuit breaker** caps gas, USD spend, and call rate — a runaway
  agent trips it instead of draining a wallet.
- **Human-in-the-loop** approval can gate any state-changing call; its WebSocket
  server binds loopback-only by design (no network exposure).
- **Pre-flight ABI validation** rejects malformed contract calls before they're
  ever signed.

## Building from source

Each subproject builds independently — see its own README.

```bash
# Engine
cd lola/lola-os/lola-core && go build -o bin/lola ./cmd/lola

# Python SDK
cd lola/lola-sdks/python && pip install -e . && pytest

# TypeScript SDK
cd lola/lola-sdks/typescript && npm install && npm run build && npm test

# Go SDK
cd lola/lola-sdks/go && go mod tidy && go build ./... && go test ./...
```

## Contributing

Issues and pull requests are welcome. Please keep the core principle intact:
**no hardcoded chain or protocol opinions** — swaps and contract calls stay thin
passthroughs. See [`CONTRIBUTING.md`](./CONTRIBUTING.md).

## License

[Apache-2.0](./LICENSE) © LOLA OS contributors.
