package server

import (
	"log/slog"
	"net/http"
)

func (s *Server) render(w http.ResponseWriter, name string, data any) {
	if err := s.templates.ExecuteTemplate(w, name, data); err != nil {
		s.logger.With(
			slog.String("component", "templates"),
			slog.String("template", name),
		).Error("render failed", slog.Any("error", err))
		http.Error(w, "template render failed", http.StatusInternalServerError)
	}
}

// PageData contains fields shared by the templates for now.
type PageData struct {
	Title              string
	View               string
	Email              string
	Error              string
	Info               string
	CSRFToken          string
	CreatedAt          string
	CreatedAtISO       string
	GoogleLoginURL     string
	GoogleLoginEnabled bool
}

func newLoginData(email, errMsg, token string) PageData {
	return PageData{Title: "Sign in · Auth Demo", View: "login", Email: email, Error: errMsg, CSRFToken: token}
}

func (s *Server) applyOAuthOptions(data PageData) PageData {
	if s.googleOAuth != nil {
		data.GoogleLoginEnabled = true
		data.GoogleLoginURL = "/login/google"
	}
	return data
}

func newUnauthorizedData(errMsg, token string) PageData {
	return PageData{
		Title:     "Access denied · Auth Demo",
		View:      "unauthorized",
		Error:     errMsg,
		CSRFToken: token,
	}
}

func newDashboardData(email, token, createdAt, createdAtISO string) PageData {
	return PageData{
		Title:        "Dashboard · Auth Demo",
		View:         "dashboard",
		Email:        email,
		CSRFToken:    token,
		CreatedAt:    createdAt,
		CreatedAtISO: createdAtISO,
	}
}

func newSignupData(email, errMsg, token string) PageData {
	return PageData{Title: "Create account · Auth Demo", View: "signup", Email: email, Error: errMsg, CSRFToken: token}
}
