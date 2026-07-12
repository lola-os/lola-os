import asyncio
from unittest.mock import patch

import pytest

from lola_os.context import get_current_overrides
from lola_os.decorator import lola_tool


@patch("lola_os.decorator.get_client")
def test_bare_decorator_sync(mock_get_client):
    @lola_tool
    def add(a, b):
        return a + b

    assert add(2, 3) == 5
    assert add.__lola_tool__ is True
    assert add.__lola_name__ == "add"
    mock_get_client.assert_called()


@patch("lola_os.decorator.get_client")
def test_decorator_with_parens_no_overrides(mock_get_client):
    @lola_tool(name="custom_name", description="does a thing")
    def fn():
        return 42

    assert fn() == 42
    assert fn.__lola_name__ == "custom_name"
    assert fn.__lola_description__ == "does a thing"


@patch("lola_os.decorator.get_client")
def test_decorator_applies_config_overrides_during_call(mock_get_client):
    captured = {}

    @lola_tool(config_overrides={"chain": "polygon", "budget_max_usd": 2.5})
    def fn():
        current = get_current_overrides()
        captured["chain"] = current.chain
        captured["budget_max_usd"] = current.budget_max_usd
        return "ok"

    assert fn() == "ok"
    assert captured["chain"] == "polygon"
    assert captured["budget_max_usd"] == 2.5
    # Overrides should not leak outside the call.
    assert get_current_overrides().chain is None


@patch("lola_os.decorator.get_client")
def test_decorator_rejects_unknown_override_keys(mock_get_client):
    with pytest.raises(TypeError):
        @lola_tool(config_overrides={"not_a_real_field": 1})
        def fn():
            return None


@patch("lola_os.decorator.get_client")
def test_decorator_works_with_async_functions(mock_get_client):
    @lola_tool
    async def fetch():
        await asyncio.sleep(0)
        return "done"

    result = asyncio.run(fetch())
    assert result == "done"


@patch("lola_os.decorator.get_client")
def test_async_decorator_applies_overrides(mock_get_client):
    captured = {}

    @lola_tool(config_overrides={"chain": "solana"})
    async def fetch():
        captured["chain"] = get_current_overrides().chain
        return "ok"

    assert asyncio.run(fetch()) == "ok"
    assert captured["chain"] == "solana"
