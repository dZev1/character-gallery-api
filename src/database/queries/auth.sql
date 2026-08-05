-- name: CreateUser :one
INSERT INTO users (id, username, password_hash)
VALUES ($1, $2, $3)
RETURNING *;

-- name: SelectUserByUsername :one
SELECT * FROM users WHERE username = $1;

-- name: SelectUserByID :one
SELECT * FROM users WHERE id = $1;
