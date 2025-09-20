package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"log"
)

const saltLen = 32

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
