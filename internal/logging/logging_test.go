package logging

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"
)

func TestParseMode(t *testing.T) {
	t.Parallel()

	cases := map[string]Mode{
		"":        ModeText,
		"text":    ModeText,
		"TEXT":    ModeText,
		"json":    ModeJSON,
		"  json ": ModeJSON,
		"unknown": ModeText,
	}

	for input, want := range cases {
		if got := ParseMode(input); got != want {
			t.Fatalf("ParseMode(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestNewTextLogger(t *testing.T) {
	var buf bytes.Buffer

	opts := &slog.HandlerOptions{ReplaceAttr: dropTime}
	logger := New(&buf, ModeText, opts)
	logger.Info("server start", slog.String("component", "http"))

	output := strings.TrimSpace(buf.String())
	if !strings.Contains(output, "level=INFO") || !strings.Contains(output, "component=http") {
		t.Fatalf("unexpected text output: %s", output)
	}
	if strings.Contains(output, slog.TimeKey) {
		t.Fatalf("expected time attribute to be stripped: %s", output)
	}
}

func TestNewJSONLogger(t *testing.T) {
	var buf bytes.Buffer

	opts := &slog.HandlerOptions{ReplaceAttr: dropTime}
	logger := New(&buf, ModeJSON, opts)
	logger.Error("save failed", slog.String("component", "auth"))

	var payload map[string]any
	if err := json.Unmarshal(buf.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode json log: %v", err)
	}

	if payload["msg"] != "save failed" {
		t.Fatalf("unexpected message: %v", payload["msg"])
	}
	if payload["component"] != "auth" {
		t.Fatalf("unexpected component: %v", payload["component"])
	}
	if payload["level"] != "ERROR" {
		t.Fatalf("unexpected level: %v", payload["level"])
	}
	if _, ok := payload[slog.TimeKey]; ok {
		t.Fatalf("expected time key to be stripped")
	}
}

func dropTime(_ []string, attr slog.Attr) slog.Attr {
	if attr.Key == slog.TimeKey {
		return slog.Attr{}
	}
	return attr
}
