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
    cooldown, capacity
) VALUES (
    $1, $2, $3, $4, $5,
    $6, $7, $8, $9, $10,
    $11, $12
)
RETURNING *;

-- name: SeedItems :copyfrom
INSERT INTO items (
    name, type, description, equippable, rarity,
    damage, defense, heal_amount, mana_cost, duration,
    cooldown, capacity
) VALUES (
    $1, $2, $3, $4, $5,
    $6, $7, $8, $9, $10,
    $11, $12
);
