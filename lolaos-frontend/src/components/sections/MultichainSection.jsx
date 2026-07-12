import { motion } from 'framer-motion'
import SectionHeading from '../ui/SectionHeading'

const CHAINS = [
  'Ethereum',
  'Polygon',
  'Arbitrum',
  'Optimism',
  'Base',
  'BNB Chain',
  'Avalanche',
  'Solana',
]

function ChainCard({ name, index }) {
  return (
    <motion.div
      initial={{ opacity: 0, y: 16 }}
      whileInView={{ opacity: 1, y: 0 }}
      viewport={{ once: true, amount: 0.4 }}
      transition={{ duration: 0.4, delay: (index % 4) * 0.06 }}
      className="flex items-center gap-3 rounded-xl border border-gray-200 bg-white px-4 py-4 transition-colors hover:border-gray-300 dark:border-gray-800 dark:bg-gray-900 dark:hover:border-gray-700"
    >
      <span className="flex h-9 w-9 items-center justify-center rounded-lg bg-gray-100 font-mono text-sm font-semibold text-gray-700 dark:bg-gray-800 dark:text-gray-300">
        {name.slice(0, 2).toUpperCase()}
      </span>
      <span className="text-sm font-medium text-gray-900 dark:text-gray-100">{name}</span>
    </motion.div>
  )
}

function MultichainSection() {
  return (
    <section id="multichain" className="bg-gray-50 py-20 dark:bg-gray-950 md:py-28">
      <div className="section-shell">
        <SectionHeading
          eyebrow="Multichain"
          title="Works with Ethereum, Polygon, and Solana — all at once"
          description="Every major EVM network and Solana behind one interface. Add a chain by name, not by rewriting your agent."
        />

        <div className="mx-auto mt-14 grid max-w-4xl grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-4">
          {CHAINS.map((chain, index) => (
            <ChainCard key={chain} name={chain} index={index} />
          ))}
        </div>

        <motion.div
          initial={{ opacity: 0, y: 16 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true, amount: 0.4 }}
          transition={{ duration: 0.5 }}
          className="mx-auto mt-10 flex max-w-2xl flex-col items-center gap-2 text-center"
        >
          <p className="font-mono text-sm text-gray-500 dark:text-gray-500">
            get_balance(<span className="text-gray-800 dark:text-gray-200">&quot;solana&quot;</span>, address)
          </p>
          <p className="text-sm text-gray-600 dark:text-gray-400">
            The same call, any supported chain — resolved by the local engine.
          </p>
        </motion.div>
      </div>
    </section>
  )
}

export default MultichainSection
