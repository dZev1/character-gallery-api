import { useEffect, useState } from 'react'
import './App.css'
import { CharacterCard } from './components/CharacterCard'
import type { CharacterData } from './types/CharacterData'

const CHARACTER_API_URL = 'http://localhost:8080/api/v0'

function App() {
  const [chars, setChars] = useState<CharacterData[]>([{
    id: 0,
    name: "Numuhukumakiaki'aia Lunamor",
    body_type: "type_a",
    species: "human",
    class: "fighter",
    stats: { strength: 1, dexterity: 0, constitution: 0, intelligence: 0, wisdom: 0, charisma: 0 },
    customization: { hair: 0, face: 0, shirt: 0, pants: 0, shoes: 0 }
  }]);

  useEffect(() => {
    fetch(`${CHARACTER_API_URL}/characters`)
      .then(res => res.json())
      .then(data => setChars(data))
      .catch(error => console.error("Error fetching characters: ", error));
  }, []);

  return (
    <>
      <main className='app-container'>
        {chars.length === 0 && <p>No characters found</p>}

        <ul className='cards-grid'>
          {
            chars.map((char) => (<CharacterCard key={char.id} data={char} />))
          }
        </ul>
      </main>
    </>
  )
}

export default App
