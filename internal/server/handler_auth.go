package server

import (
	"errors"
	"log"
	"net/http"

	"github.com/rjnemo/auth/internal/auth"
)

func (s *Server) loginHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		state := sessionFromContext(r.Context())

		if err := r.ParseForm(); err != nil {
			http.Error(w, "invalid form submission", http.StatusBadRequest)
			return
		}

		emailInput := r.FormValue("email")
		password := r.FormValue("password")

		email, err := auth.NewUserEmail(emailInput)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			s.render(w, "index.html", newIndexData("", "Email and password are required.", state.CSRFToken))
			return
		}

		account, err := s.authService.Authenticate(r.Context(), email, password)
		switch {
		case err == nil:
			state.Authenticated = true
			state.Email = account.Email.String()
			if err := s.sessions.Save(w, state); err != nil {
				log.Printf("session: save failed: %v", err)
			}
			http.Redirect(w, r, "/in", http.StatusSeeOther)

		case errors.Is(err, auth.ErrInvalidInput):
			w.WriteHeader(http.StatusBadRequest)
			s.render(w, "index.html", newIndexData(email.String(), "Email and password are required.", state.CSRFToken))
		case errors.Is(err, auth.ErrInvalidCredentials):
			s.renderLoginFailure(w, email, state.CSRFToken)
		default:
			log.Printf("auth: authenticate failed: %v", err)
			http.Error(w, "unexpected error", http.StatusInternalServerError)
		}
	}
}

func (s *Server) logoutHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.sessions.Clear(w)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func (s *Server) renderLoginFailure(w http.ResponseWriter, email auth.UserEmail, token string) {
	w.WriteHeader(http.StatusUnauthorized)
	s.render(w, "index.html", newIndexData(email.String(), "Invalid credentials.", token))
}
