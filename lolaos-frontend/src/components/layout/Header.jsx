import { useEffect, useState } from 'react'
import { Link, useLocation } from 'react-router-dom'
import { Github, Menu } from 'lucide-react'
import LolaMark from '../ui/LolaMark'
import ThemeToggle from '../theme/ThemeToggle'
import MobileMenu from './MobileMenu'

const NAV_ITEMS = [
  { name: 'Home', path: '/' },
  { name: 'Examples', path: '/examples' },
  { name: 'Docs', path: '/docs' },
]

const GITHUB_URL = 'https://github.com/0xSemantic/lola-os'

function Header() {
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false)
  const [scrolled, setScrolled] = useState(false)
  const location = useLocation()

  useEffect(() => {
    const onScroll = () => setScrolled(window.scrollY > 8)
    onScroll()
    window.addEventListener('scroll', onScroll, { passive: true })
    return () => window.removeEventListener('scroll', onScroll)
  }, [])

  return (
    <>
      <header
        className={`sticky top-0 z-50 transition-colors duration-300 ${
          scrolled
            ? 'glass-effect border-b border-gray-200/70 dark:border-gray-800/70'
            : 'border-b border-transparent bg-transparent'
        }`}
      >
        <div className="section-shell">
          <div className="flex h-16 items-center justify-between">
            <Link to="/" className="group flex items-center gap-2.5" aria-label="LOLA OS home">
              <span className="text-gray-900 transition-transform duration-200 group-hover:scale-105 dark:text-gray-50">
                <LolaMark className="h-8 w-8" />
              </span>
              <span className="flex flex-col leading-none">
                <span className="text-base font-semibold tracking-tight text-gray-900 dark:text-gray-50">LOLA OS</span>
                <span className="mt-0.5 text-[11px] text-gray-500 dark:text-gray-500">Onchain tools for AI agents</span>
              </span>
            </Link>

            <nav className="hidden items-center gap-1 md:flex">
              {NAV_ITEMS.map((item) => {
                const active = location.pathname === item.path
                return (
                  <Link
                    key={item.name}
                    to={item.path}
                    className={`rounded-lg px-3 py-2 text-sm font-medium transition-colors ${
                      active
                        ? 'text-gray-900 dark:text-gray-50'
                        : 'text-gray-600 hover:text-gray-900 dark:text-gray-400 dark:hover:text-gray-100'
                    }`}
                  >
                    {item.name}
                  </Link>
                )
              })}
            </nav>

            <div className="flex items-center gap-2">
              <a
                href={GITHUB_URL}
                target="_blank"
                rel="noopener noreferrer"
                aria-label="LOLA OS on GitHub"
                className="hidden h-9 w-9 items-center justify-center rounded-lg border border-gray-200 bg-white/60 text-gray-600 transition-colors hover:bg-gray-100 hover:text-gray-900 dark:border-gray-800 dark:bg-gray-900/60 dark:text-gray-400 dark:hover:bg-gray-800 dark:hover:text-gray-100 sm:inline-flex"
              >
                <Github className="h-[18px] w-[18px]" />
              </a>
              <ThemeToggle />
              <a
                href={GITHUB_URL}
                target="_blank"
                rel="noopener noreferrer"
                className="hidden rounded-lg bg-gray-900 px-4 py-2 text-sm font-medium text-gray-50 transition-colors hover:bg-gray-800 dark:bg-gray-50 dark:text-gray-950 dark:hover:bg-gray-200 md:inline-flex"
              >
                Get started
              </a>
              <button
                type="button"
                aria-label="Open menu"
                className="inline-flex h-9 w-9 items-center justify-center rounded-lg text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-gray-800 md:hidden"
                onClick={() => setIsMobileMenuOpen(true)}
              >
                <Menu className="h-5 w-5" />
              </button>
            </div>
          </div>
        </div>
      </header>

      <MobileMenu
        isOpen={isMobileMenuOpen}
        onClose={() => setIsMobileMenuOpen(false)}
        navItems={NAV_ITEMS}
        githubUrl={GITHUB_URL}
      />
    </>
  )
}

export default Header
