//go:build !integration

package telemetry

import (
	"context"
	"log/slog"
	"testing"

	"github.com/BhaumikTalwar/Gama/config"
	"github.com/stretchr/testify/assert"
)

func resetTelemetryState() {
	tracerProvider = nil
	initialized = false
}

func TestInitTelemetry_Disabled(t *testing.T) {
	resetTelemetryState()

	logger := slog.New(slog.NewTextHandler(discardWriter{}, nil))
	cfg := &config.TelemetryConfig{Enabled: false}

	err := InitTelemetry(context.Background(), cfg, logger)
	assert.NoError(t, err)
	assert.False(t, IsInitialized())
}

func TestIsInitialized_Default(t *testing.T) {
	resetTelemetryState()
	assert.False(t, IsInitialized())
}

func TestTracer_NilWhenNotInitialized(t *testing.T) {
	resetTelemetryState()
	tr := Tracer("test")
	assert.Nil(t, tr)
}

func TestShutdown_NilProvider(t *testing.T) {
	resetTelemetryState()
	logger := slog.New(slog.NewTextHandler(discardWriter{}, nil))
	err := Shutdown(context.Background(), logger)
	assert.NoError(t, err)
}

type discardWriter struct{}

func (discardWriter) Write(p []byte) (int, error) {
	return len(p), nil
}
