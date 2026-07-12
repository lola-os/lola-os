import { BrowserRouter, Routes, Route } from 'react-router-dom'
import Layout from '../components/layout/Layout'
import HomePage from '../pages/HomePage'
import DocsPage from '../pages/DocsPage'
import ExamplesPage from '../pages/ExamplesPage'

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Layout />}>
          <Route index element={<HomePage />} />
          <Route path="docs" element={<DocsPage />} />
          <Route path="examples" element={<ExamplesPage />} />
        </Route>
      </Routes>
    </BrowserRouter>
  )
}

export default App