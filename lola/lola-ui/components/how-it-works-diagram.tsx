import { ShieldCheck, GitBranch, Database, Bot } from "lucide-react";

const STEPS = [
  {
    icon: Bot,
    label: "Your agent",
    detail: "calls a @lola_tool function",
  },
  {
    icon: ShieldCheck,
    label: "lola-core validates",
    detail: "checks the ABI, checks the budget",
  },
  {
    icon: GitBranch,
    label: "You approve",
    detail: "console prompt or your own UI",
  },
  {
    icon: Database,
    label: "Chain + registry",
    detail: "signed, broadcast, recorded locally",
  },
];

/**
 * A quiet, icon-anchored step flow — grayscale and geometric per
 * branding.md section 7, intentionally restrained rather than a
 * boxes-and-arrows flowchart.
 */
export function HowItWorksDiagram() {
  return (
    <div className="relative">
      <div
        aria-hidden
        className="absolute left-0 right-0 top-7 hidden h-px bg-gray-200 dark:bg-gray-800 lg:block"
      />
      <div className="grid gap-10 sm:grid-cols-2 lg:grid-cols-4 lg:gap-8">
        {STEPS.map((step, i) => {
          const Icon = step.icon;
          return (
            <div key={step.label} className="relative">
              <div className="relative z-10 flex h-14 w-14 items-center justify-center rounded-full border border-gray-200 bg-gray-50 dark:border-gray-800 dark:bg-gray-950">
                <Icon className="h-5 w-5 text-gray-700 dark:text-gray-300" strokeWidth={1.75} />
              </div>
              <div className="mt-5">
                <div className="text-[0.7rem] font-semibold uppercase tracking-wide text-gray-400 dark:text-gray-600">
                  Step {i + 1}
                </div>
                <div className="mt-1.5 font-semibold text-gray-900 dark:text-gray-50">{step.label}</div>
                <div className="mt-1 text-sm leading-6 text-gray-500 dark:text-gray-400">{step.detail}</div>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}
