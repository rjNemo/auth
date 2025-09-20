package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (s *Server) registerRoutes(r chi.Router) {
	r.Get("/", s.loginPageHandler())
	r.Post("/login", s.loginHandler())
	r.Post("/logout", s.logoutHandler())
	r.Get("/signup", s.signupPageHandler())
	r.Post("/signup", s.signupHandler())
	r.Get("/dashboard", s.dashboardPageHandler())
}

// Router returns the configured HTTP router.
func (s *Server) Router() http.Handler {
	r := chi.NewRouter()

	r.Use(
		middleware.RequestID,
		middleware.RealIP,
		middleware.Logger,
		middleware.Recoverer,
		s.sessionMiddleware,
		s.csrfMiddleware,
	)

	s.registerRoutes(r)

	return r
}
