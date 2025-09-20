package config

import (
	"encoding/base64"
	"testing"

	"github.com/rjnemo/auth/internal/driver/logging"
)

func TestNewDefaults(t *testing.T) {
	t.Setenv("AUTH_SESSION_SECRET", base64.StdEncoding.EncodeToString(bytesOfLength(32)))
	cfg, err := New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.ListenAddr != ":8000" {
		t.Fatalf("expected default listen addr, got %s", cfg.ListenAddr)
	}
	if cfg.LogMode != logging.ModeText {
		t.Fatalf("expected default log mode text, got %s", cfg.LogMode)
	}
	if cfg.Environment != "development" {
		t.Fatalf("expected default environment, got %s", cfg.Environment)
	}
	if got := len(cfg.SessionSecret); got != 32 {
		t.Fatalf("expected secret length 32, got %d", got)
	}
}

func TestNewOverrides(t *testing.T) {
	secret := base64.StdEncoding.EncodeToString(bytesOfLength(40))
	t.Setenv("AUTH_SESSION_SECRET", secret)
	t.Setenv("AUTH_LISTEN_ADDR", "127.0.0.1:9000")
	t.Setenv("AUTH_LOG_MODE", "json")
	t.Setenv("AUTH_ENV", "production")

	cfg, err := New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.ListenAddr != "127.0.0.1:9000" {
		t.Fatalf("expected overridden listen addr, got %s", cfg.ListenAddr)
	}
	if cfg.LogMode != logging.ModeJSON {
		t.Fatalf("expected json mode, got %s", cfg.LogMode)
	}
	if cfg.Environment != "production" {
		t.Fatalf("expected environment production, got %s", cfg.Environment)
	}
	if len(cfg.SessionSecret) != 40 {
		t.Fatalf("expected secret length 40, got %d", len(cfg.SessionSecret))
	}
}

func TestNewMissingSecret(t *testing.T) {
	t.Setenv("AUTH_SESSION_SECRET", "")
	if _, err := New(); err == nil {
		t.Fatalf("expected error for missing secret")
	}
}

func TestNewInvalidSecret(t *testing.T) {
	t.Setenv("AUTH_SESSION_SECRET", "not-base64")
	if _, err := New(); err == nil {
		t.Fatalf("expected error for invalid secret")
	}
}

func TestNewShortSecretAccepted(t *testing.T) {
	t.Setenv("AUTH_SESSION_SECRET", base64.StdEncoding.EncodeToString(bytesOfLength(16)))
	if _, err := New(); err != nil {
		t.Fatalf("expected short secret to pass config load, got %v", err)
	}
}

func bytesOfLength(n int) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = byte(i % 255)
	}
	return b
}
