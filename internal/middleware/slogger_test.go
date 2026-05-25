package middleware

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestStructuredLoggerSlog_LogsRequestDetails(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var logOutput strings.Builder
	logger := slog.New(slog.NewTextHandler(&logOutput, &slog.HandlerOptions{}))

	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.Use(StructuredLoggerSlog(logger))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, logOutput.String(), "HTTP Request")
	assert.Contains(t, logOutput.String(), "GET")
	assert.Contains(t, logOutput.String(), "/test")
}

func TestStructuredLoggerSlog_LogsStatusCode(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var logOutput strings.Builder
	logger := slog.New(slog.NewTextHandler(&logOutput, &slog.HandlerOptions{}))

	router := gin.New()
	router.Use(StructuredLoggerSlog(logger))
	router.GET("/ok", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ok", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestStructuredLoggerSlog_LogsLatency(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var logOutput strings.Builder
	logger := slog.New(slog.NewTextHandler(&logOutput, &slog.HandlerOptions{}))

	router := gin.New()
	router.Use(StructuredLoggerSlog(logger))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestStructuredLoggerSlog_AppendsQueryString(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var logOutput strings.Builder
	logger := slog.New(slog.NewTextHandler(&logOutput, &slog.HandlerOptions{}))

	router := gin.New()
	router.Use(StructuredLoggerSlog(logger))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test?foo=bar", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestStructuredLoggerSlog_IncludesClientIP(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var logOutput strings.Builder
	logger := slog.New(slog.NewTextHandler(&logOutput, &slog.HandlerOptions{}))

	router := gin.New()
	router.Use(StructuredLoggerSlog(logger))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestStructuredLoggerSlog_IncludesUserAgent(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var logOutput strings.Builder
	logger := slog.New(slog.NewTextHandler(&logOutput, &slog.HandlerOptions{}))

	router := gin.New()
	router.Use(StructuredLoggerSlog(logger))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("User-Agent", "TestAgent/1.0")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestStructuredLoggerSlog_LogsErrorIfPresent(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var logOutput strings.Builder
	logger := slog.New(slog.NewTextHandler(&logOutput, &slog.HandlerOptions{}))

	router := gin.New()
	router.Use(StructuredLoggerSlog(logger))
	router.GET("/error", func(c *gin.Context) {
		c.Error(gin.Error{
			Err:  assert.AnError,
			Meta: "test error",
			Type: gin.ErrorTypePrivate,
		})
		c.String(http.StatusInternalServerError, "Error")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/error", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestStructuredLoggerSlog_NoErrorIfNoErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var logOutput strings.Builder
	logger := slog.New(slog.NewTextHandler(&logOutput, &slog.HandlerOptions{}))

	router := gin.New()
	router.Use(StructuredLoggerSlog(logger))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotContains(t, logOutput.String(), "Request errors")
}