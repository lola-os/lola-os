import * as assert from "assert";
import {
  ABIMismatchError,
  ApprovalDeniedError,
  BudgetExceededError,
  CODE_ABI_MISMATCH,
  CODE_APPROVAL_DENIED,
  CODE_BUDGET_EXCEEDED,
  CODE_RPC_CONNECTION,
  CODE_UNKNOWN_CHAIN,
  fromRpcError,
  LolaError,
  RPCConnectionError,
  UnknownChainError,
} from "../src/errors";

export function test_from_rpc_error_maps_known_codes(): void {
  assert.ok(fromRpcError("budget", CODE_BUDGET_EXCEEDED) instanceof BudgetExceededError);
  assert.ok(fromRpcError("abi", CODE_ABI_MISMATCH) instanceof ABIMismatchError);
  assert.ok(fromRpcError("rpc", CODE_RPC_CONNECTION) instanceof RPCConnectionError);
  assert.ok(fromRpcError("denied", CODE_APPROVAL_DENIED) instanceof ApprovalDeniedError);
  assert.ok(fromRpcError("chain", CODE_UNKNOWN_CHAIN) instanceof UnknownChainError);
}

export function test_from_rpc_error_falls_back_to_base_class(): void {
  const err = fromRpcError("something else", -32603);
  assert.ok(err instanceof LolaError);
  assert.ok(!(err instanceof BudgetExceededError));
}

export function test_error_carries_code_and_data(): void {
  const err = new BudgetExceededError("over budget", CODE_BUDGET_EXCEEDED, { limit: "gas" });
  assert.strictEqual(err.code, CODE_BUDGET_EXCEEDED);
  assert.deepStrictEqual(err.data, { limit: "gas" });
  assert.strictEqual(err.message, "over budget");
  assert.strictEqual(err.name, "BudgetExceededError");
}

export function test_error_is_instanceof_error(): void {
  const err = new LolaError("oops");
  assert.ok(err instanceof Error);
  assert.ok(err instanceof LolaError);
}
