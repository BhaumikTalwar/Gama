package middleware

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CustomSlogRecovery(logger *slog.Logger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		logger.Error("Panic recovered",
			slog.Any("error", recovered),
			slog.String("request_id", c.GetString(RequestIdKey)),
			slog.String("path", c.Request.URL.Path),
			slog.String("method", c.Request.Method),
		)
		c.AbortWithStatus(http.StatusInternalServerError)
	})
}
