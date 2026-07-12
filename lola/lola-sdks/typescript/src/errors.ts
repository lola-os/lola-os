/**
 * Typed errors for the LOLA OS TypeScript SDK.
 *
 * lola-core never throws/panics internally: every failure comes back as a
 * JSON-RPC error object with a LOLA-specific code (see
 * lola-core/internal/jsonrpc/jsonrpc.go). This module maps those codes to
 * specific, catchable error classes.
 */

export const CODE_PARSE_ERROR = -32700;
export const CODE_INVALID_REQUEST = -32600;
export const CODE_METHOD_NOT_FOUND = -32601;
export const CODE_INVALID_PARAMS = -32602;
export const CODE_INTERNAL_ERROR = -32603;

export const CODE_BUDGET_EXCEEDED = -33001;
export const CODE_ABI_MISMATCH = -33002;
export const CODE_RPC_CONNECTION = -33003;
export const CODE_APPROVAL_DENIED = -33004;
export const CODE_UNKNOWN_CHAIN = -33005;

/** Base class for all LOLA OS SDK errors. */
export class LolaError extends Error {
  public readonly code?: number;
  public readonly data?: unknown;

  constructor(message: string, code?: number, data?: unknown) {
    super(message);
    this.name = "LolaError";
    this.code = code;
    this.data = data;
    Object.setPrototypeOf(this, LolaError.prototype);
  }
}

/**
 * Thrown when a write operation would exceed the configured gas/USD/rate
 * budget and the circuit breaker's action is "pause" or "deny".
 *
 * Suggestion: check `budgetStatus()` for current spend, raise the
 * relevant limit in config.yaml, or wait for the rate window to reset.
 */
export class BudgetExceededError extends LolaError {
  constructor(message: string, code?: number, data?: unknown) {
    super(message, code, data);
    this.name = "BudgetExceededError";
    Object.setPrototypeOf(this, BudgetExceededError.prototype);
  }
}

/**
 * Thrown when a contract method name or argument types don't match the
 * ABI lola-core validated against — usually a typo in the method name or
 * an argument of the wrong type, a common failure mode for AI agents
 * guessing at a contract's interface.
 */
export class ABIMismatchError extends LolaError {
  constructor(message: string, code?: number, data?: unknown) {
    super(message, code, data);
    this.name = "ABIMismatchError";
    Object.setPrototypeOf(this, ABIMismatchError.prototype);
  }
}

/**
 * Thrown when lola-core could not reach a chain's RPC endpoint, an
 * oracle, or an external API.
 *
 * Suggestion: run `lola doctor` to check connectivity.
 */
export class RPCConnectionError extends LolaError {
  constructor(message: string, code?: number, data?: unknown) {
    super(message, code, data);
    this.name = "RPCConnectionError";
    Object.setPrototypeOf(this, RPCConnectionError.prototype);
  }
}

/** Thrown when a human-in-the-loop approval request was denied or timed out. */
export class ApprovalDeniedError extends LolaError {
  constructor(message: string, code?: number, data?: unknown) {
    super(message, code, data);
    this.name = "ApprovalDeniedError";
    Object.setPrototypeOf(this, ApprovalDeniedError.prototype);
  }
}

/** Thrown when a request referenced a chain name that isn't configured. */
export class UnknownChainError extends LolaError {
  constructor(message: string, code?: number, data?: unknown) {
    super(message, code, data);
    this.name = "UnknownChainError";
    Object.setPrototypeOf(this, UnknownChainError.prototype);
  }
}

const CODE_TO_ERROR_CLASS: Record<number, new (message: string, code?: number, data?: unknown) => LolaError> = {
  [CODE_BUDGET_EXCEEDED]: BudgetExceededError,
  [CODE_ABI_MISMATCH]: ABIMismatchError,
  [CODE_RPC_CONNECTION]: RPCConnectionError,
  [CODE_APPROVAL_DENIED]: ApprovalDeniedError,
  [CODE_UNKNOWN_CHAIN]: UnknownChainError,
};

/**
 * Constructs the most specific LolaError subclass for a JSON-RPC error
 * code, falling back to the base LolaError for unrecognized codes.
 */
export function fromRpcError(message: string, code?: number, data?: unknown): LolaError {
  const cls = code !== undefined ? CODE_TO_ERROR_CLASS[code] : undefined;
  if (cls) {
    return new cls(message, code, data);
  }
  return new LolaError(message, code, data);
}
