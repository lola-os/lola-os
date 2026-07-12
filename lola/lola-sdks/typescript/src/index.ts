/**
 * LOLA OS TypeScript/JavaScript SDK.
 *
 *   import { lolaTool, getBalance } from "lola-os";
 *
 *   const checkBalance = lolaTool(async (address: string) => {
 *     return getBalance("ethereum", address);
 *   });
 *
 * See README.md for the full quickstart and API reference.
 */

export { LolaCore, getClient, resetClient, LolaCoreOptions } from "./client";
export { lolaTool, LolaTool, LolaToolOptions, LolaToolMetadata, WrappedTool } from "./decorator";
export { withOverrides, getCurrentOverrides, Overrides } from "./context";
export {
  LolaError,
  BudgetExceededError,
  ABIMismatchError,
  RPCConnectionError,
  ApprovalDeniedError,
  UnknownChainError,
} from "./errors";
export {
  getBalance,
  getTokenBalance,
  callContract,
  executeContract,
  sendTransaction,
  transferToken,
  swapTokens,
  getPriceFromOracle,
  fetchExternalApi,
  budgetStatus,
  vaultKeyNames,
  multiOperation,
  streamLogs,
} from "./functions";
export * from "./types";

export const VERSION = "1.0.0";
