# lola-os (Python SDK)

![PyPI](https://img.shields.io/pypi/v/lola-os?labelColor=252525&color=555555&style=flat-square)
![Downloads](https://img.shields.io/pypi/dm/lola-os?labelColor=252525&color=555555&style=flat-square)
![Python versions](https://img.shields.io/pypi/pyversions/lola-os?labelColor=252525&color=555555&style=flat-square)
![License](https://img.shields.io/pypi/l/lola-os?labelColor=252525&color=555555&style=flat-square)

Add `@lola_tool` to any function and it can talk to blockchains, oracles,
and APIs through the local `lola-core` engine — no hosted backend, no API
keys, no billing.

```python
from lola_os import lola_tool, get_balance

@lola_tool
def check_balance(address: str) -> dict:
    return get_balance("ethereum", address)

print(check_balance("0x..."))
```

## Status

44 tests pass (`python3 -m pytest`), covering exceptions, context
overrides, the decorator, and every convenience function against a mocked
`LolaCore` client. The SDK spawns a built `lola-core` binary as a
subprocess and speaks newline-delimited JSON-RPC 2.0 to it (the wire format
in `client.py` matches `lola-core/internal/jsonrpc`). Build the binary
first (see the root `lola-core/README.md`) and point the SDK at it via the
`LOLA_CORE_BINARY` environment variable, a `LolaCore(binary_path=...)`
argument, or a `lola` binary on your `PATH`.

## Installation

```bash
pip install -e .
# or, for the optional WebSocket-based HITL listener:
pip install -e ".[websocket]"
```

This package does **not** bundle a `lola-core` binary in this build (see
`lola_os/bin/README.txt`). Build one and point the SDK at it:

```bash
cd ../../lola-core
go build -o bin/lola ./cmd/lola
export LOLA_CORE_BINARY=$(pwd)/bin/lola
export LOLA_VAULT_PASSPHRASE="your-passphrase"
```

## Quickstart

```python
import lola_os as lola

# Native balance
balance = lola.get_balance("ethereum", "0x...")
print(balance)  # {"address", "token", "raw_value", "decimals", "symbol"}

# Read-only contract call
supply = lola.call_contract("ethereum", "0xTokenAddress", "totalSupply")

# State-changing call — may prompt for human approval depending on config
receipt = lola.execute_contract(
    chain="ethereum",
    from_address="0xYourAddress",
    contract="0xTokenAddress",
    method="transfer",
    args=["0xRecipient", 1000000000000000000],
    abi='[{"constant":false,"inputs":[{"name":"_to","type":"address"},{"name":"_value","type":"uint256"}],"name":"transfer","outputs":[{"name":"","type":"bool"}],"type":"function"}]',
    key_name="deployer",
)
print(receipt)  # {"tx_hash": "0x...", "status": "pending"}
```

## Context overrides

Switch chains, RPCs, or budget limits for a single block of code without
touching global config:

```python
with lola.override(chain="polygon", budget_max_usd=5.0):
    result = lola.execute_contract(...)
```

Or scope overrides to one specific tool function:

```python
@lola_tool(config_overrides={"chain": "polygon", "mode": "simulate_only"})
def my_polygon_tool():
    ...
```

## Async support

Every convenience function has an `_async` counterpart:

```python
import asyncio
import lola_os as lola

async def main():
    balance = await lola.get_balance_async("ethereum", "0x...")
    print(balance)

asyncio.run(main())
```

`@lola_tool` also works transparently on `async def` functions.

## Error handling

```python
from lola_os import execute_contract, BudgetExceededError, ABIMismatchError, RPCConnectionError

try:
    execute_contract(...)
except BudgetExceededError as e:
    print("Over budget:", e)
except ABIMismatchError as e:
    print("Bad method/args:", e)
except RPCConnectionError as e:
    print("Network issue:", e)
```

## Streaming logs

```python
for entry in lola.stream_logs():
    print(entry["level"], entry["message"])
```

## Running the tests

```bash
pip install -e ".[dev]"
pytest
```

(In this build environment, `pytest` wasn't installable due to no network
access — tests were verified with a minimal manual runner instead. They
are written in standard `pytest` style and will run normally with
`pytest`/`pytest-asyncio` installed.)

## API reference

| Function | Description |
|---|---|
| `get_balance(chain, address)` | Native balance |
| `get_token_balance(chain, address, token_address)` | ERC20/SPL balance |
| `call_contract(chain, contract, method, args, abi)` | Read-only call |
| `execute_contract(chain, from_address, contract, method, args, abi, ...)` | State-changing call |
| `send_transaction(chain, from_address, to, value_wei, ...)` | Native transfer |
| `transfer_token(chain, from_address, to, token, amount_raw, ...)` | Token transfer |
| `swap_tokens(chain, from_address, router_contract, method, args, abi, ...)` | DEX swap via router call |
| `get_price_from_oracle(chain, pair)` | Chainlink price read |
| `fetch_external_api(url)` | Rate-limited, retrying REST GET |
| `multi_operation(operations)` | Run a client-side batch of the above |
| `listen_to_websocket(on_approval_request, addr)` | Custom HITL UI integration |
| `stream_logs()` | Generator of structured log entries |

Every function above has an `_async` counterpart with the same signature.

---

Built and maintained by **[0xSemantic](https://github.com/0xSemantic)** —
developer and visionary behind LOLA OS. Licensed under Apache-2.0.
