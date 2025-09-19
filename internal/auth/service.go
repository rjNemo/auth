package auth

import (
	"context"
	"errors"
)

var (
	// ErrInvalidInput indicates the caller supplied malformed credentials.
	ErrInvalidInput = errors.New("auth: invalid input")
	// ErrInvalidCredentials indicates the credentials do not match any account.
	ErrInvalidCredentials = errors.New("auth: invalid credentials")
)

// Service exposes authentication business operations to HTTP handlers.
type Service struct {
	store UserStore
}

// NewService wires a Service with the provided persistence implementation.
func NewService(store UserStore) *Service {
	return &Service{store: store}
}

// Authenticate validates the provided email/password and returns the account on success.
func (s *Service) Authenticate(ctx context.Context, email UserEmail, password string) (*User, error) {
	if email.IsZero() || password == "" {
		return nil, ErrInvalidInput
	}

	account, err := s.store.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if !VerifyPassword(password, account.PasswordSalt, account.PasswordHash) {
		return nil, ErrInvalidCredentials
	}

	return account, nil
}
