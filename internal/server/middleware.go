package server

import (
	"context"
	"crypto/subtle"
	"log/slog"
	"net/http"
)

type sessionContextKey struct{}

func (s *Server) sessionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := s.logger.With(slog.String("component", "session"))

		state := s.sessions.Load(r)
		updated, err := ensureCSRFToken(state)
		if err != nil {
			logger.Error("csrf token generation failed", slog.Any("error", err))
			http.Error(w, "session error", http.StatusInternalServerError)
			return
		}
		state = updated

		if err := s.sessions.Save(w, state); err != nil {
			logger.Warn("session save failed", slog.Any("error", err))
		}

		ctx := withSession(r.Context(), state)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *Server) csrfMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace:
			next.ServeHTTP(w, r)
			return
		}

		state := sessionFromContext(r.Context())
		if state.CSRFToken == "" {
			http.Error(w, "missing csrf token", http.StatusForbidden)
			return
		}

		token := r.Header.Get("X-CSRF-Token")
		if token == "" {
			if err := r.ParseForm(); err == nil {
				token = r.Form.Get("_csrf")
			}
		}

		if !validCSRFToken(token, state.CSRFToken) {
			http.Error(w, "invalid csrf token", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

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

func validCSRFToken(provided, expected string) bool {
	if provided == "" || expected == "" {
		return false
	}
	if len(provided) != len(expected) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(provided), []byte(expected)) == 1
}
