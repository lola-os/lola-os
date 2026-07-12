# DEVELOPMENT PROMPT – LOLA OS v1.0 (Fully Open Source, Free Forever – Maturity Edition)

I have given you the following design and planning files:

- `blueprint.md` – the complete LOLA OS Technical Blueprint (v1.0, fully open source, no paid tiers, no LOLA Chain, free forever, including all maturity features: budget enforcement, replay engine, registry, doctor, metrics, nonce management, idempotency, pre-flight validation, context overrides, and structured execution plans).
- `roadmap.md` – the single 16‑week development plan delivering all features together as v1.0, with expanded milestones for the maturity tooling and consolidated repository structure.
- `branding.md` – the grayscale branding guide (no accent colours, Inter & JetBrains Mono, calm animations, human voice).

**Your task:**

Build the **complete, operationally mature LOLA OS system** exactly as specified in the blueprint and roadmap. This is not a prototype – it is a production‑ready v1.0 that delivers the "5‑minute integration promise" for every AI developer while providing enterprise‑grade safety nets, diagnostics, and automation capabilities.

Unlike the previous version, **there are no private repositories, no "pro" features, no licence keys, no hosted API, no dashboards with billing, no marketplace**. Everything is public, Apache 2.0, and free forever. The only thing we sell later (outside this codebase) is optional enterprise support – but the software itself is complete and unrestricted.

---

## What You Must Build

The system consists of **four main repositories** (all public):

1. **`lola-core`** – Go engine (single binary) that does all blockchain/oracle operations, encrypted vault, HITL (console + WebSocket), rich logging, JSON‑RPC interface, **persistent SQLite registry**, **structured execution replay engine**, **budget enforcement (circuit breaker)**, **nonce manager**, **idempotency cache**, **pre‑flight ABI validation**, and **operational CLI tooling (`doctor`, `registry`, `metrics`)**.

2. **`lola-sdks`** – A monorepo containing all three language SDKs:
   - **Python SDK** (`python/`) – `@lola_tool` decorator, binary management, convenience functions, **context overrides**.
   - **Go SDK** (`go/`) – `Tool()` decorator, direct core import (no subprocess), **context overrides**.
   - **TypeScript/JavaScript SDK** (`typescript/`) – `@lolaTool` decorator, Promise API, Node.js + browser via WebAssembly, **context overrides**.
   - Shared root `README.md`, `CONTRIBUTING.md`, and build scripts.

3. **`lola-ui`** – The complete user‑facing web presence:
   - **Landing Page** – Marketing site explaining the 5‑minute promise, showing code snippets, linking to docs and GitHub.
   - **Developer Documentation** – Comprehensive docs with getting started, API reference, configuration, security guide, FAQ, and **dedicated pages for all operational CLI commands** (`replay`, `doctor`, `registry`, `metrics`).
   - **Embedded Examples** – All examples are written as `.md` files within the documentation pages (e.g., `/docs/examples/python-5min`, `/docs/examples/replay-plan`). No separate `lola-examples` repository.
   - Built with **React 19 + Next.js 14 (App Router)** using **shadcn/ui** components, **Tailwind CSS**, **Framer Motion** for animations. Must be fully responsive, modern, clean, and extremely fast.

4. **`lola-infra`** – Docker Compose, deployment scripts, and integration test harnesses for running LOLA in containerised environments. Also includes example `plan.json` files for testing and documentation references.

**There is no hosted API, no dashboard, no marketplace, no billing, no authentication required for using the SDK.** All SDKs work entirely locally. The only "service" is the optional WebSocket HITL server embedded in the Go binary – and that is run by the user on their own machine.

---

## Repository Structure (All Public – Consolidated)

```
lola-os/
├── lola-core/          # Go engine (with replay, registry, doctor, metrics, budget)
├── lola-sdks/          # Monorepo for all SDKs
│   ├── python/         # Python SDK (with context overrides)
│   ├── go/             # Go SDK
│   ├── typescript/     # TypeScript/JavaScript SDK
│   ├── README.md       # Root SDK docs
│   └── CONTRIBUTING.md # Contribution guide for all SDKs
├── lola-ui/            # Landing page + Developer documentation (with embedded examples)
│   ├── app/            # Next.js App Router
│   ├── components/     # shadcn/ui components
│   ├── content/        # MDX/Markdown documentation files (with embedded examples)
│   ├── public/         # Static assets (logo, favicon)
│   └── README.md       # UI docs
├── lola-infra/         # Docker Compose, deployment scripts, plan.json examples
└── README.md           # Root overview
```

