-- name: CreateLoginEvent :one
INSERT INTO login_events (user_id, provider, success, ip, user_agent)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, user_id, provider, success, ip, user_agent, created_at;

-- name: ListLoginEventsForUser :many
SELECT id, user_id, provider, success, ip, user_agent, created_at
FROM login_events
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2;
