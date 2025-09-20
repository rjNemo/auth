package logging

import (
	"io"
	"log/slog"
	"strings"
)

// Mode selects the output format for structured logs.
type Mode string

const (
	// ModeText renders human-friendly key/value lines for development.
	ModeText Mode = "text"
	// ModeJSON emits JSON objects suited for production ingestion.
	ModeJSON Mode = "json"
)

// ParseMode canonicalises textual representations of the logging mode.
func ParseMode(value string) Mode {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(ModeJSON):
		return ModeJSON
	default:
		return ModeText
	}
}

// New constructs a slog.Logger with the desired mode and handler options.
func New(out io.Writer, mode Mode, opts *slog.HandlerOptions) *slog.Logger {
	if out == nil {
		out = io.Discard
	}
	if opts == nil {
		opts = &slog.HandlerOptions{}
	}

	var handler slog.Handler
	if mode == ModeJSON {
		handler = slog.NewJSONHandler(out, opts)
	} else {
		handler = slog.NewTextHandler(out, opts)
	}

	return slog.New(handler)
}
