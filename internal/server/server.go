package server

import (
	"html/template"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/rjnemo/auth/web"
)

// Server holds HTTP dependencies for the application.
type Server struct {
	templates *template.Template
	loggedIn  bool
}

// New constructs a Server with parsed templates and default state.
func New() *Server {
	tmpl := template.Must(template.ParseFS(
		web.Templates,
		"templates/index.html",
		"templates/in.html",
		"templates/unauthorized.html",
	))

	return &Server{
		templates: tmpl,
	}
}

// Router returns the configured HTTP router.
func (s *Server) Router() http.Handler {
	r := chi.NewRouter()
	s.registerRoutes(r)
	return r
}
