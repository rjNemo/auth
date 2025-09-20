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
	Email        string
	Error        string
	Info         string
	CSRFToken    string
	CreatedAt    string
	CreatedAtISO string
}

func newLoginData(email, errMsg, token string) PageData {
	return PageData{Email: email, Error: errMsg, CSRFToken: token}
}

func newUnauthorizedData(errMsg, token string) PageData {
	return PageData{Error: errMsg, CSRFToken: token}
}

func newDashboardData(email, token, createdAt, createdAtISO string) PageData {
	return PageData{Email: email, CSRFToken: token, CreatedAt: createdAt, CreatedAtISO: createdAtISO}
}

func newSignupData(email, errMsg, token string) PageData {
	return PageData{Email: email, Error: errMsg, CSRFToken: token}
}
