# LOLA OS Technical Blueprint

**Version:** 1.0 (The Complete Bridge – Maturity Edition)  
**Mission:** Give every AI agent superpowers – talk to any blockchain, any oracle, any service in 5 minutes.  
**License:** Apache 2.0 – free forever, no hidden features.  
**Monetization:** None. (Optional enterprise support if you need a handshake.)

---

## 1. The One‑Sentence Promise

> **Add `@lola_tool` to any Python, Go, or TypeScript function – and it instantly reads from blockchains, writes transactions, fetches real‑world data via oracles, and talks to external APIs. No new language. No months of learning. Just your code, now connected to everything.**

---

## 2. What LOLA OS Solves (In Plain English)

### The Problem That Hurts Every AI Developer

You’ve built an AI agent – maybe in CrewAI, LangChain, or just a custom script. It’s smart, it’s fast. But when you want it to:

- **Pay for an API** using crypto  
- **Verify a fact** stored on a blockchain  
- **Execute a trade** when a condition is met  
- **Get real‑world data** (price feeds, weather, sports scores) via oracles  
- **Log decisions** immutably  

…you hit a wall. Suddenly you’re learning Solidity, gas mechanics, private key management, RPC endpoints, oracle protocols, signing, nonces. Weeks of pain.

**LOLA OS removes that wall.** You stay in your world (Python, Go, TypeScript). We handle the blockchain mess **and** the oracle/API glue.

### What Makes LOLA OS Different From Every Other “Web3 SDK”

| Other SDKs | LOLA OS |
|------------|---------|
| Built for blockchain developers | Built for **AI developers** |
| You learn their chain‑specific APIs | You use `@lola_tool` – one decorator |
| Lock you into one framework (e.g., LangChain) | Works with **any** Python/Go/TS code |
| Often have a backend or API key | **Runs entirely on your machine** – no third‑party |
| “Open core” – hide features behind a paywall | **Fully open source** – nothing hidden |
| Charge per request or subscription | **Free forever** – we only sell optional support |
| Oracle access requires separate SDKs | **Built‑in oracle gateway** – Chainlink, Pyth, API3, and custom REST APIs |

---

## 3. How It Works (The 30‑Second Mental Model)

Imagine your existing function:

```python
def analyze_price(token):
    price = get_price_from_some_api(token)  # Not on‑chain
    return f"{token} price is ${price}"
```

Now add one line:

```python
from lola import lola_tool, get_price_from_oracle

@lola_tool   # That's it
def analyze_price(token):
    price = get_price_from_oracle(token)  # Now it's REAL oracle data!
    return f"{token} price is ${price}"
```

Behind the scenes, LOLA OS:

1. Detects you’re asking for an oracle/blockchain operation
2. Spins up a tiny, lightning‑fast Go engine (embedded in the SDK)
3. Connects to your configured oracles (Chainlink, Pyth, or any REST API)
4. Returns the data as if it were a normal function call

**No config files required. But when you need advanced setup – wallets, custom RPCs, multiple oracles – we support `.env` and `config.yaml` with clear examples.**

---

## 4. What’s Inside LOLA OS v1.0 (Complete, No Phases)

We’re not doing “phases” or “roadmap fluff”. This is what you get **today** when you install LOLA OS.

### 4.1 Core Engine (The Brain)

- **Single Go binary** – 10MB, self‑contained, runs anywhere (Linux, macOS, Windows)
- **No backend, no API** – everything runs locally. Your keys stay with you.
- **Encrypted vault** – store private keys safely (AES‑256, scrypt)
- **Multi‑chain gateway** – one interface for EVM (all chains) + Solana
- **Oracle gateway** – fetch data from Chainlink, Pyth, API3, or any HTTP endpoint
- **Modular chain engines** – we built two engines (EVM, Solana) with clean interfaces. Adding Cosmos, Polkadot, etc. later is plug‑and‑play.
- **Persistent State Store** – LOLA maintains a local SQLite database (`~/.lola/lola.db`) to track transaction receipts, nonces, and execution history. This ensures that if your AI agent crashes, you can resume without losing state.
- **Formalized Chain Engine Interface** – The Go codebase defines a strict `ChainEngine` interface (`GetBalance`, `SendTx`, `CallContract`, `SimulateTx`, etc.) before implementing any specific chain. This forces clean separation of concerns and makes community contributions for new chains straightforward.

### 4.2 Three Language SDKs (Use What You Love)

| SDK | What You Get | Install |
|-----|--------------|---------|
| **Python** | `@lola_tool` decorator, async support, full type hints | `pip install lola-os` |
| **Go** | `Tool()` decorator, native performance | `go get github.com/lola-os/lola-go` |
| **TypeScript/JavaScript** | `@lolaTool` decorator, Promise API, Node.js + browser (via WebAssembly) | `npm install lola-os` |

