import { CardField } from './CardField';
import { StatsBox } from './StatsBox';
import type { CharacterData } from '../types/CharacterData';
import type { StatsData } from '../types/StatsData';
import HumanB from '../assets/human_b_base.svg'

export interface CharacterProps {
    data: CharacterData | null;
}

export interface StatsProps {
    data: StatsData
}

export function CharacterCard({ data }: CharacterProps) {
    if (!data) {
        return (
            <div className="flex gap-5 w-[400pt] h-[270pt] bg-slate-400 rounded-3xl p-5 border-[3pt] border-solid border-white font-[Fredoka] text-black box-border">
                Character not found
            </div>
        );
    }
    const stats = data.stats;
    const bodyType = data.body_type == "type_a" ? "Type A" : "Type B"

    return (
        <div className="flex gap-5 w-[400pt] h-[270pt] bg-slate-400 rounded-3xl p-5 border-[3pt] border-solid border-white font-[Fredoka] text-black box-border">
            <div className="flex flex-col h-full items-center w-36">
                <img
                    src={HumanB}
                    className='w-full h-[295px] flex-1 border-solid border-black border-4 rounded-3xl'
                />
                <span className='font-bold flex-1 text-xl tracking-[1px]'>
                    ID: {String(data.id).padStart(8, '0')}
                </span>
            </div>
            <div className='flex flex-col flex-1 justify-between'>
                <CardField fieldName='NAME' fieldVal={data.name} />
                <CardField fieldName='SPECIES' fieldVal={data.species} />
                <CardField fieldName='CLASS' fieldVal={data.class} />
                <CardField fieldName='BODY TYPE' fieldVal={bodyType} />
                <StatsBox data={stats} />
            </div>
        </div>
    )
}

