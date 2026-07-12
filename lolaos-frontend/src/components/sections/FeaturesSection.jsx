import { motion } from 'framer-motion'
import { FileCheck2, GaugeCircle, History, Lock, ScrollText, UserCheck } from 'lucide-react'
import SectionHeading from '../ui/SectionHeading'

const FEATURES = [
  {
    icon: Lock,
    title: 'Encrypted key vault',
    description: 'Keys are stored locally with AES-256 and never leave your machine — no hosted custody, ever.',
  },
  {
    icon: UserCheck,
    title: 'Human-in-the-loop approval',
    description: 'Every write can pause for explicit sign-off before it is broadcast to the network.',
  },
  {
    icon: GaugeCircle,
    title: 'Budget circuit breaker',
    description: 'Track gas and USD spend per session and halt automatically before an agent overspends.',
  },
  {
    icon: ScrollText,
    title: 'Transaction registry',
    description: 'A local record of every call and transaction, so you always have a full audit trail.',
  },
  {
    icon: History,
    title: 'Structured execution replay',
    description: 'Write a plan, replay it deterministically, and inspect exactly what ran and why.',
  },
  {
    icon: FileCheck2,
    title: 'Pre-flight ABI validation',
    description: 'Contract calls are checked against the ABI before submission to catch mistakes early.',
  },
]

function FeaturesSection() {
  return (
    <section id="features" className="bg-white py-20 dark:bg-gray-900 md:py-28">
      <div className="section-shell">
        <SectionHeading
          eyebrow="Operational maturity"
          title="Built for agents that touch real value"
          description="The safeguards you would otherwise assemble yourself, included and running locally by default."
        />

        <div className="mt-14 grid grid-cols-1 gap-px overflow-hidden rounded-2xl border border-gray-200 bg-gray-200 dark:border-gray-800 dark:bg-gray-800 sm:grid-cols-2 lg:grid-cols-3">
          {FEATURES.map((feature, index) => {
            const Icon = feature.icon
            return (
              <motion.div
                key={feature.title}
                initial={{ opacity: 0 }}
                whileInView={{ opacity: 1 }}
                viewport={{ once: true, amount: 0.3 }}
                transition={{ duration: 0.4, delay: (index % 3) * 0.08 }}
                className="group flex h-full flex-col bg-white p-7 transition-colors hover:bg-gray-50 dark:bg-gray-900 dark:hover:bg-gray-950"
              >
                <span className="inline-flex h-11 w-11 items-center justify-center rounded-lg border border-gray-200 bg-gray-50 text-gray-900 transition-colors group-hover:border-gray-300 dark:border-gray-800 dark:bg-gray-950 dark:text-gray-100">
                  <Icon className="h-5 w-5" />
                </span>
                <h3 className="mt-5 text-base font-semibold text-gray-900 dark:text-gray-100">{feature.title}</h3>
                <p className="mt-2 text-sm leading-relaxed text-gray-600 dark:text-gray-400">{feature.description}</p>
              </motion.div>
            )
          })}
        </div>

        <motion.p
          initial={{ opacity: 0 }}
          whileInView={{ opacity: 1 }}
          viewport={{ once: true }}
          transition={{ duration: 0.5 }}
          className="mt-8 text-center text-sm text-gray-500 dark:text-gray-500"
        >
          One Go engine — lola-core — runs all of this locally, behind the SDK you call.
        </motion.p>
      </div>
    </section>
  )
}

export default FeaturesSection
