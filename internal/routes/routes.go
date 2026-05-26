package routes

import (
	"log/slog"
	"net/http"

	"github.com/BhaumikTalwar/Gama/config"
	"github.com/BhaumikTalwar/Gama/internal/middleware"
	"github.com/BhaumikTalwar/Gama/internal/repository"
	"github.com/BhaumikTalwar/Gama/internal/service"
	"github.com/BhaumikTalwar/Gama/internal/service/Otp"
	"github.com/BhaumikTalwar/Gama/internal/telemetry"
	"github.com/BhaumikTalwar/Gama/utils"
	"github.com/gin-contrib/gzip"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
)

func SetupRouter(
	env string,
	repos *repository.Repositories,
	logger *slog.Logger,
	s3Store *service.S3Store,
	otpService otp.OtpService,
	healthChecker *service.HealthChecker,
	metrics *telemetry.Metrics,
) *gin.Engine {
	isProd := !utils.IsDevEnv(env)
	if isProd {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	r := gin.New()
	r.Use(middleware.RequestIDMiddleware())
	r.Use(middleware.StructuredLoggerSlog(logger))
	r.Use(middleware.CustomSlogRecovery(logger))
	r.Use(middleware.CORSMiddleware(config.GetAppConfig().CorsAddresses...))
	r.Use(gzip.Gzip(gzip.DefaultCompression, gzip.WithExcludedPaths([]string{"/metrics", "/debug/pprof"})))

	if metrics != nil && config.GetTelemetryConfig().EnableHTTPMetrics {
		r.Use(middleware.MetricsMiddleware(metrics))
	}

	r.GET("/health", func(c *gin.Context) {
		if healthChecker != nil {
			response := healthChecker.Check(c.Request.Context())
			status := http.StatusOK
			if response.Status == "unhealthy" {
				status = http.StatusServiceUnavailable
			}
			c.JSON(status, response)
		} else {
			c.JSON(http.StatusOK, gin.H{
				"status":  "ok",
				"service": "Gama",
			})
		}
	})

	r.GET("/health/live", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "alive",
		})
	})

	r.GET("/health/ready", func(c *gin.Context) {
		if healthChecker != nil {
			health := healthChecker.Readiness()
			if health.Status != "ready" {
				c.JSON(http.StatusServiceUnavailable, health)
				return
			}
			c.JSON(http.StatusOK, health)
		} else {
			c.JSON(http.StatusOK, gin.H{
				"status": "ready",
			})
		}
	})

	if !isProd {
		pprof.Register(r)
	}

	api := r.Group("/api/v1")
	SetupAuthRoutes(api, repos, otpService)
	return r
}
