package server

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"
)

const (
	sessionCookieName          = "auth_session"
	sessionLifetime            = 12 * time.Hour
	sessionSecretMinLength     = 32
	csrfTokenByteLength    int = 32
)

// SessionStore persists session data using secure HTTP cookies.
type SessionStore struct {
	secret []byte
}

// NewSessionStore creates a cookie-backed session store.
func NewSessionStore(secret []byte) (*SessionStore, error) {
	if len(secret) < sessionSecretMinLength {
		return nil, fmt.Errorf("session secret must be at least %d bytes", sessionSecretMinLength)
	}
	// copy secret to avoid external mutation
	buf := make([]byte, len(secret))
	copy(buf, secret)
	return &SessionStore{secret: buf}, nil
}

// SessionState holds per-request session data after loading.
type SessionState struct {
	Authenticated bool
	Email         string
	CSRFToken     string
}

// Load extracts session data from the request cookies.
func (s *SessionStore) Load(r *http.Request) SessionState {
	c, err := r.Cookie(sessionCookieName)
	if err != nil {
		return SessionState{}
	}

	payload, err := decodeSession(c.Value, s.secret)
	if err != nil {
		return SessionState{}
	}

	return payload
}

// Save persists the session state onto the response cookies.
func (s *SessionStore) Save(w http.ResponseWriter, state SessionState) error {
	serialized, err := encodeSession(state, s.secret)
	if err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    serialized,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // TODO: in production, set to true
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(sessionLifetime),
	})

	return nil
}

// Clear removes the session cookie from the client.
func (s *SessionStore) Clear(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})
}

// ensureCSRFToken returns a session state with a CSRF token present.
func ensureCSRFToken(state SessionState) (SessionState, error) {
	if state.CSRFToken != "" {
		return state, nil
	}
	token := make([]byte, csrfTokenByteLength)
	if _, err := rand.Read(token); err != nil {
		return state, err
	}
	state.CSRFToken = base64.RawURLEncoding.EncodeToString(token)
	return state, nil
}
