/**
 * Convenience functions wrapping LolaCore.call() for every operation LOLA
 * OS supports. Each returns a Promise (no separate sync/async split in
 * JS — everything here is already async by nature of the IPC call).
 *
 * All functions honor the current context overrides (set via
 * `withOverrides(...)`) by attaching them to the RPC params under
 * `_context_overrides`.
 */

import { LolaCore, getClient } from "./client";
import { getCurrentOverrides, overridesToRpcParams } from "./context";
import {
  Balance,
  BatchOperation,
  BatchResult,
  BudgetStatus,
  ExecuteContractOptions,
  PriceResult,
  SendTransactionOptions,
  TransferTokenOptions,
  TxResult,
} from "./types";

function withOverrideParams(params: Record<string, unknown>): Record<string, unknown> {
  const overrides = overridesToRpcParams(getCurrentOverrides());
  if (Object.keys(overrides).length > 0) {
    return { ...params, _context_overrides: overrides };
  }
  return params;
}

function resolveClient(client?: LolaCore): LolaCore {
  return client ?? getClient();
}

// -- balances ---------------------------------------------------------------

export async function getBalance(chain: string, address: string, client?: LolaCore): Promise<Balance> {
  return resolveClient(client).call<Balance>("get_balance", withOverrideParams({ chain, address }));
}

export async function getTokenBalance(
  chain: string,
  address: string,
  tokenAddress: string,
  client?: LolaCore
): Promise<Balance> {
  return resolveClient(client).call<Balance>(
    "get_token_balance",
    withOverrideParams({ chain, address, token_address: tokenAddress })
  );
}

// -- contract calls ---------------------------------------------------------

export async function callContract(
  chain: string,
  contract: string,
  method: string,
  args: unknown[] = [],
  abi = "",
  client?: LolaCore
): Promise<unknown> {
  const result = await resolveClient(client).call<{ result: unknown }>(
    "call_contract",
    withOverrideParams({ chain, contract, method, args, abi })
  );
  return (result as { result?: unknown })?.result ?? result;
}

export async function executeContract(
  chain: string,
  fromAddress: string,
  contract: string,
  method: string,
  options: ExecuteContractOptions = {}
): Promise<TxResult> {
  const { args = [], abi = "", valueWei = "0", idempotencyKey = "", keyName = "default" } = options;
  return getClient().call<TxResult>(
    "execute_contract",
    withOverrideParams({
      chain,
      from: fromAddress,
      contract,
      method,
      args,
      abi,
      value_wei: valueWei,
      idempotency_key: idempotencyKey,
      key_name: keyName,
    })
  );
}

// -- transactions ---------------------------------------------------------

export async function sendTransaction(
  chain: string,
  fromAddress: string,
  to: string,
  valueWei: string,
  options: SendTransactionOptions = {}
): Promise<TxResult> {
  const { idempotencyKey = "", keyName = "default" } = options;
  return getClient().call<TxResult>(
    "send_transaction",
    withOverrideParams({
      chain,
      from: fromAddress,
      to,
      value_wei: valueWei,
      idempotency_key: idempotencyKey,
      key_name: keyName,
    })
  );
}

export async function transferToken(
  chain: string,
  fromAddress: string,
  to: string,
  token: string,
  amountRaw: string,
  options: TransferTokenOptions = {}
): Promise<TxResult> {
  const { idempotencyKey = "", keyName = "default" } = options;
  return getClient().call<TxResult>(
    "transfer_token",
    withOverrideParams({
      chain,
      from: fromAddress,
      to,
      token,
      amount_raw: amountRaw,
      idempotency_key: idempotencyKey,
      key_name: keyName,
    })
  );
}

export async function swapTokens(
  chain: string,
  fromAddress: string,
  routerContract: string,
  method: string,
  args: unknown[],
  abi: string,
  options: ExecuteContractOptions = {}
): Promise<TxResult> {
  return executeContract(chain, fromAddress, routerContract, method, { ...options, args, abi });
}

