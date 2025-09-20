package config

import (
	"cmp"
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/rjnemo/auth/internal/logging"
)

const (
	envListenAddr    = "AUTH_LISTEN_ADDR"
	envLogMode       = "AUTH_LOG_MODE"
	envEnvironment   = "AUTH_ENV"
	envSessionSecret = "AUTH_SESSION_SECRET"

	defaultListenAddr  = ":8000"
	defaultEnvironment = "development"
)

// Config holds application configuration derived from environment variables.
type Config struct {
	ListenAddr    string
	LogMode       logging.Mode
	Environment   string
	SessionSecret []byte
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

	cfg := &Config{
		ListenAddr:    listenAddr,
		LogMode:       logMode,
		Environment:   environment,
		SessionSecret: secret,
	}

	return cfg, nil
}
