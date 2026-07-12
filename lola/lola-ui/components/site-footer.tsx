import Link from "next/link";
import { LolaMark } from "@/components/lola-mark";

const FOOTER_SECTIONS = [
  {
    title: "Documentation",
    links: [
      { href: "/docs/getting-started", label: "Getting started" },
      { href: "/docs/configuration", label: "Configuration" },
      { href: "/docs/security", label: "Security" },
      { href: "/docs/faq", label: "FAQ" },
    ],
  },
  {
    title: "Operational tooling",
    links: [
      { href: "/docs/operational/replay", label: "lola replay" },
      { href: "/docs/operational/doctor", label: "lola doctor" },
      { href: "/docs/operational/registry", label: "lola registry" },
      { href: "/docs/operational/metrics", label: "lola metrics" },
    ],
  },
  {
    title: "Examples",
    links: [
      { href: "/docs/examples/python-5min", label: "Python in 5 minutes" },
      { href: "/docs/examples/replay-plan", label: "Replay plan walkthrough" },
      { href: "/docs/examples/crewai", label: "CrewAI integration" },
      { href: "/docs/examples/langchain", label: "LangChain integration" },
    ],
  },
  {
    title: "Project",
    links: [
      { href: "https://github.com/lola-os", label: "GitHub" },
      { href: "https://github.com/lola-os/lola-core/blob/main/LICENSE", label: "Apache 2.0 License" },
    ],
  },
];

export function SiteFooter() {
  return (
    <footer className="border-t border-gray-200/70 dark:border-gray-800/70">
      <div className="container py-20">
        <div className="grid grid-cols-2 gap-x-8 gap-y-12 sm:grid-cols-3 lg:grid-cols-[1.3fr_repeat(4,1fr)] lg:gap-x-10">
          <div className="col-span-2 sm:col-span-3 lg:col-span-1">
            <div className="flex items-center gap-2">
              <LolaMark className="h-5 w-5 text-gray-900 dark:text-gray-50" />
              <span className="text-[0.95rem] font-bold tracking-tight text-gray-900 dark:text-gray-50">LOLA OS</span>
            </div>
            <p className="mt-4 max-w-[15rem] text-sm leading-6 text-gray-500 dark:text-gray-500">
              The bridge between AI agents and blockchains. Apache 2.0, free forever.
            </p>
          </div>
          {FOOTER_SECTIONS.map((section) => (
            <div key={section.title}>
              <h3 className="text-xs font-semibold uppercase tracking-wide text-gray-400 dark:text-gray-500">
                {section.title}
              </h3>
              <ul className="mt-4 space-y-3">
                {section.links.map((link) => (
                  <li key={link.href}>
                    <Link
                      href={link.href}
                      className="text-sm text-gray-600 transition-colors hover:text-gray-900 dark:text-gray-400 dark:hover:text-gray-100"
                    >
                      {link.label}
                    </Link>
                  </li>
                ))}
              </ul>
            </div>
          ))}
        </div>
        <div className="mt-16 flex flex-col gap-2 border-t border-gray-200/70 pt-8 text-xs text-gray-400 dark:border-gray-800/70 dark:text-gray-600 sm:flex-row sm:items-center sm:justify-between">
          <span>© {new Date().getFullYear()} LOLA OS contributors.</span>
          <span>Apache 2.0 License · No ads, no telemetry, no hosted backend.</span>
        </div>
      </div>
    </footer>
  );
}
