import { useState, useEffect } from 'react';
import type { CharacterData } from '../types/CharacterData';
import { CharacterCard } from '../components/CharacterCard';
import { CHARACTER_API_URL } from '../App';
import { Link } from 'react-router-dom';

export function GalleryPage() {
    const [chars, setChars] = useState<CharacterData[]>([]);

    useEffect(() => {
        fetch(`${CHARACTER_API_URL}/characters`)
            .then(res => res.json())
            .then(data => setChars(data))
            .catch(error => console.error("Error fetching characters: ", error));
    }, []);

    return (
        <>
            <header className='bg-stone-950 text-white mb-3 flex flex-row font-[Fredoka]'>
                <span>
                    <Link to="/"><strong>CHARACTER GALLERY</strong></Link>
                </span>
                <button className='bg-stone-800 text-white p-2 rounded-lg'>
                    <Link to="/submit">Submit Character</Link>
                </button>
            </header>
            <main className='bg-stone-900 flex items-center flex-col justify-center min-h-screen'>
                {chars.length === 0 && <p className='text-white'>No characters found</p>}

                <ul className='flex gap-2 flex-wrap flex-col justify-center list-none p-0 m-0'>
                    {
                        chars.map((char) => (
                            <li key={char.id}>
                                <CharacterCard data={char} />
                            </li>)
                        )
                    }
                </ul>
            </main>
        </>
    );
}