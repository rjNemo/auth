package server

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/rjnemo/auth/internal/config"
	"github.com/rjnemo/auth/internal/driver/logging"
	"github.com/rjnemo/auth/internal/service/auth"
)

func newTestServer(t *testing.T) *Server {
	t.Helper()

	cfg := config.Config{
		ListenAddr:    ":0",
		LogMode:       logging.ModeText,
		Environment:   "test",
		SessionSecret: bytes.Repeat([]byte("s"), 32),
		DatabaseURL:   "postgres://localhost/auth_test?sslmode=disable",
	}

	logger := logging.New(io.Discard, logging.ModeText, nil)

	store := auth.NewMemoryStore()
	service := auth.NewService(store)
	srv, err := New(cfg, service, logger)
	if err != nil {
		t.Fatalf("new server: %v", err)
	}
	return srv
}

func newGoogleTestServer(t *testing.T) *Server {
	t.Helper()

	cfg := config.Config{
		ListenAddr:    ":0",
		LogMode:       logging.ModeText,
		Environment:   "test",
		SessionSecret: bytes.Repeat([]byte("g"), 32),
		GoogleOAuth: config.GoogleOAuthConfig{
			ClientID:     "client",
			ClientSecret: "secret",
			RedirectURL:  "http://localhost/login/google/callback",
		},
		DatabaseURL: "postgres://localhost/auth_test?sslmode=disable",
	}

	logger := logging.New(io.Discard, logging.ModeText, nil)

	store := auth.NewMemoryStore()
	service := auth.NewService(store)
	srv, err := New(cfg, service, logger)
	if err != nil {
		t.Fatalf("new google server: %v", err)
	}
	return srv
}

func attachSession(req *http.Request, state SessionState) *http.Request {
	return req.WithContext(withSession(req.Context(), state))
}

