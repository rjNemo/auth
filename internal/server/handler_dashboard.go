package server

import (
	"log"
	"net/http"
)

func (s *Server) dashboardHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !s.loggedIn {
			w.WriteHeader(http.StatusUnauthorized)
			if err := s.templates.ExecuteTemplate(w, "unauthorized.html", nil); err != nil {
				log.Printf("render unauthorized: %v", err)
				http.Error(w, "template render failed", http.StatusInternalServerError)
			}
			return
		}

		if err := s.templates.ExecuteTemplate(w, "in.html", nil); err != nil {
			log.Printf("render dashboard: %v", err)
			http.Error(w, "template render failed", http.StatusInternalServerError)
		}
	}
}
