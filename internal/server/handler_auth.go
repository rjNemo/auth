package server

import (
	"log"
	"net/http"
)

func (s *Server) loginHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Login request received")
		s.loggedIn = true
		http.Redirect(w, r, "/in", http.StatusSeeOther)
	}
}
