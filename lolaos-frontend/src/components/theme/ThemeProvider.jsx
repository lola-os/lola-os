import { useCallback, useEffect, useMemo, useState } from 'react'
import { ThemeContext } from './themeContext'

function getInitialTheme() {
  if (typeof document === 'undefined') return 'light'
  // The pre-paint script in index.html already resolved and applied the
  // correct class, so trust the DOM as the source of truth.
  return document.documentElement.classList.contains('dark') ? 'dark' : 'light'
}

export function ThemeProvider({ children }) {
  const [theme, setThemeState] = useState(getInitialTheme)

  const applyTheme = useCallback((next) => {
    const root = document.documentElement
    root.classList.toggle('dark', next === 'dark')
    try {
      localStorage.setItem('lola-theme', next)
    } catch {
      /* storage may be unavailable; ignore */
    }
  }, [])

  const setTheme = useCallback(
    (next) => {
      setThemeState(next)
      applyTheme(next)
    },
    [applyTheme]
  )

  const toggleTheme = useCallback(() => {
    setThemeState((prev) => {
      const next = prev === 'dark' ? 'light' : 'dark'
      applyTheme(next)
      return next
    })
  }, [applyTheme])

  // Follow the OS preference until the user makes an explicit choice.
  useEffect(() => {
    const media = window.matchMedia('(prefers-color-scheme: dark)')
    const handleChange = (event) => {
      let stored = null
      try {
        stored = localStorage.getItem('lola-theme')
      } catch {
        stored = null
      }
      if (!stored) setTheme(event.matches ? 'dark' : 'light')
    }
    media.addEventListener('change', handleChange)
    return () => media.removeEventListener('change', handleChange)
  }, [setTheme])

  const value = useMemo(() => ({ theme, toggleTheme, setTheme }), [theme, toggleTheme, setTheme])

  return <ThemeContext.Provider value={value}>{children}</ThemeContext.Provider>
}
