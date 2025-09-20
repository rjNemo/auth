package auth

import (
	"errors"
	"testing"
)

func TestValidatePassword(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		password string
		wantErr  bool
	}{
		"valid":         {password: "Password1", wantErr: false},
		"too short":     {password: "Pw1", wantErr: true},
		"missing upper": {password: "password1", wantErr: true},
		"missing digit": {password: "Password", wantErr: true},
		"unicode upper": {password: "Åpple9xY", wantErr: false},
		"unicode digit": {password: "Passwørd７", wantErr: false},
	}

	for name, tc := range cases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := ValidatePassword(tc.password)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if !errors.Is(err, ErrWeakPassword) {
					t.Fatalf("expected ErrWeakPassword, got %v", err)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidatePasswordBoundary(t *testing.T) {
	t.Parallel()

	// Exactly eight characters, meets other requirements.
	if err := ValidatePassword("Passw0rd"); err != nil {
		t.Fatalf("expected password to be valid, got %v", err)
	}
}
