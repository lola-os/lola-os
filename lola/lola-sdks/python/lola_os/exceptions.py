"""Typed exceptions for the LOLA OS Python SDK.

lola-core never panics: every failure comes back as a JSON-RPC error with
a LOLA-specific code (see lola-core/internal/jsonrpc/jsonrpc.go). This
module maps those codes to specific, catchable exception types so callers
don't have to parse error strings.
"""

from __future__ import annotations

from typing import Any, Optional


# Mirrors the codes defined in lola-core/internal/jsonrpc/jsonrpc.go
CODE_PARSE_ERROR = -32700
CODE_INVALID_REQUEST = -32600
CODE_METHOD_NOT_FOUND = -32601
CODE_INVALID_PARAMS = -32602
CODE_INTERNAL_ERROR = -32603

CODE_BUDGET_EXCEEDED = -33001
CODE_ABI_MISMATCH = -33002
CODE_RPC_CONNECTION = -33003
CODE_APPROVAL_DENIED = -33004
CODE_UNKNOWN_CHAIN = -33005


class LolaError(Exception):
    """Base class for all LOLA OS SDK errors.

    Attributes:
        code: The JSON-RPC error code returned by lola-core, if any.
        data: Optional structured error payload from lola-core.
    """

    def __init__(self, message: str, code: Optional[int] = None, data: Any = None):
        super().__init__(message)
        self.message = message
        self.code = code
        self.data = data

    def __repr__(self) -> str:  # pragma: no cover - cosmetic
        return f"{self.__class__.__name__}(message={self.message!r}, code={self.code})"


class BudgetExceededError(LolaError):
    """Raised when a write operation would exceed the configured gas/USD/
    rate budget and the circuit breaker's action is 'pause' or 'deny'.

    Suggestion: check `budget_status()` for current spend, raise the
    relevant limit in config.yaml, or wait for the per-minute rate window
    to reset.
    """


class ABIMismatchError(LolaError):
    """Raised when a contract method name or argument types don't match
    the ABI lola-core validated against. This usually means a typo in the
    method name, the wrong number of arguments, or an argument of the
    wrong type (e.g. a string passed where a uint256 was expected) — a
    common failure mode for AI agents guessing at a contract's interface.

    Suggestion: double-check the method name and argument types against
    the contract's actual ABI before retrying.
    """


class RPCConnectionError(LolaError):
    """Raised when lola-core could not reach a chain's RPC endpoint, an
    oracle, or an external API.

    Suggestion: run `lola doctor` to check connectivity, or verify the
    RPC URL in config.yaml / LOLA_RPC_URL_<CHAIN>.
    """


class ApprovalDeniedError(LolaError):
    """Raised when a human-in-the-loop approval request was denied or
    timed out."""


class UnknownChainError(LolaError):
    """Raised when a request referenced a chain name that isn't
    configured in lola-core's config.yaml."""


class VaultError(LolaError):
    """Raised for vault-related failures: wrong passphrase, missing key
    name, or a corrupted vault file."""


_CODE_TO_EXCEPTION = {
    CODE_BUDGET_EXCEEDED: BudgetExceededError,
    CODE_ABI_MISMATCH: ABIMismatchError,
    CODE_RPC_CONNECTION: RPCConnectionError,
    CODE_APPROVAL_DENIED: ApprovalDeniedError,
    CODE_UNKNOWN_CHAIN: UnknownChainError,
}


def from_rpc_error(message: str, code: Optional[int], data: Any = None) -> LolaError:
    """Construct the most specific LolaError subclass for a JSON-RPC error
    code, falling back to the base LolaError for unrecognized codes."""
    exc_cls = _CODE_TO_EXCEPTION.get(code, LolaError)
    return exc_cls(message, code=code, data=data)
