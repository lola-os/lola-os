import asyncio
from unittest.mock import MagicMock

import pytest

from lola_os import functions
from lola_os.context import override


def make_mock_client(return_value=None, side_effect=None):
    client = MagicMock()
    if side_effect is not None:
        client.call.side_effect = side_effect
    else:
        client.call.return_value = return_value

    async def call_async(method, params=None, timeout=None):
        if side_effect is not None:
            if isinstance(side_effect, Exception):
                raise side_effect
            return side_effect(method, params, timeout)
        return return_value

    client.call_async.side_effect = call_async
    return client


def test_get_balance_calls_expected_method_and_params():
    client = make_mock_client(return_value={"raw_value": "100", "decimals": 18, "symbol": "ETH"})
    result = functions.get_balance("ethereum", "0xabc", client=client)
    client.call.assert_called_once_with("get_balance", {"chain": "ethereum", "address": "0xabc"})
    assert result["symbol"] == "ETH"


def test_get_balance_includes_context_overrides():
    client = make_mock_client(return_value={})
    with override(chain="polygon", budget_max_usd=5.0):
        functions.get_balance("ethereum", "0xabc", client=client)
    _, params = client.call.call_args[0]
    assert params["_context_overrides"] == {"chain": "polygon", "budget_max_usd": 5.0}


def test_call_contract_unwraps_result_field():
    client = make_mock_client(return_value={"result": 12345})
    result = functions.call_contract("ethereum", "0xToken", "totalSupply", client=client)
    assert result == 12345


def test_execute_contract_passes_all_fields():
    client = make_mock_client(return_value={"tx_hash": "0xhash", "status": "pending"})
    result = functions.execute_contract(
        "ethereum", "0xFrom", "0xContract", "transfer", args=["0xTo", 100],
        abi="[]", value_wei="0", idempotency_key="key1", key_name="deployer", client=client,
    )
    assert result["tx_hash"] == "0xhash"
    _, params = client.call.call_args[0]
    assert params["method"] == "transfer"
    assert params["idempotency_key"] == "key1"
    assert params["key_name"] == "deployer"


def test_send_transaction_basic():
    client = make_mock_client(return_value={"tx_hash": "0xhash", "status": "pending"})
    result = functions.send_transaction("ethereum", "0xFrom", "0xTo", "1000000000000000000", client=client)
    assert result["status"] == "pending"


def test_multi_operation_runs_sequence_and_stops_on_error():
    client = make_mock_client()
    client.call.side_effect = [
        {"raw_value": "1", "decimals": 18, "symbol": "ETH"},
        Exception("boom"),
    ]
    ops = [
        {"type": "get_balance", "chain": "ethereum", "address": "0xabc"},
        {"type": "get_balance", "chain": "ethereum", "address": "0xdef"},
        {"type": "get_balance", "chain": "ethereum", "address": "0xshould_not_run"},
    ]
    results = functions.multi_operation(ops, client=client, stop_on_error=True)
    assert len(results) == 2
    assert "result" in results[0]
    assert "error" in results[1]


def test_multi_operation_unknown_type_records_error():
    client = make_mock_client()
    results = functions.multi_operation([{"type": "not_a_real_op"}], client=client)
    assert "error" in results[0]
    assert "unknown operation type" in results[0]["error"]


@pytest.mark.asyncio
async def test_get_balance_async():
    client = make_mock_client(return_value={"symbol": "SOL"})
    result = await functions.get_balance_async("solana", "addr123", client=client)
    assert result["symbol"] == "SOL"


@pytest.mark.asyncio
async def test_multi_operation_async_runs_sequence():
    client = make_mock_client(return_value={"symbol": "ETH"})
    ops = [{"type": "get_balance", "chain": "ethereum", "address": "0xabc"}]
    results = await functions.multi_operation_async(ops, client=client)
    assert results[0]["result"]["symbol"] == "ETH"


def test_fetch_external_api_passes_url():
    client = make_mock_client(return_value={"price": 42})
    result = functions.fetch_external_api("https://example.com/api", client=client)
    assert result["price"] == 42
    client.call.assert_called_once_with("fetch_external_api", {"url": "https://example.com/api"})
