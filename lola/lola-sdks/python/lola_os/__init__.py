"""LOLA OS Python SDK.

Add ``@lola_tool`` to any function and it gains the ability to talk to
blockchains, oracles, and APIs through the local ``lola-core`` engine.

    from lola_os import lola_tool, get_balance

    @lola_tool
    def check_balance(address: str) -> dict:
        return get_balance("ethereum", address)

See README.md for the full quickstart and API reference.
"""

from .client import LolaCore, get_client
from .decorator import lola_tool
from .context import override
from .exceptions import (
    LolaError,
    BudgetExceededError,
    ABIMismatchError,
    RPCConnectionError,
    ApprovalDeniedError,
    UnknownChainError,
    VaultError,
)
from .functions import (
    get_balance,
    get_balance_async,
    get_token_balance,
    get_token_balance_async,
    call_contract,
    call_contract_async,
    execute_contract,
    execute_contract_async,
    send_transaction,
    send_transaction_async,
    transfer_token,
    transfer_token_async,
    swap_tokens,
    swap_tokens_async,
    get_price_from_oracle,
    get_price_from_oracle_async,
    fetch_external_api,
    fetch_external_api_async,
    multi_operation,
    multi_operation_async,
    listen_to_websocket,
    stream_logs,
)

__version__ = "1.0.0"

__all__ = [
    "LolaCore",
    "get_client",
    "lola_tool",
    "override",
    "LolaError",
    "BudgetExceededError",
    "ABIMismatchError",
    "RPCConnectionError",
    "ApprovalDeniedError",
    "UnknownChainError",
    "VaultError",
    "get_balance",
    "get_balance_async",
    "get_token_balance",
    "get_token_balance_async",
    "call_contract",
    "call_contract_async",
    "execute_contract",
    "execute_contract_async",
    "send_transaction",
    "send_transaction_async",
    "transfer_token",
    "transfer_token_async",
    "swap_tokens",
    "swap_tokens_async",
    "get_price_from_oracle",
    "get_price_from_oracle_async",
    "fetch_external_api",
    "fetch_external_api_async",
    "multi_operation",
    "multi_operation_async",
    "listen_to_websocket",
    "stream_logs",
    "__version__",
]
