-- name: CreateUser :one
INSERT INTO users (email, display_name)
VALUES ($1, $2)
RETURNING id, email, display_name, created_at;

-- name: GetUserByID :one
SELECT id, email, display_name, created_at
FROM users
WHERE id = $1;

-- name: GetUserByEmail :one
SELECT id, email, display_name, created_at
FROM users
WHERE email = $1;
