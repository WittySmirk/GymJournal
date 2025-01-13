import { BrowserRouter, Routes, Route } from 'react-router'
import { createRoot } from 'react-dom/client'
import App from './pages/App.tsx'
import Index from './pages/Index.tsx'

createRoot(document.getElementById('root')!).render(
  <BrowserRouter>
    <Routes>
      <Route index element={<Index />} />
      <Route path="app" element={<App />} />
    </Routes>
  </BrowserRouter>
)
