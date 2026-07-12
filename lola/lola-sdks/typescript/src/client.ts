/**
 * The LolaCore client: spawns (or connects to) the lola-core binary and
 * speaks newline-delimited JSON-RPC 2.0 over its stdin/stdout, matching
 * lola-core/internal/jsonrpc exactly.
 *
 * A process-wide singleton is exposed via getClient() for convenience —
 * most callers never need to construct a LolaCore directly.
 */

import { ChildProcess, spawn } from "child_process";
import * as readline from "readline";
import { EventEmitter } from "events";

import { getGlobalConfig, resolveBinaryPath } from "./config";
import { fromRpcError, LolaError, RPCConnectionError } from "./errors";
import { LogEntry } from "./types";

interface JsonRpcRequest {
  jsonrpc: "2.0";
  id: number;
  method: string;
  params?: unknown;
}

interface JsonRpcError {
  code: number;
  message: string;
  data?: unknown;
}

interface JsonRpcResponse {
  jsonrpc: "2.0";
  id: number;
  result?: unknown;
  error?: JsonRpcError;
}

export interface LolaCoreOptions {
  binaryPath?: string;
  vaultPassphrase?: string;
  extraArgs?: string[];
  autoStart?: boolean;
  startupTimeoutMs?: number;
}

interface PendingCall {
  resolve: (value: unknown) => void;
  reject: (err: Error) => void;
}

let nextId = 1;

export class LolaCore {
  private binaryPath?: string;
  private vaultPassphrase?: string;
  private extraArgs: string[];
  private startupTimeoutMs: number;

  private proc?: ChildProcess;
  private pending: Map<number, PendingCall> = new Map();
  private started = false;
  private startPromise?: Promise<void>;
  private logEmitter = new EventEmitter();

  constructor(options: LolaCoreOptions = {}) {
    const cfg = getGlobalConfig();
    this.binaryPath = options.binaryPath ?? cfg.binaryPath;
    this.vaultPassphrase = options.vaultPassphrase ?? cfg.vaultPassphrase;
    this.extraArgs = options.extraArgs ?? [];
    this.startupTimeoutMs = options.startupTimeoutMs ?? cfg.startupTimeoutMs;

    if (options.autoStart !== false) {
      // Fire-and-forget; call() awaits `ensureStarted()` so callers don't
      // need to manually await construction.
      this.startPromise = this.start();
    }
  }

  /** Starts the lola-core subprocess if it isn't already running. Safe to
   * call multiple times; subsequent calls are no-ops while/after the
   * first start completes. */
  async start(): Promise<void> {
    if (this.started) return;
    if (this.startPromise && this.proc) {
      await this.startPromise;
      return;
    }

    const resolvedPath = resolveBinaryPath(this.binaryPath);
    const env = { ...process.env };
    if (this.vaultPassphrase) {
      env.LOLA_VAULT_PASSPHRASE = this.vaultPassphrase;
    }

    this.proc = spawn(resolvedPath, ["serve", ...this.extraArgs], {
      env,
      stdio: ["pipe", "pipe", "pipe"],
    });

    const stdout = this.proc.stdout;
    const stderr = this.proc.stderr;
    if (!stdout || !stderr || !this.proc.stdin) {
      throw new RPCConnectionError("Failed to open stdio pipes for lola-core");
    }

    const rlOut = readline.createInterface({ input: stdout });
    rlOut.on("line", (line) => this.handleLine(line));

    const rlErr = readline.createInterface({ input: stderr });
    rlErr.on("line", (line) => this.handleLogLine(line));

    this.proc.on("exit", (code) => {
      this.started = false;
      for (const [, pending] of this.pending) {
        pending.reject(new RPCConnectionError(`lola-core exited (code ${code}) before responding`));
      }
      this.pending.clear();
    });

    await this.waitForReady();
    this.started = true;
  }

