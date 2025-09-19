package server

import (
	"net/http"

	"github.com/rjnemo/auth/internal/auth"
	"github.com/rjnemo/auth/internal/identity"
)

func (s *Server) loginHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "invalid form submission", http.StatusBadRequest)
			return
		}

		email := identity.NormalizeEmail(r.FormValue("email"))
		password := r.FormValue("password")

		if email == "" || password == "" {
			w.WriteHeader(http.StatusBadRequest)
			s.render(w, "index.html", newIndexData(email, "Email and password are required."))
			return
		}

		account, err := s.users.FindByEmail(r.Context(), email)
		if err != nil {
			s.renderLoginFailure(w, email)
			return
		}

		if !auth.VerifyPassword(password, account.PasswordSalt, account.PasswordHash) {
			s.renderLoginFailure(w, email)
			return
		}

		s.sessions.SetAuthenticated(account.Email)
		http.Redirect(w, r, "/in", http.StatusSeeOther)
	}
}

func (s *Server) logoutHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.sessions.Clear()
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func (s *Server) renderLoginFailure(w http.ResponseWriter, email string) {
	w.WriteHeader(http.StatusUnauthorized)
	s.render(w, "index.html", newIndexData(email, "Invalid credentials."))
}
