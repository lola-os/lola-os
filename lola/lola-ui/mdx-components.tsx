import type { MDXComponents } from "mdx/types";
import { CodeBlock } from "@/components/code-block";

/**
 * Required by @next/mdx: this is the global override point for every
 * element MDX content can render. We route `pre > code` (i.e. fenced
 * ```lang blocks) through our grayscale CodeBlock with copy-to-clipboard,
 * and leave everything else as default HTML so docs-prose CSS (see
 * globals.css) applies normally.
 */
export function useMDXComponents(components: MDXComponents): MDXComponents {
  return {
    pre: ({ children }) => {
      // `children` is a <code> element with a `className` like
      // "language-python" and its text content as children.
      const codeElement = children as React.ReactElement<{ className?: string; children?: string }>;
      const className = codeElement?.props?.className ?? "";
      const language = className.replace("language-", "") || "text";
      const code = String(codeElement?.props?.children ?? "").replace(/\n$/, "");
      return <CodeBlock code={code} language={language} />;
    },
    ...components,
  };
}
