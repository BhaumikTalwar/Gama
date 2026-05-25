package middleware

import (
	"net/http"
	"time"

	"github.com/gin-contrib/timeout"
	"github.com/gin-gonic/gin"
)

func TimeoutMiddleware(t time.Duration) gin.HandlerFunc {
	return timeout.New(
		timeout.WithTimeout(t),
		timeout.WithResponse(timeoutResponse),
	)
}

func timeoutResponse(c *gin.Context) {
	c.JSON(http.StatusGatewayTimeout, gin.H{
		"error":   "request_timeout",
		"code":    http.StatusGatewayTimeout,
		"message": "The server failed to respond within the expected time frame.",
	})
}
