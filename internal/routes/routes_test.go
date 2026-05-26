package routes

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/BhaumikTalwar/Gama/config"
	"github.com/BhaumikTalwar/Gama/internal/caching"
	"github.com/BhaumikTalwar/Gama/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	cfg := &config.AppConfig{
		AppName:                  "TestApp",
		Env:                      "test",
		CorsAddresses:            []string{"*"},
		JWTKey:                   "test_jwt_secret_key_that_is_32_bytes",
		AppSecret:                "test_app_secret_key_32_byte_long!!",
		AESKey:                   "test_aes_key_32_byte_long_for_aes",
		AccessTokenDuration:      15 * time.Minute,
		RefreshTokenDuration:     24 * time.Hour,
		RefreshRotationThreshold: 1 * time.Hour,
	}
	config.SetTestAppConfig(cfg)
}

func TestSetupRouter_BasicRoutes(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(discardWriter{}, nil))
	router := SetupRouter("test", nil, logger, &service.S3Store{}, nil, nil, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/health/live", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/health/ready", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSetupRouter_HealthWithoutChecker(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(discardWriter{}, nil))
	router := SetupRouter("test", nil, logger, nil, nil, nil, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSetupRouter_AuthRoutesRegistered(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(discardWriter{}, nil))
	s3Store := &service.S3Store{}

	_ = caching.NewNoOpCacheService()

	router := SetupRouter("test", nil, logger, s3Store, nil, nil, nil)

	routes := router.Routes()
	routePaths := make(map[string]string)
	for _, r := range routes {
		routePaths[r.Method+" "+r.Path] = r.Path
	}

	expectedRoutes := []string{
		"GET /health",
		"GET /health/live",
		"GET /health/ready",
		"POST /api/v1/auth/register",
		"POST /api/v1/auth/login",
		"POST /api/v1/auth/refresh",
		"POST /api/v1/auth/refresh/logout",
		"POST /api/v1/auth/mfa/setup/init",
		"POST /api/v1/auth/mfa/setup/finalize",
		"POST /api/v1/auth/mfa/verify",
		"POST /api/v1/auth/mfa/resend",
		"GET /api/v1/auth/me",
		"PUT /api/v1/auth/profile",
		"PUT /api/v1/auth/password",
	}

	for _, route := range expectedRoutes {
		assert.Contains(t, routePaths, route, "missing route: %s", route)
	}
}

func TestSetupRouter_ReleaseMode(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := slog.New(slog.NewTextHandler(discardWriter{}, nil))
	_ = SetupRouter("production", nil, logger, nil, nil, nil, nil)
	assert.Equal(t, gin.ReleaseMode, gin.Mode())
}

func TestSetupRouter_DebugMode(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := slog.New(slog.NewTextHandler(discardWriter{}, nil))
	_ = SetupRouter("dev", nil, logger, nil, nil, nil, nil)
	assert.Equal(t, gin.DebugMode, gin.Mode())
}

type discardWriter struct{}

func (discardWriter) Write(p []byte) (int, error) {
	return len(p), nil
}