func TestLoginPageHandler(t *testing.T) {
	t.Parallel()

	srv := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = attachSession(req, SessionState{CSRFToken: "token"})
	rr := httptest.NewRecorder()

	srv.loginPageHandler()(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if body := rr.Body.String(); !strings.Contains(body, "Welcome back to Nucleus") {
		t.Fatalf("expected login copy in response, got %q", body)
	}
}

func TestLoginPageHandlerIncludesGoogleLinkWhenConfigured(t *testing.T) {
	t.Parallel()

	srv := newGoogleTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = attachSession(req, SessionState{CSRFToken: "token"})
	rr := httptest.NewRecorder()

	srv.loginPageHandler()(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "id=\"google_login_form\"") {
		t.Fatalf("expected google login form in page, got %q", body)
	}
}

func TestLoginHandlerSuccess(t *testing.T) {
	t.Parallel()

	srv := newTestServer(t)

	form := url.Values{}
	form.Set("email", "user@example.com")
	form.Set("password", "Password123")
	form.Set("_csrf", "csrf-token")

	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = attachSession(req, SessionState{CSRFToken: "csrf-token"})

	rr := httptest.NewRecorder()
	srv.loginHandler()(rr, req)

	res := rr.Result()
	if res.StatusCode != http.StatusSeeOther {
		t.Fatalf("expected 303, got %d", res.StatusCode)
	}
	if loc := res.Header.Get("Location"); loc != "/dashboard" {
		t.Fatalf("expected redirect to /dashboard, got %q", loc)
	}
	foundSession := false
	for _, c := range res.Cookies() {
		if c.Name == sessionCookieName && c.Value != "" {
			foundSession = true
		}
	}
	if !foundSession {
		t.Fatal("expected session cookie to be set")
	}
}

func TestLoginHandlerInvalidCredentials(t *testing.T) {
	t.Parallel()

	srv := newTestServer(t)

	form := url.Values{}
	form.Set("email", "user@example.com")
	form.Set("password", "Password999")
	form.Set("_csrf", "csrf-token")

	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = attachSession(req, SessionState{CSRFToken: "csrf-token"})

	rr := httptest.NewRecorder()
	srv.loginHandler()(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
	if body := rr.Body.String(); !strings.Contains(body, "Unable to sign in") {
		t.Fatalf("expected failure message, got %q", body)
	}
}

func TestGoogleLoginHandlerDisabled(t *testing.T) {
	t.Parallel()

	srv := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/login/google", nil)
	rr := httptest.NewRecorder()

	srv.googleLoginHandler()(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404 when google oauth disabled, got %d", rr.Code)
	}
}

func TestGoogleLoginHandlerRedirects(t *testing.T) {
	t.Parallel()

	srv := newGoogleTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/login/google", nil)
	req = attachSession(req, SessionState{CSRFToken: "csrf"})
	rr := httptest.NewRecorder()

	srv.googleLoginHandler()(rr, req)

	res := rr.Result()
	if res.StatusCode != http.StatusFound {
		t.Fatalf("expected 302 redirect, got %d", res.StatusCode)
	}
	location := res.Header.Get("Location")
	if location == "" {
		t.Fatal("expected redirect location header")
	}
	if !strings.Contains(location, "accounts.google.com") {
		t.Fatalf("expected google authorization URL, got %q", location)
	}

	parsed, err := url.Parse(location)
	if err != nil {
		t.Fatalf("parse redirect url: %v", err)
	}
	stateParam := parsed.Query().Get("state")
	if stateParam == "" {
		t.Fatal("expected state parameter in redirect")
	}

	var sessionCookie *http.Cookie
	for _, c := range res.Cookies() {
		if c.Name == sessionCookieName {
			sessionCookie = c
			break
		}
	}
	if sessionCookie == nil {
		t.Fatal("expected session cookie to be set")
	}

	savedState, err := decodeSession(sessionCookie.Value, srv.configuration.SessionSecret)
	if err != nil {
		t.Fatalf("decode session: %v", err)
	}
	if savedState.OAuthState == "" {
		t.Fatal("expected oauth state stored in session")
	}
	if savedState.OAuthState != stateParam {
		t.Fatalf("expected oauth state %q to match redirect param %q", savedState.OAuthState, stateParam)
	}
}

func TestGoogleCallbackHandlerStateMismatch(t *testing.T) {
	t.Parallel()

	srv := newGoogleTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/login/google/callback?state=other&code=ignored", nil)
	req = attachSession(req, SessionState{OAuthState: "expected", CSRFToken: "csrf"})
	rr := httptest.NewRecorder()

	srv.googleCallbackHandler()(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for state mismatch, got %d", rr.Code)
	}

	res := rr.Result()
	var sessionCookie *http.Cookie
	for _, c := range res.Cookies() {
		if c.Name == sessionCookieName {
			sessionCookie = c
			break
		}
	}
	if sessionCookie == nil {
		t.Fatal("expected session cookie to be set")
	}

	savedState, err := decodeSession(sessionCookie.Value, srv.configuration.SessionSecret)
	if err != nil {
		t.Fatalf("decode session: %v", err)
	}
	if savedState.OAuthState != "" {
		t.Fatal("expected oauth state to be cleared after mismatch")
	}
}

func TestSignupHandlerSuccess(t *testing.T) {
	t.Parallel()

	srv := newTestServer(t)

	form := url.Values{}
	form.Set("email", "new-user@example.com")
	form.Set("password", "Password123")
	form.Set("_csrf", "csrf-token")

	req := httptest.NewRequest(http.MethodPost, "/signup", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = attachSession(req, SessionState{CSRFToken: "csrf-token"})

	rr := httptest.NewRecorder()
	srv.signupHandler()(rr, req)

	res := rr.Result()
	if res.StatusCode != http.StatusSeeOther {
		t.Fatalf("expected 303, got %d", res.StatusCode)
	}
	if loc := res.Header.Get("Location"); loc != "/dashboard" {
		t.Fatalf("expected redirect to /dashboard, got %q", loc)
	}
}

func TestSignupHandlerDuplicate(t *testing.T) {
	t.Parallel()

	srv := newTestServer(t)

	form := url.Values{}
	form.Set("email", "user@example.com")
	form.Set("password", "Password123")
	form.Set("_csrf", "csrf-token")

	req := httptest.NewRequest(http.MethodPost, "/signup", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = attachSession(req, SessionState{CSRFToken: "csrf-token"})

	rr := httptest.NewRecorder()
	srv.signupHandler()(rr, req)

	if rr.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", rr.Code)
	}
	if body := rr.Body.String(); !strings.Contains(body, "account with that email") {
		t.Fatalf("expected duplicate email message, got %q", body)
	}
}

func TestDashboardPageHandler(t *testing.T) {
	t.Parallel()

	srv := newTestServer(t)

	t.Run("unauthenticated", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
		req = attachSession(req, SessionState{CSRFToken: "csrf"})
		rr := httptest.NewRecorder()
		srv.dashboardPageHandler()(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", rr.Code)
		}
		if body := rr.Body.String(); !strings.Contains(body, "Access denied") {
			t.Fatalf("expected unauthorized template, got %q", body)
		}
	})

	t.Run("authenticated", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
		req = attachSession(req, SessionState{Authenticated: true, Email: "user@example.com", CSRFToken: "csrf"})
		rr := httptest.NewRecorder()
		srv.dashboardPageHandler()(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rr.Code)
		}
		body := rr.Body.String()
		if !strings.Contains(body, "Member since") {
			t.Fatalf("expected membership text, got %q", body)
		}
	})
}