All code must be **well documented** (inline comments, READMEs, API reference generated from docstrings). Each repository must have its own `README.md` explaining purpose, installation, and usage.

---

## Detailed Requirements by Component

### 1. `lola-core` (Go Engine) – The Mature Core

**Must implement everything from the updated blueprint, including the foundational maturity layer:**

#### Core Blockchain & Oracle Logic
- Multi‑chain gateway (EVM + Solana) with modular `ChainAdapter` interface (formalised, documented).
- EVM: native balance, token balance (ERC20), contract calls (ABI, pre‑deployed, verified, raw), transaction sending, gas estimation, nonce management.
- Solana: SOL balance, SPL token balance, transaction building, signing.
- Oracle gateway: Chainlink price feeds (read from contract), generic REST API calls (with rate limiting and retries).

#### Security & Key Management
- Encrypted vault (AES‑256‑GCM, scrypt, CLI for key management). Keys never in plaintext.
- Read‑only mode by default; write operations require explicit approval (HITL).
- Console HITL: rich terminal UI (colours, boxes, icons) that prompts for Approve/Deny/Skip. Timeout configurable.
- WebSocket HITL server: same approval requests sent over WebSocket for custom UIs (localhost only, no auth).

#### Operational Maturity (MUST Build – Non‑Negotiable)
- **Persistent SQLite Registry** (`~/.lola/lola.db`):
  - Table `transactions`: hash, chain, from, to, value, gas, status, timestamp, plan_id.
  - Table `nonces`: chain, address, nonce.
  - Table `execution_plans`: id, description, status, result.
  - Ensures crash recovery and full auditability.
- **Nonce Manager**: Atomic in‑memory nonce counter per chain/wallet, persisted to SQLite. Prevents double‑spends and nonce conflicts in concurrent environments.
- **Idempotency Cache**: Store results of write operations keyed by user‑supplied `idempotency_key`. TTL: 24 hours. Returns cached result on duplicate key, preventing duplicate broadcasts.
- **Budget Monitor (Circuit Breaker)**: Background thread that tracks gas spent and USD cost (via oracle conversion) against `budget` limits (`max_gas_spend_per_session`, `max_usd_spend_per_session`, `max_requests_per_minute`). If exceeded, LOLA automatically pauses all write operations and emits a CRITICAL log. Supports `pause` (default), `notify`, or `deny` actions.
- **Pre‑flight ABI Validation**: Before simulation or broadcast, fetch the contract's ABI (from Etherscan, Sourcify, or local) and validate the provided `method` name and `args` types against it. Throw clear `TypeError` on mismatch – catches typos and AI hallucinations early.
- **Structured Execution Replay Engine** (`lola replay <plan.json>`):
  - Ingests a JSON plan (see schema in Blueprint Appendix D).
  - Executes operations sequentially: `call_contract`, `send_transaction`, `transfer_token`, `execute_contract`, `swap_tokens`, `assert`, `wait`.
  - Supports variable interpolation (`${key}`) for chaining outputs.
  - Records the entire plan execution in the SQLite registry.
  - Flags: `--fork-url <RPC>` (execute against a specific network or fork), `--dry-run` (simulate only), `--output receipt.json`.
- **`lola doctor` Command**: Comprehensive environment health check:
  - Validates RPC connectivity (ping & latest block).
  - Verifies Oracle endpoint reachability.
  - Checks `config.yaml` syntax.
  - Tests Vault integrity (decrypts test string).
  - Detects Hardware Wallet connectivity.
  - Outputs a beautiful pass/fail table with actionable fix messages.
- **`lola registry` Commands**:
  - `lola registry list` – Shows recent operations (tx hash, chain, method, timestamp, status).
  - `lola registry show <tx_hash>` – Fetches full details from SQLite.
  - `lola registry clear` – Resets the local registry (truncates tables).
