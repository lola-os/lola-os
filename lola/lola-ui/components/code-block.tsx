"use client";

import * as React from "react";
import { Prism as SyntaxHighlighter } from "react-syntax-highlighter";
import { Check, Copy } from "lucide-react";
import { cn } from "@/lib/utils";

/**
 * A grayscale syntax theme — per branding.md, code blocks follow the
 * same "no colour other than grayscale" rule as the rest of the UI.
 * Token weight and style (italic for comments, bold for keywords) carry
 * the distinction color normally would.
 */
const grayscaleCodeTheme: Record<string, React.CSSProperties> = {
  'pre[class*="language-"]': { background: "transparent", color: "#e4e4e7" },
  'code[class*="language-"]': { background: "transparent", color: "#e4e4e7" },
  comment: { color: "#6b6b6f", fontStyle: "italic" },
  prolog: { color: "#6b6b6f" },
  doctype: { color: "#6b6b6f" },
  cdata: { color: "#6b6b6f" },
  punctuation: { color: "#9a9a9e" },
  property: { color: "#d4d4d8" },
  tag: { color: "#f4f4f5", fontWeight: 500 },
  boolean: { color: "#f4f4f5", fontWeight: 500 },
  number: { color: "#f4f4f5", fontWeight: 500 },
  constant: { color: "#f4f4f5", fontWeight: 500 },
  symbol: { color: "#d4d4d8" },
  selector: { color: "#e4e4e7" },
  "attr-name": { color: "#d4d4d8" },
  string: { color: "#fafafa" },
  char: { color: "#fafafa" },
  builtin: { color: "#f4f4f5", fontWeight: 500 },
  operator: { color: "#9a9a9e" },
  entity: { color: "#d4d4d8" },
  url: { color: "#d4d4d8", textDecoration: "underline" },
  keyword: { color: "#fafafa", fontWeight: 600 },
  function: { color: "#f4f4f5", fontWeight: 500 },
  "class-name": { color: "#f4f4f5", fontWeight: 500 },
  regex: { color: "#d4d4d8" },
  important: { color: "#fafafa", fontWeight: 600 },
};

export interface CodeBlockProps {
  code: string;
  language?: string;
  filename?: string;
  className?: string;
}

export function CodeBlock({ code, language = "python", filename, className }: CodeBlockProps) {
  const [copied, setCopied] = React.useState(false);

  async function handleCopy() {
    await navigator.clipboard.writeText(code);
    setCopied(true);
    setTimeout(() => setCopied(false), 1500);
  }

  return (
    <div
      className={cn(
        "group relative overflow-hidden rounded-xl border border-white/10 bg-[#0d0d0f] shadow-xl shadow-black/20 ring-1 ring-black/5",
        className
      )}
    >
      <div className="flex h-11 items-center justify-between border-b border-white/10 px-4">
        <div className="flex items-center gap-3">
          <div className="flex gap-1.5">
            <span className="h-2.5 w-2.5 rounded-full bg-white/15" />
            <span className="h-2.5 w-2.5 rounded-full bg-white/15" />
            <span className="h-2.5 w-2.5 rounded-full bg-white/15" />
          </div>
          {filename && <span className="text-xs font-medium text-white/40">{filename}</span>}
        </div>
        <button
          onClick={handleCopy}
          aria-label="Copy code"
          className="inline-flex h-7 w-7 items-center justify-center rounded-md text-white/40 transition-colors hover:bg-white/10 hover:text-white/80"
        >
          {copied ? <Check className="h-3.5 w-3.5" /> : <Copy className="h-3.5 w-3.5" />}
        </button>
      </div>
      <div className="overflow-x-auto">
        <SyntaxHighlighter
          language={language}
          style={grayscaleCodeTheme}
          customStyle={{
            margin: 0,
            padding: "1.25rem 1.5rem",
            background: "transparent",
            fontSize: "0.8125rem",
            lineHeight: "1.65",
          }}
          codeTagProps={{ style: { fontFamily: "var(--font-jetbrains-mono), monospace" } }}
          wrapLongLines={false}
        >
          {code}
        </SyntaxHighlighter>
      </div>
    </div>
  );
}
