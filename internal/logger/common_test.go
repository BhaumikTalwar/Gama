package logger

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"testing"
	"time"

	"github.com/BhaumikTalwar/Gama/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultReplaceAttr_RedactsSensitiveKeys(t *testing.T) {
	replace := DefaultReplaceAttr()

	tests := []struct {
		key      string
		value    string
		expected string
	}{
		{"password", "mysecret", "[REDACTED]"},
		{"secret", "supersecret", "[REDACTED]"},
		{"token", "jwt-token", "[REDACTED]"},
		{"api_key", "key123", "[REDACTED]"},
		{"auth", "basic", "[REDACTED]"},
		{"email", "user@test.com", "user@test.com"},
		{"name", "John", "John"},
		{"user_id", "123", "123"},
	}

	for _, tt := range tests {
		attr := replace(nil, slog.String(tt.key, tt.value))
		assert.Equal(t, tt.key, attr.Key, "key should remain unchanged")
		assert.Equal(t, tt.expected, attr.Value.String())
	}
}

func TestDefaultReplaceAttr_ReplacesTimestamp(t *testing.T) {
	replace := DefaultReplaceAttr()

	attr := replace(nil, slog.Time(slog.TimeKey, mustParseTime("2024-01-15T10:30:00Z")))
	assert.Equal(t, "timestamp", attr.Key)
	assert.Equal(t, "1705314600", attr.Value.String())
}

func TestDefaultReplaceAttr_RemovesEmptyStrings(t *testing.T) {
	replace := DefaultReplaceAttr()

	attr := replace(nil, slog.String("optional", ""))
	assert.Equal(t, slog.Attr{}, attr)
}

func TestDefaultReplaceAttr_CaseInsensitiveRedaction(t *testing.T) {
	replace := DefaultReplaceAttr()

	attr := replace(nil, slog.String("Password", "secret123"))
	assert.Equal(t, "[REDACTED]", attr.Value.String())

	attr = replace(nil, slog.String("TOKEN", "abc"))
	assert.Equal(t, "[REDACTED]", attr.Value.String())
}

func TestNewSlogger_JSONFormat(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		ReplaceAttr: DefaultReplaceAttr(),
	}))
	logger.Info("test message", "key", "value")

	var result map[string]any
	err := json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err)
	assert.Equal(t, "test message", result["msg"])
	assert.Equal(t, "value", result["key"])
}

func TestNewSlogger_TextFormat(t *testing.T) {
	cfg := &config.LoggingConfig{
		Level:  "DEBUG",
		Format: "text",
		Output: "stdout",
	}

	logger := NewSlogger("test-service", cfg)
	assert.NotNil(t, logger)
}

func TestNewSlogger_DefaultLogLevel(t *testing.T) {
	assert.Equal(t, slog.LevelInfo, getLogLevel("invalid"))
	assert.Equal(t, slog.LevelInfo, getLogLevel(""))
}

func TestGetLogLevel(t *testing.T) {
	tests := []struct {
		input string
		want  slog.Level
	}{
		{"DEBUG", slog.LevelDebug},
		{"debug", slog.LevelDebug},
		{"INFO", slog.LevelInfo},
		{"WARN", slog.LevelWarn},
		{"WARNING", slog.LevelWarn},
		{"ERROR", slog.LevelError},
		{"unknown", slog.LevelInfo},
		{"", slog.LevelInfo},
	}

	for _, tt := range tests {
		got := getLogLevel(tt.input)
		assert.Equal(t, tt.want, got, "getLogLevel(%q)", tt.input)
	}
}

func TestSensitiveKeyRedaction_Integration(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		ReplaceAttr: DefaultReplaceAttr(),
	}))

	logger.Info("login",
		"email", "user@test.com",
		"password", "hunter2",
	)

	var result map[string]any
	err := json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err)
	assert.Equal(t, "user@test.com", result["email"])
	assert.Equal(t, "[REDACTED]", result["password"])
}

func TestNewSloggerWithServiceInfo(t *testing.T) {
	cfg := &config.LoggingConfig{
		Level:  "INFO",
		Format: "json",
		Output: "stdout",
	}

	logger := NewSlogger("myapp", cfg)
	assert.NotNil(t, logger)
	h := logger.Handler()
	assert.NotNil(t, h)
}

func mustParseTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t
}
