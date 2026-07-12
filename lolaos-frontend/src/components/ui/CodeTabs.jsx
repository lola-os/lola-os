import { useState } from 'react'
import { Check, Copy } from 'lucide-react'
import SyntaxHighlighter from '../../lib/syntaxHighlighter'
import { monoCodeTheme } from '../../lib/codeTheme'

/**
 * A tabbed code panel — used for the Python / TypeScript / Go SDK samples.
 * `tabs` is an array of { id, label, language, filename, code }.
 */
function CodeTabs({ tabs, className = '' }) {
  const [active, setActive] = useState(tabs[0]?.id)
  const [copied, setCopied] = useState(false)

  const current = tabs.find((t) => t.id === active) ?? tabs[0]

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(current.code)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch {
      /* ignore */
    }
  }

  return (
    <div className={`overflow-hidden rounded-2xl border border-gray-800 bg-gray-950 ${className}`}>
      <div className="flex items-center justify-between border-b border-gray-800 pr-3">
        <div className="flex overflow-x-auto" role="tablist" aria-label="SDK language">
          {tabs.map((tab) => {
            const isActive = tab.id === current.id
            return (
              <button
                key={tab.id}
                role="tab"
                aria-selected={isActive}
                type="button"
                onClick={() => setActive(tab.id)}
                className={`relative whitespace-nowrap px-4 py-3 text-sm font-medium transition-colors ${
                  isActive ? 'text-gray-50' : 'text-gray-500 hover:text-gray-300'
                }`}
              >
                {tab.label}
                {isActive && <span className="absolute inset-x-3 -bottom-px h-0.5 rounded-full bg-gray-100" />}
              </button>
            )
          })}
        </div>
        <button
          type="button"
          onClick={handleCopy}
          className="inline-flex shrink-0 items-center gap-1.5 rounded-md px-2 py-1 text-xs text-gray-400 transition-colors hover:bg-gray-800 hover:text-gray-100"
        >
          {copied ? <Check className="h-3.5 w-3.5" /> : <Copy className="h-3.5 w-3.5" />}
          <span className="hidden sm:inline">{copied ? 'Copied' : 'Copy'}</span>
        </button>
      </div>

      {current.filename && (
        <div className="border-b border-gray-800/70 px-4 py-2 font-mono text-xs text-gray-500">
          {current.filename}
        </div>
      )}

      <div className="overflow-x-auto p-4 sm:p-5">
        <SyntaxHighlighter
          language={current.language}
          style={monoCodeTheme}
          customStyle={{ margin: 0, padding: 0, background: 'transparent', fontSize: '0.85rem' }}
          codeTagProps={{ style: { fontFamily: "'JetBrains Mono', ui-monospace, monospace" } }}
        >
          {current.code}
        </SyntaxHighlighter>
      </div>
    </div>
  )
}

export default CodeTabs
