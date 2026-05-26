//go:build integration

package postgres

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/BhaumikTalwar/Gama/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getTestPostgresConfig(t *testing.T) *config.PostgresConfig {
	t.Helper()
	return &config.PostgresConfig{
		Host:              getEnvDefault("TEST_PG_HOST", "localhost"),
		Port:              5432,
		User:              getEnvDefault("TEST_PG_USER", "devuser"),
		Password:          getEnvDefault("TEST_PG_PASS", "devpass"),
		Database:          getEnvDefault("TEST_PG_DB", "devdb"),
		SSLMode:           "disable",
		MaxConns:          5,
		MinConns:          1,
		MaxConnIdleTime:   30 * time.Second,
		AcquireTimeout:    10 * time.Second,
		ConnectTimeout:    5 * time.Second,
		KeepAlive:         false,
		KeepAliveInterval: 0,
	}
}

func getEnvDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func TestNewPool_Success(t *testing.T) {
	cfg := getTestPostgresConfig(t)
	pool, err := NewPool(context.Background(), "test-app", cfg)
	require.NoError(t, err)
	require.NotNil(t, pool)

	err = pool.Ping(context.Background())
	assert.NoError(t, err)

	pool.Close()
}

func TestNewPool_InvalidPort(t *testing.T) {
	cfg := getTestPostgresConfig(t)
	cfg.Port = 99999

	pool, err := NewPool(context.Background(), "test-app", cfg)
	if err == nil {
		pool.Close()
	}
	assert.Error(t, err)
}

func TestNewPool_WrongCredentials(t *testing.T) {
	cfg := getTestPostgresConfig(t)
	cfg.User = "nonexistent_user"

	pool, err := NewPool(context.Background(), "test-app", cfg)
	if err != nil {
		assert.Nil(t, pool)
		return
	}
	defer pool.Close()

	err = pool.Ping(context.Background())
	assert.Error(t, err)
}

func TestNewPool_PoolConfigValues(t *testing.T) {
	cfg := getTestPostgresConfig(t)
	cfg.MaxConns = 3
	cfg.MinConns = 1
	cfg.ConnectTimeout = 10 * time.Second
	cfg.KeepAlive = false

	pool, err := NewPool(context.Background(), "test-app", cfg)
	require.NoError(t, err)
	defer pool.Close()

	stats := pool.Stat()
	assert.GreaterOrEqual(t, stats.MaxConns(), int32(3))
}

func TestNewPool_MaxConnIdleTime(t *testing.T) {
	cfg := getTestPostgresConfig(t)
	cfg.MaxConnIdleTime = 5 * time.Minute

	pool, err := NewPool(context.Background(), "test-app", cfg)
	require.NoError(t, err)
	defer pool.Close()

	err = pool.Ping(context.Background())
	assert.NoError(t, err)
}

func TestNewPool_WithConnectTimeout(t *testing.T) {
	cfg := getTestPostgresConfig(t)
	cfg.ConnectTimeout = 1 * time.Second

	pool, err := NewPool(context.Background(), "test-app", cfg)
	require.NoError(t, err)
	defer pool.Close()

	assert.NoError(t, pool.Ping(context.Background()))
}
