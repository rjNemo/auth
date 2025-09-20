package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"
)

var (
	// ErrInvalidInput indicates the caller supplied malformed credentials.
	ErrInvalidInput = errors.New("auth: invalid input")
	// ErrInvalidCredentials indicates the credentials do not match any account.
	ErrInvalidCredentials = errors.New("auth: invalid credentials")
	// ErrEmailExists indicates an account already uses the provided email address.
	ErrEmailExists = errors.New("auth: email already registered")
)

const userIDByteLength = 16

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
		CreatedAt:    time.Now().UTC(),
	}

	if err := s.store.Create(ctx, user); err != nil {
		return nil, err
	}

	return &user, nil
}

// TODO: could be UUID. return a dedicated type
func generateUserID() (string, error) {
	buf := make([]byte, userIDByteLength)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}
