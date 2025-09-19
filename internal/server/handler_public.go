package server

import "net/http"

func (s *Server) indexHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		state := sessionFromContext(r.Context())
		s.render(w, "index.html", newIndexData(state.Email, "", state.CSRFToken))
	}
}
