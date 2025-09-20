package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"unicode"
	"unicode/utf8"
)

const saltLen = 32
const passwordMinLength = 8

var ErrWeakPassword = errors.New("auth: password does not meet complexity requirements")

// ValidatePassword ensures a password satisfies baseline complexity rules.
func ValidatePassword(password string) error {
	if utf8.RuneCountInString(password) < passwordMinLength {
		return fmt.Errorf("%w: minimum length %d", ErrWeakPassword, passwordMinLength)
	}

	var hasUpper, hasDigit bool
	for _, r := range password {
		if unicode.IsUpper(r) {
			hasUpper = true
		}
		if unicode.IsDigit(r) {
			hasDigit = true
		}
		if hasUpper && hasDigit {
			break
		}
	}

	if !hasUpper {
		return fmt.Errorf("%w: missing uppercase letter", ErrWeakPassword)
	}
	if !hasDigit {
		return fmt.Errorf("%w: missing numeric character", ErrWeakPassword)
	}

	return nil
}

// HashPassword returns a base64-encoded salt and hash for the provided plaintext.
func HashPassword(plain string) (salt string, hash string, err error) {
	if plain == "" {
		return "", "", fmt.Errorf("password cannot be empty")
	}

	rawSalt := make([]byte, saltLen)
	if _, err = rand.Read(rawSalt); err != nil {
		return "", "", fmt.Errorf("generate salt: %w", err)
	}

	salt = base64.StdEncoding.EncodeToString(rawSalt)
	hash = encodeHash(rawSalt, plain)

	return salt, hash, nil
}

// VerifyPassword reports whether the supplied plaintext matches the salt+hash pair.
func VerifyPassword(plain, salt, expectedHash string) bool {
	if plain == "" || salt == "" || expectedHash == "" {
		return false
	}

	rawSalt, err := base64.StdEncoding.DecodeString(salt)
	if err != nil {
		log.Printf("auth: invalid salt encoding: %v", err)
		return false
	}

	calculated := encodeHash(rawSalt, plain)
	return subtle.ConstantTimeCompare([]byte(calculated), []byte(expectedHash)) == 1
}

func encodeHash(salt []byte, plain string) string {
	digest := sha256.Sum256(append(salt, plain...))
	return base64.StdEncoding.EncodeToString(digest[:])
}
