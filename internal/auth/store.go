package auth

import (
	"context"
	"errors"
	"sync"
)

// ErrUserNotFound signals no user exists for the provided lookup criteria.
var (
	ErrUserNotFound  = errors.New("auth: user not found")
	ErrEmailRequired = errors.New("auth: email required")
)

// UserStore defines persistence expectations for user lookups.
type UserStore interface {
	FindByEmail(ctx context.Context, email UserEmail) (*User, error)
	Create(ctx context.Context, user User) error
}

// MemoryStore is an in-memory implementation of UserStore for development and tests.
type MemoryStore struct {
	mu    sync.RWMutex
	users map[string]User
}

// NewMemoryStore builds an empty MemoryStore instance.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{users: make(map[string]User)}
}

// FindByEmail returns a copy of the stored user.
func (s *MemoryStore) FindByEmail(_ context.Context, email UserEmail) (*User, error) {
	if email.IsZero() {
		return nil, ErrUserNotFound
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	user, ok := s.users[email.String()]
	if !ok {
		return nil, ErrUserNotFound
	}

	userCopy := user
	return &userCopy, nil
}

// Create inserts or replaces the stored user by email.

func (s *MemoryStore) Create(_ context.Context, user User) error {
	if user.Email.IsZero() {
		return ErrEmailRequired
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.users == nil {
		s.users = make(map[string]User)
	}

	s.users[user.Email.String()] = user
	return nil
}
