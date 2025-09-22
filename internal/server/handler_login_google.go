package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"golang.org/x/oauth2"

	"github.com/rjnemo/auth/internal/service/auth"
)

const (
	googleUserInfoEndpoint = "https://www.googleapis.com/oauth2/v3/userinfo"
	googleAuthFailedMsg    = "Unable to sign in with Google. Please try again."
	googleAuthCanceledMsg  = "Google sign-in was cancelled."
)

type googleUserInfo struct {
	ID            string `json:"sub"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"email_verified"`
}

func (s *Server) googleLoginHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := s.logger.With(slog.String("component", "google_oauth"))
		if s.googleOAuth == nil {
			http.NotFound(w, r)
			return
		}

		state := sessionFromContext(r.Context())
		if state.Authenticated {
			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
			return
		}

		token, err := generateOAuthState()
		if err != nil {
			logger.Error("generate oauth state failed", slog.Any("error", err))
			http.Error(w, "unexpected error", http.StatusInternalServerError)
			return
		}

		state.OAuthState = token
		if err := s.sessions.Save(w, state); err != nil {
			logger.Error("persist oauth state failed", slog.Any("error", err))
			http.Error(w, "unexpected error", http.StatusInternalServerError)
			return
		}

		redirectURL := s.googleOAuth.AuthCodeURL(token, oauth2.AccessTypeOnline)
		http.Redirect(w, r, redirectURL, http.StatusFound)
	}
}

func (s *Server) googleCallbackHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := s.logger.With(slog.String("component", "google_oauth"))
		if s.googleOAuth == nil {
			http.NotFound(w, r)
			return
		}

		state := sessionFromContext(r.Context())
		expectedState := state.OAuthState
		providedState := r.URL.Query().Get("state")
		state.OAuthState = ""

		saveState := func() bool {
			if err := s.sessions.Save(w, state); err != nil {
				logger.Error("session save failed", slog.Any("error", err))
				http.Error(w, "unexpected error", http.StatusInternalServerError)
				return false
			}
			return true
		}

		respondWithLogin := func(status int, message string) {
			if status != 0 {
				w.WriteHeader(status)
			}
			s.render(w, "login.html", s.applyOAuthOptions(newLoginData(state.Email, message, state.CSRFToken)))
		}

		if expectedState == "" || providedState == "" || providedState != expectedState {
			if !saveState() {
				return
			}
			http.Error(w, "invalid oauth state", http.StatusBadRequest)
			return
		}

		if errParam := r.URL.Query().Get("error"); errParam != "" {
			logger.Info("google oauth returned error", slog.String("google_error", errParam))
			if !saveState() {
				return
			}
			respondWithLogin(http.StatusBadRequest, googleAuthCanceledMsg)
			return
		}

		authCode := r.URL.Query().Get("code")
		if authCode == "" {
			if !saveState() {
				return
			}
			http.Error(w, "missing authorization code", http.StatusBadRequest)
			return
		}

		token, err := s.googleOAuth.Exchange(r.Context(), authCode)
		if err != nil {
			logger.Error("oauth code exchange failed", slog.Any("error", err))
			if !saveState() {
				return
			}
			respondWithLogin(http.StatusUnauthorized, googleAuthFailedMsg)
			return
		}

		info, err := s.fetchGoogleUserInfo(r.Context(), token)
		if err != nil {
			logger.Error("fetch google user info failed", slog.Any("error", err))
			if !saveState() {
				return
			}
			respondWithLogin(http.StatusUnauthorized, googleAuthFailedMsg)
			return
		}

		if !info.VerifiedEmail || info.Email == "" {
			logger.Warn("google returned unverified email", slog.Bool("verified", info.VerifiedEmail))
			if !saveState() {
				return
			}
			respondWithLogin(http.StatusUnauthorized, googleAuthFailedMsg)
			return
		}

		email, err := auth.NewUserEmail(info.Email)
		if err != nil {
			logger.Error("normalize google email failed", slog.Any("error", err))
			if !saveState() {
				return
			}
			respondWithLogin(http.StatusUnauthorized, googleAuthFailedMsg)
			return
		}

		if strings.TrimSpace(info.ID) == "" {
			logger.Warn("google returned empty subject")
			if !saveState() {
				return
			}
			respondWithLogin(http.StatusUnauthorized, googleAuthFailedMsg)
			return
		}

		account, err := s.authService.EnsureExternalUser(r.Context(), email, auth.ProviderGoogle, info.ID, info.VerifiedEmail)
		if err != nil {
			logger.Error("ensure external user failed", slog.Any("error", err))
			if !saveState() {
				return
			}
			http.Error(w, "unexpected error", http.StatusInternalServerError)
			return
		}

		state.Authenticated = true
		state.Email = account.Email.String()
		if err := s.sessions.Save(w, state); err != nil {
			logger.Error("session save failed", slog.Any("error", err))
			http.Error(w, "unexpected error", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	}
}

func (s *Server) fetchGoogleUserInfo(ctx context.Context, token *oauth2.Token) (googleUserInfo, error) {
	client := s.googleOAuth.Client(ctx, token)
	resp, err := client.Get(googleUserInfoEndpoint)
	if err != nil {
		return googleUserInfo{}, fmt.Errorf("request google userinfo: %w", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			s.logger.With(slog.String("component", "google_oauth")).Warn("close google userinfo body failed", slog.Any("error", cerr))
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return googleUserInfo{}, fmt.Errorf("google userinfo response %d: %s", resp.StatusCode, string(body))
	}

	var info googleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return googleUserInfo{}, fmt.Errorf("decode google userinfo: %w", err)
	}

	return info, nil
}
