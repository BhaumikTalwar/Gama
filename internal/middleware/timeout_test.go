package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestTimeoutMiddleware_PassesThroughNonTimeoutRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(TimeoutMiddleware(5 * time.Second))
	router.GET("/fast", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/fast", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "OK", w.Body.String())
}

func TestTimeoutMiddleware_TriggersTimeout(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(TimeoutMiddleware(10 * time.Millisecond))
	router.GET("/slow", func(c *gin.Context) {
		time.Sleep(100 * time.Millisecond)
		c.String(http.StatusOK, "OK")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/slow", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusGatewayTimeout, w.Code)
	assert.Contains(t, w.Body.String(), "request_timeout")
}

func TestTimeoutMiddleware_ReturnsCorrectJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(TimeoutMiddleware(10 * time.Millisecond))
	router.GET("/slow", func(c *gin.Context) {
		time.Sleep(100 * time.Millisecond)
		c.String(http.StatusOK, "OK")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/slow", nil)
	router.ServeHTTP(w, req)

	assert.Contains(t, w.Body.String(), "error")
	assert.Contains(t, w.Body.String(), "504")
}

func TestTimeoutMiddleware_DifferentDurations(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		timeout    time.Duration
		sleepDelay time.Duration
		expectCode int
	}{
		{
			name:       "1ms timeout should trigger for 100ms delay",
			timeout:    1 * time.Millisecond,
			sleepDelay: 100 * time.Millisecond,
			expectCode: http.StatusGatewayTimeout,
		},
		{
			name:       "10ms timeout should trigger for 50ms delay",
			timeout:    10 * time.Millisecond,
			sleepDelay: 50 * time.Millisecond,
			expectCode: http.StatusGatewayTimeout,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(TimeoutMiddleware(tt.timeout))
			router.GET("/delayed", func(c *gin.Context) {
				time.Sleep(tt.sleepDelay)
				c.String(http.StatusOK, "OK")
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/delayed", nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectCode, w.Code)
		})
	}
}
