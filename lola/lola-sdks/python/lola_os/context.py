"""Runtime context overrides: temporarily change chain, RPC URL, budget
limits, or mode for the duration of a `with` block, without mutating
global state or requiring a config file edit.

    with lola.override(chain="polygon", budget_max_usd=5.0):
        result = execute_contract(...)

Overrides are stored in a contextvar so they compose correctly with
threads/async tasks and don't leak between unrelated calls.
"""

from __future__ import annotations

import contextvars
from contextlib import contextmanager
from dataclasses import dataclass, replace
from typing import Optional

_current_overrides: contextvars.ContextVar[Optional["Overrides"]] = contextvars.ContextVar(
    "lola_overrides", default=None
)


@dataclass(frozen=True)
class Overrides:
    chain: Optional[str] = None
    rpc_url: Optional[str] = None
    mode: Optional[str] = None
    budget_max_gas: Optional[float] = None
    budget_max_usd: Optional[float] = None
    hitl_timeout_seconds: Optional[int] = None

    def merged_with(self, other: "Overrides") -> "Overrides":
        """Returns a new Overrides with `other`'s non-None fields taking
        precedence over self's. Used to compose decorator-level overrides
        with an enclosing `with lola.override(...)` block."""
        updates = {k: v for k, v in other.__dict__.items() if v is not None}
        return replace(self, **updates)

    def to_rpc_params(self) -> dict:
        """Shapes this Overrides for inclusion in an RPC call's params
        under the `_context_overrides` key, matching the shape lola-core's
        config.Overrides expects."""
        out = {}
        if self.chain is not None:
            out["chain"] = self.chain
        if self.rpc_url is not None:
            out["rpc_url"] = self.rpc_url
        if self.mode is not None:
            out["mode"] = self.mode
        if self.budget_max_gas is not None:
            out["budget_max_gas"] = self.budget_max_gas
        if self.budget_max_usd is not None:
            out["budget_max_usd"] = self.budget_max_usd
        if self.hitl_timeout_seconds is not None:
            out["hitl_timeout_seconds"] = self.hitl_timeout_seconds
        return out


def get_current_overrides() -> Overrides:
    return _current_overrides.get() or Overrides()


@contextmanager
def override(
    chain: Optional[str] = None,
    rpc_url: Optional[str] = None,
    mode: Optional[str] = None,
    budget_max_gas: Optional[float] = None,
    budget_max_usd: Optional[float] = None,
    hitl_timeout_seconds: Optional[int] = None,
):
    """Context manager that scopes config overrides to the enclosed block.

    Example:

        with lola.override(rpc_url="https://backup.alchemy.io", chain="ethereum"):
            balance = get_balance("ethereum", "0x...")
    """
    new_overrides = Overrides(
        chain=chain,
        rpc_url=rpc_url,
        mode=mode,
        budget_max_gas=budget_max_gas,
        budget_max_usd=budget_max_usd,
        hitl_timeout_seconds=hitl_timeout_seconds,
    )
    merged = get_current_overrides().merged_with(new_overrides)
    token = _current_overrides.set(merged)
    try:
        yield merged
    finally:
        _current_overrides.reset(token)
