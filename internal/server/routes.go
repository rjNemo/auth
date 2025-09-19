package server

import "github.com/go-chi/chi/v5"

func (s *Server) registerRoutes(r chi.Router) {
	r.Get("/", s.indexHandler())
	r.Get("/in", s.dashboardHandler())
	r.Post("/login", s.loginHandler())
}
