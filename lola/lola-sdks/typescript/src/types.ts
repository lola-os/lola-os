/** Shared type definitions for the LOLA OS TypeScript SDK. */

export interface Balance {
  address: string;
  token: string;
  raw_value: string;
  decimals: number;
  symbol: string;
}

export interface TxResult {
  tx_hash: string;
  status: string;
}

export interface ContractCallOptions {
  args?: unknown[];
  abi?: string;
}

export interface ExecuteContractOptions extends ContractCallOptions {
  valueWei?: string;
  idempotencyKey?: string;
  keyName?: string;
}

export interface SendTransactionOptions {
  idempotencyKey?: string;
  keyName?: string;
}

export interface TransferTokenOptions {
  idempotencyKey?: string;
  keyName?: string;
}

export interface BudgetStatus {
  gas_spent: number;
  usd_spent: number;
  requests_this_minute: number;
  paused: boolean;
  paused_reason: string;
}

export interface PriceResult {
  pair: string;
  price: number;
  decimals: number;
  updated_at: string;
}

export interface LogEntry {
  time?: string;
  level?: string;
  icon?: string;
  color?: string;
  message?: string;
  fields?: Record<string, unknown>;
  raw?: string;
}

/** A single batch operation for multiOperation(). */
export interface BatchOperation {
  type:
    | "get_balance"
    | "call_contract"
    | "execute_contract"
    | "send_transaction"
    | "transfer_token"
    | "get_price"
    | "fetch_external_api";
  [key: string]: unknown;
}

export interface BatchResult {
  result?: unknown;
  error?: string;
}