- **`lola metrics` Endpoint**: Exposes JSON lines (and optional Prometheus format) with operational metrics: total requests, avg latency per chain, rate‑limit hits, gas spent, success/failure counts.

#### Integration & Communication
- JSON‑RPC server over stdin/stdout (for Python/TS SDKs) and also a local TCP socket (optional, for the Go SDK direct import).
- Rich logs: structured JSON with `color`, `icon` fields; when `format: rich` in config, output ANSI colour codes and Unicode icons.
- Configuration: YAML (`~/.lola/config.yaml`) + env vars + sensible defaults, including the `budget` section.

#### Testing
- Unit tests for each module.
- Integration tests against public testnets (Sepolia, Solana devnet).
- **Specific tests for**: budget circuit breaker triggers, replay engine correctness (multi‑step plan with assertions), nonce manager concurrency, pre‑flight ABI validation (catching mismatches), and `lola doctor` diagnostics.

---

### 2. `lola-sdks` (Monorepo – All SDKs)

The `lola-sdks` repository contains three SDKs in subdirectories. Each SDK must be independently installable and documented.

#### Root `lola-sdks` Requirements
- `README.md` – overview of all SDKs, installation instructions for each, and links to their individual docs.
- `CONTRIBUTING.md` – guidelines for adding features, fixing bugs, and maintaining consistency across SDKs.
- Shared build scripts (e.g., `./scripts/build-all.sh`, `./scripts/test-all.sh`) to reduce duplication.

---

#### 2.1 Python SDK (`lola-sdks/python/`)

- One `pip install lola-os` command.
- Binary management: download the correct `lola-core` binary for the user's platform (macOS, Linux, Windows, both amd64/arm64). Fallback to bundled binary in the wheel.
- Singleton `LolaCore` class that manages the subprocess and JSON‑RPC communication.
- `@lola_tool` decorator: inspects function signature, detects blockchain/oracle parameters, maps to the appropriate core method. Works with sync and async functions.
- **Context Overrides (MUST Build)**:
  - Decorator override: `@lola_tool(config_overrides={"chain": "polygon", "mode": "simulate_only"})`.
  - Context manager override: `with lola.override(rpc_url="backup.alchemy.io", budget_max_gas="0.1 ETH"): result = execute_contract(...)`.
  - These allow runtime switching of chains, RPCs, and budget limits without changing the global config.
- Convenience functions: `get_balance`, `call_contract`, `multi_operation`, `get_price_from_oracle`, `fetch_external_api`, `send_transaction`, `transfer_token`, `swap_tokens`, `listen_to_websocket`.
- `stream_logs()` generator that yields parsed JSON logs as they arrive (for custom dashboards).
- Rich error messages with actionable suggestions, including specific exceptions: `BudgetExceededError`, `ABIMismatchError`, `RPCConnectionError`.
- Async support (all functions have `_async` versions or native `async def`).
- Type hints everywhere, thoroughly tested with `pytest`.
- `python/README.md` – quickstart, API reference, and example usage.

---

#### 2.2 Go SDK (`lola-sdks/go/`)

- Importable package `github.com/lola-os/lola-go` (or `lola-sdks/go`).
- Direct use of `lola-core` as a library (no subprocess) – the Go SDK imports the core engine and calls it directly.
- `Tool()` decorator (higher‑order function) that wraps a Go function and detects blockchain operations.
- Client struct with all convenience methods.
- **Context Overrides** in the Tool function and client methods.
- Example usage in a Go‑based AI agent (e.g., using `github.com/joaorufino/gollm` or custom).
- `go/README.md` – installation, usage, and example.

---

#### 2.3 TypeScript/JavaScript SDK (`lola-sdks/typescript/`)

- NPM package `lola-os`.
- Spawns `lola-core` as a subprocess (Node.js) or uses WebAssembly build (optional).
- `@lolaTool` decorator (or higher‑order function for JS).
- Promise‑based API.
- **Context Overrides** via options object in tool calls.
- Works in Node.js (v18+) and optionally in the browser via WebAssembly.
- Full type definitions.
- `typescript/README.md` – installation, usage, and example.

---

### 3. `lola-ui` (Landing + Developer Docs with Embedded Examples)

**Tech stack:** Next.js 14 (App Router) with TypeScript, Tailwind CSS, shadcn/ui components, Framer Motion.

