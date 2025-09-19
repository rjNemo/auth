package identity

import "strings"

// NormalizeEmail trims whitespace and lowercases an email for canonical comparisons.
func NormalizeEmail(email string) string {
	return strings.TrimSpace(strings.ToLower(email))
}
