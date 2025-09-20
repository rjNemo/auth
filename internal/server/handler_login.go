package server

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/rjnemo/auth/internal/service/auth"
)

func (s *Server) loginPageHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		state := sessionFromContext(r.Context())
		if state.Authenticated {
			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
			return
		}
		s.render(w, "login.html", newLoginData(state.Email, "", state.CSRFToken))
	}
}

func (s *Server) loginHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := s.logger.With(slog.String("component", "login"))
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
			s.render(w, "login.html", newLoginData("", credentialRequiredMsg, state.CSRFToken))
			return
		}

		account, err := s.authService.Authenticate(r.Context(), email, password)
		switch {
		case err == nil:
			state.Authenticated = true
			state.Email = account.Email.String()
			if err := s.sessions.Save(w, state); err != nil {
				logger.Warn("session save failed", slog.Any("error", err))
			}
			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)

		case errors.Is(err, auth.ErrWeakPassword):
			w.WriteHeader(http.StatusBadRequest)
			s.render(w, "login.html", newLoginData(email.String(), weakPasswordMsg, state.CSRFToken))
		case errors.Is(err, auth.ErrInvalidInput):
			w.WriteHeader(http.StatusBadRequest)
			s.render(w, "login.html", newLoginData(email.String(), credentialRequiredMsg, state.CSRFToken))
		case errors.Is(err, auth.ErrInvalidCredentials):
			s.renderLoginFailure(w, email, state.CSRFToken)
		default:
			logger.Error("authenticate failed", slog.Any("error", err))
			http.Error(w, "unexpected error", http.StatusInternalServerError)
		}
	}
}

func (s *Server) renderLoginFailure(w http.ResponseWriter, email auth.UserEmail, token string) {
	w.WriteHeader(http.StatusUnauthorized)
	s.render(w, "login.html", newLoginData(email.String(), invalidCredentialsMsg, token))
}
