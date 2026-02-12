import { GalleryPage } from './pages/GalleryPage'
import { BrowserRouter, Routes, Route } from 'react-router-dom'
import { CharacterPage } from './pages/CharacterPage'

export const CHARACTER_API_URL = 'http://localhost:8080/api/v0'

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<GalleryPage />} />
        <Route path="/character/:id" element={<CharacterPage />} />
      </Routes>
    </BrowserRouter>
  )
}

export default App
