import { motion } from 'framer-motion'
import { ArrowRight, Boxes, Radio, ShieldCheck } from 'lucide-react'
import SectionHeading from '../ui/SectionHeading'
import CodeBlock from '../ui/CodeBlock'

const SOLUTION_CODE = `from lola_os import lola_tool, get_balance

# Add one decorator to a plain function.
@lola_tool
def check_wallet(address: str) -> dict:
    # Live, multichain balances — no RPC wiring, no key handling.
    return get_balance("ethereum", address)

# Hand it to any agent framework as a normal tool.`

const POINTS = [
  {
    icon: Boxes,
    title: 'One interface, every chain',
    description: 'The same call resolves across all major EVM networks and Solana through a single adapter.',
  },
  {
    icon: Radio,
    title: 'Reads, writes, oracles, APIs',
    description: 'Query state, send transactions, read Chainlink price feeds, and call any REST endpoint.',
  },
  {
    icon: ShieldCheck,
    title: 'Safe by default',
    description: 'Reads are free; writes go through approval and spend limits. Your keys stay on your machine.',
  },
]

function SolutionSection() {
  return (
    <section id="solution" className="bg-gray-50 py-20 dark:bg-gray-950 md:py-28">
      <div className="section-shell">
        <SectionHeading
          eyebrow="The idea"
          title="Turn any function into an onchain tool"
          description="LOLA OS is not another agent framework. It is the layer that lets the framework you already use reach the chain."
        />

        <div className="mt-14 grid grid-cols-1 items-center gap-10 lg:grid-cols-2">
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true, amount: 0.3 }}
            transition={{ duration: 0.5 }}
            className="order-2 lg:order-1"
          >
            <ul className="space-y-6">
              {POINTS.map((point) => {
                const Icon = point.icon
                return (
                  <li key={point.title} className="flex gap-4">
                    <span className="mt-0.5 inline-flex h-10 w-10 shrink-0 items-center justify-center rounded-lg border border-gray-200 bg-white text-gray-900 dark:border-gray-800 dark:bg-gray-900 dark:text-gray-100">
                      <Icon className="h-5 w-5" />
                    </span>
                    <div>
                      <h3 className="text-base font-semibold text-gray-900 dark:text-gray-100">{point.title}</h3>
                      <p className="mt-1 text-sm leading-relaxed text-gray-600 dark:text-gray-400">
                        {point.description}
                      </p>
                    </div>
                  </li>
                )
              })}
            </ul>

            <a
              href="#how-it-works"
              className="group mt-8 inline-flex items-center gap-2 text-sm font-medium text-gray-900 hover:text-gray-600 dark:text-gray-100 dark:hover:text-gray-300"
            >
              See the three steps
              <ArrowRight className="h-4 w-4 transition-transform group-hover:translate-x-0.5" />
            </a>
          </motion.div>

          <motion.div
            initial={{ opacity: 0, y: 20 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true, amount: 0.3 }}
            transition={{ duration: 0.5, delay: 0.1 }}
            className="order-1 lg:order-2"
          >
            <CodeBlock code={SOLUTION_CODE} language="python" title="check_wallet.py" />
          </motion.div>
        </div>
      </div>
    </section>
  )
}

export default SolutionSection
