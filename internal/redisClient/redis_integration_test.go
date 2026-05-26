//go:build integration

package redisClient

import (
	"context"
	"os"
	"testing"

	"github.com/BhaumikTalwar/Gama/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getTestRedisConfig(t *testing.T) *config.RedisConfig {
	t.Helper()
	return &config.RedisConfig{
		RedisHost:     getEnvDefault("TEST_REDIS_HOST", "localhost"),
		RedisPort:     6379,
		RedisPassword: "",
		RedisDB:       1,
		RedisPoolSize: 5,
		RedisUseTLS:   false,
	}
}

func getEnvDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func TestNewRedisClient_Success(t *testing.T) {
	cfg := getTestRedisConfig(t)
	client, err := NewRedisClient(context.Background(), cfg)
	require.NoError(t, err)
	require.NotNil(t, client)

	err = client.Ping(context.Background()).Err()
	assert.NoError(t, err)

	client.Close()
}

func TestNewRedisClient_WrongPort(t *testing.T) {
	cfg := getTestRedisConfig(t)
	cfg.RedisPort = 6380

	client, err := NewRedisClient(context.Background(), cfg)
	if err == nil {
		client.Close()
	}
	assert.Error(t, err)
}

func TestNewRedisClient_CustomDB(t *testing.T) {
	cfg := getTestRedisConfig(t)
	cfg.RedisDB = 2

	client, err := NewRedisClient(context.Background(), cfg)
	require.NoError(t, err)
	defer client.Close()

	err = client.Ping(context.Background()).Err()
	assert.NoError(t, err)
}

func TestNewRedisClient_WithPassword(t *testing.T) {
	cfg := getTestRedisConfig(t)
	cfg.RedisPassword = "somepassword"

	client, err := NewRedisClient(context.Background(), cfg)
	if err != nil {
		assert.Nil(t, client)
		return
	}
	defer client.Close()

	err = client.Ping(context.Background()).Err()
	assert.NoError(t, err)
}
