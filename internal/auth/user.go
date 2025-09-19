package auth

import "time"

// User represents authenticated account details.
type User struct {
	ID           string
	Email        string
	PasswordSalt string
	PasswordHash string
	CreatedAt    time.Time
}
