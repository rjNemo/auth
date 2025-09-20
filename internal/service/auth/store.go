package auth

import (
	"context"
	"errors"
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
