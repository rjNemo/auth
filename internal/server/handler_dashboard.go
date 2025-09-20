package server

import (
	"log"
	"net/http"
	"time"

	"github.com/rjnemo/auth/internal/service/auth"
)

const dashboardTimeDisplayLayout = "02 Jan 2006 15:04 MST"

func (s *Server) dashboardPageHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		state := sessionFromContext(r.Context())

		if !state.Authenticated {
			w.WriteHeader(http.StatusUnauthorized)
			s.render(w, "unauthorized.html", newUnauthorizedData("Sign in to continue.", state.CSRFToken))
			return
		}

		email, err := auth.NewUserEmail(state.Email)
		if err != nil {
			log.Printf("dashboard: invalid session email: %v", err)
			http.Error(w, "session invalid", http.StatusUnauthorized)
			return
		}

		account, err := s.authService.LookupByEmail(r.Context(), email)
		if err != nil {
			log.Printf("dashboard: lookup failed: %v", err)
			http.Error(w, "unable to load account", http.StatusInternalServerError)
			return
		}

		createdAtISO := account.CreatedAt.Format(time.RFC3339)
		createdAtDisplay := account.CreatedAt.Format(dashboardTimeDisplayLayout)

		s.render(w, "dashboard.html", newDashboardData(state.Email, state.CSRFToken, createdAtDisplay, createdAtISO))
	}
}
