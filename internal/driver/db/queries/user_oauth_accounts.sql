-- name: CreateUserOAuthAccount :one
INSERT INTO user_oauth_accounts (user_id, provider, subject, email, email_verified, profile)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, user_id, provider, subject, email, email_verified, profile, created_at, updated_at;

-- name: GetUserOAuthAccountByProviderSubject :one
SELECT id, user_id, provider, subject, email, email_verified, profile, created_at, updated_at
FROM user_oauth_accounts
WHERE provider = $1 AND subject = $2;

-- name: ListUserOAuthAccountsByUserID :many
SELECT id, user_id, provider, subject, email, email_verified, profile, created_at, updated_at
FROM user_oauth_accounts
WHERE user_id = $1
ORDER BY created_at DESC;
