import { createContext, useContext } from 'react'

export const ThemeContext = createContext({
  theme: 'light',
  toggleTheme: () => {},
  setTheme: () => {},
})

export function useTheme() {
  return useContext(ThemeContext)
}
