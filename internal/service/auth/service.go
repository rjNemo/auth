package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	// ErrInvalidInput indicates the caller supplied malformed credentials.
	ErrInvalidInput = errors.New("auth: invalid input")
	// ErrInvalidCredentials indicates the credentials do not match any account.
	ErrInvalidCredentials = errors.New("auth: invalid credentials")
	// ErrEmailExists indicates an account already uses the provided email address.
	ErrEmailExists = errors.New("auth: email already registered")
	// ErrProviderRequired indicates the external provider identifier was missing.
	ErrProviderRequired = errors.New("auth: provider required")
)

const (
	userIDByteLength = 16
	// ProviderPassword identifies accounts managed via email/password.
	ProviderPassword = "password"
	// ProviderGoogle identifies accounts authenticated via Google OAuth2.
	ProviderGoogle = "google"
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
	if err := ValidatePassword(password); err != nil {
		return nil, err
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

// LookupByEmail fetches a user by canonical email.
func (s *Service) LookupByEmail(ctx context.Context, email UserEmail) (*User, error) {
	if email.IsZero() {
		return nil, ErrInvalidInput
	}

	return s.store.FindByEmail(ctx, email)
}

// Register provisions a new user account for the provided credentials.
func (s *Service) Register(ctx context.Context, email UserEmail, password string) (*User, error) {
	if email.IsZero() || password == "" {
		return nil, ErrInvalidInput
	}
	if err := ValidatePassword(password); err != nil {
		return nil, err
	}

	if existing, err := s.store.FindByEmail(ctx, email); err == nil && existing != nil {
		return nil, ErrEmailExists
	} else if err != nil && !errors.Is(err, ErrUserNotFound) {
		return nil, err
	}

	id, err := generateUserID()
	if err != nil {
		return nil, fmt.Errorf("generate user id: %w", err)
	}

	salt, hash, err := HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := User{
		ID:           id,
		Email:        email,
		PasswordSalt: salt,
		PasswordHash: hash,
		Provider:     ProviderPassword,
		CreatedAt:    time.Now().UTC(),
	}

	if err := s.store.Create(ctx, user); err != nil {
		return nil, err
	}

	return &user, nil
}

// EnsureExternalUser retrieves or provisions an account authenticated by an external provider.
func (s *Service) EnsureExternalUser(ctx context.Context, email UserEmail, provider string) (*User, error) {
	if email.IsZero() {
		return nil, ErrInvalidInput
	}
	if strings.TrimSpace(provider) == "" {
		return nil, ErrProviderRequired
	}

	account, err := s.store.FindByEmail(ctx, email)
	switch {
	case err == nil:
		return account, nil
	case !errors.Is(err, ErrUserNotFound):
		return nil, err
	}

	id, err := generateUserID()
	if err != nil {
		return nil, fmt.Errorf("generate user id: %w", err)
	}

	user := User{
		ID:        id,
		Email:     email,
		Provider:  provider,
		CreatedAt: time.Now().UTC(),
	}

	if err := s.store.Create(ctx, user); err != nil {
		return nil, err
	}

	return &user, nil
}
