import * as assert from "assert";

// Mock the client module's getClient before requiring decorator, so
// lolaTool's internal getClient() call never tries to spawn a real
// lola-core subprocess during these pure-logic tests.
// eslint-disable-next-line @typescript-eslint/no-var-requires
const clientModule = require("../src/client");
const originalGetClient = clientModule.getClient;
clientModule.getClient = () => ({ mocked: true });

import { lolaTool } from "../src/decorator";
import { getCurrentOverrides } from "../src/context";

export function test_lolaTool_wraps_sync_function_metadata(): void {
  function add(a: number, b: number): number {
    return a + b;
  }
  const wrapped = lolaTool(add, { name: "adder", description: "adds two numbers" });
  assert.strictEqual(wrapped.lolaTool, true);
  assert.strictEqual(wrapped.lolaName, "adder");
  assert.strictEqual(wrapped.lolaDescription, "adds two numbers");
  assert.strictEqual(wrapped(2, 3), 5);
}

export function test_lolaTool_defaults_name_to_function_name(): void {
  function myFunction(): number {
    return 1;
  }
  const wrapped = lolaTool(myFunction);
  assert.strictEqual(wrapped.lolaName, "myFunction");
}

export function test_lolaTool_config_overrides_scope_only_during_call(): void {
  let seenChain: string | undefined;
  const fn = lolaTool(
    () => {
      seenChain = getCurrentOverrides().chain;
      return "ok";
    },
    { configOverrides: { chain: "polygon" } }
  );

  const result = fn();
  assert.strictEqual(result, "ok");
  assert.strictEqual(seenChain, "polygon");
  assert.strictEqual(getCurrentOverrides().chain, undefined);
}

export function test_lolaTool_without_overrides_leaves_context_untouched(): void {
  let seen: string | undefined = "untouched";
  const fn = lolaTool(() => {
    seen = getCurrentOverrides().chain;
    return null;
  });
  fn();
  assert.strictEqual(seen, undefined);
}

export function teardown(): void {
  clientModule.getClient = originalGetClient;
}
