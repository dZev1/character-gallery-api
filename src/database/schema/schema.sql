CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'user',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS items (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    type TEXT NOT NULL CHECK (
        type IN (
            'armor', 'ring', 'weapon', 'shield', 'tool',
            'adventuring_gear', 'rod', 'staff', 'wand',
            'scroll', 'potion', 'ammo', 'consumable', 'wondrous_item'
        )
    ),
    description TEXT NOT NULL,
    equippable BOOLEAN NOT NULL DEFAULT FALSE,
    rarity INT NOT NULL,

    damage INT,
    defense INT,
    heal_amount INT,
    mana_cost INT,
    duration INT,
    cooldown INT,
    capacity INT,

    CONSTRAINT items_name_rarity_unique UNIQUE (name, rarity)
);

CREATE TABLE IF NOT EXISTS characters (
    id BIGSERIAL PRIMARY KEY,
    owner_id UUID NOT NULL REFERENCES users(id),
    name TEXT NOT NULL,
    body_type TEXT NOT NULL CHECK (body_type IN ('type_a', 'type_b')),
    species TEXT NOT NULL CHECK (
        species IN (
            'aasimar', 'dragonborn', 'dwarf', 'elf', 'gnome',
            'goliath', 'halfling', 'human', 'orc', 'tiefling'
        )
    ),
    class TEXT NOT NULL CHECK (
        class IN (
            'barbarian', 'bard', 'cleric', 'druid', 'fighter',
            'monk', 'paladin', 'ranger', 'rogue', 'sorcerer',
            'warlock', 'wizard'
        )
    ),
    level       INT  NOT NULL DEFAULT 1,
    xp          INT  NOT NULL DEFAULT 0,
    hp_max      INT  NOT NULL DEFAULT 1,
    hp_current  INT  NOT NULL DEFAULT 1,
    prestige    INT  NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS inventory (
    character_id BIGINT NOT NULL REFERENCES characters (id) ON DELETE CASCADE,
    item_id BIGINT NOT NULL REFERENCES items (id) ON DELETE CASCADE,
    quantity INT NOT NULL CHECK (quantity > 0),
    is_equipped BOOLEAN NOT NULL DEFAULT FALSE,

    PRIMARY KEY (character_id, item_id)
);

CREATE TABLE IF NOT EXISTS stats (
    character_id BIGINT PRIMARY KEY REFERENCES characters (id) ON DELETE CASCADE,
    strength SMALLINT NOT NULL CHECK (strength BETWEEN 1 AND 99),
    dexterity SMALLINT NOT NULL CHECK (dexterity BETWEEN 1 AND 99),
    constitution SMALLINT NOT NULL CHECK (constitution BETWEEN 1 AND 99),
    intelligence SMALLINT NOT NULL CHECK (intelligence BETWEEN 1 AND 99),
    wisdom SMALLINT NOT NULL CHECK (wisdom BETWEEN 1 AND 99),
    charisma SMALLINT NOT NULL CHECK (charisma BETWEEN 1 AND 99)
);

CREATE TABLE IF NOT EXISTS customizations (
    character_id BIGINT PRIMARY KEY REFERENCES characters (id) ON DELETE CASCADE,
    hair SMALLINT NOT NULL CHECK (hair BETWEEN 0 AND 20),
    beard SMALLINT NOT NULL CHECK (beard BETWEEN 0 AND 15),
    skin_tone SMALLINT NOT NULL CHECK (skin_tone BETWEEN 0 AND 15),
    hair_color SMALLINT NOT NULL CHECK (hair_color BETWEEN 0 AND 15),
    cloak SMALLINT NOT NULL CHECK (cloak BETWEEN 0 AND 15),
    accent_color SMALLINT NOT NULL CHECK (accent_color BETWEEN 0 AND 15)
);

CREATE TABLE IF NOT EXISTS enemies (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    type TEXT NOT NULL CHECK (
        type IN ('normal', 'elite', 'mini_boss', 'boss')
    ),
    tier SMALLINT NOT NULL CHECK (tier BETWEEN 1 AND 5),
    description TEXT NOT NULL,
    hp_max INT NOT NULL CHECK (hp_max > 0),
    ac INT NOT NULL CHECK (ac > 0),
    damage INT NOT NULL CHECK (damage > 0),
    xp_reward INT NOT NULL DEFAULT 0,
    score_reward INT NOT NULL DEFAULT 0,

    CONSTRAINT enemies_name_tier_unique UNIQUE (name, tier)
);

CREATE TABLE IF NOT EXISTS daily_runs (
    id BIGSERIAL PRIMARY KEY,
    run_date DATE NOT NULL,
    tier SMALLINT NOT NULL CHECK (tier BETWEEN 1 AND 5),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    character_id BIGINT NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    total_floors SMALLINT NOT NULL CHECK (total_floors IN (10, 15, 20, 25, 30)),
    current_floor SMALLINT NOT NULL DEFAULT 1,
    current_node_id TEXT NOT NULL,
    visited_nodes TEXT[] NOT NULL DEFAULT '{}',
    turns_taken INT NOT NULL DEFAULT 0,
    score INT NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'in_progress' CHECK (
        status IN ('in_progress', 'in_combat', 'victory', 'defeat')
    ),
    current_enemy_id BIGINT REFERENCES enemies(id) ON DELETE SET NULL,
    current_enemy_hp INT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE (run_date, tier, character_id)
);

CREATE TABLE IF NOT EXISTS daily_scores (
    id BIGSERIAL PRIMARY KEY,
    run_date DATE NOT NULL,
    tier SMALLINT NOT NULL CHECK (tier BETWEEN 1 AND 5),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    character_id BIGINT NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    character_name TEXT NOT NULL,
    character_class TEXT NOT NULL,
    character_level INT NOT NULL,
    character_prestige INT NOT NULL DEFAULT 0,
    score INT NOT NULL,
    floors_cleared SMALLINT NOT NULL,
    total_floors SMALLINT NOT NULL,
    turns_taken INT NOT NULL,
    hp_remaining INT NOT NULL,
    completed BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE (run_date, tier, character_id)
);

CREATE INDEX IF NOT EXISTS idx_daily_scores_leaderboard
    ON daily_scores (run_date, tier, score DESC, turns_taken ASC);
