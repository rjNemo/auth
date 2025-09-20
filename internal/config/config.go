package config

import (
	"cmp"
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/rjnemo/auth/internal/driver/logging"
)

const (
	envListenAddr         = "AUTH_LISTEN_ADDR"
	envLogMode            = "AUTH_LOG_MODE"
	envEnvironment        = "AUTH_ENV"
	envSessionSecret      = "AUTH_SESSION_SECRET"
	envDatabaseURL        = "AUTH_DATABASE_URL"
	envGoogleClientID     = "AUTH_GOOGLE_CLIENT_ID"
	envGoogleClientSecret = "AUTH_GOOGLE_CLIENT_SECRET"
	envGoogleRedirectURL  = "AUTH_GOOGLE_REDIRECT_URL"

	defaultListenAddr  = ":8000"
	defaultEnvironment = "development"
)

// Config holds application configuration derived from environment variables.
type Config struct {
	ListenAddr    string
	LogMode       logging.Mode
	Environment   string
	SessionSecret []byte
	DatabaseURL   string
	GoogleOAuth   GoogleOAuthConfig
}

// GoogleOAuthConfig holds configuration for Google OAuth2 login.
type GoogleOAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

// Enabled reports whether Google OAuth2 is fully configured.
func (g GoogleOAuthConfig) Enabled() bool {
	return g.ClientID != "" && g.ClientSecret != "" && g.RedirectURL != ""
}

// New loads configuration from environment variables, applying defaults and validation.
func New() (*Config, error) {
	listenAddr := cmp.Or(strings.TrimSpace(os.Getenv(envListenAddr)), defaultListenAddr)
	environment := cmp.Or(strings.TrimSpace(os.Getenv(envEnvironment)), defaultEnvironment)
	logMode := logging.ModeText
	if rawMode := strings.TrimSpace(os.Getenv(envLogMode)); rawMode != "" {
		logMode = logging.ParseMode(rawMode)
	}

	secretRaw, ok := os.LookupEnv(envSessionSecret)
	if !ok || strings.TrimSpace(secretRaw) == "" {
		return nil, fmt.Errorf("missing required configuration: set %s to a base64-encoded secret", envSessionSecret)
	}

	secret, err := base64.StdEncoding.DecodeString(secretRaw)
	if err != nil {
		return nil, fmt.Errorf("invalid %s: %w", envSessionSecret, err)
	}

	databaseURL := strings.TrimSpace(os.Getenv(envDatabaseURL))
	if databaseURL == "" {
		return nil, fmt.Errorf("missing required configuration: set %s", envDatabaseURL)
	}

	googleOAuth := GoogleOAuthConfig{
		ClientID:     strings.TrimSpace(os.Getenv(envGoogleClientID)),
		ClientSecret: strings.TrimSpace(os.Getenv(envGoogleClientSecret)),
		RedirectURL:  strings.TrimSpace(os.Getenv(envGoogleRedirectURL)),
	}

	if partiallyConfigured(googleOAuth) {
		return nil, fmt.Errorf("incomplete google oauth configuration: set %s, %s, and %s", envGoogleClientID, envGoogleClientSecret, envGoogleRedirectURL)
	}

	cfg := &Config{
		ListenAddr:    listenAddr,
		LogMode:       logMode,
		Environment:   environment,
		SessionSecret: secret,
		DatabaseURL:   databaseURL,
		GoogleOAuth:   googleOAuth,
	}

	return cfg, nil
}

func partiallyConfigured(cfg GoogleOAuthConfig) bool {
	switch {
	case cfg.ClientID == "" && cfg.ClientSecret == "" && cfg.RedirectURL == "":
		return false
	case cfg.ClientID == "" || cfg.ClientSecret == "" || cfg.RedirectURL == "":
		return true
	default:
		return false
	}
}
