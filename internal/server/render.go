package server

import (
	"log"
	"net/http"
)

func (s *Server) render(w http.ResponseWriter, name string, data any) {
	if err := s.templates.ExecuteTemplate(w, name, data); err != nil {
		log.Printf("render %s: %v", name, err)
		http.Error(w, "template render failed", http.StatusInternalServerError)
	}
}
