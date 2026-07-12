import * as React from "react";

/**
 * The LOLA OS mark: two parallel horizontal lines (AI, top; blockchain,
 * bottom) connected by a diagonal slash (LOLA, the bridge). Per
 * branding.md: always grayscale, never colored, distorted, or shadowed.
 * Uses `currentColor` so it inherits gray-900 (light) / gray-100 (dark)
 * from its container automatically.
 */
export function LolaMark({ className }: { className?: string }) {
  return (
    <svg
      viewBox="0 0 48 48"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      className={className}
      role="img"
      aria-label="LOLA OS"
    >
      <line x1="6" y1="14" x2="30" y2="14" stroke="currentColor" strokeWidth="3" strokeLinecap="round" />
      <line x1="18" y1="34" x2="42" y2="34" stroke="currentColor" strokeWidth="3" strokeLinecap="round" />
      <line x1="14" y1="14" x2="34" y2="34" stroke="currentColor" strokeWidth="3" strokeLinecap="round" />
    </svg>
  );
}
