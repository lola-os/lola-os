import { ArrowRight, BookOpen } from 'lucide-react'
import Button from '../components/ui/Button'
import CodeBlock from '../components/ui/CodeBlock'
import CodeTabs from '../components/ui/CodeTabs'

const GITHUB_URL = 'https://github.com/0xSemantic/lola-os'

const INSTALL_TABS = [
  { id: 'python', label: 'Python', language: 'bash', filename: 'terminal', code: 'pip install lola-os' },
  { id: 'ts', label: 'TypeScript', language: 'bash', filename: 'terminal', code: 'npm install lola-os' },
  { id: 'go', label: 'Go', language: 'bash', filename: 'terminal', code: 'go get github.com/lola-os/lola-go' },
]

const QUICKSTART = `from lola_os import lola_tool, get_balance

@lola_tool
def check_balance(address: str) -> dict:
    return get_balance("ethereum", address)

# Pass check_balance to your agent framework as a tool,
# or call it directly:
print(check_balance("0x742d35Cc6634C0532925a3b844Bc9e90F1A6B1E7"))`

const STEPS = [
  {
    title: '1. Install the SDK',
    body: 'Add LOLA OS to your project. Everything runs locally — no account, no API key.',
  },
  {
    title: '2. Decorate a function',
    body: 'Wrap any function with @lola_tool. LOLA figures out the chains and calls it needs.',
  },
  {
    title: '3. Configure chains',
    body: 'Point LOLA at RPC endpoints through environment variables. Reads work out of the box.',
  },
  {
    title: '4. Add guardrails for writes',
    body: 'Enable approval and a spend budget before letting an agent send transactions.',
  },
]

function DocsPage() {
  return (
    <div className="bg-gray-50 py-16 dark:bg-gray-950 md:py-24">
      <div className="section-shell">
        <div className="mx-auto max-w-3xl">
          <span className="eyebrow">
            <BookOpen className="h-3.5 w-3.5" />
            Quickstart
          </span>
          <h1 className="mt-5 text-3xl font-bold tracking-tight text-gray-900 dark:text-gray-50 sm:text-4xl">
            Get running in five minutes
          </h1>
          <p className="mt-4 text-lg leading-relaxed text-gray-600 dark:text-gray-400">
            Install the SDK, decorate a function, and your agent can reach the chain. The full reference and guides live
            in the documentation site.
          </p>

          <div className="mt-10">
            <h2 className="text-sm font-semibold uppercase tracking-wider text-gray-500">Install</h2>
            <div className="mt-4">
              <CodeTabs tabs={INSTALL_TABS} />
            </div>
          </div>

          <div className="mt-12">
            <h2 className="text-sm font-semibold uppercase tracking-wider text-gray-500">Your first tool</h2>
            <div className="mt-4">
              <CodeBlock code={QUICKSTART} language="python" title="quickstart.py" />
            </div>
          </div>

          <div className="mt-12">
            <h2 className="text-sm font-semibold uppercase tracking-wider text-gray-500">The path</h2>
            <ol className="mt-4 space-y-4">
              {STEPS.map((step) => (
                <li
                  key={step.title}
                  className="rounded-2xl border border-gray-200 bg-white p-5 dark:border-gray-800 dark:bg-gray-900"
                >
                  <h3 className="text-base font-semibold text-gray-900 dark:text-gray-100">{step.title}</h3>
                  <p className="mt-1 text-sm leading-relaxed text-gray-600 dark:text-gray-400">{step.body}</p>
                </li>
              ))}
            </ol>
          </div>

          <div className="mt-12 flex flex-col gap-3 rounded-2xl border border-gray-200 bg-white p-6 dark:border-gray-800 dark:bg-gray-900 sm:flex-row sm:items-center sm:justify-between">
            <div>
              <h3 className="text-base font-semibold text-gray-900 dark:text-gray-100">Full documentation</h3>
              <p className="mt-1 text-sm text-gray-600 dark:text-gray-400">
                API reference, security model, and end-to-end examples.
              </p>
            </div>
            <Button onClick={() => window.open(GITHUB_URL, '_blank', 'noopener')}>
              Read the docs
              <ArrowRight className="h-4 w-4" />
            </Button>
          </div>
        </div>
      </div>
    </div>
  )
}

export default DocsPage
