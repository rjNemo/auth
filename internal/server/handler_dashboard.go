package server

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/rjnemo/auth/internal/service/auth"
)

const dashboardTimeDisplayLayout = "02 Jan 2006 15:04 MST"

func (s *Server) dashboardPageHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := s.logger.With(slog.String("component", "dashboard"))
		state := sessionFromContext(r.Context())

		if !state.Authenticated {
			w.WriteHeader(http.StatusUnauthorized)
			s.render(w, "unauthorized.html", newUnauthorizedData("Sign in to continue.", state.CSRFToken))
			return
		}

		email, err := auth.NewUserEmail(state.Email)
		if err != nil {
			logger.Warn("invalid session email", slog.Any("error", err))
			http.Error(w, "session invalid", http.StatusUnauthorized)
			return
		}

		account, err := s.authService.LookupByEmail(r.Context(), email)
		if err != nil {
			logger.Error("lookup failed", slog.Any("error", err))
			http.Error(w, "unable to load account", http.StatusInternalServerError)
			return
		}

		createdAtISO := account.CreatedAt.Format(time.RFC3339)
		createdAtDisplay := account.CreatedAt.Format(dashboardTimeDisplayLayout)

		s.render(w, "dashboard.html", newDashboardData(state.Email, state.CSRFToken, createdAtDisplay, createdAtISO))
	}
}
