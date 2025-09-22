package auth

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rjnemo/auth/internal/driver/db"
)

const passwordAlgorithm = "sha256"

// SQLStore persists users in PostgreSQL via generated sqlc queries.
type SQLStore struct {
	pool    *pgxpool.Pool
	queries *db.Queries
}

// NewSQLStore builds a SQL-backed user store.
func NewSQLStore(pool *pgxpool.Pool) *SQLStore {
	return &SQLStore{
		pool:    pool,
		queries: db.New(pool),
	}
}

// FindByEmail returns the stored user aggregate by canonical email address.
func (s *SQLStore) FindByEmail(ctx context.Context, email UserEmail) (*User, error) {
	row, err := s.queries.GetUserByEmail(ctx, email.String())
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("lookup user: %w", err)
	}

	normalizedEmail, err := NewUserEmail(row.Email)
	if err != nil {
		return nil, fmt.Errorf("normalize email: %w", err)
	}

	user := &User{
		ID:        row.ID.String(),
		Email:     normalizedEmail,
		CreatedAt: timestamptzValue(row.CreatedAt),
	}

	if pw, err := s.queries.GetUserPassword(ctx, row.ID); err == nil {
		user.PasswordSalt = base64.StdEncoding.EncodeToString(pw.PasswordSalt)
		user.PasswordHash = base64.StdEncoding.EncodeToString(pw.PasswordHash)
		user.Provider = ProviderPassword
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("load password: %w", err)
	}

	oauthAccounts, err := s.queries.ListUserOAuthAccountsByUserID(ctx, row.ID)
	if err != nil {
		return nil, fmt.Errorf("load oauth accounts: %w", err)
	}

	if len(oauthAccounts) > 0 {
		acct := oauthAccounts[0]
		if user.Provider == "" {
			user.Provider = acct.Provider
		}
		user.OAuthSubject = acct.Subject
		user.OAuthEmailVerified = acct.EmailVerified
	}

	if user.Provider == "" {
		user.Provider = ProviderPassword
	}

	return user, nil
}

// Create writes a new user aggregate to persistent storage.
func (s *SQLStore) Create(ctx context.Context, user User) error {
	if user.Email.IsZero() {
		return ErrEmailRequired
	}

	id, err := uuid.Parse(user.ID)
	if err != nil {
		return fmt.Errorf("parse user id: %w", err)
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	qtx := s.queries.WithTx(tx)

	if _, err = qtx.CreateUser(ctx, db.CreateUserParams{ID: id, Email: user.Email.String()}); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrEmailExists
		}
		return fmt.Errorf("insert user: %w", err)
	}

	switch user.Provider {
	case ProviderPassword:
		if user.PasswordHash == "" || user.PasswordSalt == "" {
			return fmt.Errorf("password credentials required")
		}
		hashBytes, err := base64.StdEncoding.DecodeString(user.PasswordHash)
		if err != nil {
			return fmt.Errorf("decode password hash: %w", err)
		}
		saltBytes, err := base64.StdEncoding.DecodeString(user.PasswordSalt)
		if err != nil {
			return fmt.Errorf("decode password salt: %w", err)
		}

		if err := qtx.CreateUserPassword(ctx, db.CreateUserPasswordParams{
			UserID:       id,
			PasswordHash: hashBytes,
			PasswordSalt: saltBytes,
			Algorithm:    passwordAlgorithm,
		}); err != nil {
			return fmt.Errorf("insert password: %w", err)
		}
	default:
		if user.OAuthSubject == "" {
			return ErrSubjectRequired
		}

		var emailValue pgtype.Text
		if !user.Email.IsZero() {
			emailValue = pgtype.Text{String: user.Email.String(), Valid: true}
		}

		if _, err := qtx.CreateUserOAuthAccount(ctx, db.CreateUserOAuthAccountParams{
			UserID:        id,
			Provider:      user.Provider,
			Subject:       user.OAuthSubject,
			Email:         emailValue,
			EmailVerified: user.OAuthEmailVerified,
			Profile:       nil,
		}); err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				return ErrEmailExists
			}
			return fmt.Errorf("insert oauth account: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

func timestamptzValue(ts pgtype.Timestamptz) time.Time {
	if !ts.Valid {
		return time.Time{}
	}
	return ts.Time
}
