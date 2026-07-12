export interface DocsNavLink {
  title: string;
  href: string;
}

export interface DocsNavSection {
  title: string;
  items: DocsNavLink[];
}

export const DOCS_NAV: DocsNavSection[] = [
  {
    title: "Introduction",
    items: [{ title: "Overview", href: "/docs" }],
  },
  {
    title: "Start here",
    items: [
      { title: "Getting Started", href: "/docs/getting-started" },
      { title: "Configuration Reference", href: "/docs/configuration" },
      { title: "Security Guide", href: "/docs/security" },
    ],
  },
  {
    title: "Operational tooling",
    items: [
      { title: "lola replay", href: "/docs/operational/replay" },
      { title: "lola doctor", href: "/docs/operational/doctor" },
      { title: "lola registry", href: "/docs/operational/registry" },
      { title: "lola metrics", href: "/docs/operational/metrics" },
    ],
  },
  {
    title: "Examples",
    items: [
      { title: "Python in 5 minutes", href: "/docs/examples/python-5min" },
      { title: "Replay plan walkthrough", href: "/docs/examples/replay-plan" },
      { title: "CrewAI integration", href: "/docs/examples/crewai" },
      { title: "LangChain integration", href: "/docs/examples/langchain" },
      { title: "Custom trading bot with HITL", href: "/docs/examples/custom-bot" },
    ],
  },
  {
    title: "Reference",
    items: [
      { title: "API Reference", href: "/docs/api-reference" },
      { title: "FAQ", href: "/docs/faq" },
    ],
  },
];
