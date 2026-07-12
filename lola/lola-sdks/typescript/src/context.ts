/**
 * Runtime context overrides: temporarily change chain, RPC URL, budget
 * limits, or mode for the duration of a callback, without mutating
 * global state.
 *
 *   await lola.withOverrides({ chain: "polygon", budgetMaxUsd: 5.0 }, async () => {
 *     const result = await executeContract(...);
 *   });
 *
 * Backed by Node's AsyncLocalStorage, so overrides correctly scope to the
 * async call tree of the callback — including across awaited promises —
 * without leaking into unrelated concurrent requests. In environments
 * without AsyncLocalStorage (older runtimes, some bundlers), this module
 * degrades to a simple module-level stack, which is still correct for
 * single-threaded synchronous nesting but cannot guarantee isolation
 * across concurrently-running async operations — see the README's
 * "Browser support" section.
 */

export interface Overrides {
  chain?: string;
  rpcUrl?: string;
  mode?: string;
  budgetMaxGas?: number;
  budgetMaxUsd?: number;
  hitlTimeoutSeconds?: number;
}

function mergeOverrides(base: Overrides, next: Overrides): Overrides {
  const merged: Overrides = { ...base };
  for (const [key, value] of Object.entries(next)) {
    if (value !== undefined) {
      (merged as Record<string, unknown>)[key] = value;
    }
  }
  return merged;
}

/** Shapes Overrides for inclusion in an RPC call's params, matching the
 * shape lola-core's config.Overrides expects on the wire. */
export function overridesToRpcParams(o: Overrides): Record<string, unknown> {
  const out: Record<string, unknown> = {};
  if (o.chain !== undefined) out.chain = o.chain;
  if (o.rpcUrl !== undefined) out.rpc_url = o.rpcUrl;
  if (o.mode !== undefined) out.mode = o.mode;
  if (o.budgetMaxGas !== undefined) out.budget_max_gas = o.budgetMaxGas;
  if (o.budgetMaxUsd !== undefined) out.budget_max_usd = o.budgetMaxUsd;
  if (o.hitlTimeoutSeconds !== undefined) out.hitl_timeout_seconds = o.hitlTimeoutSeconds;
  return out;
}

// -- storage backend selection ---------------------------------------------

interface OverrideStorage {
  getStore(): Overrides | undefined;
  run<T>(value: Overrides, fn: () => T): T;
}

class AsyncLocalStorageBackend implements OverrideStorage {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  private als: any;

  constructor() {
    // Loaded lazily/conditionally so bundlers targeting browsers (which
    // don't have node:async_hooks) don't choke on this import — see
    // FallbackStackBackend below for that environment.
    // eslint-disable-next-line @typescript-eslint/no-var-requires
    const { AsyncLocalStorage } = require("async_hooks");
    this.als = new AsyncLocalStorage();
  }

  getStore(): Overrides | undefined {
    return this.als.getStore();
  }

  run<T>(value: Overrides, fn: () => T): T {
    return this.als.run(value, fn);
  }
}

class FallbackStackBackend implements OverrideStorage {
  private stack: Overrides[] = [];

  getStore(): Overrides | undefined {
    return this.stack[this.stack.length - 1];
  }

  run<T>(value: Overrides, fn: () => T): T {
    this.stack.push(value);
    try {
      return fn();
    } finally {
      this.stack.pop();
    }
  }
}

function createBackend(): OverrideStorage {
  try {
    return new AsyncLocalStorageBackend();
  } catch {
    return new FallbackStackBackend();
  }
}

const backend = createBackend();

/** Returns the currently active overrides (merged across nested calls),
 * or an empty object if none are set. */
export function getCurrentOverrides(): Overrides {
  return backend.getStore() ?? {};
}

/**
 * Runs `fn` (sync or async) with `overrides` applied for its entire
 * execution, composing with any enclosing `withOverrides` call (the
 * innermost override wins for any field both set).
 */
export function withOverrides<T>(overrides: Overrides, fn: () => T): T {
  const merged = mergeOverrides(getCurrentOverrides(), overrides);
  return backend.run(merged, fn);
}
