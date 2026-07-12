"""The `@lola_tool` decorator.

In its simplest form it's a transparent passthrough that just marks a
function as a LOLA tool (useful for agent frameworks like CrewAI/LangChain
that discover tools by inspecting decorated callables) and ensures the
LolaCore client is started before the function runs.

It also accepts `config_overrides` to scope chain/RPC/budget overrides to
every call of that specific function, composing with any enclosing
`lola.override(...)` context manager (the decorator's overrides win for
any field present in both).
"""

from __future__ import annotations

import asyncio
import functools
import inspect
from typing import Any, Callable, Dict, Optional, TypeVar

from .client import get_client
from .context import Overrides, override

F = TypeVar("F", bound=Callable[..., Any])


def lola_tool(
    func: Optional[F] = None,
    *,
    config_overrides: Optional[Dict[str, Any]] = None,
    name: Optional[str] = None,
    description: Optional[str] = None,
) -> Callable[..., Any]:
    """Marks `func` as a LOLA tool.

    Usage as a bare decorator:

        @lola_tool
        def check_balance(address: str) -> dict:
            return get_balance("ethereum", address)

    Usage with overrides:

        @lola_tool(config_overrides={"chain": "polygon", "budget_max_usd": 5.0})
        def check_balance_polygon(address: str) -> dict:
            return get_balance("polygon", address)

    Works transparently with both sync and async functions. The wrapped
    function gains a few introspectable attributes agent frameworks can
    use for tool discovery: `__lola_tool__`, `__lola_name__`,
    `__lola_description__`, and `__lola_signature__`.
    """

    def decorate(fn: F) -> F:
        overrides_obj = _overrides_from_dict(config_overrides or {})
        tool_name = name or fn.__name__
        tool_description = description or (inspect.getdoc(fn) or "").strip()
        signature = inspect.signature(fn)

        if asyncio.iscoroutinefunction(fn):
            @functools.wraps(fn)
            async def async_wrapper(*args: Any, **kwargs: Any) -> Any:
                get_client()  # ensure the engine is up before the tool body runs
                if config_overrides:
                    with override(**overrides_obj.__dict__):
                        return await fn(*args, **kwargs)
                return await fn(*args, **kwargs)

            wrapper = async_wrapper
        else:
            @functools.wraps(fn)
            def sync_wrapper(*args: Any, **kwargs: Any) -> Any:
                get_client()
                if config_overrides:
                    with override(**overrides_obj.__dict__):
                        return fn(*args, **kwargs)
                return fn(*args, **kwargs)

            wrapper = sync_wrapper

        wrapper.__lola_tool__ = True
        wrapper.__lola_name__ = tool_name
        wrapper.__lola_description__ = tool_description
        wrapper.__lola_signature__ = signature
        return wrapper  # type: ignore[return-value]

    if func is not None:
        # Used as @lola_tool with no parentheses.
        return decorate(func)
    # Used as @lola_tool(...) with arguments.
    return decorate


def _overrides_from_dict(d: Dict[str, Any]) -> Overrides:
    known_fields = {"chain", "rpc_url", "mode", "budget_max_gas", "budget_max_usd", "hitl_timeout_seconds"}
    unknown = set(d.keys()) - known_fields
    if unknown:
        raise TypeError(
            f"lola_tool(config_overrides=...) received unknown key(s): {sorted(unknown)}. "
            f"Valid keys are: {sorted(known_fields)}"
        )
    return Overrides(**d)
