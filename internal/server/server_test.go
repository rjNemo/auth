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
)

func newTestServer(t *testing.T) *Server {
	t.Helper()

	cfg := config.Config{
		ListenAddr:    ":0",
		LogMode:       logging.ModeText,
		Environment:   "test",
		SessionSecret: bytes.Repeat([]byte("s"), 32),
	}

	logger := logging.New(io.Discard, logging.ModeText, nil)

	srv, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("new server: %v", err)
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
