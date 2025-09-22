-- name: CreateUser :one
INSERT INTO users (id, email)
VALUES ($1, $2)
RETURNING id, email, created_at;

-- name: GetUserByID :one
SELECT id, email, created_at
FROM users
WHERE id = $1;

-- name: GetUserByEmail :one
SELECT id, email, created_at
FROM users
WHERE email = $1;