All three SDKs talk to the same Go core engine. Your code stays clean. We do the heavy lifting.

### 4.3 What You Can Do (100% Read + Write + Oracle)

#### Read Operations (No approval needed)

- `get_balance(address)` – native balance on any chain
- `get_token_balance(address, token)` – ERC20 / SPL tokens
- `call_contract(address, method, args)` – read any contract (ABI, pre‑deployed, or auto‑fetched)
- `multi_operation(addresses)` – batch balances across 100+ wallets
- `get_transaction(tx_hash)` – details of any tx (checks local DB first, then chain)
- `get_block(block_number)` – block data

#### Write Operations (Requires signing – but LOLA makes it safe)

- `send_transaction(to, amount)` – transfer native currency
- `transfer_token(to, amount, token)` – send ERC20 / SPL tokens
- `execute_contract(address, method, args, value)` – call any contract method that changes state
- `swap_tokens(token_in, token_out, amount)` – DEX swap (Uniswap, Pancake, Raydium)

#### Oracle & External Service Operations (Real‑world data)

- `get_price_from_oracle(token, source)` – price feeds (Chainlink, Pyth, etc.)
- `fetch_external_api(url, method, headers, body)` – any REST API
- `listen_to_websocket(url, handler)` – real‑time data streams
- `query_subgraph(endpoint, query)` – The Graph protocol

#### Structured Execution & Replay (The Force Multiplier)

LOLA OS can ingest a **structured execution plan** (JSON) and replay it deterministically on any chain or forked environment.

- **Command:** `lola replay <plan.json> --fork-url <RPC>`
- **Capabilities:**
  - Reads a sequence of operations (e.g., `["call_contract", "swap_tokens", "send_transaction"]`).
  - Executes them step-by-step with automatic nonce management.
  - Returns a detailed receipt (success/failure, balance deltas, gas used, tx hashes).
- **Why this matters:** It turns theoretical attack vectors or complex DeFi strategies into verifiable, executable proofs. Security scanners can export a plan, and LOLA proves it works.

**All operations go through safety checks:**  
- Simulation first for writes  
- Rate limiting for external APIs  
- Retry logic with exponential backoff  
- Optional human‑in‑the‑loop (HITL) for sensitive operations

### 4.4 Rich Console Logs – Designed for Humans, Not Just Machines

We threw away boring, monochrome logs. LOLA OS outputs **beautiful, colour‑coded, structured logs** that are a joy to read.

**Example of what you see in your terminal:**

```
┌─────────────────────────────────────────────────────────────────┐
│ 2026-06-10 14:32:15  INFO  [LOLA] Engine started                │
│ ─────────────────────────────────────────────────────────────── │
│ 🟢 Chain ethereum   │ Connected │ 1.2 ms │ RPC: infura          │
│ 🟢 Chain polygon    │ Connected │ 2.1 ms │ RPC: alchemy         │
│ 🟢 Chain solana     │ Connected │ 4.5 ms │ RPC: helius          │
│ 🔵 Oracle chainlink │ Ready     │        │ Feed: ETH/USD        │
│ ─────────────────────────────────────────────────────────────── │
│ 📦 Tool 'analyze_price' invoked with args: {'token': 'ETH'}     │
│ 🔍 Detected oracle operation → get_price_from_oracle            │
│ ⏳ Fetching from Chainlink... (127.0.0.1:8545)                  │
│ ✅ Price received: $3,245.67                                    │
│ └────────────────────────────────────────────────────────────── │
```

**Why this matters:**  
- **Colour hints** – 🟢 green = success, 🔴 red = error, 🟡 yellow = warning, 🔵 blue = info  
- **Icons** – fast visual scanning  
- **Tables** – easy to read configuration status  
- **Progressive disclosure** – you can toggle verbosity with `LOLA_LOG_LEVEL`  

**You can also stream these logs as JSON** to your own dashboard, monitoring system (Datadog, Grafana), or even a Slack bot. We don’t lock you into our UI.

### 4.5 Human‑in‑the‑Loop (HITL) – Fully Flexible

LOLA OS does **not** force a specific approval UI. Instead, we emit **rich JSON logs** (same as above) that you can:

- Stream into your own dashboard (React, Vue, whatever)
- Forward to monitoring systems
- Show in a beautiful terminal prompt (built‑in, but you can replace it)
- Integrate with Slack, Discord, Telegram bots

**Example log output (JSON) with extra context:**

```json
{
  "level": "info",
  "event": "hitl_approval_required",
  "timestamp": "2026-06-10T14:32:20Z",
  "operation": "send_transaction",
  "from": "0x742d...",
  "to": "0x1234...",
  "amount": "1.5 ETH",
  "value_usd": 4868.50,
  "approval_url": "ws://localhost:8080/approve/abc123",
  "timeout_seconds": 60,
  "color": "yellow",
  "icon": "⚠️"
}
```

