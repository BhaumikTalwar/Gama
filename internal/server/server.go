package server

import (
	"context"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/BhaumikTalwar/Gama/config"
	"github.com/BhaumikTalwar/Gama/internal/caching"
	"github.com/BhaumikTalwar/Gama/internal/logger"
	"github.com/BhaumikTalwar/Gama/internal/postgres"
	"github.com/BhaumikTalwar/Gama/internal/redisClient"
	"github.com/BhaumikTalwar/Gama/internal/repository"
	"github.com/BhaumikTalwar/Gama/internal/routes"
	"github.com/BhaumikTalwar/Gama/internal/service"
	"github.com/BhaumikTalwar/Gama/internal/service/Otp"
	"github.com/BhaumikTalwar/Gama/internal/telemetry"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func RunServer() {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)

	defer stop()
	appConfig := config.GetAppConfig()
	slogger := logger.NewSlogger(appConfig.AppName, config.GetLogConfig())
	slog.SetDefault(slogger)

	telemetryConfig := config.GetTelemetryConfig()
	metrics := telemetry.NewMetrics(appConfig.AppName)

	if err := telemetry.InitTelemetry(ctx, telemetryConfig, slogger); err != nil {
		slog.Error("Failed to initialize telemetry", slog.Any("error", err))
		return
	}
	defer func() {
		telemetryShutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := telemetry.Shutdown(telemetryShutdownCtx, slogger); err != nil {
			slog.Error("Failed to shutdown telemetry", slog.Any("error", err))
		}
	}()

	pgxPool, err := postgres.NewPool(ctx, appConfig.AppName, config.GetPostgresConfig())
	if err != nil {
		slog.Error("Cant Get a Postgres Connection Check Your Config", slog.Any("Error", err))
		return
	}
	defer func() {
		pgxPool.Close()
		slogger.Info("Postgres connection pool closed")
	}()

	redisCl, err := redisClient.NewRedisClient(ctx, config.GetRedisConfig())
	if err != nil {
		slog.Error("Cant Get a Redis Connection Check Your Config", slog.Any("Error", err))
		return
	}
	defer func() {
		if err := redisCl.Close(); err != nil {
			slogger.Error("Error closing Redis client", slog.Any("error", err))
		} else {
			slogger.Info("Redis connection closed")
		}
	}()

	cacheInstance, err := caching.NewJetCacheInstance(redisCl, config.GetCacheConfig(), pgx.ErrNoRows)
	if err != nil {
		slog.Error("Cant Get a Jet Cache Go Instance", slog.Any("Error", err))
		return
	}

	cacheService := caching.NewCacheService(cacheInstance, redisCl)
	versionMgr := caching.NewVersionManager(redisCl)

	repos := repository.SetupPostgresRepositories(pgxPool, cacheService, versionMgr, appConfig.AppName)

	s3Config := config.GetS3Config()
	s3Store, err := service.NewS3Store(s3Config.Endpoint, s3Config.Region, s3Config.AccessKey, s3Config.SecretKey, s3Config.PublicBucket, s3Config.PrivateBucket, s3Config.PublicBaseURL)
	if err != nil {
		slog.Error("Cant Initialize S3 Store", slog.Any("Error", err))
		return
	}

	healthChecker := service.NewHealthChecker(pgxPool, redisCl, s3Store)
	otpService := otp.NewOtpService(config.GetOTPConfig(), slogger)

	httpConfig := config.GetHttpServerConfig()
	router := routes.SetupRouter(appConfig.Env, repos, slogger, s3Store, otpService, healthChecker, metrics)

	if telemetryConfig.Enabled && telemetryConfig.EnableHTTPMetrics {
		router.GET(telemetryConfig.PrometheusPath, prometheusHandler(metrics))
	}

	if telemetryConfig.Enabled && telemetryConfig.EnableDBMetrics {
		telemetry.StartDBPoolCollector(ctx, metrics, func() telemetry.DBPoolStats {
			stats := pgxPool.Stat()
			return telemetry.DBPoolStats{
				AcquiredConnections:     int64(stats.AcquiredConns()),
				IdleConnections:         int64(stats.IdleConns()),
				ConstructingConnections: int64(stats.ConstructingConns()),
				EmptyAttempts:           stats.EmptyAcquireCount(),
				TotalAcquired:           stats.AcquireCount(),
			}
		}, 15*time.Second)
	}

	srv := NewHttpServer(router, httpConfig)

	go func() {
		slogger.Info("Starting server", slog.String("port", httpConfig.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Failed to start server", slog.Any("error", err))
			stop()
		}
	}()

	<-ctx.Done()
	slogger.Info("Shutdown signal received")

	slogger.Info("Shutting down server")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), httpConfig.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("Server forced to shutdown", slog.Any("error", err))
	} else {
		slogger.Info("HTTP server stopped gracefully")
	}

	slogger.Info("Server exited")
}

func prometheusHandler(m *telemetry.Metrics) gin.HandlerFunc {
	return func(c *gin.Context) {
		promhttp.HandlerFor(m.Registry, promhttp.HandlerOpts{}).ServeHTTP(c.Writer, c.Request)
	}
}
