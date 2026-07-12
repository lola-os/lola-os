"""Convenience functions wrapping LolaCore.call() for every operation LOLA
OS supports. Each has a sync version and an `_async` counterpart.

All functions honor the current context overrides (set via
`lola.override(...)`) by attaching them to the RPC params under
`_context_overrides`; lola-core applies them for the duration of that
single call.
"""

from __future__ import annotations

import asyncio
import json
import time
from typing import Any, AsyncIterator, Callable, Dict, Iterator, List, Optional

from .client import LolaCore, get_client
from .context import get_current_overrides


def _params_with_overrides(params: Dict[str, Any]) -> Dict[str, Any]:
    overrides = get_current_overrides().to_rpc_params()
    if overrides:
        params = {**params, "_context_overrides": overrides}
    return params


def _client(client: Optional[LolaCore]) -> LolaCore:
    return client or get_client()


# -- balances ---------------------------------------------------------------

def get_balance(chain: str, address: str, *, client: Optional[LolaCore] = None) -> dict:
    """Returns the native asset balance for `address` on `chain`.

    Returns a dict: {"address", "token", "raw_value", "decimals", "symbol"}.
    `raw_value` is a decimal string in the smallest unit (wei, lamports);
    convert using `decimals` for a human-readable amount.
    """
    c = _client(client)
    return c.call("get_balance", _params_with_overrides({"chain": chain, "address": address}))


async def get_balance_async(chain: str, address: str, *, client: Optional[LolaCore] = None) -> dict:
    c = _client(client)
    return await c.call_async("get_balance", _params_with_overrides({"chain": chain, "address": address}))


def get_token_balance(chain: str, address: str, token_address: str, *, client: Optional[LolaCore] = None) -> dict:
    """Returns an ERC20/SPL token balance for `address`."""
    c = _client(client)
    return c.call(
        "get_token_balance",
        _params_with_overrides({"chain": chain, "address": address, "token_address": token_address}),
    )


async def get_token_balance_async(chain: str, address: str, token_address: str, *, client: Optional[LolaCore] = None) -> dict:
    c = _client(client)
    return await c.call_async(
        "get_token_balance",
        _params_with_overrides({"chain": chain, "address": address, "token_address": token_address}),
    )


# -- contract calls ---------------------------------------------------------

def call_contract(
    chain: str,
    contract: str,
    method: str,
    args: Optional[List[Any]] = None,
    abi: Optional[str] = None,
    *,
    client: Optional[LolaCore] = None,
) -> Any:
    """Performs a read-only contract call. If `abi` is omitted, lola-core
    attempts to fetch a verified ABI from the configured block explorer.

    Raises ABIMismatchError if `method`/`args` don't match the ABI.
    """
    c = _client(client)
    result = c.call(
        "call_contract",
        _params_with_overrides({"chain": chain, "contract": contract, "method": method, "args": args or [], "abi": abi or ""}),
    )
    return result.get("result") if isinstance(result, dict) else result


async def call_contract_async(
    chain: str, contract: str, method: str, args: Optional[List[Any]] = None, abi: Optional[str] = None,
    *, client: Optional[LolaCore] = None,
) -> Any:
    c = _client(client)
    result = await c.call_async(
        "call_contract",
        _params_with_overrides({"chain": chain, "contract": contract, "method": method, "args": args or [], "abi": abi or ""}),
    )
    return result.get("result") if isinstance(result, dict) else result


def execute_contract(
    chain: str,
    from_address: str,
    contract: str,
    method: str,
    args: Optional[List[Any]] = None,
    abi: Optional[str] = None,
    value_wei: str = "0",
    idempotency_key: str = "",
    key_name: str = "default",
    *,
    client: Optional[LolaCore] = None,
) -> dict:
    """Builds, signs, and broadcasts a state-changing contract call. May
    trigger a human-in-the-loop approval prompt depending on config.

    Returns {"tx_hash": str, "status": str}.
    """
    c = _client(client)
    return c.call(
        "execute_contract",
        _params_with_overrides(
            {
                "chain": chain, "from": from_address, "contract": contract, "method": method,
                "args": args or [], "abi": abi or "", "value_wei": value_wei,
                "idempotency_key": idempotency_key, "key_name": key_name,
            }
        ),
    )


async def execute_contract_async(
    chain: str, from_address: str, contract: str, method: str, args: Optional[List[Any]] = None,
    abi: Optional[str] = None, value_wei: str = "0", idempotency_key: str = "", key_name: str = "default",
    *, client: Optional[LolaCore] = None,
) -> dict:
    c = _client(client)
    return await c.call_async(
        "execute_contract",
        _params_with_overrides(
            {
                "chain": chain, "from": from_address, "contract": contract, "method": method,
                "args": args or [], "abi": abi or "", "value_wei": value_wei,
                "idempotency_key": idempotency_key, "key_name": key_name,
            }
        ),
    )


