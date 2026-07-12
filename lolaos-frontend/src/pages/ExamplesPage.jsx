import { useMemo, useState } from 'react'
import { motion } from 'framer-motion'
import { ExternalLink, Search } from 'lucide-react'
import Button from '../components/ui/Button'
import CodeBlock from '../components/ui/CodeBlock'

const GITHUB_URL = 'https://github.com/0xSemantic/lola-os'

const EXAMPLES = [
  {
    id: 'crewai-treasury',
    title: 'CrewAI treasury analyst',
    description: 'Give a CrewAI agent a tool that reads live balances across chains.',
    framework: 'crewai',
    level: 'Beginner',
    language: 'python',
    code: `from crewai import Agent
from lola_os import lola_tool, get_balance

@lola_tool
def wallet_balance(address: str) -> dict:
    return get_balance("ethereum", address)

analyst = Agent(
    role="Treasury analyst",
    goal="Report wallet health",
    tools=[wallet_balance],
)`,
  },
  {
    id: 'langchain-prices',
    title: 'LangChain price lookup',
    description: 'Read a Chainlink price feed from inside a LangChain tool.',
    framework: 'langchain',
    level: 'Beginner',
    language: 'python',
    code: `from langchain.tools import Tool
from lola_os import lola_tool, get_price

@lola_tool
def eth_price(pair: str = "ETH/USD") -> float:
    # Reads directly from the Chainlink feed contract.
    return get_price(pair)

price_tool = Tool(
    name="eth_price",
    func=eth_price,
    description="Current ETH price in USD",
)`,
  },
  {
    id: 'custom-transfer',
    title: 'Guarded transfer',
    description: 'A write that pauses for human approval and respects the spend budget.',
    framework: 'custom',
    level: 'Intermediate',
    language: 'python',
    code: `from lola_os import lola_tool, send_transaction

@lola_tool(requires_approval=True)
def pay(to: str, amount: str) -> dict:
    # The circuit breaker and approval prompt run before broadcast.
    return send_transaction(
        chain="base",
        to=to,
        value=amount,
    )`,
  },
  {
    id: 'replay-plan',
    title: 'Replay a plan',
    description: 'Run a structured plan deterministically and keep a full audit trail.',
    framework: 'custom',
    level: 'Intermediate',
    language: 'bash',
    code: `# Replay a saved plan; every step lands in the local registry.
lola replay plan.json

# Inspect what ran, in order, with results.
lola registry list --last 20`,
  },
]

const FILTERS = [
  { id: 'all', name: 'All' },
  { id: 'crewai', name: 'CrewAI' },
  { id: 'langchain', name: 'LangChain' },
  { id: 'custom', name: 'Custom' },
]

function ExamplesPage() {
  const [activeFilter, setActiveFilter] = useState('all')
  const [query, setQuery] = useState('')

  const filtered = useMemo(() => {
    return EXAMPLES.filter((example) => {
      if (activeFilter !== 'all' && example.framework !== activeFilter) return false
      if (query) {
        const q = query.toLowerCase()
        return (
          example.title.toLowerCase().includes(q) ||
          example.description.toLowerCase().includes(q) ||
          example.framework.includes(q)
        )
      }
      return true
    })
  }, [activeFilter, query])

  return (
    <div className="bg-gray-50 py-16 dark:bg-gray-950 md:py-24">
      <div className="section-shell">
        <div className="mx-auto max-w-2xl text-center">
          <span className="eyebrow">Examples</span>
          <h1 className="mt-5 text-3xl font-bold tracking-tight text-gray-900 dark:text-gray-50 sm:text-4xl">
            Copy, adapt, ship
          </h1>
          <p className="mt-4 text-lg text-gray-600 dark:text-gray-400">
            Small, real snippets that add blockchain, oracles, and approvals to agents you already run.
          </p>
        </div>

        <div className="mt-10 flex flex-col items-stretch gap-4 sm:flex-row sm:items-center sm:justify-between">
          <div className="flex flex-wrap gap-2">
            {FILTERS.map((filter) => (
              <button
                key={filter.id}
                type="button"
                onClick={() => setActiveFilter(filter.id)}
                className={`rounded-lg px-4 py-2 text-sm font-medium transition-colors ${
                  activeFilter === filter.id
                    ? 'bg-gray-900 text-gray-50 dark:bg-gray-50 dark:text-gray-950'
                    : 'bg-gray-100 text-gray-700 hover:bg-gray-200 dark:bg-gray-900 dark:text-gray-300 dark:hover:bg-gray-800'
                }`}
              >
                {filter.name}
              </button>
            ))}
          </div>

          <div className="relative sm:w-64">
            <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400" />
            <input
              type="text"
              value={query}
              onChange={(event) => setQuery(event.target.value)}
              placeholder="Search examples"
              className="w-full rounded-lg border border-gray-300 bg-white py-2 pl-9 pr-3 text-sm text-gray-900 placeholder:text-gray-400 focus:border-gray-400 focus:outline-none focus-visible:ring-2 focus-visible:ring-gray-900 dark:border-gray-800 dark:bg-gray-900 dark:text-gray-100 dark:focus-visible:ring-gray-100"
            />
          </div>
        </div>

        <div className="mt-8 grid grid-cols-1 gap-6 lg:grid-cols-2">
          {filtered.map((example, index) => (
            <motion.article
              key={example.id}
              initial={{ opacity: 0, y: 16 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.4, delay: index * 0.05 }}
              className="flex h-full flex-col rounded-2xl border border-gray-200 bg-white p-6 dark:border-gray-800 dark:bg-gray-900"
            >
              <div className="flex items-start justify-between gap-4">
                <div>
                  <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-100">{example.title}</h2>
                  <p className="mt-1 text-sm text-gray-600 dark:text-gray-400">{example.description}</p>
                </div>
                <span className="shrink-0 rounded-full border border-gray-200 px-2.5 py-1 text-xs font-medium text-gray-600 dark:border-gray-700 dark:text-gray-400">
                  {example.level}
                </span>
              </div>

              <div className="mt-5">
                <CodeBlock code={example.code} language={example.language} title={`${example.id}`} />
              </div>

              <div className="mt-5">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => window.open(GITHUB_URL, '_blank', 'noopener')}
                >
                  <ExternalLink className="h-4 w-4" />
                  View on GitHub
                </Button>
              </div>
            </motion.article>
          ))}
        </div>

        {filtered.length === 0 && (
          <div className="py-16 text-center">
            <p className="text-gray-600 dark:text-gray-400">No examples match that search.</p>
            <Button
              className="mt-4"
              variant="outline"
              onClick={() => {
                setActiveFilter('all')
                setQuery('')
              }}
            >
              Clear filters
            </Button>
          </div>
        )}
      </div>
    </div>
  )
}

export default ExamplesPage
