import { type ClassValue, clsx } from "clsx";
import { twMerge } from "tailwind-merge";

/** Merges conditional class names and resolves Tailwind conflicts —
 * the standard shadcn/ui utility, used by every component in
 * components/ui/. */
export function cn(...inputs: ClassValue[]): string {
  return twMerge(clsx(inputs));
}
