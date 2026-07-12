import * as assert from "assert";
import { withOverrides } from "../src/context";
import * as functions from "../src/functions";
import { LolaCore } from "../src/client";

function mockClient(impl: (method: string, params: unknown) => unknown): LolaCore {
  const calls: Array<{ method: string; params: unknown }> = [];
  const fake = {
    calls,
    call: async (method: string, params: unknown) => {
      calls.push({ method, params });
      const result = impl(method, params);
      if (result instanceof Error) throw result;
      return result;
    },
  };
  return fake as unknown as LolaCore;
}

export async function test_getBalance_calls_expected_method_and_params(): Promise<void> {
  const client = mockClient(() => ({ raw_value: "100", decimals: 18, symbol: "ETH" }));
  const result = await functions.getBalance("ethereum", "0xabc", client);
  assert.strictEqual((result as { symbol: string }).symbol, "ETH");
  const calls = (client as unknown as { calls: Array<{ method: string; params: unknown }> }).calls;
  assert.strictEqual(calls[0].method, "get_balance");
  assert.deepStrictEqual(calls[0].params, { chain: "ethereum", address: "0xabc" });
}

export async function test_getBalance_includes_context_overrides(): Promise<void> {
  const client = mockClient(() => ({}));
  await withOverrides({ chain: "polygon", budgetMaxUsd: 5.0 }, async () => {
    await functions.getBalance("ethereum", "0xabc", client);
  });
  const calls = (client as unknown as { calls: Array<{ method: string; params: Record<string, unknown> }> }).calls;
  assert.deepStrictEqual(calls[0].params._context_overrides, { chain: "polygon", budget_max_usd: 5.0 });
}

export async function test_callContract_unwraps_result_field(): Promise<void> {
  const client = mockClient(() => ({ result: 12345 }));
  const result = await functions.callContract("ethereum", "0xToken", "totalSupply", [], "", client);
  assert.strictEqual(result, 12345);
}

export async function test_multiOperation_runs_sequence_and_stops_on_error(): Promise<void> {
  let call = 0;
  const client = mockClient(() => {
    call += 1;
    if (call === 1) return { raw_value: "1", decimals: 18, symbol: "ETH" };
    return new Error("boom");
  });

  const results = await functions.multiOperation(
    [
      { type: "get_balance", chain: "ethereum", address: "0xabc" },
      { type: "get_balance", chain: "ethereum", address: "0xdef" },
      { type: "get_balance", chain: "ethereum", address: "0xshould_not_run" },
    ],
    { client, stopOnError: true }
  );

  assert.strictEqual(results.length, 2);
  assert.ok(results[0].result);
  assert.ok(results[1].error);
}

export async function test_multiOperation_unknown_type_records_error(): Promise<void> {
  const client = mockClient(() => ({}));
  const results = await functions.multiOperation([{ type: "not_a_real_op" } as never], { client });
  assert.ok(results[0].error?.includes("unknown operation type"));
}

export async function test_fetchExternalApi_passes_url(): Promise<void> {
  const client = mockClient(() => ({ price: 42 }));
  const result = await functions.fetchExternalApi<{ price: number }>("https://example.com/api", client);
  assert.strictEqual(result.price, 42);
}
