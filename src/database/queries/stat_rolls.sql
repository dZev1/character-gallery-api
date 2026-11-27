-- name: GetUserStatRoll :one
SELECT * FROM user_stat_rolls
WHERE user_id = $1;

-- name: UpsertUserStatRoll :one
INSERT INTO user_stat_rolls (
    user_id, rolls, last_reroll_at, last_character_created_at, updated_at
) VALUES (
    $1, $2, $3, $4, NOW()
)
ON CONFLICT (user_id) DO UPDATE SET
    rolls = EXCLUDED.rolls,
    last_reroll_at = EXCLUDED.last_reroll_at,
    last_character_created_at = EXCLUDED.last_character_created_at,
    updated_at = NOW()
RETURNING *;

-- name: ConsumeUserStatRoll :one
UPDATE user_stat_rolls
SET rolls = '{}',
    last_character_created_at = NOW(),
    updated_at = NOW()
WHERE user_id = $1
RETURNING *;
