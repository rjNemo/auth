package server

import "net/http"

func (s *Server) indexHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		state := sessionFromContext(r.Context())
		if state.Authenticated {
			http.Redirect(w, r, "/in", http.StatusSeeOther)
			return
		}
		s.render(w, "index.html", newIndexData(state.Email, "", state.CSRFToken))
	}
}
