import { useState } from 'react'
import { Check, Copy } from 'lucide-react'
import SyntaxHighlighter from '../../lib/syntaxHighlighter'
import { monoCodeTheme } from '../../lib/codeTheme'

function CodeBlock({ code, language = 'python', title, showLineNumbers = false, className = '' }) {
  const [copied, setCopied] = useState(false)

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(code)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch {
      /* clipboard may be blocked; fail quietly */
    }
  }

  return (
    <div className={`overflow-hidden rounded-xl border border-gray-800 bg-gray-950 ${className}`}>
      <div className="flex items-center justify-between border-b border-gray-800 px-4 py-2.5">
        <div className="flex items-center gap-2">
          <span className="flex gap-1.5">
            <span className="h-2.5 w-2.5 rounded-full bg-gray-700" />
            <span className="h-2.5 w-2.5 rounded-full bg-gray-700" />
            <span className="h-2.5 w-2.5 rounded-full bg-gray-700" />
          </span>
          {title && <span className="ml-2 font-mono text-xs text-gray-500">{title}</span>}
        </div>
        <button
          type="button"
          onClick={handleCopy}
          className="inline-flex items-center gap-1.5 rounded-md px-2 py-1 text-xs text-gray-400 transition-colors hover:bg-gray-800 hover:text-gray-100"
        >
          {copied ? <Check className="h-3.5 w-3.5" /> : <Copy className="h-3.5 w-3.5" />}
          <span>{copied ? 'Copied' : 'Copy'}</span>
        </button>
      </div>

      <div className="overflow-x-auto p-4">
        <SyntaxHighlighter
          language={language}
          style={monoCodeTheme}
          showLineNumbers={showLineNumbers}
          lineNumberStyle={{ color: '#525252', minWidth: '2.25em' }}
          customStyle={{
            margin: 0,
            padding: 0,
            background: 'transparent',
            fontSize: '0.85rem',
          }}
          codeTagProps={{ style: { fontFamily: "'JetBrains Mono', ui-monospace, monospace" } }}
        >
          {code}
        </SyntaxHighlighter>
      </div>
    </div>
  )
}

export default CodeBlock
