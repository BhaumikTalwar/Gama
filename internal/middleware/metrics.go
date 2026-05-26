package middleware

import (
	"strconv"
	"time"

	"github.com/BhaumikTalwar/Gama/internal/telemetry"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

func MetricsMiddleware(metrics *telemetry.Metrics) gin.HandlerFunc {
	tracer := otel.Tracer("gin-middleware")

	return func(c *gin.Context) {
		if metrics == nil {
			c.Next()
			return
		}

		start := time.Now()
		path := c.FullPath()
		if path == "" {
			path = "unknown"
		}

		metrics.IncInFlight()

		ctx, span := tracer.Start(c.Request.Context(), c.Request.Method+" "+path,
			trace.WithAttributes(
				attribute.String("http.method", c.Request.Method),
				attribute.String("http.route", path),
				attribute.String("http.user_agent", c.Request.UserAgent()),
				attribute.String("http.client_ip", c.ClientIP()),
			),
		)
		c.Request = c.Request.WithContext(ctx)

		c.Next()

		span.SetAttributes(
			attribute.Int("http.status_code", c.Writer.Status()),
			attribute.Int64("http.response_size", int64(c.Writer.Size())),
		)
		span.End()

		duration := time.Since(start)
		metrics.DecInFlight()
		metrics.RecordHTTPRequest(c.Request.Method, path, c.Writer.Status(), duration)

		propagator := otel.GetTextMapPropagator()
		headers := make(map[string]string)
		propagator.Inject(ctx, propagation.MapCarrier(headers))
	}
}

func TracingMiddleware(serviceName string) gin.HandlerFunc {
	tracer := otel.Tracer(serviceName)
	propagator := otel.GetTextMapPropagator()

	return func(c *gin.Context) {
		ctx := propagator.Extract(c.Request.Context(), propagation.HeaderCarrier(c.Request.Header))
		path := c.FullPath()
		if path == "" {
			path = "unknown"
		}

		ctx, span := tracer.Start(ctx, c.Request.Method+" "+path,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				attribute.String("http.method", c.Request.Method),
				attribute.String("http.url", c.Request.URL.String()),
				attribute.String("http.route", path),
				attribute.String("http.host", c.Request.Host),
				attribute.String("http.user_agent", c.Request.UserAgent()),
				attribute.String("http.client_ip", c.ClientIP()),
			),
		)
		defer span.End()

		c.Request = c.Request.WithContext(ctx)
		c.Next()

		span.SetAttributes(
			attribute.Int("http.status_code", c.Writer.Status()),
			attribute.Int64("http.response_size", int64(c.Writer.Size())),
		)

		if len(c.Errors) > 0 {
			span.SetAttributes(attribute.String("error.message", c.Errors.String()))
		}
	}
}

func getPathTemplate(c *gin.Context) string {
	return c.FullPath()
}

func normalizeStatusCode(code int) string {
	switch {
	case code >= 200 && code < 300:
		return "2xx"
	case code >= 300 && code < 400:
		return "3xx"
	case code >= 400 && code < 500:
		return "4xx"
	case code >= 500:
		return "5xx"
	default:
		return "unknown"
	}
}

func getStatusCodeLabel(c *gin.Context) string {
	return strconv.Itoa(c.Writer.Status())
}
