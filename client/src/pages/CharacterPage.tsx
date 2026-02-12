import { CharacterCard } from '../components/CharacterCard';
import type { CharacterData } from '../types/CharacterData';
import { useState, useEffect } from 'react';
import { Link, useParams } from 'react-router-dom';
import { CHARACTER_API_URL } from '../App';

export function CharacterPage() {
    const id = Number(useParams<{ id: string }>().id);
    const [character, setCharacter] = useState<CharacterData | null>(null);

    useEffect(() => {
        fetch(`${CHARACTER_API_URL}/characters/${id}`)
            .then(res => res.json())
            .then(data => setCharacter(data))
            .catch(error => console.error("Error fetching character: ", error));
    }, [id]);

    return (
        <>
            <header className='bg-stone-950 text-white mb-3 flex flex-row font-[Fredoka]'>
                <span>
                    <Link to="/"><strong>CHARACTER GALLERY</strong></Link>
                </span>
                <button>
                    Submit Character
                </button>
            </header>
            <main className='bg-stone-900 flex items-center flex-col justify-center min-h-screen'>
                {character && <CharacterCard data={character} />}
            </main>
        </>
    )
}