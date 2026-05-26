package redisClient

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"time"

	"github.com/BhaumikTalwar/Gama/config"
	"github.com/redis/go-redis/v9"
)

func NewRedisClient(ctx context.Context, cfg *config.RedisConfig) (*redis.Client, error) {
	opts := &redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.RedisHost, cfg.RedisPort),
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,

		PoolSize:     cfg.RedisPoolSize,
		MinIdleConns: cfg.RedisMinIdleConns,

		DialTimeout:  cfg.RedisDialTimeout,
		ReadTimeout:  cfg.RedisReadTimeout,
		WriteTimeout: cfg.RedisWriteTimeout,
	}

	if cfg.RedisUseTLS {
		tlsConfig, err := getRedisTLSConfig(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to create redis TLS config: %w", err)
		}
		opts.TLSConfig = tlsConfig
	}

	client := redis.NewClient(opts)

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := client.Ping(pingCtx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return client, nil
}

func getRedisTLSConfig(cfg *config.RedisConfig) (*tls.Config, error) {
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
		ServerName: cfg.RedisHost,
	}

	if cfg.RedisTLSCAFile != "" {
		caCert, err := os.ReadFile(cfg.RedisTLSCAFile)
		if err != nil {
			return nil, fmt.Errorf("reading redis CA file: %w", err)
		}

		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("failed to append redis CA cert")
		}

		tlsConfig.RootCAs = caCertPool
	}

	if cfg.RedisTLSCertFile != "" && cfg.RedisTLSKeyFile != "" {
		cert, err := tls.LoadX509KeyPair(cfg.RedisTLSCertFile, cfg.RedisTLSKeyFile)
		if err != nil {
			return nil, fmt.Errorf("loading redis client cert/key: %w", err)
		}

		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	return tlsConfig, nil
}
