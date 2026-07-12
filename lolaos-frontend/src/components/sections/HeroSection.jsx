import { lazy, Suspense, useState } from 'react'
import { motion } from 'framer-motion'
import { ArrowRight, Check, Copy } from 'lucide-react'

// The three.js scene is the heaviest dependency on the site; load it on
// its own chunk so the hero copy paints first.
const HeroScene = lazy(() => import('../3d/HeroScene'))

const GITHUB_URL = 'https://github.com/0xSemantic/lola-os'
const INSTALL = 'pip install lola-os'

function InstallPill() {
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
    <button
      type="button"
      onClick={copy}
      className="group inline-flex items-center gap-3 rounded-full border border-gray-200 bg-white/70 px-4 py-2 font-mono text-sm text-gray-700 backdrop-blur transition-colors hover:bg-white dark:border-gray-800 dark:bg-gray-900/70 dark:text-gray-300 dark:hover:bg-gray-900"
    >
      <span className="select-none text-gray-400">$</span>
      <span>{INSTALL}</span>
      <span className="text-gray-400 transition-colors group-hover:text-gray-700 dark:group-hover:text-gray-200">
        {copied ? <Check className="h-4 w-4" /> : <Copy className="h-4 w-4" />}
      </span>
    </button>
  )
}

function HeroSection() {
  return (
    <section className="relative flex min-h-[92vh] items-center overflow-hidden">
      {/* 3D bridge / network scene */}
      <div className="absolute inset-0" aria-hidden="true">
        <Suspense fallback={null}>
          <HeroScene />
        </Suspense>
      </div>
      {/* Readability wash — keeps the monochrome scene subordinate to the copy */}
      <div className="absolute inset-0 bg-gradient-to-b from-gray-50/60 via-gray-50/75 to-gray-50 dark:from-gray-950/60 dark:via-gray-950/75 dark:to-gray-950" />

      <div className="section-shell relative py-24">
        <div className="mx-auto max-w-3xl text-center">
          <motion.span
            initial={{ opacity: 0, y: 12 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5 }}
            className="eyebrow"
          >
            <span className="h-1.5 w-1.5 rounded-full bg-gray-500" />
            Live Onchain Logical Agents
          </motion.span>

          <motion.h1
            initial={{ opacity: 0, y: 16 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.6, delay: 0.05 }}
            className="text-balance mt-6 text-4xl font-bold leading-[1.05] tracking-tight text-gray-900 dark:text-gray-50 sm:text-6xl"
          >
            Five minutes. One decorator.
            <br className="hidden sm:block" /> Any blockchain.
          </motion.h1>

          <motion.p
            initial={{ opacity: 0, y: 16 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.6, delay: 0.15 }}
            className="text-balance mx-auto mt-6 max-w-2xl text-lg leading-relaxed text-gray-600 dark:text-gray-400"
          >
            Wrap any function with <span className="font-mono text-gray-900 dark:text-gray-100">@lola_tool</span> and your
            AI agents can read and write blockchains, read oracle prices, and call APIs. Works with Ethereum, Polygon,
            and Solana — all at once.
          </motion.p>

          <motion.div
            initial={{ opacity: 0, y: 16 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.6, delay: 0.25 }}
            className="mt-9 flex flex-col items-center justify-center gap-3 sm:flex-row"
          >
            <a
              href={GITHUB_URL}
              target="_blank"
              rel="noopener noreferrer"
              className="group inline-flex w-full items-center justify-center gap-2 rounded-lg bg-gray-900 px-6 py-3 text-sm font-medium text-gray-50 transition-colors hover:bg-gray-800 dark:bg-gray-50 dark:text-gray-950 dark:hover:bg-gray-200 sm:w-auto"
            >
              Get started
              <ArrowRight className="h-4 w-4 transition-transform group-hover:translate-x-0.5" />
            </a>
            <a
              href="#how-it-works"
              className="inline-flex w-full items-center justify-center gap-2 rounded-lg border border-gray-300 px-6 py-3 text-sm font-medium text-gray-800 transition-colors hover:bg-gray-100 dark:border-gray-700 dark:text-gray-200 dark:hover:bg-gray-800 sm:w-auto"
            >
              See how it works
            </a>
          </motion.div>

          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ duration: 0.6, delay: 0.35 }}
            className="mt-8 flex flex-col items-center gap-4"
          >
            <InstallPill />
            <p className="text-sm text-gray-500 dark:text-gray-500">
              Free forever · Apache-2.0 · Your keys never leave your machine
            </p>
          </motion.div>
        </div>
      </div>

      {/* Scroll cue */}
      <motion.a
        href="#problem"
        aria-label="Scroll to content"
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        transition={{ delay: 1 }}
        className="absolute bottom-6 left-1/2 hidden -translate-x-1/2 sm:block"
      >
        <span className="flex h-9 w-5 items-start justify-center rounded-full border border-gray-400/60 p-1 dark:border-gray-600/60">
          <motion.span
            animate={{ y: [0, 10, 0] }}
            transition={{ duration: 1.6, repeat: Infinity }}
            className="h-2 w-1 rounded-full bg-gray-500"
          />
        </span>
      </motion.a>
    </section>
  )
}

export default HeroSection
