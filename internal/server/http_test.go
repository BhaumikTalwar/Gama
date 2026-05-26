package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/BhaumikTalwar/Gama/config"
	"github.com/stretchr/testify/assert"
)

func TestNewHttpServer_SetsAllFields(t *testing.T) {
	handler := http.NewServeMux()
	cfg := &config.HTTPServerConfig{
		Host:              "0.0.0.0",
		Port:              "8080",
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       2 * time.Minute,
		MaxHeaderBytes:    1 << 20,
		ShutdownTimeout:   10 * time.Second,
		KeepAlive:         true,
	}

	srv := NewHttpServer(handler, cfg)

	assert.Equal(t, "0.0.0.0:8080", srv.Addr)
	assert.Equal(t, handler, srv.Handler)
	assert.Equal(t, 10*time.Second, srv.ReadTimeout)
	assert.Equal(t, 5*time.Second, srv.ReadHeaderTimeout)
	assert.Equal(t, 15*time.Second, srv.WriteTimeout)
	assert.Equal(t, 2*time.Minute, srv.IdleTimeout)
	assert.Equal(t, 1<<20, srv.MaxHeaderBytes)
}

func TestNewHttpServer_KeepAliveDisabled(t *testing.T) {
	handler := http.NewServeMux()
	cfg := &config.HTTPServerConfig{Host: "0.0.0.0", Port: "8080", KeepAlive: false}
	srv := NewHttpServer(handler, cfg)
	assert.NotNil(t, srv)
}

func TestNewHttpServer_ServesRequests(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	cfg := &config.HTTPServerConfig{Host: "127.0.0.1", Port: "0", KeepAlive: false}
	srv := NewHttpServer(mux, cfg)
	assert.NotNil(t, srv)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	mux.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestNewHttpServer_IPv6HostPort(t *testing.T) {
	handler := http.NewServeMux()
	cfg := &config.HTTPServerConfig{Host: "::1", Port: "9090", KeepAlive: false}
	srv := NewHttpServer(handler, cfg)
	assert.Equal(t, "[::1]:9090", srv.Addr)
}

func TestNewHttpServer_ZeroValues(t *testing.T) {
	handler := http.NewServeMux()
	cfg := &config.HTTPServerConfig{Host: "", Port: "8080"}
	srv := NewHttpServer(handler, cfg)
	assert.Equal(t, ":8080", srv.Addr)
	assert.Equal(t, time.Duration(0), srv.ReadTimeout)
	assert.Equal(t, time.Duration(0), srv.ReadHeaderTimeout)
}
