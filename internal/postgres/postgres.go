package postgres

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"os"

	"github.com/BhaumikTalwar/Gama/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(ctx context.Context, appName string, cfg *config.PostgresConfig) (*pgxpool.Pool, error) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database, cfg.SSLMode)

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	poolConfig.ConnConfig.RuntimeParams["statement_timeout"] = fmt.Sprintf("%d", int64(cfg.StatementTimeout.Milliseconds()))
	poolConfig.ConnConfig.RuntimeParams["idle_in_transaction_session_timeout"] = fmt.Sprintf("%d", int64(cfg.IdleTxnTimeout.Milliseconds()))
	poolConfig.ConnConfig.RuntimeParams["application_name"] = appName

	if cfg.ConnectTimeout > 0 {
		poolConfig.ConnConfig.ConnectTimeout = cfg.ConnectTimeout
	}

	if cfg.SSLMode != "" && cfg.SSLMode != "disable" && (cfg.SSLCAPath != "" || (cfg.SSLCertPath != "" && cfg.SSLKeyPath != "")) {
		tlsConfig, err := getTLSConfig(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to load TLS config: %w", err)
		}

		poolConfig.ConnConfig.TLSConfig = tlsConfig
	}

	if cfg.KeepAlive && cfg.KeepAliveInterval > 0 && cfg.ConnectTimeout > 0 {
		dialer := &net.Dialer{
			KeepAlive: cfg.KeepAliveInterval,
			Timeout:   cfg.ConnectTimeout,
		}

		poolConfig.ConnConfig.DialFunc = dialer.DialContext
	}

	if cfg.MaxConns > 0 {
		poolConfig.MaxConns = int32(cfg.MaxConns)
	}
	if cfg.MinConns > 0 {
		poolConfig.MinConns = int32(cfg.MinConns)
	}

	if cfg.MaxConnIdleTime > 0 {
		poolConfig.MaxConnIdleTime = cfg.MaxConnIdleTime
	}

	if cfg.KeepAliveInterval > 0 {
		poolConfig.HealthCheckPeriod = cfg.KeepAliveInterval
	}

	return pgxpool.NewWithConfig(ctx, poolConfig)
}

func getTLSConfig(cfg *config.PostgresConfig) (*tls.Config, error) {
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
		ServerName: cfg.Host,
	}

	if cfg.SSLCAPath != "" {
		rootCertPool := x509.NewCertPool()
		pem, err := os.ReadFile(cfg.SSLCAPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA cert: %w", err)
		}

		if ok := rootCertPool.AppendCertsFromPEM(pem); !ok {
			return nil, fmt.Errorf("failed to append CA cert")
		}

		tlsConfig.RootCAs = rootCertPool
	}

	if cfg.SSLCertPath != "" && cfg.SSLKeyPath != "" {
		cert, err := tls.LoadX509KeyPair(cfg.SSLCertPath, cfg.SSLKeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load client cert/key pair: %w", err)
		}

		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	return tlsConfig, nil
}
