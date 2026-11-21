-- name: CreateDailyRun :one
INSERT INTO daily_runs (
    run_date,
    tier,
    user_id,
    character_id,
    total_floors,
    current_floor,
    current_node_id,
    visited_nodes,
    turns_taken,
    score,
    status,
    current_enemy_id,
    current_enemy_hp
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
)
RETURNING *;

-- name: GetDailyRun :one
SELECT * FROM daily_runs
WHERE run_date = $1 AND tier = $2 AND character_id = $3;

-- name: GetDailyRunByID :one
SELECT * FROM daily_runs
WHERE id = $1;

-- name: GetActiveDailyRunForCharacter :one
SELECT * FROM daily_runs
WHERE run_date = $1
  AND character_id = $2
  AND status IN ('in_progress', 'in_combat')
LIMIT 1;

-- name: ListActiveDailyRunsByUser :many
SELECT * FROM daily_runs
WHERE run_date = $1
  AND user_id = $2
  AND status IN ('in_progress', 'in_combat')
ORDER BY updated_at DESC;

-- name: UpdateDailyRunState :one
UPDATE daily_runs
SET current_floor    = $2,
    current_node_id  = $3,
    visited_nodes    = $4,
    turns_taken      = $5,
    score            = $6,
    status           = $7,
    current_enemy_id = $8,
    current_enemy_hp = $9,
    updated_at       = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateDailyRunEnemyHP :one
UPDATE daily_runs
SET current_enemy_hp = $2,
    turns_taken      = $3,
    updated_at       = NOW()
WHERE id = $1
RETURNING *;

-- name: UpsertDailyScore :one
INSERT INTO daily_scores (
    run_date,
    tier,
    user_id,
    character_id,
    character_name,
    character_class,
    character_level,
    character_prestige,
    score,
    floors_cleared,
    total_floors,
    turns_taken,
    hp_remaining,
    completed
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
)
ON CONFLICT (run_date, tier, character_id)
DO UPDATE SET
    character_level    = EXCLUDED.character_level,
    character_prestige = EXCLUDED.character_prestige,
    score              = EXCLUDED.score,
    floors_cleared     = EXCLUDED.floors_cleared,
    turns_taken        = EXCLUDED.turns_taken,
    hp_remaining       = EXCLUDED.hp_remaining,
    completed          = EXCLUDED.completed
RETURNING *;

-- name: GetDailyLeaderboard :many
SELECT ds.id,
       ds.run_date,
       ds.tier,
       ds.user_id,
       u.username,
       ds.character_id,
       ds.character_name,
       ds.character_class,
       ds.character_level,
       ds.character_prestige,
       ds.score,
       ds.floors_cleared,
       ds.total_floors,
       ds.turns_taken,
       ds.hp_remaining,
       ds.completed,
       ds.created_at
FROM daily_scores ds
JOIN users u ON u.id = ds.user_id
WHERE ds.run_date = $1 AND ds.tier = $2
ORDER BY ds.score DESC, ds.turns_taken ASC, ds.created_at ASC
LIMIT $3 OFFSET $4;

-- name: CountDailyLeaderboard :one
SELECT COUNT(*) FROM daily_scores
WHERE run_date = $1 AND tier = $2;

-- name: GetUserDailyScores :many
SELECT * FROM daily_scores
WHERE run_date = $1 AND user_id = $2
ORDER BY tier ASC, score DESC;
