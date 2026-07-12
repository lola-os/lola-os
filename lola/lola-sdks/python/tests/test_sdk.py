"""Tests for the LOLA OS Python SDK.

These tests never spawn a real lola-core subprocess: they substitute a
FakeLolaCore that records calls and returns canned responses, so the
suite runs anywhere with just Python installed.
"""

from __future__ import annotations

import asyncio
import sys
from pathlib import Path
from typing import Any, Dict, List, Optional

import pytest

sys.path.insert(0, str(Path(__file__).resolve().parents[1]))

from lola_os import context as context_mod  # noqa: E402
from lola_os import exceptions as exc_mod  # noqa: E402
from lola_os import functions as fn_mod  # noqa: E402
from lola_os.decorator import lola_tool  # noqa: E402
from lola_os.context import override, get_current_overrides, Overrides  # noqa: E402


class FakeLolaCore:
    """Drop-in stand-in for lola_os.client.LolaCore in tests."""

    def __init__(self, canned: Optional[Dict[str, Any]] = None, raise_on: Optional[Dict[str, Exception]] = None):
        self.canned = canned or {}
        self.raise_on = raise_on or {}
        self.calls: List[Dict[str, Any]] = []

    def call(self, method: str, params: Optional[dict] = None, timeout: Optional[float] = None) -> Any:
        self.calls.append({"method": method, "params": params})
        if method in self.raise_on:
            raise self.raise_on[method]
        return self.canned.get(method, {})

    async def call_async(self, method: str, params: Optional[dict] = None, timeout: Optional[float] = None) -> Any:
        return self.call(method, params, timeout)

    def subscribe_logs(self):
        import queue

        return queue.Queue()

    def unsubscribe_logs(self, q):
        pass


# ---------------------------------------------------------------------------
# Exceptions
# ---------------------------------------------------------------------------

def test_from_rpc_error_maps_known_codes():
    err = exc_mod.from_rpc_error("budget exceeded", exc_mod.CODE_BUDGET_EXCEEDED)
    assert isinstance(err, exc_mod.BudgetExceededError)
    assert err.code == exc_mod.CODE_BUDGET_EXCEEDED

    err2 = exc_mod.from_rpc_error("bad abi", exc_mod.CODE_ABI_MISMATCH)
    assert isinstance(err2, exc_mod.ABIMismatchError)


def test_from_rpc_error_falls_back_to_base_class():
    err = exc_mod.from_rpc_error("something else", -32603)
    assert isinstance(err, exc_mod.LolaError)
    assert not isinstance(err, exc_mod.BudgetExceededError)


# ---------------------------------------------------------------------------
# Context overrides
# ---------------------------------------------------------------------------

def test_override_scopes_to_with_block():
    assert get_current_overrides() == Overrides()
    with override(chain="polygon", budget_max_usd=5.0):
        ov = get_current_overrides()
        assert ov.chain == "polygon"
        assert ov.budget_max_usd == 5.0
    # Overrides must not leak outside the block.
    assert get_current_overrides() == Overrides()


def test_nested_overrides_compose_with_inner_winning():
    with override(chain="ethereum", budget_max_usd=10.0):
        with override(chain="polygon"):
            ov = get_current_overrides()
            # Inner override wins for `chain`...
            assert ov.chain == "polygon"
            # ...but the outer override's other field is preserved.
            assert ov.budget_max_usd == 10.0


def test_overrides_to_rpc_params_only_includes_set_fields():
    ov = Overrides(chain="polygon")
    params = ov.to_rpc_params()
    assert params == {"chain": "polygon"}


# ---------------------------------------------------------------------------
# Convenience functions
# ---------------------------------------------------------------------------

def test_get_balance_calls_expected_method_and_params():
    fake = FakeLolaCore(canned={"get_balance": {"address": "0xabc", "raw_value": "100", "decimals": 18, "symbol": "ETH"}})
    result = fn_mod.get_balance("ethereum", "0xabc", client=fake)
    assert result["raw_value"] == "100"
    assert fake.calls[0]["method"] == "get_balance"
    assert fake.calls[0]["params"]["chain"] == "ethereum"
    assert fake.calls[0]["params"]["address"] == "0xabc"


def test_get_balance_includes_context_overrides_in_params():
    fake = FakeLolaCore(canned={"get_balance": {}})
    with override(chain="polygon", rpc_url="https://backup.example.com"):
        fn_mod.get_balance("ethereum", "0xabc", client=fake)
    sent = fake.calls[0]["params"]
    assert sent["_context_overrides"]["chain"] == "polygon"
    assert sent["_context_overrides"]["rpc_url"] == "https://backup.example.com"


def test_get_balance_without_override_omits_context_overrides_key():
    fake = FakeLolaCore(canned={"get_balance": {}})
    fn_mod.get_balance("ethereum", "0xabc", client=fake)
    assert "_context_overrides" not in fake.calls[0]["params"]


