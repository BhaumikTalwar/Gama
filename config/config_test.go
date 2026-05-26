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

func TestLoggingConfig_ValidateValid(t *testing.T) {
	cfg := &LoggingConfig{Level: "debug", Format: "json", Output: "stdout"}
	assert.NoError(t, cfg.validate())
}

func TestLoggingConfig_ValidateValidText(t *testing.T) {
	cfg := &LoggingConfig{Level: "info", Format: "text", Output: "file"}
	assert.NoError(t, cfg.validate())
}

func TestLoggingConfig_ValidateInvalidLevel(t *testing.T) {
	cfg := &LoggingConfig{Level: "trace", Format: "json", Output: "stdout"}
	err := cfg.validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid Logging Level")
}

func TestLoggingConfig_ValidateEmptyLevel(t *testing.T) {
	cfg := &LoggingConfig{Level: "", Format: "json", Output: "stdout"}
	err := cfg.validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid Logging Level")
}

func TestLoggingConfig_ValidateInvalidFormat(t *testing.T) {
	cfg := &LoggingConfig{Level: "info", Format: "yaml", Output: "stdout"}
	err := cfg.validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid Log Format")
}

func TestLoggingConfig_ValidateInvalidOutput(t *testing.T) {
	cfg := &LoggingConfig{Level: "info", Format: "json", Output: "console"}
	err := cfg.validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid Log Output")
}

func TestLoggingConfig_ValidateCaseInsensitiveLevel(t *testing.T) {
	cfg := &LoggingConfig{Level: "DEBUG", Format: "json", Output: "stdout"}
	assert.NoError(t, cfg.validate())
}

func TestTelemetryConfig_ValidateValid(t *testing.T) {
	cfg := &TelemetryConfig{TraceSampleRate: 0.5}
	assert.NoError(t, cfg.validate())
}

func TestTelemetryConfig_ValidateSampleRateZero(t *testing.T) {
	cfg := &TelemetryConfig{TraceSampleRate: 0.0}
	assert.NoError(t, cfg.validate())
}

func TestTelemetryConfig_ValidateSampleRateOne(t *testing.T) {
	cfg := &TelemetryConfig{TraceSampleRate: 1.0}
	assert.NoError(t, cfg.validate())
}

func TestTelemetryConfig_ValidateSampleRateNegative(t *testing.T) {
	cfg := &TelemetryConfig{TraceSampleRate: -0.1}
	err := cfg.validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "trace_sample_rate")
}

func TestTelemetryConfig_ValidateSampleRateTooHigh(t *testing.T) {
	cfg := &TelemetryConfig{TraceSampleRate: 1.1}
	err := cfg.validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "trace_sample_rate")
}

func TestPostgresConfig_ValidateValid(t *testing.T) {
	cfg := &PostgresConfig{Port: 5432}
	assert.NoError(t, cfg.validate())
}

func TestPostgresConfig_ValidatePortZero(t *testing.T) {
	cfg := &PostgresConfig{Port: 0}
	assert.NoError(t, cfg.validate())
}

func TestPostgresConfig_ValidatePortNegative(t *testing.T) {
	cfg := &PostgresConfig{Port: -1}
	err := cfg.validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid Postgres Port")
}

func TestPostgresConfig_ValidatePortTooHigh(t *testing.T) {
	cfg := &PostgresConfig{Port: 65536}
	err := cfg.validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid Postgres Port")
}

func TestPostgresConfig_ValidateMaxPort(t *testing.T) {
	cfg := &PostgresConfig{Port: 65535}
	assert.NoError(t, cfg.validate())
}

func TestRedisConfig_ValidateValid(t *testing.T) {
	cfg := &RedisConfig{RedisPort: 6379}
	assert.NoError(t, cfg.validate())
}

func TestRedisConfig_ValidatePortNegative(t *testing.T) {
	cfg := &RedisConfig{RedisPort: -1}
	err := cfg.validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid Redis Port")
}

func TestRedisConfig_ValidatePortTooHigh(t *testing.T) {
	cfg := &RedisConfig{RedisPort: 65536}
	err := cfg.validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid Redis Port")
}

func TestRedisConfig_ValidatePortZero(t *testing.T) {
	cfg := &RedisConfig{RedisPort: 0}
	assert.NoError(t, cfg.validate())
}

func TestS3Config_ValidateValid(t *testing.T) {
	cfg := &S3Config{
		Endpoint: "http://localhost:9000", Region: "us-east-1",
		AccessKey: "key", SecretKey: "secret",
		PublicBucket: "pub", PrivateBucket: "priv",
	}
	assert.NoError(t, cfg.validate())
}

func TestS3Config_ValidateEmptyEndpoint(t *testing.T) {
	cfg := &S3Config{Region: "us-east-1", AccessKey: "key", SecretKey: "secret", PublicBucket: "pub", PrivateBucket: "priv"}
	err := cfg.validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "endpoint")
}

func TestS3Config_ValidateEmptyRegion(t *testing.T) {
	cfg := &S3Config{Endpoint: "http://localhost:9000", AccessKey: "key", SecretKey: "secret", PublicBucket: "pub", PrivateBucket: "priv"}
	err := cfg.validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "region")
}

func TestS3Config_ValidateEmptyAccessKey(t *testing.T) {
	cfg := &S3Config{Endpoint: "http://localhost:9000", Region: "us-east-1", SecretKey: "secret", PublicBucket: "pub", PrivateBucket: "priv"}
	err := cfg.validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "access key")
}

func TestS3Config_ValidateEmptySecretKey(t *testing.T) {
	cfg := &S3Config{Endpoint: "http://localhost:9000", Region: "us-east-1", AccessKey: "key", PublicBucket: "pub", PrivateBucket: "priv"}
	err := cfg.validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "secret key")
}

func TestS3Config_ValidateEmptyPublicBucket(t *testing.T) {
	cfg := &S3Config{Endpoint: "http://localhost:9000", Region: "us-east-1", AccessKey: "key", SecretKey: "secret", PrivateBucket: "priv"}
	err := cfg.validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "public bucket")
}

func TestS3Config_ValidateEmptyPrivateBucket(t *testing.T) {
	cfg := &S3Config{Endpoint: "http://localhost:9000", Region: "us-east-1", AccessKey: "key", SecretKey: "secret", PublicBucket: "pub"}
	err := cfg.validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "private bucket")
}

func TestS3Config_ValidateMissingOptionalFields(t *testing.T) {
	cfg := &S3Config{
		Endpoint: "http://localhost:9000", Region: "us-east-1",
		AccessKey: "key", SecretKey: "secret",
		PublicBucket: "pub", PrivateBucket: "priv",
		UseSSL: true, PresignedURLTTL: 3600, PublicBaseURL: "https://cdn.example.com",
	}
	assert.NoError(t, cfg.validate())
}

func TestOTPConfig_ValidateValid(t *testing.T) {
	cfg := &OTPConfig{APIKey: "some-api-key"}
	assert.NoError(t, cfg.validate())
}

func TestOTPConfig_ValidateEmptyKey(t *testing.T) {
	cfg := &OTPConfig{APIKey: ""}
	err := cfg.validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "OTP API key")
}

func TestSqliteConfig_ValidateNoOp(t *testing.T) {
	cfg := &SqliteConfig{}
	err := cfg.validate()
	assert.NoError(t, err)
}

func TestSqliteConfig_ValidateWithPathStillNoOp(t *testing.T) {
	cfg := &SqliteConfig{Path: "/data/test.db"}
	err := cfg.validate()
	assert.NoError(t, err)
}
