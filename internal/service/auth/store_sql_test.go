package auth

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	schemaUpSQL = `
CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email CITEXT NOT NULL UNIQUE,
    display_name TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE user_passwords (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    password_hash BYTEA NOT NULL,
    password_salt BYTEA NOT NULL,
    algorithm TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE user_oauth_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider TEXT NOT NULL,
    subject TEXT NOT NULL,
    email TEXT,
    email_verified BOOLEAN NOT NULL DEFAULT false,
    profile JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX user_oauth_accounts_provider_subject_idx
    ON user_oauth_accounts (provider, subject);

CREATE INDEX user_oauth_accounts_user_id_idx
    ON user_oauth_accounts (user_id);

CREATE TABLE login_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    provider TEXT,
    success BOOLEAN NOT NULL,
    ip INET,
    user_agent TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX login_events_user_id_idx ON login_events (user_id);
CREATE INDEX login_events_created_at_idx ON login_events (created_at);
`

	schemaDownSQL = `
DROP TABLE IF EXISTS login_events;
DROP TABLE IF EXISTS user_oauth_accounts;
DROP TABLE IF EXISTS user_passwords;
DROP TABLE IF EXISTS users;
DROP EXTENSION IF EXISTS citext;
DROP EXTENSION IF EXISTS pgcrypto;
`
)

func TestSQLStoreIntegration(t *testing.T) {
	dsn := os.Getenv("AUTH_DATABASE_URL")
	if strings.TrimSpace(dsn) == "" {
		t.Skip("AUTH_DATABASE_URL not set")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("connect database: %v", err)
	}
	t.Cleanup(func() { pool.Close() })

	resetDatabase(t, ctx, pool)

	t.Run("register and authenticate", func(t *testing.T) {
		resetDatabase(t, ctx, pool)

		store := NewSQLStore(pool)
		service := NewService(store)

		email := MustUserEmail("sql-user@example.com")

		user, err := service.Register(ctx, email, "Password123")
		if err != nil {
			t.Fatalf("register user: %v", err)
		}
		if user.ID == "" {
			t.Fatal("expected user id")
		}
		if user.Provider != ProviderPassword {
			t.Fatalf("expected provider %q, got %q", ProviderPassword, user.Provider)
		}

		authenticated, err := service.Authenticate(ctx, email, "Password123")
		if err != nil {
			t.Fatalf("authenticate user: %v", err)
		}
		if authenticated.ID != user.ID {
			t.Fatalf("expected matching user id, got %q", authenticated.ID)
		}
		if authenticated.PasswordHash == "" || authenticated.PasswordSalt == "" {
			t.Fatal("expected persisted password credentials")
		}
	})

	t.Run("ensure external user", func(t *testing.T) {
		resetDatabase(t, ctx, pool)

		store := NewSQLStore(pool)
		service := NewService(store)

		email := MustUserEmail("sql-google@example.com")
		subject := "google-subject-123"

		account, err := service.EnsureExternalUser(ctx, email, ProviderGoogle, subject, true)
		if err != nil {
			t.Fatalf("ensure external user: %v", err)
		}
		if account.Provider != ProviderGoogle {
			t.Fatalf("expected provider %q, got %q", ProviderGoogle, account.Provider)
		}
		if account.OAuthSubject != subject {
			t.Fatalf("expected oauth subject %q, got %q", subject, account.OAuthSubject)
		}

		again, err := service.EnsureExternalUser(ctx, email, ProviderGoogle, subject, true)
		if err != nil {
			t.Fatalf("ensure existing external user: %v", err)
		}
		if again.ID != account.ID {
			t.Fatalf("expected same user id, got %q vs %q", again.ID, account.ID)
		}
	})
}

func resetDatabase(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()

	execStatements := func(stmts []string) {
		for _, stmt := range stmts {
			if strings.TrimSpace(stmt) == "" {
				continue
			}
			if _, execErr := pool.Exec(ctx, stmt); execErr != nil {
				t.Fatalf("exec statement %q: %v", stmt, execErr)
			}
		}
	}

	execStatements(splitSQLStatements(schemaDownSQL))
	execStatements(splitSQLStatements(schemaUpSQL))
}

func splitSQLStatements(section string) []string {
	section = strings.TrimSpace(section)
	if section == "" {
		return nil
	}

	parts := strings.Split(section, ";")
	statements := make([]string, 0, len(parts))
	for _, part := range parts {
		stmt := strings.TrimSpace(part)
		if stmt == "" {
			continue
		}
		statements = append(statements, stmt+";")
	}

	return statements
}
