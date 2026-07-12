import { Link } from 'react-router-dom'
import { Github } from 'lucide-react'
import LolaMark from '../ui/LolaMark'

const GITHUB_URL = 'https://github.com/0xSemantic/lola-os'

const LINKS = {
  Product: [
    { label: 'How it works', href: '/#how-it-works' },
    { label: 'Multichain', href: '/#multichain' },
    { label: 'Features', href: '/#features' },
    { label: 'SDKs', href: '/#sdks' },
  ],
  Learn: [
    { label: 'Documentation', to: '/docs' },
    { label: 'Examples', to: '/examples' },
    { label: 'GitHub', href: GITHUB_URL },
  ],
  Install: [
    { label: 'pip install lola-os', href: GITHUB_URL },
    { label: 'npm install lola-os', href: GITHUB_URL },
    { label: 'Go module', href: GITHUB_URL },
  ],
}

function Footer() {
  const year = new Date().getFullYear()

  return (
    <footer className="border-t border-gray-200 bg-gray-50 dark:border-gray-800 dark:bg-gray-950">
      <div className="section-shell py-14">
        <div className="grid grid-cols-2 gap-10 md:grid-cols-4">
          <div className="col-span-2 md:col-span-1">
            <Link to="/" className="flex items-center gap-2.5" aria-label="LOLA OS home">
              <span className="text-gray-900 dark:text-gray-50">
                <LolaMark className="h-7 w-7" />
              </span>
              <span className="text-base font-semibold tracking-tight text-gray-900 dark:text-gray-50">LOLA OS</span>
            </Link>
            <p className="mt-4 max-w-xs text-sm leading-relaxed text-gray-600 dark:text-gray-400">
              One decorator connects any AI agent to blockchains, oracles, and APIs. Runs entirely on your machine.
            </p>
            <a
              href={GITHUB_URL}
              target="_blank"
              rel="noopener noreferrer"
              className="mt-5 inline-flex items-center gap-2 rounded-lg border border-gray-200 bg-white px-3 py-2 text-sm font-medium text-gray-700 transition-colors hover:bg-gray-100 dark:border-gray-800 dark:bg-gray-900 dark:text-gray-300 dark:hover:bg-gray-800"
            >
              <Github className="h-4 w-4" />
              Star on GitHub
            </a>
          </div>

          {Object.entries(LINKS).map(([heading, items]) => (
            <div key={heading}>
              <h3 className="text-sm font-semibold text-gray-900 dark:text-gray-100">{heading}</h3>
              <ul className="mt-4 space-y-3">
                {items.map((item) => (
                  <li key={item.label}>
                    {item.to ? (
                      <Link
                        to={item.to}
                        className="text-sm text-gray-600 transition-colors hover:text-gray-900 dark:text-gray-400 dark:hover:text-gray-100"
                      >
                        {item.label}
                      </Link>
                    ) : (
                      <a
                        href={item.href}
                        target={item.href?.startsWith('http') ? '_blank' : undefined}
                        rel={item.href?.startsWith('http') ? 'noopener noreferrer' : undefined}
                        className={`text-sm text-gray-600 transition-colors hover:text-gray-900 dark:text-gray-400 dark:hover:text-gray-100 ${
                          heading === 'Install' ? 'font-mono text-[13px]' : ''
                        }`}
                      >
                        {item.label}
                      </a>
                    )}
                  </li>
                ))}
              </ul>
            </div>
          ))}
        </div>

        <div className="mt-12 flex flex-col items-center justify-between gap-3 border-t border-gray-200 pt-8 text-sm text-gray-500 dark:border-gray-800 sm:flex-row">
          <p>© {year} LOLA OS · Apache-2.0</p>
          <p>Free forever. No hosted backend, no API keys, no billing.</p>
        </div>
      </div>
    </footer>
  )
}

export default Footer