#### 3.1 Landing Page

**Sections:**

- Hero: "Add `@lola_tool` to any function – blockchain in 5 minutes". Code snippet showing the decorator.
- Features: bullet points from the blueprint (multi‑chain, oracles, HITL, rich logs, free forever, **budget safety, replay automation**).
- How it works: diagram (static or animated).
- Comparison table: LOLA OS vs other SDKs.
- Call to action: buttons for "Get Started" (docs) and "GitHub".
- Footer: links to docs, GitHub, license.

**No fake numbers** – if no real stats, show "Join the waitlist" or simply omit counters.

---

#### 3.2 Developer Documentation

**Pages (MUST include all of these):**

- **Home (docs home)** – overview, 5‑minute promise, quick install.
- **Getting started** – step‑by‑step with code snippets, including running `lola doctor` to verify the environment.
- **Operational Tooling (MUST have dedicated pages for)**:
  - `lola replay` – how to write a `plan.json`, execute it, and interpret the receipt. Include the full JSON schema with a visual breakdown.
  - `lola doctor` – diagnostic tool usage and common fixes.
  - `lola registry` – managing transaction history and audit trails.
  - `lola metrics` – monitoring and observability.
- **API reference** – generated from docstrings (manual pages for main functions is acceptable, but aim for completeness).
- **Examples** – all examples are embedded as `.md` files within the documentation:
  - `docs/examples/python-5min.mdx` – 5‑minute demo walkthrough.
  - `docs/examples/replay-plan.mdx` – `plan.json` walkthrough with a real example.
  - `docs/examples/crewai.mdx` – CrewAI integration.
  - `docs/examples/langchain.mdx` – LangChain integration.
  - `docs/examples/custom-bot.mdx` – custom trading bot with HITL.
- **Security guide** – how to use the vault, hardware wallets, HITL, and the budget circuit breaker.
- **Configuration reference** – all environment variables and YAML keys, including the `budget` section.
- **FAQ** (expanded with questions about budget, replay, idempotency, and registry).
- **Blog / news** (optional, placeholder).

**Design Requirements:**

