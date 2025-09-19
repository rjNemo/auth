package server

import "net/http"

func (s *Server) dashboardHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !s.sessions.IsAuthenticated() {
			w.WriteHeader(http.StatusUnauthorized)
			s.render(w, "unauthorized.html", newUnauthorizedData("Sign in to continue."))
			return
		}

		s.render(w, "in.html", PageData{Email: s.sessions.CurrentAccount()})
	}
}
