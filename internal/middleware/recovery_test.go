package middleware

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCustomSlogRecovery_Returns500OnPanic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := slog.New(slog.NewTextHandler(&testWriter{}, &slog.HandlerOptions{}))

	router := gin.New()
	router.Use(CustomSlogRecovery(logger))
	router.GET("/panic", func(c *gin.Context) {
		panic("test panic")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/panic", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCustomSlogRecovery_DoesNotAbortIfNoPanic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := slog.New(slog.NewTextHandler(&testWriter{}, &slog.HandlerOptions{}))

	router := gin.New()
	router.Use(CustomSlogRecovery(logger))
	router.GET("/ok", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ok", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "OK", w.Body.String())
}

func TestCustomSlogRecovery_RecoversFromDifferentPanicTypes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := slog.New(slog.NewTextHandler(&testWriter{}, &slog.HandlerOptions{}))

	testCases := []struct {
		name        string
		panicValue  interface{}
	}{
		{"string panic", "something went wrong"},
		{"nil panic", nil},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			router := gin.New()
			router.Use(CustomSlogRecovery(logger))
			router.GET("/panic", func(c *gin.Context) {
				panic(tc.panicValue)
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/panic", nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusInternalServerError, w.Code)
		})
	}
}

type testWriter struct{}

func (testWriter) Write(p []byte) (int, error) {
	return len(p), nil
}