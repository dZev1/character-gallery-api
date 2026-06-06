-- name: ValidateAPIKey :one
SELECT EXISTS(
    SELECT 1 FROM api_keys
    WHERE key_hash = $1 AND is_active = true
);

-- name: UpdateLastUsed :exec
UPDATE api_keys
SET last_used_at = NOW()
WHERE key_hash = $1;

-- name: CreateAPIKey :one
INSERT INTO api_keys (key_hash, name, is_active)
VALUES ($1, $2, true)
RETURNING *;
