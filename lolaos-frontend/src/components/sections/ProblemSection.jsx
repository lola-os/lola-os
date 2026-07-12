import { motion } from 'framer-motion'
import { KeyRound, Layers, Network, ShieldAlert } from 'lucide-react'
import SectionHeading from '../ui/SectionHeading'

const PROBLEMS = [
  {
    icon: Layers,
    title: 'Every chain is its own stack',
    description: 'Different libraries, RPC quirks, and signing flows for Ethereum, each L2, and Solana.',
  },
  {
    icon: KeyRound,
    title: 'Key handling is risky',
    description: 'Private keys, nonce management, and gas estimation are easy to get wrong and hard to audit.',
  },
  {
    icon: ShieldAlert,
    title: 'No guardrails by default',
    description: 'A model that can send transactions needs approval, spend limits, and a record of what it did.',
  },
  {
    icon: Network,
    title: 'Agents stay offchain',
    description: 'So most AI frameworks read static data instead of live balances, prices, and contract state.',
  },
]

function ProblemSection() {
  return (
    <section id="problem" className="bg-white py-20 dark:bg-gray-900 md:py-28">
      <div className="section-shell">
        <SectionHeading
          eyebrow="The gap"
          title="Connecting agents to chains is the hard part"
          description="The AI and blockchain ecosystems move fast, but wiring them together still costs weeks of specialist work."
        />

        <div className="mt-14 grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
          {PROBLEMS.map((problem, index) => {
            const Icon = problem.icon
            return (
              <motion.div
                key={problem.title}
                initial={{ opacity: 0, y: 20 }}
                whileInView={{ opacity: 1, y: 0 }}
                viewport={{ once: true, amount: 0.4 }}
                transition={{ duration: 0.5, delay: index * 0.08 }}
                className="flex h-full flex-col rounded-2xl border border-gray-200 bg-gray-50 p-6 transition-colors hover:border-gray-300 dark:border-gray-800 dark:bg-gray-950 dark:hover:border-gray-700"
              >
                <span className="inline-flex h-10 w-10 items-center justify-center rounded-lg bg-gray-900 text-gray-50 dark:bg-gray-100 dark:text-gray-950">
                  <Icon className="h-5 w-5" />
                </span>
                <h3 className="mt-4 text-base font-semibold text-gray-900 dark:text-gray-100">{problem.title}</h3>
                <p className="mt-2 text-sm leading-relaxed text-gray-600 dark:text-gray-400">{problem.description}</p>
              </motion.div>
            )
          })}
        </div>
      </div>
    </section>
  )
}

export default ProblemSection
