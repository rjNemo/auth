-- name: CreateUserPassword :exec
INSERT INTO user_passwords (user_id, password_hash, password_salt, algorithm)
VALUES ($1, $2, $3, $4);

-- name: UpdateUserPassword :exec
UPDATE user_passwords
SET password_hash = $1,
    password_salt = $2,
    algorithm = $3,
    updated_at = now()
WHERE user_id = $4;

-- name: GetUserPassword :one
SELECT user_id, password_hash, password_salt, algorithm, created_at, updated_at
FROM user_passwords
WHERE user_id = $1;
