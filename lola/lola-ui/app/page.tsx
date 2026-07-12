import Link from "next/link";
import {
  Network,
  Radio,
  UserCheck,
  ShieldAlert,
  History,
  Gift,
  ArrowRight,
  Check,
  Minus,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { CodeBlock } from "@/components/code-block";
import { HeroFade, ScrollFade } from "@/components/motion";
import { HowItWorksDiagram } from "@/components/how-it-works-diagram";

const HERO_CODE = `from lola_os import lola_tool, get_balance

@lola_tool
def check_balance(address: str) -> dict:
    return get_balance("ethereum", address)
`;

const FEATURES = [
  {
    icon: Network,
    title: "Works with any chain",
    description:
      "Ethereum, Polygon, Solana, and any EVM-compatible network — through one modular adapter interface.",
  },
  {
    icon: Radio,
    title: "Oracles and APIs",
    description:
      "Read Chainlink price feeds directly from their contracts, or fetch any REST API with rate limiting and retries built in.",
  },
  {
    icon: UserCheck,
    title: "Human-in-the-loop, your way",
    description:
      "A rich console prompt by default, or a local WebSocket feed if you'd rather build your own approval UI.",
  },
  {
    icon: ShieldAlert,
    title: "Budget safety",
    description:
      "A circuit breaker tracks gas and USD spend per session, and pauses writes automatically before you overspend.",
  },
  {
    icon: History,
    title: "Replay automation",
    description:
      "Write a structured plan.json, run it with lola replay, and get a full audit trail in a local SQLite registry.",
  },
  {
    icon: Gift,
    title: "Free forever",
    description:
      "Fully open source under Apache 2.0. No paid tiers, no hosted API, no license keys — ever.",
  },
];

const COMPARISON_ROWS = [
  { feature: "Runs entirely on your machine", lola: true, others: false },
  { feature: "No API keys or billing", lola: true, others: false },
  { feature: "Budget circuit breaker", lola: true, others: false },
  { feature: "Structured replay with an audit trail", lola: true, others: false },
  { feature: "Pre-flight ABI validation", lola: true, others: false },
  { feature: "Apache 2.0, no paid tier", lola: true, others: false },
];

export default function HomePage() {
  return (
    <>
      {/* Hero */}
      <section className="relative overflow-hidden">
        <div
          aria-hidden
          className="absolute inset-0 -z-10 [background-image:radial-gradient(circle_at_1px_1px,theme(colors.gray.300)_1px,transparent_0)] [background-size:32px_32px] opacity-[0.25] dark:[background-image:radial-gradient(circle_at_1px_1px,theme(colors.gray.700)_1px,transparent_0)] dark:opacity-[0.15]"
        />
        <div className="container pb-20 pt-20 sm:pb-28 sm:pt-28 lg:pb-36 lg:pt-36">
          <HeroFade>
            <div className="mx-auto max-w-3xl text-center">
              <div className="inline-flex items-center gap-2 rounded-full border border-gray-200 bg-white/60 px-3.5 py-1.5 text-xs font-medium text-gray-600 backdrop-blur dark:border-gray-800 dark:bg-gray-900/60 dark:text-gray-400">
                <span className="h-1.5 w-1.5 rounded-full bg-gray-900 dark:bg-gray-100" />
                Apache 2.0 · free forever · no hosted backend
              </div>
              <h1 className="mt-7 text-display-sm font-bold text-gray-900 dark:text-gray-50 sm:text-display-md lg:text-display-lg">
                Add one decorator.
                <br />
                Talk to any blockchain.
              </h1>
              <p className="mx-auto mt-6 max-w-xl text-lg leading-7 text-gray-500 dark:text-gray-400">
                LOLA OS connects AI agents to blockchains, oracles, and APIs — so you stay in your
                world and we handle the messy parts. No hype, no hosted backend. Infrastructure
                that just works.
              </p>
              <div className="mt-9 flex flex-wrap items-center justify-center gap-3">
                <Button size="lg" asChild>
                  <Link href="/docs/getting-started">
                    Get started
                    <ArrowRight className="h-4 w-4" />
                  </Link>
                </Button>
                <Button size="lg" variant="outline" asChild>
                  <Link href="https://github.com/lola-os">View on GitHub</Link>
                </Button>
              </div>
            </div>
          </HeroFade>

          <HeroFade>
            <div className="mx-auto mt-16 max-w-2xl">
              <CodeBlock code={HERO_CODE} language="python" filename="agent.py" />
            </div>
          </HeroFade>
        </div>
      </section>

      {/* Features */}
      <section className="border-t border-gray-200/70 py-24 dark:border-gray-800/70 sm:py-32">
        <div className="container">
          <ScrollFade>
            <div className="mx-auto max-w-2xl text-center">
              <h2 className="text-3xl font-bold sm:text-4xl">
                Everything you need. Nothing you don&apos;t.
              </h2>
              <p className="mt-4 text-base leading-7 text-gray-500 dark:text-gray-400">
                Multi-chain, oracles, human approval, budget safety, and replay automation —
                operationally mature from the first commit.
              </p>
            </div>
          </ScrollFade>

          <div className="mt-16 grid gap-x-8 gap-y-12 sm:grid-cols-2 lg:grid-cols-3">
            {FEATURES.map((feature, i) => {
              const Icon = feature.icon;
              return (
                <ScrollFade key={feature.title} delay={i * 0.06}>
                  <div className="flex h-10 w-10 items-center justify-center rounded-lg border border-gray-200 bg-white dark:border-gray-800 dark:bg-gray-900">
                    <Icon className="h-[1.1rem] w-[1.1rem] text-gray-700 dark:text-gray-300" strokeWidth={1.75} />
                  </div>
                  <h3 className="mt-4 font-semibold text-gray-900 dark:text-gray-50">{feature.title}</h3>
                  <p className="mt-2 text-sm leading-6 text-gray-500 dark:text-gray-400">
                    {feature.description}
                  </p>
                </ScrollFade>
              );
            })}
          </div>
        </div>
      </section>

      {/* How it works */}
      <section className="border-t border-gray-200/70 py-24 dark:border-gray-800/70 sm:py-32">
        <div className="container">
          <ScrollFade>
            <div className="mx-auto max-w-2xl text-center">
              <h2 className="text-3xl font-bold sm:text-4xl">How it works</h2>
              <p className="mt-4 text-base leading-7 text-gray-500 dark:text-gray-400">
                Your agent calls a decorated function. LOLA OS handles validation, approval,
                signing, and broadcasting — and records everything locally.
              </p>
            </div>
          </ScrollFade>
          <ScrollFade delay={0.1}>
            <div className="mx-auto mt-16 max-w-5xl">
              <HowItWorksDiagram />
            </div>
          </ScrollFade>
        </div>
      </section>

      {/* Comparison */}
      <section className="border-t border-gray-200/70 py-24 dark:border-gray-800/70 sm:py-32">
        <div className="container">
          <ScrollFade>
            <div className="mx-auto max-w-2xl text-center">
              <h2 className="text-3xl font-bold sm:text-4xl">LOLA OS vs. a typical hosted SDK</h2>
            </div>
          </ScrollFade>
          <ScrollFade delay={0.1}>
            <div className="mx-auto mt-12 max-w-2xl overflow-hidden rounded-2xl border border-gray-200 dark:border-gray-800">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-gray-200 dark:border-gray-800">
                    <th className="px-6 py-4 text-left font-medium text-gray-500 dark:text-gray-400">Feature</th>
                    <th className="px-6 py-4 text-center font-semibold text-gray-900 dark:text-gray-50">LOLA OS</th>
                    <th className="px-6 py-4 text-center font-medium text-gray-400 dark:text-gray-600">
                      Hosted SDK
                    </th>
                  </tr>
                </thead>
                <tbody>
                  {COMPARISON_ROWS.map((row, i) => (
                    <tr
                      key={row.feature}
                      className={i !== COMPARISON_ROWS.length - 1 ? "border-b border-gray-100 dark:border-gray-900" : ""}
                    >
                      <td className="px-6 py-4 text-gray-700 dark:text-gray-300">{row.feature}</td>
                      <td className="px-6 py-4 text-center">
                        <Check className="mx-auto h-4 w-4 text-gray-900 dark:text-gray-50" strokeWidth={2.25} />
                      </td>
                      <td className="px-6 py-4 text-center">
                        <Minus className="mx-auto h-4 w-4 text-gray-300 dark:text-gray-700" />
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </ScrollFade>
        </div>
      </section>

      {/* CTA */}
      <section className="border-t border-gray-200/70 py-24 dark:border-gray-800/70 sm:py-32">
        <div className="container">
          <ScrollFade>
            <div className="mx-auto max-w-xl text-center">
              <h2 className="text-3xl font-bold sm:text-4xl">
                Five minutes. One decorator. Any blockchain.
              </h2>
              <p className="mt-4 text-base leading-7 text-gray-500 dark:text-gray-400">
                Read the docs, run <code className="rounded-md border border-gray-200 bg-gray-100 px-1.5 py-0.5 text-[0.85em] font-medium text-gray-700 dark:border-gray-800 dark:bg-gray-900 dark:text-gray-300">lola doctor</code> to check your setup, and you&apos;re live.
              </p>
              <div className="mt-8 flex flex-wrap items-center justify-center gap-3">
                <Button size="lg" asChild>
                  <Link href="/docs/getting-started">
                    Get started
                    <ArrowRight className="h-4 w-4" />
                  </Link>
                </Button>
                <Button size="lg" variant="outline" asChild>
                  <Link href="https://github.com/lola-os">GitHub</Link>
                </Button>
              </div>
            </div>
          </ScrollFade>
        </div>
      </section>
    </>
  );
}
