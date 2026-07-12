/**
 * `lolaTool`: marks a function as a LOLA tool.
 *
 * TypeScript/JavaScript doesn't have Python's bare-decorator ergonomics
 * for plain functions (decorators only attach to class members in the
 * stable spec), so this is offered as a higher-order function — the
 * idiomatic equivalent for wrapping standalone functions:
 *
 *   const checkBalance = lolaTool(async (address: string) => {
 *     return getBalance("ethereum", address);
 *   }, { name: "check_balance" });
 *
 * A `@lolaTool()` method decorator is also provided for classes (e.g. a
 * LangChain.js Tool subclass) where the standard decorator syntax does
 * apply.
 */

import { getClient } from "./client";
import { Overrides, withOverrides } from "./context";

export interface LolaToolOptions {
  name?: string;
  description?: string;
  configOverrides?: Overrides;
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
type AnyFn = (...args: any[]) => any;

export interface LolaToolMetadata {
  lolaTool: true;
  lolaName: string;
  lolaDescription: string;
}

export type WrappedTool<F extends AnyFn> = F & LolaToolMetadata;

/**
 * Wraps `fn` (sync or async) as a named LOLA tool: ensures the LolaCore
 * client is started before the function runs, and — if
 * `options.configOverrides` is set — scopes those overrides to every
 * invocation of the wrapped function.
 */
export function lolaTool<F extends AnyFn>(fn: F, options: LolaToolOptions = {}): WrappedTool<F> {
  const name = options.name ?? fn.name ?? "anonymous_tool";
  const description = options.description ?? "";

  const wrapped = ((...args: Parameters<F>) => {
    getClient();
    if (options.configOverrides) {
      return withOverrides(options.configOverrides, () => fn(...args));
    }
    return fn(...args);
  }) as WrappedTool<F>;

  wrapped.lolaTool = true;
  wrapped.lolaName = name;
  wrapped.lolaDescription = description;
  Object.defineProperty(wrapped, "name", { value: name, configurable: true });

  return wrapped;
}

/**
 * Method decorator form, for classes:
 *
 *   class MyTools {
 *     @LolaTool({ name: "check_balance" })
 *     async checkBalance(address: string) {
 *       return getBalance("ethereum", address);
 *     }
 *   }
 *
 * Compatible with the legacy (experimentalDecorators) TypeScript
 * decorator proposal, which is what most current Node/TS agent
 * frameworks (LangChain.js, etc.) target as of this writing.
 */
export function LolaTool(options: LolaToolOptions = {}) {
  return function decorate(
    target: object,
    propertyKey: string,
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    descriptor: TypedPropertyDescriptor<AnyFn>
  ): TypedPropertyDescriptor<AnyFn> | void {
    const original = descriptor.value;
    if (!original) return descriptor;

    descriptor.value = function (this: unknown, ...args: unknown[]) {
      getClient();
      if (options.configOverrides) {
        return withOverrides(options.configOverrides, () => original.apply(this, args));
      }
      return original.apply(this, args);
    };
    return descriptor;
  };
}
