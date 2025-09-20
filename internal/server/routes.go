package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (s *Server) registerRoutes(r chi.Router) {
	r.Get("/", s.indexHandler())
	r.Post("/login", s.loginHandler())
	r.Post("/logout", s.logoutHandler())
	r.Get("/signup", s.signupHandler())
	r.Post("/signup", s.registerHandler())
	r.Get("/in", s.dashboardHandler())
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
