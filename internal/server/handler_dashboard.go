package server

import "net/http"

func (s *Server) dashboardHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		state := sessionFromContext(r.Context())

		if !state.Authenticated {
			w.WriteHeader(http.StatusUnauthorized)
			s.render(w, "unauthorized.html", newUnauthorizedData("Sign in to continue."))
			return
		}

		s.render(w, "in.html", PageData{Email: state.Email})
	}
}