**You decide how to present this to your user.** We provide a dead‑simple terminal prompt out of the box, but you can override it completely.

### 4.6 Smart Contract Calling – Four Ways, Pick What’s Easiest

| Method | When to Use | Example |
|--------|-------------|---------|
| **ABI file** | You have the contract’s ABI locally | `call_contract(…, abi_path="./MyContract.abi")` |
| **Pre‑deployed type** | Standard contracts (ERC20, Uniswap, Aave) | `call_contract(…, contract_type="erc20")` |
| **Verified (auto‑fetch)** | Contract verified on Etherscan/Polygonscan | `call_contract(…, method="balanceOf(address)")` – no ABI needed! |
| **Raw calldata** | Expert mode, full control | `call_contract(…, calldata="0xa9059cbb…")` |

**LOLA auto‑detects** the best method. You don’t have to think about it.

#### 4.6.1 Pre‑flight ABI Validation (Safety Net)
Before simulating or broadcasting any transaction, LOLA performs a **pre‑flight validation**:
- It fetches the contract's ABI (from Etherscan, Sourcify, or local file).
- It compares the provided `method` name and `args` types against the ABI.
- If the method doesn’t exist, or the argument types don’t match (e.g., you passed a `string` but the method expects `uint256`), LOLA throws a clear `TypeError` **before** wasting gas on an RPC simulation.
- **Why this matters:** It catches typos and AI hallucinations early, saving you time and RPC credits.

### 4.7 Configuration – Simple by Default, Powerful When Needed

#### Zero‑Config Mode (for 90% of users)

Set **one** environment variable and you’re done:

```bash
export ETH_RPC_URL="https://mainnet.infura.io/v3/your-key"
```

LOLA automatically:
- Uses that RPC for all EVM chains
- Enables read‑only mode (safe)
- Stores nothing on disk (stateless)

#### Advanced Configuration (for power users)

Create `~/.lola/config.yaml` (auto‑detected) to define:

```yaml
# ~/.lola/config.yaml
version: "1.0"

chains:
  ethereum:
    rpc_url: "${ETH_RPC_URL}"   # references env var
    chain_id: 1
    enabled: true
    max_gas_price: "50 gwei"
  polygon:
    rpc_url: "${POLYGON_RPC_URL}"
    chain_id: 137
    enabled: true
  solana:
    rpc_url: "${SOLANA_RPC_URL}"
    enabled: true

oracles:
  chainlink:
    enabled: true
    endpoints:
      eth_usd: "0x5f4eC3Df9cbd43714FE2740f5E3616155c5b8419"
  pyth:
    enabled: true
    api_key: "${PYTH_API_KEY}"
  custom_apis:
    weather:
      url: "https://api.weather.gov"
      rate_limit: 10

wallets:
  # Never put private keys in YAML – use env vars with "encrypted_env" type
  ethereum_trading:
    type: "encrypted_env"
    env_var: "LOLA_ETH_KEY_ENCRYPTED"   # set via `lola encrypt-key`

budget:  # NEW - Session Budget & Circuit Breaker
  max_gas_spend_per_session: "0.5 ETH"
  max_usd_spend_per_session: 1000
  max_requests_per_minute: 60
  action: "pause"  # pause | notify | deny

hitl:
  default_timeout: 60
  default_action: "deny"
  console_ui: true    # use our beautiful terminal UI
  webhook_url: "${HITL_WEBHOOK_URL}"   # optional Slack/Discord

logging:
  level: "info"
  format: "rich"      # rich = colour + icons, json = structured
  output: "stdout"    # or file
```

**Priority order (higher wins):**
1. Command line flags (`lola --rpc-url ...`)
2. Environment variables (`.env` file or system env)
3. `config.yaml` values
4. Hardcoded defaults

**Secrets management:**  
- Private keys **never** go into YAML.  
- Use `lola encrypt-key --env MY_KEY` to store an encrypted key in your `.env`.  
- Or use hardware wallets (Ledger/Trezor) – no keys on disk at all.

#### 4.7.1 Context Overrides (Runtime Modes)
Sometimes you want to temporarily override the config without editing the file or restarting the engine.

- **Decorator override:**
  ```python
  @lola_tool(config_overrides={"chain": "polygon", "mode": "simulate_only"})
  def attack_vault(...):
      ...
  ```
- **Context manager override:**
  ```python
  with lola.override(rpc_url="backup.alchemy.io", budget_max_gas="0.1 ETH"):
      result = execute_contract(...)
  ```
- **Why this matters:** You can test risky operations on a local fork (`simulate_only`) or switch to a backup RPC if the primary is rate-limited, all without changing your global config.

### 4.8 Operational Tooling & Diagnostics (Maturity Layer)

