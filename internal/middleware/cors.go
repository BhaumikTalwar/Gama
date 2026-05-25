package middleware

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

const (
	AllowCredentials = true
	MaxAge           = 12 * time.Hour
)

func CORSMiddleware(origins ...string) gin.HandlerFunc {
	config := cors.DefaultConfig()
	config.AllowOrigins = origins
	config.AllowHeaders = []string{
		"Origin",
		"Content-Type",
		"Authorization",
		"X-Request-ID",
		"X-XSRF-Token",
		"X-XSRF-TOKEN",
		"X-CSRF-Token",
		"Accept",
	}
	config.ExposeHeaders = []string{"Content-Length", "X-Request-ID"}
	config.AllowCredentials = AllowCredentials
	config.MaxAge = MaxAge

	return cors.New(config)
}
