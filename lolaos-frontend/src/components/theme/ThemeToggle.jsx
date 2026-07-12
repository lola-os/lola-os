import { Moon, Sun } from 'lucide-react'
import { useTheme } from './themeContext'

function ThemeToggle({ className = '' }) {
  const { theme, toggleTheme } = useTheme()
  const isDark = theme === 'dark'

  return (
    <button
      type="button"
      onClick={toggleTheme}
      aria-label={isDark ? 'Switch to light theme' : 'Switch to dark theme'}
      title={isDark ? 'Switch to light theme' : 'Switch to dark theme'}
      className={`inline-flex h-9 w-9 items-center justify-center rounded-lg border border-gray-200 bg-white/60 text-gray-600 transition-colors hover:bg-gray-100 hover:text-gray-900 dark:border-gray-800 dark:bg-gray-900/60 dark:text-gray-400 dark:hover:bg-gray-800 dark:hover:text-gray-100 ${className}`}
    >
      {isDark ? <Sun className="h-[18px] w-[18px]" /> : <Moon className="h-[18px] w-[18px]" />}
    </button>
  )
}

export default ThemeToggle