def test_call_contract_unwraps_result_key():
    fake = FakeLolaCore(canned={"call_contract": {"result": 42}})
    out = fn_mod.call_contract("ethereum", "0xToken", "totalSupply", client=fake)
    assert out == 42


def test_call_contract_raises_abi_mismatch_error():
    fake = FakeLolaCore(raise_on={"call_contract": exc_mod.ABIMismatchError("bad method", code=exc_mod.CODE_ABI_MISMATCH)})
    with pytest.raises(exc_mod.ABIMismatchError):
        fn_mod.call_contract("ethereum", "0xToken", "trasnfer", client=fake)


def test_send_transaction_returns_tx_hash():
    fake = FakeLolaCore(canned={"send_transaction": {"tx_hash": "0xdead", "status": "pending"}})
    out = fn_mod.send_transaction("ethereum", "0xfrom", "0xto", "1000", client=fake)
    assert out["tx_hash"] == "0xdead"


def test_send_transaction_raises_budget_exceeded():
    fake = FakeLolaCore(raise_on={"send_transaction": exc_mod.BudgetExceededError("over budget", code=exc_mod.CODE_BUDGET_EXCEEDED)})
    with pytest.raises(exc_mod.BudgetExceededError):
        fn_mod.send_transaction("ethereum", "0xfrom", "0xto", "999999999999999999999", client=fake)


def test_multi_operation_runs_in_order_and_stops_on_error():
    fake = FakeLolaCore(
        canned={"get_balance": {"raw_value": "1"}},
        raise_on={"call_contract": RuntimeError("boom")},
    )
    ops = [
        {"type": "get_balance", "chain": "ethereum", "address": "0xabc"},
        {"type": "call_contract", "chain": "ethereum", "contract": "0xC", "method": "x"},
        {"type": "get_balance", "chain": "ethereum", "address": "0xdef"},
    ]
    results = fn_mod.multi_operation(ops, client=fake, stop_on_error=True)
    assert len(results) == 2  # stopped after the error on step 2
    assert "result" in results[0]
    assert "error" in results[1]


def test_multi_operation_continues_when_stop_on_error_false():
    fake = FakeLolaCore(
        canned={"get_balance": {"raw_value": "1"}},
        raise_on={"call_contract": RuntimeError("boom")},
    )
    ops = [
        {"type": "call_contract", "chain": "ethereum", "contract": "0xC", "method": "x"},
        {"type": "get_balance", "chain": "ethereum", "address": "0xdef"},
    ]
    results = fn_mod.multi_operation(ops, client=fake, stop_on_error=False)
    assert len(results) == 2
    assert "error" in results[0]
    assert "result" in results[1]


# ---------------------------------------------------------------------------
# Async variants
# ---------------------------------------------------------------------------

def test_get_balance_async_works():
    fake = FakeLolaCore(canned={"get_balance": {"raw_value": "7"}})

    async def run():
        return await fn_mod.get_balance_async("ethereum", "0xabc", client=fake)

    result = asyncio.run(run())
    assert result["raw_value"] == "7"


def test_multi_operation_async():
    fake = FakeLolaCore(canned={"get_balance": {"raw_value": "1"}})
    ops = [{"type": "get_balance", "chain": "ethereum", "address": "0xabc"}]

    async def run():
        return await fn_mod.multi_operation_async(ops, client=fake)

    results = asyncio.run(run())
    assert "result" in results[0]


# ---------------------------------------------------------------------------
# @lola_tool decorator
# ---------------------------------------------------------------------------

def test_lola_tool_bare_decorator_marks_function(monkeypatch):
    monkeypatch.setattr("lola_os.decorator.get_client", lambda: None)

    @lola_tool
    def my_tool(x: int) -> int:
        """Doubles x."""
        return x * 2

    assert my_tool.__lola_tool__ is True
    assert my_tool.__lola_name__ == "my_tool"
    assert "Doubles" in my_tool.__lola_description__
    assert my_tool(21) == 42


def test_lola_tool_with_overrides_applies_them_during_call(monkeypatch):
    monkeypatch.setattr("lola_os.decorator.get_client", lambda: None)
    captured = {}

    @lola_tool(config_overrides={"chain": "polygon", "budget_max_usd": 3.0})
    def my_tool() -> None:
        captured["overrides"] = get_current_overrides()

    my_tool()
    assert captured["overrides"].chain == "polygon"
    assert captured["overrides"].budget_max_usd == 3.0
    # And overrides must not leak after the call returns.
    assert get_current_overrides() == Overrides()


def test_lola_tool_rejects_unknown_override_keys():
    with pytest.raises(TypeError):
        @lola_tool(config_overrides={"not_a_real_field": 1})
        def my_tool():
            pass


def test_lola_tool_works_with_async_functions(monkeypatch):
    monkeypatch.setattr("lola_os.decorator.get_client", lambda: None)

    @lola_tool
    async def my_async_tool(x: int) -> int:
        await asyncio.sleep(0)
        return x + 1

    assert my_async_tool.__lola_tool__ is True
    result = asyncio.run(my_async_tool(41))
    assert result == 42
