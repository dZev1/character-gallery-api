-- name: FindAllItems :many
SELECT * FROM items
ORDER BY id
LIMIT $1 OFFSET $2;

-- name: FindItem :one
SELECT * FROM items
WHERE id = $1;

-- name: CountAllItems :one
SELECT COUNT(*) FROM items;

-- name: CreateItem :one
INSERT INTO items (
    name, type, description, equippable, rarity,
    damage, defense, heal_amount, mana_cost, duration,
    cooldown, capacity, image_url
) VALUES (
    $1, $2, $3, $4, $5,
    $6, $7, $8, $9, $10,
    $11, $12, $13
)
RETURNING *;

-- name: SeedItems :copyfrom
INSERT INTO items (
    name, type, description, equippable, rarity,
    damage, defense, heal_amount, mana_cost, duration,
    cooldown, capacity, image_url
) VALUES (
    $1, $2, $3, $4, $5,
    $6, $7, $8, $9, $10,
    $11, $12, $13
);

-- name: FindItemsByRarityRange :many
SELECT * FROM items
WHERE rarity BETWEEN $1 AND $2
ORDER BY id;

-- name: FindRandomItemByRarityRange :one
SELECT * FROM items
WHERE rarity BETWEEN $1 AND $2
ORDER BY RANDOM()
LIMIT 1;

-- name: UpdateItem :one
UPDATE items
SET name        = $2,
    type        = $3,
    description = $4,
    equippable  = $5,
    rarity      = $6,
    damage      = $7,
    defense     = $8,
    heal_amount = $9,
    mana_cost   = $10,
    duration    = $11,
    cooldown    = $12,
    capacity    = $13,
    image_url   = $14
WHERE id = $1
RETURNING *;

-- name: DeleteItem :execrows
DELETE FROM items
WHERE id = $1;
