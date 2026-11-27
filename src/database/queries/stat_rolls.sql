-- name: GetUserStatRoll :one
SELECT * FROM user_stat_rolls
WHERE user_id = $1;

-- name: UpsertUserStatRoll :one
INSERT INTO user_stat_rolls (
    user_id, rolls, rerolls_used, updated_at
) VALUES (
    $1, $2, $3, NOW()
)
ON CONFLICT (user_id) DO UPDATE SET
    rolls = EXCLUDED.rolls,
    rerolls_used = EXCLUDED.rerolls_used,
    updated_at = NOW()
RETURNING *;

-- name: ConsumeUserStatRoll :one
UPDATE user_stat_rolls
SET rolls = '{}',
    rerolls_used = 0,
    last_character_created_at = NOW(),
    updated_at = NOW()
WHERE user_id = $1
RETURNING *;

-- name: RecordUserCharacterDelete :one
INSERT INTO user_stat_rolls (
    user_id, rolls, rerolls_used, last_character_deleted_at, updated_at
) VALUES (
    $1, '{}', 0, NOW(), NOW()
)
ON CONFLICT (user_id) DO UPDATE SET
    last_character_deleted_at = NOW(),
    updated_at = NOW()
RETURNING *;
