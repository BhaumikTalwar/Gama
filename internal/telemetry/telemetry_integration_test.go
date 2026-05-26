//go:build integration

package telemetry

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/BhaumikTalwar/Gama/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
)

func discardOutput() *slog.Logger {
	return slog.New(slog.NewTextHandler(discardWriter{}, nil))
}

type discardWriter struct{}

func (discardWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

func resetTelemetryGlobals() {
	tracerProvider = nil
	initialized = false
}

func TestInitTelemetry_Enabled(t *testing.T) {
	resetTelemetryGlobals()

	cfg := &config.TelemetryConfig{
		Enabled:         true,
		ServiceName:     "test-service",
		Environment:     "test",
		OTLPEndpoint:    "localhost:4317",
		TraceSampleRate: 1.0,
	}

	err := InitTelemetry(context.Background(), cfg, discardOutput())
	require.NoError(t, err)
	assert.True(t, IsInitialized())

	tr := Tracer("test-tracer")
	require.NotNil(t, tr)

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = Shutdown(shutdownCtx, discardOutput())
	assert.NoError(t, err)
}

func TestInitTelemetry_TraceSampleRateRatio(t *testing.T) {
	resetTelemetryGlobals()

	cfg := &config.TelemetryConfig{
		Enabled:         true,
		ServiceName:     "test-service",
		Environment:     "test",
		OTLPEndpoint:    "localhost:4317",
		TraceSampleRate: 0.5,
	}

	err := InitTelemetry(context.Background(), cfg, discardOutput())
	require.NoError(t, err)

	tr := Tracer("test-tracer")
	require.NotNil(t, tr)

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	Shutdown(shutdownCtx, discardOutput())
}

func TestInitTelemetry_NeverSample(t *testing.T) {
	resetTelemetryGlobals()

	cfg := &config.TelemetryConfig{
		Enabled:         true,
		ServiceName:     "test-service",
		Environment:     "test",
		OTLPEndpoint:    "localhost:4317",
		TraceSampleRate: 0.0,
	}

	err := InitTelemetry(context.Background(), cfg, discardOutput())
	require.NoError(t, err)

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	Shutdown(shutdownCtx, discardOutput())
}

func TestInitTelemetry_BadEndpoint(t *testing.T) {
	resetTelemetryGlobals()

	cfg := &config.TelemetryConfig{
		Enabled:         true,
		ServiceName:     "test-service",
		Environment:     "test",
		OTLPEndpoint:    "localhost:1",
		TraceSampleRate: 1.0,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := InitTelemetry(ctx, cfg, discardOutput())
	if err != nil {
		assert.False(t, IsInitialized())
	}
}

func TestInitTelemetry_CreateAndEndSpan(t *testing.T) {
	resetTelemetryGlobals()

	cfg := &config.TelemetryConfig{
		Enabled:         true,
		ServiceName:     "test-service",
		Environment:     "test",
		OTLPEndpoint:    "localhost:4317",
		TraceSampleRate: 1.0,
	}

	err := InitTelemetry(context.Background(), cfg, discardOutput())
	require.NoError(t, err)

	tr := Tracer("test-tracer")
	require.NotNil(t, tr)

	_, span := tr.Start(context.Background(), "test-op")
	span.SetAttributes(attribute.String("key", "value"))
	span.End()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	Shutdown(shutdownCtx, discardOutput())
}
