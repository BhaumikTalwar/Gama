package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAppConfig_ValidateValidDev(t *testing.T) {
	cfg := &AppConfig{
		AppName:                  "TestApp",
		Env:                      "dev",
		CorsAddresses:            []string{"http://localhost:3000"},
		AppSecret:                "12345678901234567890123456789012",
		JWTKey:                   "12345678901234567890123456789012",
		AESKey:                   "12345678901234567890123456789012",
		AccessTokenDuration:      15 * time.Minute,
		RefreshTokenDuration:     24 * time.Hour,
		RefreshRotationThreshold: 1 * time.Hour,
	}

	err := cfg.validate()
	assert.NoError(t, err)
}

func TestAppConfig_ValidateValidProd(t *testing.T) {
	cfg := &AppConfig{
		AppName:                  "TestApp",
		Env:                      "PROD",
		CorsAddresses:            []string{"http://localhost:3000"},
		AppSecret:                "12345678901234567890123456789012",
		JWTKey:                   "12345678901234567890123456789012",
		AESKey:                   "12345678901234567890123456789012",
		AccessTokenDuration:      15 * time.Minute,
		RefreshTokenDuration:     24 * time.Hour,
		RefreshRotationThreshold: 1 * time.Hour,
	}

	err := cfg.validate()
	assert.NoError(t, err)
}

func TestAppConfig_ValidateInvalidEnv(t *testing.T) {
	cfg := &AppConfig{
		Env:       "staging",
		AppSecret: "12345678901234567890123456789012",
		JWTKey:    "12345678901234567890123456789012",
		AESKey:    "12345678901234567890123456789012",
	}

	err := cfg.validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid Env")
}

func TestAppConfig_ValidateEmptyAppSecret(t *testing.T) {
	cfg := &AppConfig{
		Env:       "dev",
		AppSecret: "",
		JWTKey:    "12345678901234567890123456789012",
		AESKey:    "12345678901234567890123456789012",
	}

	err := cfg.validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "app_secret")
}

func TestAppConfig_ValidateEmptyCorsAddresses(t *testing.T) {
	cfg := &AppConfig{
		Env:           "dev",
		CorsAddresses: []string{},
		AppSecret:     "12345678901234567890123456789012",
		JWTKey:        "12345678901234567890123456789012",
		AESKey:        "12345678901234567890123456789012",
	}

	err := cfg.validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Cors")
}

func TestAppConfig_ValidateShortAppSecret(t *testing.T) {
	cfg := &AppConfig{
		Env:           "dev",
		CorsAddresses: []string{"http://localhost:3000"},
		AppSecret:     "short",
		JWTKey:        "12345678901234567890123456789012",
		AESKey:        "12345678901234567890123456789012",
	}

	err := cfg.validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "32 bytes")
}

func TestAppConfig_ValidateEmptyJWTKey(t *testing.T) {
	cfg := &AppConfig{
		Env:           "dev",
		CorsAddresses: []string{"http://localhost:3000"},
		AppSecret:     "12345678901234567890123456789012",
		JWTKey:        "",
		AESKey:        "12345678901234567890123456789012",
	}

	err := cfg.validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "jwt_key")
}

func TestAppConfig_ValidateShortJWTKey(t *testing.T) {
	cfg := &AppConfig{
		Env:           "dev",
		CorsAddresses: []string{"http://localhost:3000"},
		AppSecret:     "12345678901234567890123456789012",
		JWTKey:        "short",
		AESKey:        "12345678901234567890123456789012",
	}

	err := cfg.validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "JWT")
}

func TestAppConfig_ValidateEmptyAESKey(t *testing.T) {
	cfg := &AppConfig{
		Env:           "dev",
		CorsAddresses: []string{"http://localhost:3000"},
		AppSecret:     "12345678901234567890123456789012",
		JWTKey:        "12345678901234567890123456789012",
		AESKey:        "",
	}

	err := cfg.validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "aes_key")
}

func TestAppConfig_ValidateShortAESKey(t *testing.T) {
	cfg := &AppConfig{
		Env:           "dev",
		CorsAddresses: []string{"http://localhost:3000"},
		AppSecret:     "12345678901234567890123456789012",
		JWTKey:        "12345678901234567890123456789012",
		AESKey:        "short",
	}

	err := cfg.validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "AES")
}

func TestHTTPServerConfig_ValidateValid(t *testing.T) {
	cfg := &HTTPServerConfig{
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

	err := cfg.validate()
	assert.NoError(t, err)
}

func TestHTTPServerConfig_ValidateEmptyHost(t *testing.T) {
	cfg := &HTTPServerConfig{
		Host:              "",
		Port:              "8080",
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       2 * time.Minute,
		MaxHeaderBytes:    1 << 20,
		ShutdownTimeout:   10 * time.Second,
	}

	err := cfg.validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "host")
}