LOLA OS ships with a suite of CLI commands to manage, monitor, and validate the bridge.

#### 4.8.1 `lola doctor` – Environment Health Check
Validates the entire LOLA environment in seconds:
- Checks RPC connectivity (ping & latest block number).
- Verifies Oracle endpoint reachability.
- Validates `config.yaml` syntax.
- Checks Vault integrity (decrypts test string).
- Detects Hardware Wallet connectivity (if configured).
- **Output:** Prints a beautiful table with ✅ (pass) or ❌ (fail) and actionable fix messages.

#### 4.8.2 `lola registry` – Transaction & State Registry
Tracks everything LOLA has executed.
- **`lola registry list`** – Shows recent operations (tx hash, chain, method, timestamp, status).
- **`lola registry show <tx_hash>`** – Fetches full details from the local SQLite DB.
- **`lola registry clear`** – Resets the local registry (keeps the DB file, just truncates).
- **Why this matters:** You can audit exactly what your AI agent did on-chain. Critical for debugging, compliance, and proving execution to clients.

#### 4.8.3 `lola replay <plan.json>` – Structured Execution
Ingests a JSON execution plan and runs it deterministically.
- **Input Schema:** See Appendix D.
- **Output:** Detailed execution receipt (success, balance deltas, tx hashes, gas used).
- **Flags:**
  - `--fork-url <RPC>` – Execute against a specific network or forked state.
  - `--dry-run` – Simulate the entire plan without broadcasting.
  - `--output receipt.json` – Save the execution receipt.
- **Why this matters:** This is the bridge between security scanners (like Hawk-i) or complex DeFi strategies and actual on-chain execution. It transforms theoretical analysis into verifiable action.

#### 4.8.4 `lola metrics` – Observability
Exposes real-time operational metrics (JSON lines or Prometheus format).
- **Tracks:** Total requests, average latency per chain, rate-limit hits, estimated gas spent in the session, vault unlock attempts, successful vs. failed transactions.
- **Example:** `lola metrics --format prometheus` pipes data directly to Grafana or Datadog.
- **Why this matters:** You can't manage what you don't measure. This is essential for enterprise deployments and cost optimization.

---

## 5. Security – Built for Real Money, Not Just Demos

We know blockchain involves real value. Security is not an afterthought.

### What’s Inside the Box

- **Encrypted vault** – private keys stored with AES‑256‑GCM, key derived via scrypt (memory‑hard)
- **Never store plaintext keys** – even in memory, we zero them after use
- **Transaction simulation** – every write is simulated locally before broadcast
- **Gas price protection** – you set max gas; we enforce it
- **Rate limiting** – prevent accidental spam or abuse (for RPCs and external APIs)
- **Circuit breakers** – if an RPC or oracle fails, we fail fast and try alternatives
- **Audit logs** – every operation is logged (rich or JSON) – you decide retention
- **Session Budget Enforcement** – a background thread monitors gas spend and USD cost (via oracle conversion) against `budget` limits. If exceeded, LOLA automatically pauses all write operations and emits a CRITICAL log.
- **Idempotency & Nonce Management** – LOLA maintains an in-memory (and persistent) nonce counter per chain/wallet to prevent double-spends. Additionally, users can pass an `idempotency_key` to any write operation. LOLA caches the result for 24 hours; if the same key is used again, it returns the cached transaction hash instead of broadcasting a duplicate.
- **Pre-flight ABI Validation** – catches mismatched function signatures and argument types before gas is consumed.

### Hardware Wallet Support

- Ledger / Trezor – sign transactions without exposing private keys
- Works via USB or WebUSB (for browser)

### What You Control (Because It’s Local)

- **All RPC endpoints** – use Infura, Alchemy, your own node, or public ones. We don’t care.
- **All oracle endpoints** – use public or private data feeds.
- **Approval rules** – set thresholds: “ask me for any transfer above 0.1 ETH”
- **Logging** – pipe logs anywhere, or disable them entirely.
- **Spending limits** – define hard caps for gas and USD per session.

---

## 6. Writing Custom Tools – Your Functions Become Blockchain/Native Tools

One of the most powerful features: **you can write normal Python/Go/TS functions, decorate them with `@lola_tool`, and they instantly become LOLA‑aware tools** that can call any blockchain operation, oracle, or even other tools.

### Example: A Custom Risk Assessment Tool

```python
from lola import lola_tool, get_balance, get_price_from_oracle

@lola_tool(name="risk_assessment")   # custom name
def assess_risk(address: str, min_balance_eth: float = 0.1):
    """
    This function is now a LOLA tool.
    It can be called directly or used by AI agents.
    """
    balance = get_balance(address)["ethereum"]
    eth_price = get_price_from_oracle("ETH/USD")
    value_usd = balance * eth_price
    
    if value_usd < min_balance_eth * eth_price:
        return {"risk": "high", "reason": f"Balance only ${value_usd:.2f}"}
    return {"risk": "low", "balance_usd": value_usd}
```

