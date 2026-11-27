-- name: FindAllEnemies :many
SELECT * FROM enemies
ORDER BY tier ASC, id ASC
LIMIT $1 OFFSET $2;

-- name: CountAllEnemies :one
SELECT COUNT(*) FROM enemies;

-- name: FindEnemy :one
SELECT * FROM enemies
WHERE id = $1;

-- name: FindEnemiesByTier :many
SELECT * FROM enemies
WHERE tier = $1
ORDER BY id ASC;

-- name: FindEnemiesByTierAndType :many
SELECT * FROM enemies
WHERE tier = $1 AND type = $2
ORDER BY id ASC;

-- name: CreateEnemy :one
INSERT INTO enemies (
    name,
    type,
    tier,
    description,
    hp_max,
    ac,
    damage,
    xp_reward,
    score_reward,
    image_url
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
)
RETURNING *;

-- name: CreateEnemyIfNotExists :one
INSERT INTO enemies (
    name,
    type,
    tier,
    description,
    hp_max,
    ac,
    damage,
    xp_reward,
    score_reward,
    image_url
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
)
ON CONFLICT (name, tier) DO NOTHING
RETURNING *;

-- name: UpdateEnemy :one
UPDATE enemies
SET name         = $2,
    type         = $3,
    tier         = $4,
    description  = $5,
    hp_max       = $6,
    ac           = $7,
    damage       = $8,
    xp_reward    = $9,
    score_reward = $10,
    image_url    = $11
WHERE id = $1
RETURNING *;

-- name: DeleteEnemy :execrows
DELETE FROM enemies
WHERE id = $1;

-- name: SeedEnemies :copyfrom
INSERT INTO enemies (
    name,
    type,
    tier,
    description,
    hp_max,
    ac,
    damage,
    xp_reward,
    score_reward,
    image_url
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
);
