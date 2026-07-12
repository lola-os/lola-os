import { Outlet } from 'react-router-dom'
import { ThemeProvider } from '../theme/ThemeProvider'
import Header from './Header'
import Footer from './Footer'

function Layout() {
  return (
    <ThemeProvider>
      <div className="flex min-h-screen flex-col bg-gray-50 text-gray-900 dark:bg-gray-950 dark:text-gray-100">
        <Header />
        <main className="flex-1">
          <Outlet />
        </main>
        <Footer />
      </div>
    </ThemeProvider>
  )
}

export default Layout