// -- oracles / external APIs ---------------------------------------------------------

export async function getPriceFromOracle(chain: string, pair: string, client?: LolaCore): Promise<PriceResult> {
  return resolveClient(client).call<PriceResult>("get_price", withOverrideParams({ chain, pair }));
}

export async function fetchExternalApi<T = unknown>(url: string, client?: LolaCore): Promise<T> {
  return resolveClient(client).call<T>("fetch_external_api", withOverrideParams({ url }));
}

// -- budget / vault introspection ---------------------------------------------------------

export async function budgetStatus(client?: LolaCore): Promise<BudgetStatus> {
  return resolveClient(client).call<BudgetStatus>("budget_status", {});
}

export async function vaultKeyNames(client?: LolaCore): Promise<string[]> {
  const result = await resolveClient(client).call<{ entries: string[] }>("vault_list", {});
  return result.entries;
}

// -- batch ---------------------------------------------------------

type OpFn = (op: BatchOperation, client?: LolaCore) => Promise<unknown>;

const dispatch: Record<string, OpFn> = {
  get_balance: (op, c) => getBalance(op.chain as string, op.address as string, c),
  call_contract: (op, c) =>
    callContract(op.chain as string, op.contract as string, op.method as string, (op.args as unknown[]) ?? [], (op.abi as string) ?? "", c),
  execute_contract: (op) =>
    executeContract(op.chain as string, op.from as string, op.contract as string, op.method as string, op as ExecuteContractOptions),
  send_transaction: (op) =>
    sendTransaction(op.chain as string, op.from as string, op.to as string, op.value_wei as string, op as SendTransactionOptions),
  transfer_token: (op) =>
    transferToken(op.chain as string, op.from as string, op.to as string, op.token as string, op.amount_raw as string, op as TransferTokenOptions),
  get_price: (op, c) => getPriceFromOracle(op.chain as string, op.pair as string, c),
  fetch_external_api: (op, c) => fetchExternalApi(op.url as string, c),
};

/**
 * Runs a sequence of operations client-side, in order, returning each
 * result. For on-chain ordering guarantees and a replayable audit trail,
 * prefer `lola replay <plan.json>` over multiOperation, which is intended
 * for lighter-weight, code-driven batches inside an agent's own logic.
 */
export async function multiOperation(
  operations: BatchOperation[],
  options: { client?: LolaCore; stopOnError?: boolean } = {}
): Promise<BatchResult[]> {
  const { client, stopOnError = true } = options;
  const results: BatchResult[] = [];
  for (const op of operations) {
    const fn = dispatch[op.type];
    if (!fn) {
      results.push({ error: `unknown operation type: ${op.type}` });
      if (stopOnError) break;
      continue;
    }
    try {
      results.push({ result: await fn(op, client) });
    } catch (err) {
      results.push({ error: err instanceof Error ? err.message : String(err) });
      if (stopOnError) break;
    }
  }
  return results;
}

// -- log streaming ---------------------------------------------------------

/**
 * Async generator yielding parsed structured log entries from lola-core
 * as they arrive.
 *
 *   for await (const entry of streamLogs()) {
 *     console.log(entry.level, entry.message);
 *   }
 */
export async function* streamLogs(client?: LolaCore): AsyncGenerator<import("./types").LogEntry> {
  const c = resolveClient(client);
  await c.start();

  const queue: import("./types").LogEntry[] = [];
  let resolveNext: (() => void) | null = null;

  const unsubscribe = c.onLog((entry) => {
    queue.push(entry);
    if (resolveNext) {
      const r = resolveNext;
      resolveNext = null;
      r();
    }
  });

  try {
    while (true) {
      if (queue.length > 0) {
        yield queue.shift() as import("./types").LogEntry;
      } else {
        await new Promise<void>((resolve) => {
          resolveNext = resolve;
        });
      }
    }
  } finally {
    unsubscribe();
  }
}
