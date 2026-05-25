package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type HealthChecker struct {
	db       *pgxpool.Pool
	redis    *redis.Client
	s3Client *S3Store
}

type ComponentHealth struct {
	Status  string      `json:"status"`
	Latency string      `json:"latency,omitempty"`
	Error   string      `json:"error,omitempty"`
	Details interface{} `json:"details,omitempty"`
}

type HealthResponse struct {
	Status     string                     `json:"status"`
	Timestamp  string                     `json:"timestamp"`
	Components map[string]ComponentHealth `json:"components"`
}

func (h *HealthResponse) ComputeStatus() string {
	if h.Components == nil {
		return h.Status
	}

	hasUnhealthy := false
	hasDegraded := false

	for _, comp := range h.Components {
		if comp.Status == "unhealthy" {
			hasUnhealthy = true
			break
		}
		if comp.Status == "degraded" {
			hasDegraded = true
		}
	}

	if hasUnhealthy {
		return "unhealthy"
	}
	if hasDegraded {
		return "degraded"
	}
	return "healthy"
}

func NewHealthChecker(db *pgxpool.Pool, redisClient *redis.Client, s3Client *S3Store) *HealthChecker {
	return &HealthChecker{
		db:       db,
		redis:    redisClient,
		s3Client: s3Client,
	}
}

func (h *HealthChecker) Check(ctx context.Context) HealthResponse {
	response := HealthResponse{
		Status:     "healthy",
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
		Components: make(map[string]ComponentHealth),
	}

	dbHealth := h.checkDatabase(ctx)
	response.Components["database"] = dbHealth
	if dbHealth.Status != "healthy" {
		response.Status = "unhealthy"
	}

	redisHealth := h.checkRedis(ctx)
	response.Components["redis"] = redisHealth
	if redisHealth.Status != "healthy" {
		response.Status = "unhealthy"
	}

	s3Health := h.checkS3(ctx)
	response.Components["s3"] = s3Health
	if s3Health.Status != "healthy" {
		response.Status = "degraded"
	}

	return response
}

func (h *HealthChecker) checkDatabase(ctx context.Context) ComponentHealth {
	start := time.Now()

	err := h.db.Ping(ctx)
	latency := time.Since(start)

	if err != nil {
		return ComponentHealth{
			Status: "unhealthy",
			Error:  fmt.Sprintf("database ping failed: %v", err),
		}
	}

	var dbSize int64
	err = h.db.QueryRow(ctx, "SELECT COUNT(*) FROM pg_database WHERE datname = current_database()").Scan(&dbSize)

	return ComponentHealth{
		Status:  "healthy",
		Latency: latency.String(),
		Details: map[string]interface{}{
			"database": "connected",
			"exists":   dbSize > 0,
		},
	}
}

func (h *HealthChecker) checkRedis(ctx context.Context) ComponentHealth {
	start := time.Now()

	err := h.redis.Ping(ctx).Err()
	latency := time.Since(start)

	if err != nil {
		return ComponentHealth{
			Status: "unhealthy",
			Error:  fmt.Sprintf("redis ping failed: %v", err),
		}
	}

	info, _ := h.redis.Info(ctx, "server").Result()
	redisVersion := ""
	for _, line := range strings.Split(info, "\n") {
		if strings.HasPrefix(line, "redis_version:") {
			redisVersion = strings.TrimSpace(strings.TrimPrefix(line, "redis_version:"))
			break
		}
	}

	return ComponentHealth{
		Status:  "healthy",
		Latency: latency.String(),
		Details: map[string]interface{}{
			"connected": true,
			"version":   redisVersion,
		},
	}
}

func (h *HealthChecker) checkS3(ctx context.Context) ComponentHealth {
	start := time.Now()

	if h.s3Client == nil || h.s3Client.client == nil {
		return ComponentHealth{
			Status: "degraded",
			Details: map[string]interface{}{
				"connected": false,
				"reason":    "S3 client not configured",
			},
		}
	}

	_, err := h.s3Client.client.ListBuckets(ctx, nil)
	latency := time.Since(start)

	if err != nil {
		return ComponentHealth{
			Status: "degraded",
			Error:  fmt.Sprintf("S3 connection failed: %v", err),
		}
	}

	return ComponentHealth{
		Status:  "healthy",
		Latency: latency.String(),
		Details: map[string]interface{}{
			"connected": true,
			"endpoint":  h.s3Client.GetEndpoint(),
			"public":    h.s3Client.GetPublicBucket(),
			"private":   h.s3Client.GetPrivateBucket(),
		},
	}
}

func (h *HealthChecker) CheckDatabaseOnly(ctx context.Context) error {
	return h.db.Ping(ctx)
}

func (h *HealthChecker) Liveness() ComponentHealth {
	return ComponentHealth{
		Status: "alive",
	}
}

func (h *HealthChecker) Readiness() ComponentHealth {
	if h.db == nil {
		return ComponentHealth{
			Status: "not ready",
			Error:  "database not configured",
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	health := h.Check(ctx)

	if health.Status == "healthy" {
		return ComponentHealth{
			Status: "ready",
		}
	}

	return ComponentHealth{
		Status: "not ready",
		Error:  "one or more components unhealthy",
	}
}

func CheckTCPConnection(host string, port string, timeout time.Duration) error {
	address := net.JoinHostPort(host, port)
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return err
	}
	conn.Close()
	return nil
}

type HealthHandler struct {
	healthChecker *HealthChecker
}

func NewHealthHandler(hc *HealthChecker) *HealthHandler {
	return &HealthHandler{healthChecker: hc}
}

func (hh *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	response := hh.healthChecker.Check(r.Context())

	status := http.StatusOK
	if response.Status == "unhealthy" {
		status = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(response)
}

func (hh *HealthHandler) Live(w http.ResponseWriter, r *http.Request) {
	health := hh.healthChecker.Liveness()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

func (hh *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	health := hh.healthChecker.Readiness()

	w.Header().Set("Content-Type", "application/json")

	if health.Status != "ready" {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(health)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(health)
}