func TestHTTPServerConfig_ValidateInvalidPort(t *testing.T) {
	cfg := &HTTPServerConfig{
		Host:              "0.0.0.0",
		Port:              "invalid",
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       2 * time.Minute,
		MaxHeaderBytes:    1 << 20,
		ShutdownTimeout:   10 * time.Second,
	}

	err := cfg.validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid http.port")
}

func TestHTTPServerConfig_ValidatePortZero(t *testing.T) {
	cfg := &HTTPServerConfig{
		Host:              "0.0.0.0",
		Port:              "0",
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       2 * time.Minute,
		MaxHeaderBytes:    1 << 20,
		ShutdownTimeout:   10 * time.Second,
	}

	err := cfg.validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid http.port")
}

func TestHTTPServerConfig_ValidatePortTooHigh(t *testing.T) {
	cfg := &HTTPServerConfig{
		Host:              "0.0.0.0",
		Port:              "70000",
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       2 * time.Minute,
		MaxHeaderBytes:    1 << 20,
		ShutdownTimeout:   10 * time.Second,
	}

	err := cfg.validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid http.port")
}

func TestHTTPServerConfig_ValidateZeroReadTimeout(t *testing.T) {
	cfg := &HTTPServerConfig{
		Host:              "0.0.0.0",
		Port:              "8080",
		ReadTimeout:       0,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       2 * time.Minute,
		MaxHeaderBytes:    1 << 20,
		ShutdownTimeout:   10 * time.Second,
	}

	err := cfg.validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "read_timeout")
}

func TestHTTPServerConfig_ValidateZeroMaxHeaderBytes(t *testing.T) {
	cfg := &HTTPServerConfig{
		Host:              "0.0.0.0",
		Port:              "8080",
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       2 * time.Minute,
		MaxHeaderBytes:    0,
		ShutdownTimeout:   10 * time.Second,
	}

	err := cfg.validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "max_header_bytes")
}

func TestHTTPServerConfig_ValidateMaxHeaderBytesTooLarge(t *testing.T) {
	cfg := &HTTPServerConfig{
		Host:              "0.0.0.0",
		Port:              "8080",
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       2 * time.Minute,
		MaxHeaderBytes:    20 << 20,
		ShutdownTimeout:   10 * time.Second,
	}

	err := cfg.validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "too large")
}

func TestCacheConfig_ValidateValid(t *testing.T) {
	cfg := &CacheConfig{
		CacheNamespace:   "app_cache",
		EnableLocalCache: true,
		LocalCacheSize:   400,
		LocalCacheTTL:    5,
		Codec:            "json",
	}

	err := cfg.validate()
	assert.NoError(t, err)
}

func TestCacheConfig_ValidateEmptyNamespace(t *testing.T) {
	cfg := &CacheConfig{
		CacheNamespace:   "   ",
		EnableLocalCache: true,
		LocalCacheSize:   400,
		LocalCacheTTL:    5,
		Codec:            "json",
	}

	err := cfg.validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "namespace")
}

func TestCacheConfig_ValidateZeroLocalCacheSize(t *testing.T) {
	cfg := &CacheConfig{
		CacheNamespace:   "app_cache",
		EnableLocalCache: true,
		LocalCacheSize:   0,
		LocalCacheTTL:    5,
		Codec:            "json",
	}

	err := cfg.validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "local_size")
}

func TestCacheConfig_ValidateTooSmallLocalCacheSize(t *testing.T) {
	cfg := &CacheConfig{
		CacheNamespace:   "app_cache",
		EnableLocalCache: true,
		LocalCacheSize:   16,
		LocalCacheTTL:    5,
		Codec:            "json",
	}

	err := cfg.validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "32MB")
}

func TestCacheConfig_ValidateZeroLocalCacheTTL(t *testing.T) {
	cfg := &CacheConfig{
		CacheNamespace:   "app_cache",
		EnableLocalCache: true,
		LocalCacheSize:   400,
		LocalCacheTTL:    0,
		Codec:            "json",
	}

	err := cfg.validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "local_ttl")
}

func TestCacheConfig_ValidateInvalidCodec(t *testing.T) {
	cfg := &CacheConfig{
		CacheNamespace:   "app_cache",
		EnableLocalCache: true,
		LocalCacheSize:   400,
		LocalCacheTTL:    5,
		Codec:            "xml",
	}

	err := cfg.validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "codec")
}

func TestCacheConfig_ValidateMsgpackCodec(t *testing.T) {
	cfg := &CacheConfig{
		CacheNamespace:   "app_cache",
		EnableLocalCache: true,
		LocalCacheSize:   400,
		LocalCacheTTL:    5,
		Codec:            "msgpack",
	}

	err := cfg.validate()
	assert.NoError(t, err)
}

func TestCacheConfig_ValidateLocalCacheDisabled(t *testing.T) {
	cfg := &CacheConfig{
		CacheNamespace:   "app_cache",
		EnableLocalCache: false,
		LocalCacheSize:   0,
		LocalCacheTTL:    0,
		Codec:            "json",
	}

	err := cfg.validate()
	assert.NoError(t, err)
}
