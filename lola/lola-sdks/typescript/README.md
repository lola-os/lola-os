# lola-os (TypeScript/JavaScript SDK)

![npm](https://img.shields.io/npm/v/lola-os?labelColor=252525&color=555555&style=flat-square)
![downloads](https://img.shields.io/npm/dm/lola-os?labelColor=252525&color=555555&style=flat-square)
![types](https://img.shields.io/npm/types/lola-os?labelColor=252525&color=555555&style=flat-square)
![License](https://img.shields.io/npm/l/lola-os?labelColor=252525&color=555555&style=flat-square)

Add `lolaTool` to any function and it can talk to blockchains, oracles,
and APIs through the local `lola-core` engine.

```ts
import { lolaTool, getBalance } from "lola-os";

const checkBalance = lolaTool(async (address: string) => {
  return getBalance("ethereum", address);
});

console.log(await checkBalance("0x..."));
```

## Status of this build

This is the one SDK in the project that was **fully compiled and tested**
in the same environment it was written in:

```
$ tsc -p tsconfig.json   ->  clean compile, zero errors
$ tsc -p tsconfig.test.json && node run-tests.js
21/21 passed
```

(The `tsc` available in that environment happened to be a newer 6.x
release than the `^5.4.0` this package declares as a dependency, which
required a couple of compatibility flags during verification — those
flags are **not** baked into the shipped `tsconfig.json`, since v5.4 users
don't need them. If you hit a `moduleResolution=node10 is deprecated`
error on a very new TypeScript install, add
`"ignoreDeprecations": "6.0"` to your tsconfig.)

What's **not** verified end-to-end is the real subprocess boundary in
`client.ts` talking to an actual compiled `lola-core` binary, since no
such binary exists in the build environment. The JSON-RPC wire format
matches `lola-core/internal/jsonrpc` exactly (newline-delimited JSON-RPC
2.0), so this should work once you have a built `lola-core` — see the root
`lola-core/README.md`.

## Installation

```bash
npm install
npm run build
```

This package does **not** bundle a `lola-core` binary (see `bin/README.txt`).
Build one and point the SDK at it:

```bash
cd ../../lola-core
go build -o bin/lola ./cmd/lola
export LOLA_CORE_BINARY=$(pwd)/bin/lola
export LOLA_VAULT_PASSPHRASE="your-passphrase"
```

## Quickstart

```ts
import * as lola from "lola-os";

const balance = await lola.getBalance("ethereum", "0x...");
console.log(balance); // { address, token, raw_value, decimals, symbol }

const supply = await lola.callContract("ethereum", "0xTokenAddress", "totalSupply");

const receipt = await lola.executeContract(
  "ethereum",
  "0xYourAddress",
  "0xTokenAddress",
  "transfer",
  {
    args: ["0xRecipient", "1000000000000000000"],
    abi: '[{"constant":false,"inputs":[{"name":"_to","type":"address"},{"name":"_value","type":"uint256"}],"name":"transfer","outputs":[{"name":"","type":"bool"}],"type":"function"}]',
    keyName: "deployer",
  }
);
console.log(receipt); // { tx_hash, status }
```

## Context overrides

Backed by Node's `AsyncLocalStorage`, so overrides correctly scope across
`await` boundaries without leaking into concurrent requests:

```ts
await lola.withOverrides({ chain: "polygon", budgetMaxUsd: 5.0 }, async () => {
  const result = await lola.executeContract(...);
});
```

Or scope overrides to one specific tool:

```ts
const polygonTool = lolaTool(myFn, { configOverrides: { chain: "polygon" } });
```

### Browser support note

`AsyncLocalStorage` comes from `node:async_hooks` and isn't available in
browsers. This SDK detects that and falls back to a simple stack-based
implementation, which is correct for synchronous nesting but can't
guarantee isolation across concurrently-running async operations in a
browser context. The WebAssembly browser build mentioned in the blueprint
would additionally need its own transport layer (no subprocess spawning
in a browser) — that transport is not implemented in this build; `client.ts`
as written assumes a Node.js `child_process` environment.

## The `lolaTool` function and decorator

```ts
// Higher-order function (works anywhere):
const checkBalance = lolaTool(async (address: string) => getBalance("ethereum", address));

// Method decorator (classes, experimentalDecorators):
class MyTools {
  @LolaTool({ name: "check_balance" })
  async checkBalance(address: string) {
    return getBalance("ethereum", address);
  }
}
```

## Error handling

```ts
import { executeContract, BudgetExceededError, ABIMismatchError, RPCConnectionError } from "lola-os";

try {
  await executeContract(...);
} catch (e) {
  if (e instanceof BudgetExceededError) console.log("Over budget:", e.message);
  else if (e instanceof ABIMismatchError) console.log("Bad method/args:", e.message);
  else if (e instanceof RPCConnectionError) console.log("Network issue:", e.message);
  else throw e;
}
```

## Streaming logs

```ts
for await (const entry of lola.streamLogs()) {
  console.log(entry.level, entry.message);
}
```

## Running the tests

```bash
npm test
```

No external test framework dependency is required — `run-tests.js` is a
small custom runner since this environment had no network access to
install Jest/Vitest. Test bodies use plain Node `assert`, so they're
drop-in compatible with Jest/Mocha/Vitest if you'd rather use one of
those; just point it at `test/*.test.ts`.

## API reference

| Function | Description |
|---|---|
| `getBalance(chain, address)` | Native balance |
| `getTokenBalance(chain, address, tokenAddress)` | ERC20/SPL balance |
| `callContract(chain, contract, method, args, abi)` | Read-only call |
| `executeContract(chain, from, contract, method, options)` | State-changing call |
| `sendTransaction(chain, from, to, valueWei, options)` | Native transfer |
| `transferToken(chain, from, to, token, amountRaw, options)` | Token transfer |
| `swapTokens(chain, from, routerContract, method, args, abi, options)` | DEX swap via router call |
| `getPriceFromOracle(chain, pair)` | Chainlink price read |
| `fetchExternalApi(url)` | Rate-limited, retrying REST GET |
| `multiOperation(operations)` | Run a client-side batch of the above |
| `streamLogs()` | Async generator of structured log entries |
| `budgetStatus()` | Current session spend snapshot |
| `vaultKeyNames()` | List stored key names (never values) |

---

Built and maintained by **[0xSemantic](https://github.com/0xSemantic)** —
developer and visionary behind LOLA OS. Licensed under Apache-2.0.
