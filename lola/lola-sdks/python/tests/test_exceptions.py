from lola_os.exceptions import (
    ABIMismatchError,
    ApprovalDeniedError,
    BudgetExceededError,
    CODE_ABI_MISMATCH,
    CODE_APPROVAL_DENIED,
    CODE_BUDGET_EXCEEDED,
    CODE_RPC_CONNECTION,
    CODE_UNKNOWN_CHAIN,
    LolaError,
    RPCConnectionError,
    UnknownChainError,
    from_rpc_error,
)


def test_from_rpc_error_maps_known_codes():
    assert isinstance(from_rpc_error("budget", CODE_BUDGET_EXCEEDED), BudgetExceededError)
    assert isinstance(from_rpc_error("abi", CODE_ABI_MISMATCH), ABIMismatchError)
    assert isinstance(from_rpc_error("rpc", CODE_RPC_CONNECTION), RPCConnectionError)
    assert isinstance(from_rpc_error("denied", CODE_APPROVAL_DENIED), ApprovalDeniedError)
    assert isinstance(from_rpc_error("chain", CODE_UNKNOWN_CHAIN), UnknownChainError)


def test_from_rpc_error_falls_back_to_base_class():
    exc = from_rpc_error("something else", -32603)
    assert isinstance(exc, LolaError)
    assert not isinstance(exc, BudgetExceededError)


def test_exception_carries_code_and_data():
    exc = BudgetExceededError("over budget", code=CODE_BUDGET_EXCEEDED, data={"limit": "gas"})
    assert exc.code == CODE_BUDGET_EXCEEDED
    assert exc.data == {"limit": "gas"}
    assert str(exc) == "over budget"