# -- transactions ---------------------------------------------------------

def send_transaction(
    chain: str,
    from_address: str,
    to: str,
    value_wei: str,
    idempotency_key: str = "",
    key_name: str = "default",
    *,
    client: Optional[LolaCore] = None,
) -> dict:
    """Sends a native asset transfer. Returns {"tx_hash": str, "status": str}."""
    c = _client(client)
    return c.call(
        "send_transaction",
        _params_with_overrides(
            {"chain": chain, "from": from_address, "to": to, "value_wei": value_wei,
             "idempotency_key": idempotency_key, "key_name": key_name}
        ),
    )


async def send_transaction_async(
    chain: str, from_address: str, to: str, value_wei: str, idempotency_key: str = "", key_name: str = "default",
    *, client: Optional[LolaCore] = None,
) -> dict:
    c = _client(client)
    return await c.call_async(
        "send_transaction",
        _params_with_overrides(
            {"chain": chain, "from": from_address, "to": to, "value_wei": value_wei,
             "idempotency_key": idempotency_key, "key_name": key_name}
        ),
    )


def transfer_token(
    chain: str,
    from_address: str,
    to: str,
    token: str,
    amount_raw: str,
    idempotency_key: str = "",
    key_name: str = "default",
    *,
    client: Optional[LolaCore] = None,
) -> dict:
    """Transfers an ERC20/SPL token. `amount_raw` is in the token's
    smallest unit (i.e. already multiplied by 10**decimals)."""
    c = _client(client)
    return c.call(
        "transfer_token",
        _params_with_overrides(
            {"chain": chain, "from": from_address, "to": to, "token": token, "amount_raw": amount_raw,
             "idempotency_key": idempotency_key, "key_name": key_name}
        ),
    )


async def transfer_token_async(
    chain: str, from_address: str, to: str, token: str, amount_raw: str,
    idempotency_key: str = "", key_name: str = "default", *, client: Optional[LolaCore] = None,
) -> dict:
    c = _client(client)
    return await c.call_async(
        "transfer_token",
        _params_with_overrides(
            {"chain": chain, "from": from_address, "to": to, "token": token, "amount_raw": amount_raw,
             "idempotency_key": idempotency_key, "key_name": key_name}
        ),
    )


def swap_tokens(
    chain: str,
    from_address: str,
    router_contract: str,
    method: str,
    args: List[Any],
    abi: str,
    value_wei: str = "0",
    idempotency_key: str = "",
    key_name: str = "default",
    *,
    client: Optional[LolaCore] = None,
) -> dict:
    """Executes a DEX swap by calling a router contract method directly
    (e.g. Uniswap's `swapExactTokensForTokens`). LOLA OS does not hardcode
    any specific router's ABI — pass the router's method/args/ABI
    explicitly, same as `execute_contract`."""
    return execute_contract(
        chain, from_address, router_contract, method, args, abi, value_wei, idempotency_key, key_name, client=client
    )


async def swap_tokens_async(
    chain: str, from_address: str, router_contract: str, method: str, args: List[Any], abi: str,
    value_wei: str = "0", idempotency_key: str = "", key_name: str = "default", *, client: Optional[LolaCore] = None,
) -> dict:
    return await execute_contract_async(
        chain, from_address, router_contract, method, args, abi, value_wei, idempotency_key, key_name, client=client
    )


# -- oracles / external APIs ---------------------------------------------------------

def get_price_from_oracle(chain: str, pair: str, *, client: Optional[LolaCore] = None) -> dict:
    """Reads the latest Chainlink price for `pair` (e.g. "ETH/USD")."""
    c = _client(client)
    return c.call("get_price", _params_with_overrides({"chain": chain, "pair": pair}))


async def get_price_from_oracle_async(chain: str, pair: str, *, client: Optional[LolaCore] = None) -> dict:
    c = _client(client)
    return await c.call_async("get_price", _params_with_overrides({"chain": chain, "pair": pair}))


def fetch_external_api(url: str, *, client: Optional[LolaCore] = None) -> Any:
    """Fetches and JSON-decodes `url` through lola-core's rate-limited,
    retrying REST gateway."""
    c = _client(client)
    return c.call("fetch_external_api", _params_with_overrides({"url": url}))


async def fetch_external_api_async(url: str, *, client: Optional[LolaCore] = None) -> Any:
    c = _client(client)
    return await c.call_async("fetch_external_api", _params_with_overrides({"url": url}))


