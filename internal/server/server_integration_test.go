//go:build integration

package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/BhaumikTalwar/Gama/internal/telemetry"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestPrometheusHandler_ServesMetrics(t *testing.T) {
	m := telemetry.NewMetrics("testapp")
	handler := prometheusHandler(m)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/metrics", handler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/metrics", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	body := w.Body.String()
	assert.Contains(t, body, "testapp_http_requests_in_flight")
}

func TestPrometheusHandler_WithCustomMetric(t *testing.T) {
	reg := prometheus.NewRegistry()
	g := prometheus.NewGauge(prometheus.GaugeOpts{Name: "custom_metric"})
	g.Set(99)
	reg.MustRegister(g)

	m := &telemetry.Metrics{Registry: reg}
	handler := prometheusHandler(m)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/metrics", handler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/metrics", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	body := w.Body.String()
	assert.Contains(t, body, "custom_metric")
	assert.Contains(t, body, "99")
}

func TestPrometheusHandler_IncludesGoMetrics(t *testing.T) {
	m := telemetry.NewMetrics("testapp")
	handler := prometheusHandler(m)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/metrics", handler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/metrics", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	body := w.Body.String()
	assert.Contains(t, body, "go_goroutines")
}
