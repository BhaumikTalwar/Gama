package otp

import (
	"log/slog"
	"testing"

	"github.com/BhaumikTalwar/Gama/config"
	"github.com/stretchr/testify/assert"
)

func TestMockOtpService_Send(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(discardWriter{}, nil))
	svc := NewMockOtpService(logger)

	sessionID, err := svc.Send("123456", "+1234567890")
	assert.NoError(t, err)
	assert.Equal(t, "mock-session-id", sessionID)
}

func TestNewOtpService_WithAPIKey(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(discardWriter{}, nil))
	cfg := &config.OTPConfig{APIKey: "test_key"}
	svc := NewOtpService(cfg, logger)
	assert.NotNil(t, svc)

	sessionID, err := svc.Send("654321", "+1234567890")
	assert.NoError(t, err)
	assert.Equal(t, "mock-session-id", sessionID)
}

type discardWriter struct{}

func (discardWriter) Write(p []byte) (int, error) {
	return len(p), nil
}
