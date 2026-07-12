import { useEffect } from 'react'
import { Link, useLocation } from 'react-router-dom'
import { Github, X } from 'lucide-react'
import { AnimatePresence, motion } from 'framer-motion'
import LolaMark from '../ui/LolaMark'

function MobileMenu({ isOpen, onClose, navItems, githubUrl }) {
  const location = useLocation()

  useEffect(() => {
    document.body.style.overflow = isOpen ? 'hidden' : ''
    return () => {
      document.body.style.overflow = ''
    }
  }, [isOpen])

  useEffect(() => {
    const handleEscape = (e) => {
      if (e.key === 'Escape') onClose()
    }
    document.addEventListener('keydown', handleEscape)
    return () => document.removeEventListener('keydown', handleEscape)
  }, [onClose])

  return (
    <AnimatePresence>
      {isOpen && (
        <>
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            transition={{ duration: 0.2 }}
            className="fixed inset-0 z-40 bg-gray-950/40 backdrop-blur-sm md:hidden"
            onClick={onClose}
          />

          <motion.div
            initial={{ x: '100%' }}
            animate={{ x: 0 }}
            exit={{ x: '100%' }}
            transition={{ type: 'tween', duration: 0.28, ease: 'easeInOut' }}
            className="fixed inset-y-0 right-0 z-50 flex w-full max-w-xs flex-col bg-white shadow-2xl dark:bg-gray-900 md:hidden"
          >
            <div className="flex items-center justify-between border-b border-gray-200 p-5 dark:border-gray-800">
              <Link to="/" className="flex items-center gap-2.5" onClick={onClose}>
                <span className="text-gray-900 dark:text-gray-50">
                  <LolaMark className="h-7 w-7" />
                </span>
                <span className="text-base font-semibold tracking-tight text-gray-900 dark:text-gray-50">LOLA OS</span>
              </Link>
              <button
                type="button"
                onClick={onClose}
                aria-label="Close menu"
                className="inline-flex h-9 w-9 items-center justify-center rounded-lg text-gray-600 hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-gray-800"
              >
                <X className="h-5 w-5" />
              </button>
            </div>

            <nav className="flex flex-col gap-1 p-4">
              {navItems.map((item) => {
                const active = location.pathname === item.path
                return (
                  <Link
                    key={item.name}
                    to={item.path}
                    onClick={onClose}
                    className={`rounded-lg px-4 py-3 text-sm font-medium transition-colors ${
                      active
                        ? 'bg-gray-100 text-gray-900 dark:bg-gray-800 dark:text-gray-50'
                        : 'text-gray-600 hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-gray-800'
                    }`}
                  >
                    {item.name}
                  </Link>
                )
              })}
              <a
                href={githubUrl}
                target="_blank"
                rel="noopener noreferrer"
                onClick={onClose}
                className="mt-2 flex items-center gap-3 rounded-lg px-4 py-3 text-sm font-medium text-gray-600 hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-gray-800"
              >
                <Github className="h-4 w-4" />
                GitHub
              </a>
            </nav>

            <div className="mt-auto border-t border-gray-200 p-4 dark:border-gray-800">
              <a
                href={githubUrl}
                target="_blank"
                rel="noopener noreferrer"
                onClick={onClose}
                className="block w-full rounded-lg bg-gray-900 px-4 py-3 text-center text-sm font-medium text-gray-50 transition-colors hover:bg-gray-800 dark:bg-gray-50 dark:text-gray-950 dark:hover:bg-gray-200"
              >
                Get started
              </a>
              <p className="mt-3 text-center text-xs text-gray-500">Free forever · Apache-2.0</p>
            </div>
          </motion.div>
        </>
      )}
    </AnimatePresence>
  )
}

export default MobileMenu
