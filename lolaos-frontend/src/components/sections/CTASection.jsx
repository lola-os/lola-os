import { useState } from 'react'
import { motion } from 'framer-motion'
import { ArrowRight, Check, Copy, Github } from 'lucide-react'
import LolaMark from '../ui/LolaMark'

const GITHUB_URL = 'https://github.com/0xSemantic/lola-os'
const INSTALL = 'pip install lola-os'

function CTASection() {
  const [copied, setCopied] = useState(false)

  const copy = async () => {
    try {
      await navigator.clipboard.writeText(INSTALL)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch {
      /* ignore */
    }
  }

  return (
    <section className="bg-white py-20 dark:bg-gray-900 md:py-28">
      <div className="section-shell">
        <motion.div
          initial={{ opacity: 0, y: 24 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true, amount: 0.3 }}
          transition={{ duration: 0.6 }}
          className="relative mx-auto flex max-w-4xl flex-col items-center overflow-hidden rounded-3xl border border-gray-800 bg-gray-950 px-6 py-16 text-center sm:px-12"
        >
          <span className="text-gray-100">
            <LolaMark className="h-10 w-10" />
          </span>

          <h2 className="text-balance mt-6 text-3xl font-bold tracking-tight text-gray-50 sm:text-4xl">
            Start building in five minutes
          </h2>
          <p className="text-balance mt-4 max-w-xl text-lg text-gray-400">
            Install the SDK, add one decorator, and give your agent the chain. Free forever, Apache-2.0, running on your
            machine.
          </p>

          <div className="mt-8 flex w-full flex-col items-center justify-center gap-3 sm:flex-row">
            <button
              type="button"
              onClick={copy}
              className="group inline-flex w-full items-center justify-center gap-3 rounded-lg border border-gray-700 bg-gray-900 px-5 py-3 font-mono text-sm text-gray-200 transition-colors hover:bg-gray-800 sm:w-auto"
            >
              <span className="text-gray-500">$</span>
              {INSTALL}
              <span className="text-gray-500 transition-colors group-hover:text-gray-200">
                {copied ? <Check className="h-4 w-4" /> : <Copy className="h-4 w-4" />}
              </span>
            </button>
            <a
              href={GITHUB_URL}
              target="_blank"
              rel="noopener noreferrer"
              className="group inline-flex w-full items-center justify-center gap-2 rounded-lg bg-gray-50 px-5 py-3 text-sm font-medium text-gray-950 transition-colors hover:bg-gray-200 sm:w-auto"
            >
              <Github className="h-4 w-4" />
              View on GitHub
              <ArrowRight className="h-4 w-4 transition-transform group-hover:translate-x-0.5" />
            </a>
          </div>

          <div className="mt-8 flex flex-wrap items-center justify-center gap-x-6 gap-y-2 text-sm text-gray-500">
            <span>No hosted backend</span>
            <span className="hidden h-1 w-1 rounded-full bg-gray-700 sm:block" />
            <span>No API keys</span>
            <span className="hidden h-1 w-1 rounded-full bg-gray-700 sm:block" />
            <span>No billing</span>
          </div>
        </motion.div>
      </div>
    </section>
  )
}

export default CTASection
