-- name: CreateCharacter :one
INSERT INTO characters (
    name, body_type, species, class
) VALUES (
    $1, $2, $3, $4
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
    character_id, hair, face, shirt, pants, shoes
) VALUES (
    $1, $2, $3, $4, $5, $6
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

-- name: SelectAllCharacters :many
SELECT * FROM characters
ORDER BY id
LIMIT $1 OFFSET $2;

-- name: CountAllCharacters :one
SELECT COUNT(*) FROM characters;

-- name: UpdateCharacter :one
UPDATE characters
SET name = $2, body_type = $3, species = $4, class = $5
WHERE id = $1
RETURNING *;

-- name: UpdateStats :one
UPDATE stats
SET strength = $2, dexterity = $3, constitution = $4,
    intelligence = $5, wisdom = $6, charisma = $7
WHERE character_id = $1
RETURNING *;

-- name: UpdateCustomization :one
UPDATE customizations
SET hair = $2, face = $3, shirt = $4, pants = $5, shoes = $6
WHERE character_id = $1
RETURNING *;

-- name: DeleteCharacter :exec
DELETE FROM characters
WHERE id = $1;
