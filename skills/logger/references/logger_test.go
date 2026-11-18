package logger

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseLevel(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want slog.Level
	}{
		{name: "debug", in: "debug", want: slog.LevelDebug},
		{name: "info", in: "INFO", want: slog.LevelInfo},
		{name: "warn", in: "warn", want: slog.LevelWarn},
		{name: "warning", in: "warning", want: slog.LevelWarn},
		{name: "error", in: "error", want: slog.LevelError},
		{name: "unknown", in: "verbose", want: slog.LevelInfo},
		{name: "empty", in: "", want: slog.LevelInfo},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, parseLevel(tt.in))
		})
	}
}

func TestNewLogger_JSONAttrs(t *testing.T) {
	var buf bytes.Buffer
	log := newLogger(Config{
		Level:       "info",
		Format:      "json",
		ServiceName: "api",
		Version:     "1.2.3",
		Environment: "production",
	}, &buf)

	log.Info("started", "port", 8080)

	out := buf.String()
	require.Contains(t, out, `"service":"api"`)
	require.Contains(t, out, `"version":"1.2.3"`)
	require.Contains(t, out, `"env":"production"`)
	require.Contains(t, out, `"msg":"started"`)
	require.Contains(t, out, `"port":8080`)
	require.NotContains(t, out, `"source"`)
}

func TestNewLogger_LevelFilters(t *testing.T) {
	var buf bytes.Buffer
	log := newLogger(Config{Level: "warn"}, &buf)

	log.Info("hidden")
	log.Warn("shown")

	out := buf.String()
	require.NotContains(t, out, "hidden")
	require.Contains(t, out, "shown")
}

func TestNewLogger_DebugAddsSource(t *testing.T) {
	var buf bytes.Buffer
	log := newLogger(Config{Level: "debug"}, &buf)

	log.Debug("trace")

	require.Contains(t, buf.String(), `"source"`)
}

func TestNewLogger_TextFormat(t *testing.T) {
	var buf bytes.Buffer
	log := newLogger(Config{
		Level:       "info",
		Format:      "text",
		ServiceName: "api",
	}, &buf)

	log.Info("started")

	out := buf.String()
	require.Contains(t, out, "level=INFO")
	require.Contains(t, out, "service=api")
}

func TestNewLogger_UnknownValuesFallBack(t *testing.T) {
	var buf bytes.Buffer
	log := newLogger(Config{Level: "verbose", Format: "xml"}, &buf)

	log.Info("started")

	out := buf.String()
	require.Contains(t, out, `"level":"INFO"`)
	require.Contains(t, out, `"msg":"started"`)
}
