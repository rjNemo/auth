package auth

import (
	"context"
	"errors"
	"testing"
)

func TestServiceAuthenticate(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := NewMemoryStore()
	service := NewService(store)

	email := MustUserEmail("user@example.com")
	salt, hash, err := HashPassword("Password123")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	if err := store.Create(ctx, User{Email: email, PasswordSalt: salt, PasswordHash: hash, Provider: ProviderPassword}); err != nil {
		t.Fatalf("seed user: %v", err)
	}

	tests := map[string]struct {
		email    UserEmail
		password string
		wantErr  error
	}{
		"invalid input":   {email: email, password: "", wantErr: ErrInvalidInput},
		"weak password":   {email: email, password: "short1", wantErr: ErrWeakPassword},
		"unknown account": {email: MustUserEmail("missing@example.com"), password: "Password123", wantErr: ErrInvalidCredentials},
		"wrong password":  {email: email, password: "Password999", wantErr: ErrInvalidCredentials},
		"success":         {email: email, password: "Password123", wantErr: nil},
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			account, err := service.Authenticate(ctx, tc.email, tc.password)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected %v, got %v", tc.wantErr, err)
				}
				if account != nil {
					t.Fatalf("expected no account, got %#v", account)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if account == nil {
				t.Fatalf("expected account")
			}
			if account.Email != email {
				t.Fatalf("expected email %q, got %q", email, account.Email)
			}
		})
	}
}

func TestServiceLookupByEmail(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := NewMemoryStore()
	service := NewService(store)

	email := MustUserEmail("lookup@example.com")
	if err := store.Create(ctx, User{Email: email, Provider: ProviderPassword}); err != nil {
		t.Fatalf("seed user: %v", err)
	}

	cases := map[string]struct {
		email   UserEmail
		wantErr error
	}{
		"zero":    {email: UserEmail(""), wantErr: ErrInvalidInput},
		"missing": {email: MustUserEmail("none@example.com"), wantErr: ErrUserNotFound},
		"found":   {email: email, wantErr: nil},
	}

	for name, tc := range cases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			user, err := service.LookupByEmail(ctx, tc.email)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected %v, got %v", tc.wantErr, err)
				}
				if user != nil {
					t.Fatalf("expected nil user, got %#v", user)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if user == nil || user.Email != email {
				t.Fatalf("expected user with email %q", email)
			}
		})
	}
}

func TestServiceRegister(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := NewMemoryStore()
	service := NewService(store)

	email := MustUserEmail("taken@example.com")
	if err := store.Create(ctx, User{Email: email, Provider: ProviderPassword}); err != nil {
		t.Fatalf("seed user: %v", err)
	}

	tests := map[string]struct {
		email    UserEmail
		password string
		wantErr  error
	}{
		"invalid input": {email: UserEmail(""), password: "", wantErr: ErrInvalidInput},
		"weak password": {email: MustUserEmail("weak@example.com"), password: "weak", wantErr: ErrWeakPassword},
		"duplicate":     {email: email, password: "Password123", wantErr: ErrEmailExists},
		"success":       {email: MustUserEmail("new@example.com"), password: "Password123", wantErr: nil},
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			user, err := service.Register(ctx, tc.email, tc.password)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected %v, got %v", tc.wantErr, err)
				}
				if user != nil {
					t.Fatalf("expected nil user, got %#v", user)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if user.Email != tc.email {
				t.Fatalf("expected email %q, got %q", tc.email, user.Email)
			}
			if user.CreatedAt.IsZero() {
				t.Fatal("expected CreatedAt to be set")
			}
			// Ensure the user is persisted with hashed credentials.
			persisted, err := store.FindByEmail(ctx, tc.email)
			if err != nil {
				t.Fatalf("expected persisted user: %v", err)
			}
			if persisted.PasswordSalt == "" || persisted.PasswordHash == "" {
				t.Fatal("expected password salt/hash to be stored")
			}
		})
	}
}

func TestServiceEnsureExternalUser(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := NewMemoryStore()
	service := NewService(store)

	googleEmail := MustUserEmail("google@example.com")
	if err := store.Create(ctx, User{ID: "existing-google", Email: googleEmail, Provider: ProviderGoogle, OAuthSubject: "existing-sub"}); err != nil {
		t.Fatalf("seed external user: %v", err)
	}

	tests := map[string]struct {
		email    UserEmail
		provider string
		subject  string
		verified bool
		wantErr  error
		wantNew  bool
	}{
		"missing email":    {email: UserEmail(""), provider: ProviderGoogle, subject: "sub", wantErr: ErrInvalidInput},
		"missing provider": {email: MustUserEmail("new@example.com"), provider: "", subject: "sub", wantErr: ErrProviderRequired},
		"missing subject":  {email: MustUserEmail("new@example.com"), provider: ProviderGoogle, subject: "", wantErr: ErrSubjectRequired},
		"existing":         {email: googleEmail, provider: ProviderGoogle, subject: "existing-sub", wantErr: nil, wantNew: false},
		"provision":        {email: MustUserEmail("brandnew@example.com"), provider: ProviderGoogle, subject: "new-sub", verified: true, wantErr: nil, wantNew: true},
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			user, err := service.EnsureExternalUser(ctx, tc.email, tc.provider, tc.subject, tc.verified)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected %v, got %v", tc.wantErr, err)
				}
				if user != nil {
					t.Fatalf("expected nil user, got %#v", user)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if user == nil {
				t.Fatal("expected user")
			}
			if user.Email != tc.email {
				t.Fatalf("expected email %q, got %q", tc.email, user.Email)
			}
			if user.Provider != tc.provider {
				t.Fatalf("expected provider %q, got %q", tc.provider, user.Provider)
			}
			persisted, err := store.FindByEmail(ctx, tc.email)
			if err != nil {
				t.Fatalf("expected user persisted: %v", err)
			}
			if persisted.OAuthSubject != "" && persisted.OAuthSubject != tc.subject {
				t.Fatalf("expected oauth subject %q, got %q", tc.subject, persisted.OAuthSubject)
			}
			if tc.wantNew && persisted.CreatedAt.IsZero() {
				t.Fatal("expected created at timestamp for new user")
			}
		})
	}
}
