package auth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"strings"
	"time"
)

// User represents authenticated account details.
type User struct {
	ID           string
	Email        UserEmail
	PasswordSalt string
	PasswordHash string
	CreatedAt    time.Time
}

// UserEmail represents a canonical email string.
type UserEmail string

// NewUserEmail constructs a canonical email value or reports an error if empty.
func NewUserEmail(raw string) (UserEmail, error) {
	normalized := strings.TrimSpace(strings.ToLower(raw))
	if normalized == "" {
		return "", errors.New("auth: email required")
	}
	return UserEmail(normalized), nil
}

// MustUserEmail is a helper for trusted inputs when failure is non-recoverable.
func MustUserEmail(raw string) UserEmail {
	email, err := NewUserEmail(raw)
	if err != nil {
		panic(err)
	}
	return email
}

// String exposes the underlying string.
func (e UserEmail) String() string {
	return string(e)
}

// IsZero reports whether the email is unset.
func (e UserEmail) IsZero() bool {
	return e == ""
}

// TODO: could be UUID. return a dedicated type
func generateUserID() (string, error) {
	buf := make([]byte, userIDByteLength)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}