  private async waitForReady(): Promise<void> {
    const deadline = Date.now() + this.startupTimeoutMs;
    let lastErr: unknown;
    while (Date.now() < deadline) {
      if (this.proc?.exitCode !== null && this.proc?.exitCode !== undefined) {
        throw new RPCConnectionError(`lola-core exited immediately (code ${this.proc.exitCode})`);
      }
      try {
        await this.callRaw("budget_status", {}, 1500);
        return;
      } catch (err) {
        lastErr = err;
        await sleep(200);
      }
    }
    throw new RPCConnectionError(`lola-core did not become ready within ${this.startupTimeoutMs}ms: ${lastErr}`);
  }

  /** Terminates the subprocess. */
  async stop(): Promise<void> {
    if (this.proc && this.proc.exitCode === null) {
      this.proc.kill();
    }
    this.started = false;
  }

  private handleLine(line: string): void {
    if (!line.trim()) return;
    let msg: JsonRpcResponse;
    try {
      msg = JSON.parse(line);
    } catch {
      return;
    }
    const pending = this.pending.get(msg.id);
    if (!pending) return;
    this.pending.delete(msg.id);

    if (msg.error) {
      pending.reject(fromRpcError(msg.error.message, msg.error.code, msg.error.data));
    } else {
      pending.resolve(msg.result);
    }
  }

  private handleLogLine(line: string): void {
    if (!line.trim()) return;
    let entry: LogEntry;
    try {
      entry = JSON.parse(line);
    } catch {
      entry = { raw: line };
    }
    this.logEmitter.emit("log", entry);
  }

  /** Subscribes to structured log entries. Returns an unsubscribe function. */
  onLog(handler: (entry: LogEntry) => void): () => void {
    this.logEmitter.on("log", handler);
    return () => this.logEmitter.off("log", handler);
  }

  private async ensureStarted(): Promise<void> {
    if (this.started) return;
    if (this.startPromise) {
      await this.startPromise;
      return;
    }
    await this.start();
  }

  private callRaw(method: string, params: unknown, timeoutMs: number): Promise<unknown> {
    return new Promise((resolve, reject) => {
      if (!this.proc || !this.proc.stdin) {
        reject(new RPCConnectionError("lola-core process is not running"));
        return;
      }
      const id = nextId++;
      const timer = setTimeout(() => {
        this.pending.delete(id);
        reject(new RPCConnectionError(`lola-core did not respond to '${method}' within ${timeoutMs}ms`));
      }, timeoutMs);

      this.pending.set(id, {
        resolve: (value) => {
          clearTimeout(timer);
          resolve(value);
        },
        reject: (err) => {
          clearTimeout(timer);
          reject(err);
        },
      });

      const payload: JsonRpcRequest = { jsonrpc: "2.0", id, method, params };
      this.proc.stdin.write(JSON.stringify(payload) + "\n", (err) => {
        if (err) {
          this.pending.delete(id);
          clearTimeout(timer);
          reject(new RPCConnectionError(`lola-core pipe write failed: ${err.message}`));
        }
      });
    });
  }

  /** Performs a JSON-RPC call, starting the subprocess first if needed.
   * Rejects with a LolaError subclass on failure. */
  async call<T = unknown>(method: string, params: Record<string, unknown> = {}, timeoutMs?: number): Promise<T> {
    await this.ensureStarted();
    const cfg = getGlobalConfig();
    return this.callRaw(method, params, timeoutMs ?? cfg.requestTimeoutMs) as Promise<T>;
  }
}

function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

// -- process-wide singleton ------------------------------------------------

let singleton: LolaCore | undefined;

/** Returns the process-wide LolaCore singleton, creating (and starting)
 * it on first use. Pass options only on the very first call. */
export function getClient(options?: LolaCoreOptions): LolaCore {
  if (!singleton) {
    singleton = new LolaCore(options);
  }
  return singleton;
}

/** Stops and clears the singleton client. Mainly useful for tests. */
export async function resetClient(): Promise<void> {
  if (singleton) {
    await singleton.stop();
  }
  singleton = undefined;
}

export { LolaError };