- Follow `branding.md` exactly: grayscale, Inter font, JetBrains Mono for code, no accent colours.
- Dark mode toggle (light/dark using Tailwind's dark mode).
- Mobile responsive: sidebar collapses to hamburger menu, code blocks scroll horizontally.
- Framer Motion: subtle fade‑in on page load, slide‑up for cards, smooth scroll.
- Code snippets with copy button, syntax highlighting (using `prism.js` or `shiki`).
- Include the structured plan JSON schema in a dedicated "Replay" page with a visual breakdown.

---

### 4. `lola-infra` (Containerisation & Test Harnesses)

- `docker-compose.yml` that runs a test environment with:
  - A local `lola-core` instance (exposing WebSocket HITL port 8080).
  - A simple demo dashboard (optional) – not required for v1.0, but can be a minimal HTML page.
- `examples/plan.json` – sample structured execution plans for testing and documentation references.
- `scripts/` – deployment scripts for different environments (optional).
- `README.md` – how to use Docker, run the test harness, and execute `lola replay` examples.

---

## Documentation Site & Landing Page Details

### React 19 + shadcn/ui + Framer Motion

- Use the **Next.js 14 App Router** for both docs and landing (within `lola-ui`).
- Install `shadcn/ui` (choose components: button, card, sheet, code block, table, avatar, etc.)
- Tailwind CSS configuration must include the grayscale colour palette as defined in `branding.md`. Extend theme with `gray` shades exactly.
- Implement dark mode: use `next-themes` with a toggle.
- Framer Motion animations:
  - Hero: fade-in-up.
  - Cards: hover scale + shadow.
  - Scroll‑triggered fade‑ins (use `whileInView`).
  - Smooth page transitions (optional).
- Code highlighting: use `react-syntax-highlighter` with a grayscale theme (e.g., `atomOneDark` but converted to grayscale).
- All text must follow the brand voice: clear, jargon‑free, helpful.

### Mobile Responsiveness

- Navigation: top bar with logo and menu button on mobile, sliding sheet.
- Layout: single column on mobile, multi‑column on desktop.
- Code blocks: horizontal scroll, font size adjusts.
- Touch targets: at least 44x44px.

---

## Quality Requirements

- **No placeholders** – every function must be fully implemented, no `// TODO` or `pass` unless it's a clear extension point (e.g., plugin interfaces for new chains).
- **Comprehensive tests** – unit tests for core logic, integration tests that call real RPCs (using testnet environment variables). **Specific tests for**: budget circuit breaker, replay engine correctness, nonce manager concurrency, pre‑flight ABI validation, and `lola doctor` diagnostics.
- **Documentation** – each repository has a `README.md` with install, usage, and development instructions. The `lola-ui` site must be self‑contained (you can run `npm run build` and serve static files). All examples are embedded as `.md`/`.mdx` files within the docs.
- **Error handling** – the Go binary never panics; it returns JSON‑RPC errors. Python SDK translates those into specific exceptions (`BudgetExceededError`, `ABIMismatchError`, etc.).
- **Security** – the vault must be auditable. Use only well‑vetted crypto libraries (`crypto/aes`, `golang.org/x/crypto/scrypt`). No custom crypto.
- **Performance** – balance check under 200ms, contract read under 300ms (on fast RPC). Use caching (in‑memory) with TTL.

---

## What You Must Deliver

A single zipped archive named `lola-os-v1.0.zip` containing:

1. **All source code** organised as described above:
   - `lola-core/`
   - `lola-sdks/` (with `python/`, `go/`, `typescript/` subdirectories)
   - `lola-ui/` (with landing + docs, all examples embedded as `.md`/`.mdx` files)
   - `lola-infra/`

2. **`README.md`** at the root explaining how to build and run everything (including setting up the UI locally).

3. **`guide.md`** – a detailed guide covering:
   - System prerequisites (Go, Python, Node, Docker – optional).
   - How to build the Go binary for all platforms (or use pre‑built releases).
   - How to install the Python SDK from source (`pip install -e lola-sdks/python`).
   - How to run the 5‑minute demo (including running `lola doctor` first).
   - **How to use the operational CLI**: `lola doctor`, `lola replay`, `lola registry`, `lola metrics`.
   - How to write a structured `plan.json` and execute it.
   - How to start the UI (`cd lola-ui && npm run dev`).
   - How to run all tests across all repositories.

4. **Environment example file** (`.env.example`) for the UI (if needed – may only need RPC URLs for demos).

**The archive must be ready to unzip and run on a fresh Ubuntu 22.04 or macOS machine with Go, Python, Node installed.** No missing files.

---

## Branding Adherence Check

- All web pages (landing and docs in `lola-ui`) must use the grayscale palette – no blue, green, red, etc. for accents. Only grayscale. (Logos can be grayscale, icons can be gray.)
- Typography: Inter for body, headings; JetBrains Mono for code. Use `font-sans` for Inter, `font-mono` for code.
- Animations: only the allowed ones (`fade-in`, `slide-up`, `pulse-slow`, `float`). No bouncy or flashy effects.
- Voice: documentation and UI copy must be simple, direct, and helpful. Avoid hype words like "revolutionary", "game‑changing". Use "simple", "fast", "works with any framework", "operationally mature".

---

## Final Note

This is a **complete, operationally mature system** – not a prototype. The code you write must be of production quality, well structured, and easy to extend. The maturity features (budgeting, replay, registry, doctor, metrics, nonce management, idempotency, pre‑flight validation, context overrides) are **core requirements**, not optional extras. They transform LOLA OS from a simple bridge into an enterprise‑grade execution engine.

The repository structure is intentionally consolidated:
- **`lola-core`** – the engine.
- **`lola-sdks`** – all SDKs in one monorepo for consistency and shared tooling.
- **`lola-ui`** – landing + docs with embedded examples (no separate `lola-examples` repo).
- **`lola-infra`** – containerisation and test harnesses.

The goal is that after unzipping, a developer can run the 5‑minute demo, then immediately explore `lola replay` with a provided `plan.json` (found in `lola-infra/examples/` and documented in `lola-ui`), and confidently run `lola doctor` to validate their setup.

**Now build the entire LOLA OS v1.0 system exactly as described – complete, mature, and production‑ready. Package everything into the zip file and deliver it.**