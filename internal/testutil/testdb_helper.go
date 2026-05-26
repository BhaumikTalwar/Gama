//go:build integration

package testutil

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func GetTestDBURI() string {
	if uri := os.Getenv("TEST_POSTGRES_URI"); uri != "" {
		return uri
	}
	return "postgres://devuser:devpass@localhost:5432/devdb?sslmode=disable"
}

func SetupTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	ctx := context.Background()
	uri := GetTestDBURI()
	pool, err := pgxpool.New(ctx, uri)
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("failed to ping test database: %v", err)
	}
	return pool
}

func GetTestRedisAddr() string {
	if addr := os.Getenv("TEST_REDIS_ADDR"); addr != "" {
		return addr
	}
	return "localhost:6379"
}

func CleanTestDB(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	tables := []string{
		"user_logs",
		"verification_tokens",
		"refresh_tokens",
		"user_mfa_settings",
		"user_roles",
		"role_permissions",
		"permissions",
		"roles",
		"users",
	}

	for _, table := range tables {
		_, err := pool.Exec(context.Background(), fmt.Sprintf("DELETE FROM %s", table))
		if err != nil {
			t.Fatalf("failed to clean table %s: %v", table, err)
		}
	}
}
