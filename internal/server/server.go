package server

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"time"

	"github.com/rjnemo/auth/internal/config"
	"github.com/rjnemo/auth/internal/logging"
	"github.com/rjnemo/auth/internal/service/auth"
	"github.com/rjnemo/auth/web"
)

const (
	seedEmail    = "user@example.com"
	seedPassword = "Password123"
)

// Server holds HTTP dependencies for the application.
type Server struct {
	templates     *template.Template
	authService   *auth.Service
	sessions      *SessionStore
	logger        *slog.Logger
	configuration config.Config
}

// New constructs a Server with parsed templates and default state.
func New(cfg config.Config, logger *slog.Logger) (*Server, error) {
	tmpl, err := template.ParseFS(
		web.Templates,
		"templates/login.html",
		"templates/dashboard.html",
		"templates/signup.html",
		"templates/unauthorized.html",
	)
	if err != nil {
		return nil, fmt.Errorf("parse templates: %w", err)
	}

	store := auth.NewMemoryStore()
	if err := seedUser(store); err != nil {
		return nil, fmt.Errorf("seed user: %w", err)
	}

	sessionStore, err := NewSessionStore(cfg.SessionSecret)
	if err != nil {
		return nil, fmt.Errorf("session store: %w", err)
	}

	if logger == nil {
		logger = logging.New(io.Discard, logging.ModeText, nil)
	}
	logger = logger.With(slog.String("service", "http"))

	return &Server{
		templates:     tmpl,
		authService:   auth.NewService(store),
		sessions:      sessionStore,
		logger:        logger,
		configuration: cfg,
	}, nil
}

func seedUser(store auth.UserStore) error {
	salt, hash, err := auth.HashPassword(seedPassword)
	if err != nil {
		return err
	}

	email := auth.MustUserEmail(seedEmail)

	ctx := context.Background()
	return store.Create(ctx, auth.User{
		ID:           "seed-user",
		Email:        email,
		PasswordSalt: salt,
		PasswordHash: hash,
		CreatedAt:    time.Now().UTC(),
	})
}
