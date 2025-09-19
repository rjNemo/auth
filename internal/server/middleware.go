package server

import (
	"context"
	"net/http"
)

func (s *Server) sessionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		state := s.sessions.Snapshot()
		ctx := withSession(r.Context(), state)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type sessionContextKey struct{}

func withSession(ctx context.Context, state SessionState) context.Context {
	return context.WithValue(ctx, sessionContextKey{}, state)
}

func sessionFromContext(ctx context.Context) SessionState {
	if ctx == nil {
		return SessionState{}
	}
	if state, ok := ctx.Value(sessionContextKey{}).(SessionState); ok {
		return state
	}
	return SessionState{}
}
