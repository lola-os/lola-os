import * as assert from "assert";
import { withOverrides, getCurrentOverrides, overridesToRpcParams } from "../src/context";

export function test_no_overrides_by_default(): void {
  assert.deepStrictEqual(getCurrentOverrides(), {});
}

export function test_with_overrides_sets_and_restores(): void {
  withOverrides({ chain: "polygon", budgetMaxUsd: 5.0 }, () => {
    const current = getCurrentOverrides();
    assert.strictEqual(current.chain, "polygon");
    assert.strictEqual(current.budgetMaxUsd, 5.0);
  });
  assert.deepStrictEqual(getCurrentOverrides(), {});
}

export function test_nested_overrides_compose(): void {
  withOverrides({ chain: "ethereum", budgetMaxUsd: 10.0 }, () => {
    withOverrides({ chain: "polygon" }, () => {
      const current = getCurrentOverrides();
      assert.strictEqual(current.chain, "polygon");
      assert.strictEqual(current.budgetMaxUsd, 10.0);
    });
    assert.strictEqual(getCurrentOverrides().chain, "ethereum");
  });
}

export async function test_overrides_survive_across_awaits(): Promise<void> {
  await withOverrides({ chain: "solana" }, async () => {
    await new Promise((resolve) => setTimeout(resolve, 5));
    assert.strictEqual(getCurrentOverrides().chain, "solana");
  });
}

export function test_overrides_to_rpc_params_only_set_fields(): void {
  const params = overridesToRpcParams({ chain: "solana" });
  assert.deepStrictEqual(params, { chain: "solana" });
}

export function test_overrides_to_rpc_params_empty_when_nothing_set(): void {
  assert.deepStrictEqual(overridesToRpcParams({}), {});
}

export function test_overrides_to_rpc_params_maps_all_fields(): void {
  const params = overridesToRpcParams({
    chain: "ethereum",
    rpcUrl: "https://example.com",
    mode: "live",
    budgetMaxGas: 1.0,
    budgetMaxUsd: 2.0,
    hitlTimeoutSeconds: 30,
  });
  assert.deepStrictEqual(params, {
    chain: "ethereum",
    rpc_url: "https://example.com",
    mode: "live",
    budget_max_gas: 1.0,
    budget_max_usd: 2.0,
    hitl_timeout_seconds: 30,
  });
}
