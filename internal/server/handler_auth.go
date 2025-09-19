package server

import (
	"errors"
	"log"
	"net/http"

	"github.com/rjnemo/auth/internal/auth"
)

func (s *Server) loginHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "invalid form submission", http.StatusBadRequest)
			return
		}

		emailInput := r.FormValue("email")
		password := r.FormValue("password")

		email, err := auth.NewUserEmail(emailInput)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			s.render(w, "index.html", newIndexData("", "Email and password are required."))
			return
		}

		account, err := s.authService.Authenticate(r.Context(), email, password)
		switch {
		case err == nil:
			s.sessions.SetAuthenticated(account.Email.String())
			http.Redirect(w, r, "/in", http.StatusSeeOther)
		case errors.Is(err, auth.ErrInvalidInput):
			w.WriteHeader(http.StatusBadRequest)
			s.render(w, "index.html", newIndexData(email.String(), "Email and password are required."))
		case errors.Is(err, auth.ErrInvalidCredentials):
			s.renderLoginFailure(w, email)
		default:
			log.Printf("auth: authenticate failed: %v", err)
			http.Error(w, "unexpected error", http.StatusInternalServerError)
		}
	}
}

func (s *Server) logoutHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.sessions.Clear()
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func (s *Server) renderLoginFailure(w http.ResponseWriter, email auth.UserEmail) {
	w.WriteHeader(http.StatusUnauthorized)
	s.render(w, "index.html", newIndexData(email.String(), "Invalid credentials."))
}
