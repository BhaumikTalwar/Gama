package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRequestIDMiddleware_GeneratesNewIDWhenHeaderAbsent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.NotEmpty(t, w.Header().Get(Header))
	assert.Len(t, w.Header().Get(Header), 20)
}

func TestRequestIDMiddleware_ForwardsExistingID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set(Header, "existing-request-id-123")
	router.ServeHTTP(w, req)

	assert.Equal(t, "existing-request-id-123", w.Header().Get(Header))
}

func TestRequestIDMiddleware_SetsContextValue(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var capturedID string
	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.GET("/test", func(c *gin.Context) {
		capturedID = c.GetString(RequestIdKey)
		c.String(http.StatusOK, "OK")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.NotEmpty(t, capturedID)
	assert.Equal(t, w.Header().Get(Header), capturedID)
}

func TestRequestIDMiddleware_SetsHeaderAndContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.GET("/test", func(c *gin.Context) {
		contextID := c.GetString(RequestIdKey)
		assert.NotEmpty(t, contextID)
		c.String(http.StatusOK, "OK")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, w.Header().Get(Header))
}

func TestRequestIDMiddleware_IDUniqueness(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)
		ids[w.Header().Get(Header)] = true
	}

	assert.Greater(t, len(ids), 1, "Request IDs should be unique across requests")
}
