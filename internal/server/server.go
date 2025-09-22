package server

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log/slog"

	"github.com/rjnemo/auth/internal/config"
	"github.com/rjnemo/auth/internal/driver/logging"
	"github.com/rjnemo/auth/internal/service/auth"
	"github.com/rjnemo/auth/web"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
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
	googleOAuth   *oauth2.Config
}

// New constructs a Server with parsed templates and default state using the provided service.
func New(cfg config.Config, authService *auth.Service, logger *slog.Logger) (*Server, error) {
	if authService == nil {
		return nil, fmt.Errorf("auth service must be provided")
	}

	if err := seedUser(context.Background(), authService); err != nil {
		return nil, fmt.Errorf("seed user: %w", err)
	}

	tmpl, err := template.ParseFS(
		web.Templates,
		"templates/auth_base.html",
		"templates/login.html",
		"templates/dashboard.html",
		"templates/signup.html",
		"templates/unauthorized.html",
	)
	if err != nil {
		return nil, fmt.Errorf("parse templates: %w", err)
	}

	sessionStore, err := NewSessionStore(cfg.SessionSecret)
	if err != nil {
		return nil, fmt.Errorf("session store: %w", err)
	}

	if logger == nil {
		logger = logging.New(io.Discard, logging.ModeText, nil)
	}
	logger = logger.With(slog.String("service", "http"))

	var googleOAuthConfig *oauth2.Config
	if cfg.GoogleOAuth.Enabled() {
		googleOAuthConfig = &oauth2.Config{
			ClientID:     cfg.GoogleOAuth.ClientID,
			ClientSecret: cfg.GoogleOAuth.ClientSecret,
			RedirectURL:  cfg.GoogleOAuth.RedirectURL,
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/userinfo.profile",
			},
			Endpoint: google.Endpoint,
		}
	}

	return &Server{
		templates:     tmpl,
		authService:   authService,
		sessions:      sessionStore,
		logger:        logger,
		configuration: cfg,
		googleOAuth:   googleOAuthConfig,
	}, nil
}

func seedUser(ctx context.Context, service *auth.Service) error {
	email := auth.MustUserEmail(seedEmail)
	if _, err := service.Register(ctx, email, seedPassword); err != nil {
		if errors.Is(err, auth.ErrEmailExists) {
			return nil
		}
		return err
	}
	return nil
}
