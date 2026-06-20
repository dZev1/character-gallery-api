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
    )
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
    hair SMALLINT NOT NULL CHECK (hair BETWEEN 0 AND 30),
    face SMALLINT NOT NULL CHECK (face BETWEEN 0 AND 30),
    shirt SMALLINT NOT NULL CHECK (shirt BETWEEN 0 AND 30),
    pants SMALLINT NOT NULL CHECK (pants BETWEEN 0 AND 30),
    shoes SMALLINT NOT NULL CHECK (shoes BETWEEN 0 AND 30)
);

