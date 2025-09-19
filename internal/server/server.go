package server

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/rjnemo/auth/internal/auth"
	"github.com/rjnemo/auth/web"
)

const (
	seedEmail    = "user@example.com"
	seedPassword = "password123"
)

// Server holds HTTP dependencies for the application.
type Server struct {
	templates *template.Template
	users     auth.UserStore
	sessions  *SessionManager
}

// New constructs a Server with parsed templates and default state.
func New() (*Server, error) {
	tmpl, err := template.ParseFS(
		web.Templates,
		"templates/index.html",
		"templates/in.html",
		"templates/unauthorized.html",
	)
	if err != nil {
		return nil, fmt.Errorf("parse templates: %w", err)
	}

	users := auth.NewMemoryStore()
	if err := seedUser(users); err != nil {
		return nil, fmt.Errorf("seed user: %w", err)
	}

	return &Server{
		templates: tmpl,
		users:     users,
		sessions:  NewSessionManager(),
	}, nil
}

// Router returns the configured HTTP router.
func (s *Server) Router() http.Handler {
	r := chi.NewRouter()
	s.registerRoutes(r)
	return r
}

func seedUser(store auth.UserStore) error {
	salt, hash, err := auth.HashPassword(seedPassword)
	if err != nil {
		return err
	}

	ctx := context.Background()
	return store.Create(ctx, auth.User{
		ID:           "seed-user",
		Email:        seedEmail,
		PasswordSalt: salt,
		PasswordHash: hash,
		CreatedAt:    time.Now().UTC(),
	})
}
