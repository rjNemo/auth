package server

import (
	"log"
	"net/http"
)

func (s *Server) indexHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := s.templates.ExecuteTemplate(w, "index.html", nil); err != nil {
			log.Printf("render index: %v", err)
			http.Error(w, "template render failed", http.StatusInternalServerError)
		}
	}
}
