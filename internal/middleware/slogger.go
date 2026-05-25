package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

func StructuredLoggerSlog(logger *slog.Logger) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		if raw != "" {
			path = path + "?" + raw
		}

		logger.Info("HTTP Request",
			slog.String("method", c.Request.Method),
			slog.String("path", path),
			slog.Int("status", c.Writer.Status()),
			slog.Duration("latency", latency),
			slog.String("client_ip", c.ClientIP()),
			slog.Int("body_size", c.Writer.Size()),
			slog.String("user_agent", c.Request.UserAgent()),
			slog.String("request_id", c.GetString(RequestIdKey)),
		)

		if len(c.Errors) > 0 {
			logger.Error("Request errors",
				slog.String("request_id", c.GetString(RequestIdKey)),
				slog.String("errors", c.Errors.String()),
			)
		}
	})
}
