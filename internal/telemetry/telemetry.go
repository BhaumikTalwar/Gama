package telemetry

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/BhaumikTalwar/Gama/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
)

var (
	tracerProvider *sdktrace.TracerProvider
	initialized    bool
)

func InitTelemetry(ctx context.Context, cfg *config.TelemetryConfig, logger *slog.Logger) error {
	if !cfg.Enabled {
		logger.Info("Telemetry is disabled")
		return nil
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion("1.0.0"),
			semconv.DeploymentEnvironment(cfg.Environment),
		),
	)
	if err != nil {
		return fmt.Errorf("failed to create resource: %w", err)
	}

	var sampler sdktrace.Sampler
	if cfg.TraceSampleRate >= 1.0 {
		sampler = sdktrace.AlwaysSample()
	} else if cfg.TraceSampleRate <= 0 {
		sampler = sdktrace.NeverSample()
	} else {
		sampler = sdktrace.TraceIDRatioBased(cfg.TraceSampleRate)
	}

	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(cfg.OTLPEndpoint),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	tracerProvider = sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter,
			sdktrace.WithBatchTimeout(time.Second*5),
		),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
	)

	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	logger.Info("Telemetry initialized",
		slog.String("service", cfg.ServiceName),
		slog.String("environment", cfg.Environment),
		slog.String("otlp_endpoint", cfg.OTLPEndpoint),
		slog.Float64("sample_rate", cfg.TraceSampleRate),
	)

	initialized = true
	return nil
}

func Shutdown(ctx context.Context, logger *slog.Logger) error {
	if tracerProvider == nil {
		return nil
	}

	logger.Info("Shutting down telemetry...")
	if err := tracerProvider.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown tracer provider: %w", err)
	}

	logger.Info("Telemetry shutdown complete")
	return nil
}

func IsInitialized() bool {
	return initialized
}

func Tracer(name string) trace.Tracer {
	if tracerProvider == nil {
		return nil
	}
	return tracerProvider.Tracer(name)
}
