import { motion } from 'framer-motion'
import SectionHeading from '../ui/SectionHeading'
import CodeTabs from '../ui/CodeTabs'

const SDK_TABS = [
  {
    id: 'python',
    label: 'Python',
    language: 'python',
    filename: 'agent.py  ·  pip install lola-os',
    code: `from lola_os import lola_tool, get_balance

@lola_tool
def check_balance(address: str) -> dict:
    return get_balance("ethereum", address)`,
  },
  {
    id: 'typescript',
    label: 'TypeScript',
    language: 'typescript',
    filename: 'agent.ts  ·  npm install lola-os',
    code: `import { lolaTool, getBalance } from "lola-os";

export const checkBalance = lolaTool(
  async (address: string) => {
    return getBalance("ethereum", address);
  }
);`,
  },
  {
    id: 'go',
    label: 'Go',
    language: 'go',
    filename: 'agent.go  ·  go get github.com/lola-os/lola-go',
    code: `package main

import "github.com/lola-os/lola-go"

var CheckBalance = lola.Tool(
    func(address string) (any, error) {
        return lola.GetBalance("ethereum", address)
    },
)`,
  },
]

function SdkSection() {
  return (
    <section id="sdks" className="bg-gray-50 py-20 dark:bg-gray-950 md:py-28">
      <div className="section-shell">
        <SectionHeading
          eyebrow="SDKs"
          title="One decorator, in your language"
          description="Python, TypeScript, and Go SDKs speak to the same local Go engine. Same idea, idiomatic in each."
        />

        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true, amount: 0.3 }}
          transition={{ duration: 0.5 }}
          className="mx-auto mt-12 max-w-3xl"
        >
          <CodeTabs tabs={SDK_TABS} />
          <p className="mt-5 text-center text-sm text-gray-500 dark:text-gray-500">
            Works alongside CrewAI, LangChain, and any custom agent loop.
          </p>
        </motion.div>
      </div>
    </section>
  )
}

export default SdkSection
