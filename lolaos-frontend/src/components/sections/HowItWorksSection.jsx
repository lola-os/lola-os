import { useState } from 'react'
import { motion } from 'framer-motion'
import { Boxes, Play, Wand2 } from 'lucide-react'
import SectionHeading from '../ui/SectionHeading'
import CodeBlock from '../ui/CodeBlock'

const STEPS = [
  {
    id: 'decorate',
    icon: Wand2,
    title: 'Decorate a function',
    description: 'Add @lola_tool to any function. LOLA detects the addresses, chains, and calls it needs.',
    filename: 'step_1_decorate.py',
    code: `from lola_os import lola_tool, get_balance

@lola_tool
def check_balance(address: str) -> dict:
    return get_balance("ethereum", address)`,
  },
  {
    id: 'run',
    icon: Play,
    title: 'Run your agent as usual',
    description: 'Pass the tool to CrewAI, LangChain, or your own loop. Nothing about your setup changes.',
    filename: 'step_2_run.py',
    code: `# Your existing agent, unchanged.
agent = Agent(
    role="Treasury analyst",
    tools=[check_balance],
)

result = agent.run("How much ETH does 0x742d… hold?")`,
  },
  {
    id: 'result',
    icon: Boxes,
    title: 'Get live onchain data',
    description: 'The local Go engine resolves the call, applies your guardrails, and returns real chain data.',
    filename: 'result.json',
    code: `{
  "address": "0x742d…",
  "chain": "ethereum",
  "balance": "1.542",
  "symbol": "ETH",
  "block": 20431187
}`,
  },
]

function HowItWorksSection() {
  const [active, setActive] = useState(0)
  const current = STEPS[active]

  return (
    <section id="how-it-works" className="bg-white py-20 dark:bg-gray-900 md:py-28">
      <div className="section-shell">
        <SectionHeading
          eyebrow="How it works"
          title="Three steps to onchain agents"
          description="No migration, no new framework to learn. Decorate, run, and read the data back."
        />

        <div className="mt-14 grid grid-cols-1 gap-8 lg:grid-cols-[minmax(0,1fr)_1.35fr]">
          <div className="flex flex-col gap-3">
            {STEPS.map((step, index) => {
              const Icon = step.icon
              const isActive = index === active
              return (
                <button
                  key={step.id}
                  type="button"
                  onClick={() => setActive(index)}
                  className={`flex w-full items-start gap-4 rounded-2xl border p-5 text-left transition-all ${
                    isActive
                      ? 'border-gray-300 bg-gray-50 shadow-sm dark:border-gray-700 dark:bg-gray-950'
                      : 'border-transparent hover:bg-gray-50 dark:hover:bg-gray-950/60'
                  }`}
                >
                  <span
                    className={`inline-flex h-10 w-10 shrink-0 items-center justify-center rounded-lg transition-colors ${
                      isActive
                        ? 'bg-gray-900 text-gray-50 dark:bg-gray-100 dark:text-gray-950'
                        : 'bg-gray-100 text-gray-500 dark:bg-gray-800 dark:text-gray-400'
                    }`}
                  >
                    <Icon className="h-5 w-5" />
                  </span>
                  <div>
                    <div className="flex items-center gap-2">
                      <span className="font-mono text-xs text-gray-400">0{index + 1}</span>
                      <h3 className="text-base font-semibold text-gray-900 dark:text-gray-100">{step.title}</h3>
                    </div>
                    <p className="mt-1 text-sm leading-relaxed text-gray-600 dark:text-gray-400">{step.description}</p>
                  </div>
                </button>
              )
            })}
          </div>

          <motion.div
            key={current.id}
            initial={{ opacity: 0, y: 12 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.35 }}
            className="lg:sticky lg:top-24 lg:self-start"
          >
            <CodeBlock
              code={current.code}
              language={current.filename.endsWith('.json') ? 'json' : 'python'}
              title={current.filename}
            />
          </motion.div>
        </div>
      </div>
    </section>
  )
}

export default HowItWorksSection