Now you can:
- Call `assess_risk("0x...")` directly in your code – it just works.
- Use it as a tool in CrewAI/LangChain – the framework sees it as a native tool.
- Combine it with other LOLA tools – they can call each other.

**No extra boilerplate. No registration. Just decorate and go.**

---

## 7. Integration Examples (Real Code, Real Simple)

### Example 1: CrewAI Agent That Monitors Wallet Health + Oracle Price

```python
from crewai import Agent, Task
from lola import lola_tool, get_balance, get_price_from_oracle

@lola_tool
def check_health(address: str):
    balances = get_balance(address)
    eth_price = get_price_from_oracle("ETH/USD")
    return {
        "address": address,
        "balances": balances,
        "eth_usd": eth_price
    }

agent = Agent(role="Wallet Monitor", tools=[check_health])
task = Task(description="Check 0x742... health")
crew = Crew(agents=[agent], tasks=[task])
result = crew.kickoff()
```

### Example 2: LangChain Arbitrage Bot With HITL Logs

```typescript
import { lolaTool, getPriceFromOracle, executeSwap } from "lola-os";

@lolaTool({ requireApproval: true })
async function arbitrage(pair: string, amount: number) {
  const priceEth = await getPriceFromOracle(pair, "ethereum", "chainlink");
  const pricePoly = await getPriceFromOracle(pair, "polygon", "pyth");
  
  if (pricePoly < priceEth * 0.98) {
    // LOLA will emit rich HITL logs – you show your own UI
    return await executeSwap(pair, amount, "polygon", "ethereum");
  }
  return "No opportunity";
}
```

### Example 3: Custom Trading Bot With Your Own Approval UI

```python
import asyncio
from lola import lola_tool, stream_logs

@lola_tool(require_approval=True)
async def trade(token, amount):
    result = await execute_swap(token, amount)
    return result

# Stream logs to your own React dashboard
async for log in stream_logs(format="json"):
    if log["event"] == "hitl_required":
        send_to_websocket(log)   # your UI listens
```

---

## 8. What LOLA OS Is NOT (Clarity Matters)

- **Not a blockchain** – we don’t have our own chain. We connect to existing ones.
- **Not an AI framework** – we don’t replace CrewAI, LangChain, etc. We enhance them.
- **Not a hosted service** – we don’t run a backend. You run the binary.
- **Not a subscription** – free forever. No “pro” tier. No feature gating.
- **Not a wallet** – we manage keys, but you stay in control.
- **Not an oracle provider** – we connect to existing oracles (Chainlink, Pyth, etc.) and any REST API.

**We are the bridge.** That’s it. And we’re damn good at it.

---

## 9. For the Business People (Monetization – Or Lack Thereof)

### Our Commitment to You

> **LOLA OS core will always be 100% open source (Apache 2.0), free to use, with no artificial limits. We will never add a paywall to read‑, write‑, or oracle operations, never require an API key, never limit the number of requests, and never hide features behind a “Pro” tier.**

### How We Stay Alive (Without Selling Out)

We offer **optional enterprise support** for companies that need:

- **SLAs** – guaranteed response times
- **Training** – teach your team LOLA in a day
- **Custom SDKs** – we’ll build the Cosmos / Polkadot / etc. engine for you
- **On‑premise deployment** – air‑gapped environments
- **Compliance packages** – SOC2, GDPR, audit trails

But the software itself? **Free.** No tricks. No “open core” nonsense.

### Why This Works (The Honest Business Model)

Great infrastructure makes money from **support, not lock‑in**. HashiCorp, Red Hat, MongoDB (before they pivoted) – they proved it. When you become indispensable, enterprises pay for the handshake, not the software.

**We will be indispensable to every AI agent builder.** Then we’ll help the biggest ones with premium support. Everyone else enjoys the free bridge forever.

---

## 10. Getting Started (The 5‑Minute Challenge)

We dare you: from zero to on‑chain data + oracle price in 5 minutes.

### Step 1 – Install

```bash
pip install lola-os
# or
npm install lola-os
# or
go get github.com/lola-os/lola-go
```

### Step 2 – Set one environment variable (optional – works without it too)

```bash
export ETH_RPC_URL="https://mainnet.infura.io/v3/your-key"
```

(If you don’t set any, LOLA uses public fallback RPCs – limited but fine for testing.)

### Step 3 – Write code

```python
from lola import lola_tool, get_balance, get_price_from_oracle

@lola_tool
def my_agent(address):
    balance = get_balance(address)
    price = get_price_from_oracle("ETH/USD")
    return {"balance": balance, "eth_usd": price}

print(my_agent("0x742d35Cc6634C0532925a3b844Bc9e90F1A6B1E7"))
```