# -- batch / multi-operation ---------------------------------------------------------

def multi_operation(operations: List[Dict[str, Any]], *, client: Optional[LolaCore] = None, stop_on_error: bool = True) -> List[dict]:
    """Runs a sequence of operations client-side, in order, returning each
    result. Each operation dict must have a "type" key matching one of:
    "get_balance", "call_contract", "execute_contract", "send_transaction",
    "transfer_token", "get_price", "fetch_external_api" — plus that
    function's other keyword arguments.

    For operations that need on-chain ordering guarantees and replayable
    audit trails, prefer `lola replay <plan.json>` (see the CLI and the
    "Replay" docs page) over multi_operation, which is intended for
    lighter-weight, code-driven batches inside an agent's own logic.
    """
    dispatch = {
        "get_balance": get_balance,
        "call_contract": call_contract,
        "execute_contract": execute_contract,
        "send_transaction": send_transaction,
        "transfer_token": transfer_token,
        "swap_tokens": swap_tokens,
        "get_price": get_price_from_oracle,
        "fetch_external_api": fetch_external_api,
    }
    results = []
    for op in operations:
        op = dict(op)
        op_type = op.pop("type", None)
        fn = dispatch.get(op_type)
        if fn is None:
            results.append({"error": f"unknown operation type: {op_type}"})
            if stop_on_error:
                break
            continue
        try:
            results.append({"result": fn(**op, client=client)})
        except Exception as exc:  # noqa: BLE001 - we deliberately capture all SDK errors here
            results.append({"error": str(exc)})
            if stop_on_error:
                break
    return results


async def multi_operation_async(operations: List[Dict[str, Any]], *, client: Optional[LolaCore] = None, stop_on_error: bool = True) -> List[dict]:
    dispatch = {
        "get_balance": get_balance_async,
        "call_contract": call_contract_async,
        "execute_contract": execute_contract_async,
        "send_transaction": send_transaction_async,
        "transfer_token": transfer_token_async,
        "swap_tokens": swap_tokens_async,
        "get_price": get_price_from_oracle_async,
        "fetch_external_api": fetch_external_api_async,
    }
    results = []
    for op in operations:
        op = dict(op)
        op_type = op.pop("type", None)
        fn = dispatch.get(op_type)
        if fn is None:
            results.append({"error": f"unknown operation type: {op_type}"})
            if stop_on_error:
                break
            continue
        try:
            results.append({"result": await fn(**op, client=client)})
        except Exception as exc:  # noqa: BLE001
            results.append({"error": str(exc)})
            if stop_on_error:
                break
    return results


# -- websocket / HITL custom UI ---------------------------------------------------------

def listen_to_websocket(
    on_approval_request: Callable[[dict], str],
    addr: str = "127.0.0.1:8765",
    *,
    max_messages: Optional[int] = None,
) -> None:
    """Connects to lola-core's HITL WebSocket server and calls
    `on_approval_request(request_dict)` for every approval request,
    sending back whatever decision string it returns ("approve", "deny",
    or "skip"). Blocks until the connection closes or `max_messages` is
    reached.

    Requires the `websocket-client` package (a lightweight, optional
    dependency — see pyproject.toml's `[project.optional-dependencies]`).
    """
    try:
        import websocket  # type: ignore
    except ImportError as exc:
        raise ImportError(
            "listen_to_websocket requires the 'websocket-client' package. "
            "Install it with: pip install lola-os[websocket]"
        ) from exc

    count = 0

    def _on_message(ws, message):
        nonlocal count
        try:
            msg = json.loads(message)
        except json.JSONDecodeError:
            return
        if msg.get("type") == "approval_request":
            decision = on_approval_request(msg.get("request", {}))
            ws.send(json.dumps({"type": "approval_response", "id": msg.get("id"), "decision": decision}))
        count += 1
        if max_messages is not None and count >= max_messages:
            ws.close()

    ws_app = websocket.WebSocketApp(f"ws://{addr}/", on_message=_on_message)
    ws_app.run_forever()


def stream_logs(*, client: Optional[LolaCore] = None, poll_interval: float = 0.1) -> Iterator[dict]:
    """Generator yielding parsed structured log entries from lola-core as
    they arrive, for building custom dashboards or monitoring.

        for entry in lola.stream_logs():
            print(entry["level"], entry["message"])

    This is a blocking generator; run it in its own thread if used
    alongside other synchronous work.
    """
    c = _client(client)
    q = c.subscribe_logs()
    try:
        while True:
            try:
                entry = q.get(timeout=poll_interval)
                yield entry
            except Exception:
                continue
    finally:
        c.unsubscribe_logs(q)
