/**
 * Global SDK configuration: where the lola-core binary lives, and the
 * platform/arch detection used by binary management.
 */

import * as fs from "fs";
import * as os from "os";
import * as path from "path";

export interface GlobalConfig {
  binaryPath?: string;
  vaultPassphrase?: string;
  startupTimeoutMs: number;
  requestTimeoutMs: number;
}

const globalConfig: GlobalConfig = {
  vaultPassphrase: process.env.LOLA_VAULT_PASSPHRASE,
  startupTimeoutMs: 15000,
  requestTimeoutMs: 60000,
};

export function getGlobalConfig(): GlobalConfig {
  return globalConfig;
}

/** Returns the expected lola-core binary filename for the current
 * platform/architecture, matching lola-core's cross-compilation targets
 * (see lola-core/README.md). */
export function platformBinaryName(): string {
  const platform = os.platform();
  const archRaw = os.arch();

  let arch: string;
  if (archRaw === "x64") {
    arch = "amd64";
  } else if (archRaw === "arm64") {
    arch = "arm64";
  } else {
    arch = archRaw;
  }

  if (platform === "darwin") {
    return `lola-darwin-${arch}`;
  }
  if (platform === "win32") {
    return `lola-windows-${arch}.exe`;
  }
  return `lola-linux-${arch}`;
}

export function bundledBinaryPath(): string {
  return path.join(__dirname, "..", "bin", platformBinaryName());
}

function isExecutable(p: string): boolean {
  try {
    fs.accessSync(p, fs.constants.X_OK);
    return fs.statSync(p).isFile();
  } catch {
    return false;
  }
}

function findOnPath(binaryName: string): string | undefined {
  const pathEnv = process.env.PATH || "";
  const sep = os.platform() === "win32" ? ";" : ":";
  for (const dir of pathEnv.split(sep)) {
    if (!dir) continue;
    const candidate = path.join(dir, binaryName);
    if (isExecutable(candidate)) {
      return candidate;
    }
  }
  return undefined;
}

/**
 * Resolution order:
 *   1. An explicit path passed to LolaCore({ binaryPath }).
 *   2. The LOLA_CORE_BINARY environment variable.
 *   3. A binary bundled alongside the installed package (bin/).
 *   4. `lola` (or `lola.exe` on Windows) found on $PATH.
 *
 * Throws an Error with an actionable message if none is found.
 */
export function resolveBinaryPath(explicit?: string): string {
  const candidates: string[] = [];
  if (explicit) candidates.push(explicit);
  if (globalConfig.binaryPath) candidates.push(globalConfig.binaryPath);
  if (process.env.LOLA_CORE_BINARY) candidates.push(process.env.LOLA_CORE_BINARY);

  const bundled = bundledBinaryPath();
  candidates.push(bundled);

  for (const candidate of candidates) {
    if (isExecutable(candidate)) {
      return candidate;
    }
  }

  const onPath = findOnPath(os.platform() === "win32" ? "lola.exe" : "lola");
  if (onPath) {
    return onPath;
  }

  throw new Error(
    "Could not locate a lola-core binary.\n" +
      `Looked for: ${candidates.join(", ")}, and 'lola' on $PATH.\n\n` +
      "Fix this by either:\n" +
      "  1. Building lola-core yourself: `cd lola-core && go build -o bin/lola ./cmd/lola`\n" +
      "     then set LOLA_CORE_BINARY=/path/to/bin/lola, or\n" +
      `  2. Placing a prebuilt binary at ${bundled} (this package ships without one — see lola-sdks/typescript/README.md), or\n` +
      "  3. Installing `lola` on your $PATH."
  );
}