**You’re done.** Blockchain + oracles integrated.

---

## 11. FAQ (Frequently Asked Questions)

### General

**Q1: Is LOLA OS really free forever?**  
Yes. The core engine and all SDKs are Apache 2.0. No “open core” tricks. We only sell optional support and custom development.

**Q2: Do I need to run a server or pay for hosting?**  
No. LOLA runs entirely on your machine (or your agent’s machine). No backend, no API keys, no monthly fees.

**Q3: What programming languages are supported?**  
Python, Go, and TypeScript/JavaScript in v1.0. More (Rust, Java, C#) will come via community contributions.

**Q4: Can I use LOLA with my existing AI framework?**  
Absolutely. CrewAI, LangChain, AutoGen, custom code – if it runs Python/Go/TS, it works.

**Q5: Does LOLA work with private blockchains (Hyperledger, etc.)?**  
Not yet. v1.0 focuses on public EVM + Solana. But because the engine is modular, adding custom chains is straightforward.

### Configuration & Setup

**Q6: Do I have to create a config.yaml?**  
No. For 90% of use cases, just set `ETH_RPC_URL` or use the public fallbacks. `config.yaml` is only for advanced setups (multiple chains, oracles, custom wallets).

**Q7: How do I add my private key securely?**  
Use `lola encrypt-key` to store an encrypted version in your `.env` file, or use a hardware wallet (Ledger/Trezor). Never put plaintext keys in YAML.

**Q8: Can I use multiple RPC providers for the same chain?**  
Yes. In `config.yaml`, list multiple `rpc_urls` – LOLA will round‑robin or fallback if one fails.

**Q9: What oracles are supported out of the box?**  
Chainlink (price feeds), Pyth, API3, and any generic REST/WebSocket API. We plan to add more based on community demand.

**Q10: How do I add a custom REST API as an oracle?**  
Just call `fetch_external_api()` in your tool. For advanced caching/rate limiting, define it in `config.yaml` under `oracles.custom_apis`.

### Operations & Performance

**Q11: How fast is LOLA?**  
Balance checks: ~200ms. Contract reads: ~300ms. Oracle calls: depends on the source (typically 100‑500ms). All operations are parallelised across chains.

**Q12: Can LOLA handle thousands of requests per second?**  
Yes. The Go core is highly concurrent. You can also run multiple instances. Rate limiting is built‑in to protect you and the RPC providers.

**Q13: Does LOLA cache results?**  
Yes. Read operations are cached with configurable TTL (default 5 minutes). You can disable or adjust in `config.yaml`.

**Q14: What happens if an RPC or oracle goes down?**  
LOLA has circuit breakers and automatic fallback to secondary RPCs (if configured). You’ll see a yellow warning in the logs.

### Security & HITL

**Q15: Is it safe to use LOLA with real funds?**  
Yes, if you follow best practices: use hardware wallets for large amounts, enable transaction simulation, set HITL thresholds, and keep your `.env` file secure.

**Q16: Can I disable HITL entirely?**  
Yes. Set `require_approval: false` in your tool decorator or globally in `config.yaml`. Not recommended for production with real funds.

**Q17: How do I see what LOLA is doing?**  
The rich console logs show every step. Set `LOLA_LOG_LEVEL=debug` for even more detail.

**Q18: Can I send logs to my own monitoring system?**  
Absolutely. Use `format: json` in `config.yaml` and pipe stdout to your collector (e.g., `lola-core rpc | your-log-consumer`).

### New Maturity Features (Added)

**Q19: What does `lola doctor` do?**  
It runs a comprehensive health check on your LOLA environment – verifying RPC connectivity, oracle reachability, configuration syntax, and wallet status. It prints a clear pass/fail table.

**Q20: How does the session budget work?**  
Set `budget.max_gas_spend_per_session` or `budget.max_usd_spend_per_session` in `config.yaml`. A background thread monitors spending. If the limit is hit, LOLA pauses all write operations to prevent runaway costs.

**Q21: What is the `lola replay` command?**  
It ingests a structured JSON execution plan and executes it step-by-step. This allows you to take a complex strategy (or an exploit plan generated by a security tool) and replay it deterministically on a fork or mainnet.

**Q22: Where are transaction receipts stored?**  
In `~/.lola/lola.db` (SQLite). You can query this with `lola registry list` or `lola registry show <tx_hash>`.

**Q23: How does idempotency work?**  
Pass an `idempotency_key` to any write operation. LOLA caches the result for 24 hours. If you call the same operation with the same key, it returns the cached result without broadcasting a duplicate transaction, preventing double-spends.

### Troubleshooting

**Q24: LOLA says “RPC not found” – what do I do?**  
Set `ETH_RPC_URL` (or `POLYGON_RPC_URL`, etc.) to a working endpoint. Run `lola doctor` to verify connectivity.

**Q25: My transaction failed with “insufficient funds” – but I have enough.**  
Check that you’re using the correct chain and that you have enough native currency for gas (e.g., ETH for Ethereum, MATIC for Polygon). LOLA’s logs will show the exact gas estimate.

**Q26: Can I run LOLA in a Docker container?**  
Yes. We provide official Docker images. The Go binary runs standalone, and you can mount your `~/.lola` directory for config.

**Q27: How do I update to the latest version?**  
`pip install --upgrade lola-os` (or `npm update`, `go get -u`). The Go binary auto‑updates itself when the SDK detects a newer version (opt‑out available).

**Q28: What’s the difference between LOLA and Web3.py / ethers.js?**  
Those are low‑level libraries for blockchain developers. LOLA is a **high‑level bridge for AI developers** – you write business logic, we handle the complexity.

**Q29: Can I contribute to LOLA?**  
Yes! GitHub issues, PRs, and discussions are welcome. We’re especially looking for help with new chain engines, SDKs for other languages, and documentation.

---

## 12. Call to Action

**If you’re an AI developer:**  
Install LOLA OS today. Add `@lola_tool` to one function. See the magic. Then tell your friends.

**If you’re a blockchain ecosystem (Ethereum, Solana, etc.):**  
We already speak your language. Want more AI agents on your chain? Help us spread the word – or fund a grant to build deeper integration.

**If you’re an open source contributor:**  
We need help with SDKs, documentation, and new chain engines. The repo is at `github.com/lola-os/lola-os`. First PR gets a sticker.

**If you’re an enterprise:**  
You’ll love LOLA. When you need a support contract, email `enterprise@lola.dev`. Otherwise, enjoy the free ride.

---

## Appendix A: Complete Configuration Examples

### A.1 Minimal `.env` for a Single Chain

```bash
# .env
ETH_RPC_URL=https://mainnet.infura.io/v3/your-key
LOLA_LOG_LEVEL=info
```

### A.2 Advanced `config.yaml` with Multiple Chains, Oracles, HITL, and Budget

```yaml
# ~/.lola/config.yaml
version: "1.0"

chains:
  ethereum:
    rpc_urls:
      - "${ETH_MAIN_RPC}"
      - "${ETH_BACKUP_RPC}"
    chain_id: 1
    enabled: true
    max_gas_price: "50 gwei"
  polygon:
    rpc_url: "${POLYGON_RPC_URL}"
    chain_id: 137
    enabled: true
  solana:
    rpc_url: "${SOLANA_RPC_URL}"
    enabled: true

oracles:
  chainlink:
    enabled: true
    feeds:
      ETH/USD: "0x5f4eC3Df9cbd43714FE2740f5E3616155c5b8419"
      BTC/USD: "0xF4030086522a5bEEa4988F8cA5B36dbC97BeE88c"
  pyth:
    enabled: true
    api_key: "${PYTH_API_KEY}"
  custom_apis:
    weather:
      url: "https://api.openweathermap.org/data/2.5/weather"
      api_key_env: "WEATHER_API_KEY"
      rate_limit: 5

wallets:
  trading:
    type: "encrypted_env"
    env_var: "LOLA_TRADING_KEY_ENCRYPTED"
  cold_storage:
    type: "hardware"
    device: "ledger"
    path: "m/44'/60'/0'/0/0"

budget:
  max_gas_spend_per_session: "0.5 ETH"
  max_usd_spend_per_session: 1000
  max_requests_per_minute: 60
  action: "pause"

hitl:
  default_timeout: 120
  default_action: "deny"
  console_ui: true
  webhook_url: "${SLACK_WEBHOOK}"

logging:
  level: "debug"
  format: "rich"   # or "json"
  output: "stdout"
```

---

## Appendix B: Log Format Specification (for Custom UIs)

LOLA’s logs are emitted as **JSON Lines** (one JSON object per line) when `format: json` is set. Each log contains:

| Field | Type | Description |
|-------|------|-------------|
| `timestamp` | string | ISO 8601 |
| `level` | string | `debug`, `info`, `warn`, `error` |
| `event` | string | e.g., `chain_connected`, `hitl_approval_required`, `transaction_sent`, `budget_exceeded` |
| `color` | string | `green`, `red`, `yellow`, `blue`, `gray` (for rich UI) |
| `icon` | string | Unicode emoji (🔵, 🟢, ⚠️, ❌, ✅) |
| `message` | string | Human‑readable summary |
| `data` | object | Extra structured data (tx hash, balance, gas used, etc.) |
| `tool_name` | string | Name of the tool that triggered the log (if applicable) |

**Example:**

```json
{"timestamp":"2026-06-10T14:32:20Z","level":"info","event":"hitl_approval_required","color":"yellow","icon":"⚠️","message":"Approval needed: send 1.5 ETH","data":{"from":"0x742d...","to":"0x1234...","amount":"1.5","value_usd":4868.50},"tool_name":"transfer_funds"}
```

You can consume this stream and render it in any UI.

---

## Appendix C: Troubleshooting Common Issues

| Symptom | Likely Cause | Solution |
|---------|--------------|----------|
| `RPC not found` | No RPC URL set | Set `ETH_RPC_URL` or other chain env var. Run `lola doctor`. |
| `Transaction simulation failed: out of gas` | Gas limit too low | LOLA auto‑estimates, but you can increase via `options={"gas_limit": 500000}` |
| `HITL timeout` | User didn’t respond in time | Increase `hitl.default_timeout` in config or handle timeout in your UI. |
| `Oracle returned stale data` | Oracle feed not updated | Check the oracle’s health. For Chainlink, use `latestRoundData()` and check `updatedAt`. |
| `Cannot find contract ABI` | Contract not verified | Use `abi_path` with a local ABI file, or use the `contract_type` parameter if standard. |
| `Go binary not found` | Installation incomplete | Reinstall the SDK. On Linux, you may need `chmod +x ~/.lola/bin/lola-core`. |
| `Private key decryption failed` | Wrong password or corrupted vault | Run `lola vault reset` (warning: deletes all keys). Re‑encrypt your keys. |
| `Budget exceeded` | You hit `max_gas_spend_per_session` | Either increase the budget in config, or wait for the session to reset. |

---

## Appendix D: Structured Execution Plan Schema (for `lola replay`)

The `lola replay <plan.json>` command accepts a JSON array of operations. This is the bridge for external tools to drive LOLA deterministically.

**Schema:**

```json
{
  "version": "1.0",
  "description": "Optional description of the plan",
  "chain": "ethereum",
  "operations": [
    {
      "type": "call_contract",
      "address": "0x...",
      "method": "balanceOf(address)",
      "args": ["0x..."],
      "output_key": "initial_balance"
    },
    {
      "type": "send_transaction",
      "to": "0x...",
      "amount": "1.5",
      "unit": "eth"
    },
    {
      "type": "execute_contract",
      "address": "0x...",
      "method": "withdraw()",
      "args": []
    },
    {
      "type": "call_contract",
      "address": "0x...",
      "method": "balanceOf(address)",
      "args": ["0x..."],
      "output_key": "final_balance"
    },
    {
      "type": "assert",
      "left": "${final_balance}",
      "operator": ">",
      "right": "${initial_balance}"
    }
  ]
}
```

**Supported Operation Types:**
- `call_contract` – Read-only view calls.
- `send_transaction` – Native currency transfers.
- `transfer_token` – ERC20/SPL token transfers.
- `execute_contract` – State-changing contract calls.
- `swap_tokens` – DEX swaps.
- `assert` – Logical checks on previous outputs (useful for testing).
- `wait` – Sleep for N seconds (for time-sensitive transactions).

**Variable Interpolation:**
Use `${key}` syntax to reference outputs from previous operations.

**Why this matters:** It allows any external system (a security scanner, a trading bot orchestrator, an audit framework) to generate a standard plan that LOLA can execute blindly and safely.

---

## Appendix E: Ecosystem Integrations (Optional)

While LOLA OS is a standalone bridge, it is designed to be the ideal execution engine for security tools and AI frameworks.

**Security Scanners:** A scanner can export a `plan.json` containing the exact sequence of transactions to reproduce a vulnerability. `lola replay` then executes it, proving the vulnerability exists.

**AI Agents:** Frameworks like CrewAI or LangChain can call `@lola_tool` functions directly. The `lola registry` provides a perfect audit trail of every decision the agent made on-chain.

**Observability:** The `lola metrics` endpoint integrates seamlessly with Prometheus, Datadog, or any custom dashboard, providing real-time visibility into the bridge's health and performance.

---

## Conclusion

**LOLA OS is not just another SDK.** It’s a declaration that AI developers should never have to become blockchain experts to build powerful, on‑chain agents.

We’ve built a bridge that is:

- **Simple** – one decorator, five minutes, zero config.
- **Powerful** – read/write any chain, call any oracle, talk to any API.
- **Operationally Mature** – diagnostics (`doctor`), persistence (`registry`), budget enforcement, and structured replay (`replay`).
- **Secure** – encrypted vault, hardware wallets, transaction simulation, HITL, idempotency, and nonce management.
- **Free** – truly open source, no paywalls, no subscriptions.
- **Beautiful** – logs you’ll actually enjoy reading.

The era of isolated AI agents is over. The era of agents that can pay, verify, trade, and act on real‑world data is here.

**LOLA OS – Blockchain for Every AI Agent.**  
No lock‑in. No subscription. Just a bridge.

*Now go build something amazing.*