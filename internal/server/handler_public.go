package server

import "net/http"

func (s *Server) indexHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.render(w, "index.html", newIndexData("", ""))
	}
}
