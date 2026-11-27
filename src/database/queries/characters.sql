-- name: CreateCharacter :one
INSERT INTO characters (
    owner_id, name, body_type, species, class, level, xp, hp_max, hp_current
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
)
RETURNING *;

-- name: CreateStats :one
INSERT INTO stats (
    character_id, strength, dexterity, constitution, intelligence, wisdom, charisma
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: CreateCustomization :one
INSERT INTO customizations (
    character_id, hair, beard, skin_tone, hair_color, cloak, accent_color
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: SelectCharacter :one
SELECT * FROM characters
WHERE id = $1;

-- name: SelectCharacterStats :one
SELECT * FROM stats
WHERE character_id = $1;

-- name: SelectCharacterCustomization :one
SELECT * FROM customizations
WHERE character_id = $1;

-- name: SelectCharacterFull :one
SELECT c.id, c.owner_id, c.name, c.body_type, c.species, c.class,
       c.level, c.xp, c.hp_max, c.hp_current, c.prestige,
       s.strength, s.dexterity, s.constitution, s.intelligence, s.wisdom, s.charisma,
       cu.hair, cu.beard, cu.skin_tone, cu.hair_color, cu.cloak, cu.accent_color
FROM characters c
JOIN stats s ON s.character_id = c.id
JOIN customizations cu ON cu.character_id = c.id
WHERE c.id = $1;

-- name: SelectAllCharactersFull :many
SELECT c.id, c.owner_id, c.name, c.body_type, c.species, c.class,
       c.level, c.xp, c.hp_max, c.hp_current, c.prestige,
       s.strength, s.dexterity, s.constitution, s.intelligence, s.wisdom, s.charisma,
       cu.hair, cu.beard, cu.skin_tone, cu.hair_color, cu.cloak, cu.accent_color
FROM characters c
JOIN stats s ON s.character_id = c.id
JOIN customizations cu ON cu.character_id = c.id
ORDER BY c.id
LIMIT $1 OFFSET $2;

-- name: SelectCharactersByOwnerFull :many
SELECT c.id, c.owner_id, c.name, c.body_type, c.species, c.class,
       c.level, c.xp, c.hp_max, c.hp_current, c.prestige,
       s.strength, s.dexterity, s.constitution, s.intelligence, s.wisdom, s.charisma,
       cu.hair, cu.beard, cu.skin_tone, cu.hair_color, cu.cloak, cu.accent_color
FROM characters c
JOIN stats s ON s.character_id = c.id
JOIN customizations cu ON cu.character_id = c.id
WHERE c.owner_id = $1
ORDER BY c.id
LIMIT $2 OFFSET $3;

-- name: CountAllCharacters :one
SELECT COUNT(*) FROM characters;

-- name: UpdateCharacter :one
UPDATE characters
SET name = $2, body_type = $3, species = $4, class = $5,
    level = $6, xp = $7, hp_max = $8, hp_current = $9
WHERE id = $1
RETURNING *;

-- name: UpdateCharacterOwned :one
UPDATE characters
SET name = $3, body_type = $4, species = $5, class = $6,
    level = $7, xp = $8, hp_max = $9, hp_current = $10
WHERE id = $1 AND owner_id = $2
RETURNING *;

-- name: UpdateStats :one
UPDATE stats
SET strength = $2, dexterity = $3, constitution = $4,
    intelligence = $5, wisdom = $6, charisma = $7
WHERE character_id = $1
RETURNING *;

-- name: UpdateCustomization :one
UPDATE customizations
SET hair = $2, beard = $3, skin_tone = $4, hair_color = $5, cloak = $6, accent_color = $7
WHERE character_id = $1
RETURNING *;

-- name: DeleteCharacter :execrows
DELETE FROM characters
WHERE id = $1;

-- name: DeleteCharacterOwned :execrows
DELETE FROM characters
WHERE id = $1 AND owner_id = $2;

-- name: CountCharactersByOwner :one
SELECT COUNT(*) FROM characters
WHERE owner_id = $1;

-- name: UpdateCharacterHP :one
UPDATE characters
SET hp_current = LEAST(GREATEST($2, 0), hp_max)
WHERE id = $1
RETURNING *;

-- name: HealCharacter :one
UPDATE characters
SET hp_current = LEAST(hp_current + $2, hp_max)
WHERE id = $1
RETURNING *;

-- name: UpdateCharacterProgress :one
UPDATE characters
SET level = $2,
    xp = $3,
    hp_max = $4,
    hp_current = $5
WHERE id = $1
RETURNING *;

-- name: PrestigeCharacter :one
UPDATE characters
SET level = 1,
    xp = 0,
    hp_max = $2,
    hp_current = $2,
    prestige = prestige + 1
WHERE id = $1 AND owner_id = $3 AND level >= 10
RETURNING *;

-- name: BoostAllStatsOnPrestige :one
UPDATE stats
SET strength     = LEAST(strength + 1, 99),
    dexterity    = LEAST(dexterity + 1, 99),
    constitution = LEAST(constitution + 1, 99),
    intelligence = LEAST(intelligence + 1, 99),
    wisdom       = LEAST(wisdom + 1, 99),
    charisma     = LEAST(charisma + 1, 99)
WHERE character_id = $1
RETURNING *;
