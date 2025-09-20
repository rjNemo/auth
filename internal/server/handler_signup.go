package server

import (
	"errors"
	"log"
	"net/http"

	"github.com/rjnemo/auth/internal/service/auth"
)

const (
	credentialRequiredMsg = "Email and password are required."
	invalidCredentialsMsg = "Invalid credentials."
	duplicateEmailMsg     = "An account with that email already exists."
	weakPasswordMsg       = "Password must be at least 8 characters, include an uppercase letter, and contain a number."
)

func (s *Server) signupPageHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		state := sessionFromContext(r.Context())

		if state.Authenticated {
			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
			return
		}

		s.render(w, "signup.html", newSignupData(state.Email, "", state.CSRFToken))
	}
}

func (s *Server) signupHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		state := sessionFromContext(r.Context())

		if err := r.ParseForm(); err != nil {
			http.Error(w, "invalid form submission", http.StatusBadRequest)
			return
		}

		emailValue := r.FormValue("email")
		password := r.FormValue("password")

		email, err := auth.NewUserEmail(emailValue)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			s.render(w, "signup.html", newSignupData("", credentialRequiredMsg, state.CSRFToken))
			return
		}

		account, err := s.authService.Register(r.Context(), email, password)
		switch {
		case err == nil:
			state.Authenticated = true
			state.Email = account.Email.String()
			if err := s.sessions.Save(w, state); err != nil {
				log.Printf("session: save failed: %v", err)
			}
			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		case errors.Is(err, auth.ErrWeakPassword):
			w.WriteHeader(http.StatusBadRequest)
			s.render(w, "signup.html", newSignupData(email.String(), weakPasswordMsg, state.CSRFToken))
		case errors.Is(err, auth.ErrInvalidInput):
			w.WriteHeader(http.StatusBadRequest)
			s.render(w, "signup.html", newSignupData(email.String(), credentialRequiredMsg, state.CSRFToken))
		case errors.Is(err, auth.ErrEmailExists):
			w.WriteHeader(http.StatusConflict)
			s.render(w, "signup.html", newSignupData(email.String(), duplicateEmailMsg, state.CSRFToken))
		default:
			log.Printf("auth: register failed: %v", err)
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
