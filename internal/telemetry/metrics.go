package telemetry

import (
	"context"
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	Registry *prometheus.Registry

	HTTPRequestsTotal   *prometheus.CounterVec
	HTTPRequestDuration *prometheus.HistogramVec
	HTTPInFlight        prometheus.Gauge

	DBPoolConnections   *prometheus.GaugeVec
	DBPoolEmptyCurrent  prometheus.Gauge
	DBPoolAcquiresTotal prometheus.Counter
	DBPoolAcquireDuration prometheus.Histogram
	DBPoolEmptyAttempts  prometheus.Counter

	CacheHits   prometheus.Counter
	CacheMisses prometheus.Counter
	CacheHitRatio prometheus.Gauge

	SystemMemAlloc    prometheus.Gauge
	SystemMemSys      prometheus.Gauge
	SystemGoroutines  prometheus.Gauge
	SystemGCDuration  prometheus.Gauge
}

func NewMetrics(serviceName string) *Metrics {
	m := &Metrics{
		Registry: prometheus.NewRegistry(),
	}

	m.Registry.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{
		Namespace: serviceName,
	}))

	m.Registry.MustRegister(prometheus.NewGoCollector())

	m.HTTPRequestsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: serviceName,
		Subsystem: "http",
		Name:      "requests_total",
		Help:      "Total number of HTTP requests",
	}, []string{"method", "path", "status"})
	m.Registry.MustRegister(m.HTTPRequestsTotal)

	m.HTTPRequestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: serviceName,
		Subsystem: "http",
		Name:      "request_duration_seconds",
		Help:      "HTTP request duration in seconds",
		Buckets:   []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
	}, []string{"method", "path"})
	m.Registry.MustRegister(m.HTTPRequestDuration)

	m.HTTPInFlight = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: serviceName,
		Subsystem: "http",
		Name:      "requests_in_flight",
		Help:      "Number of HTTP requests currently being processed",
	})
	m.Registry.MustRegister(m.HTTPInFlight)

	m.DBPoolConnections = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: serviceName,
		Subsystem: "db",
		Name:      "pool_connections",
		Help:      "Number of connections in the pool",
	}, []string{"state"})
	m.Registry.MustRegister(m.DBPoolConnections)

	m.DBPoolEmptyCurrent = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: serviceName,
		Subsystem: "db",
		Name:      "pool_empty_current",
		Help:      "Current number of failed attempts to acquire due to empty pool",
	})
	m.Registry.MustRegister(m.DBPoolEmptyCurrent)

	m.DBPoolAcquiresTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: serviceName,
		Subsystem: "db",
		Name:      "pool_acquires_total",
		Help:      "Total number of connection acquisitions from pool",
	})
	m.Registry.MustRegister(m.DBPoolAcquiresTotal)

	m.DBPoolAcquireDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: serviceName,
		Subsystem: "db",
		Name:      "pool_acquire_duration_seconds",
		Help:      "Time to acquire a connection from pool",
		Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
	})
	m.Registry.MustRegister(m.DBPoolAcquireDuration)

	m.DBPoolEmptyAttempts = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: serviceName,
		Subsystem: "db",
		Name:      "pool_empty_attempts_total",
		Help:      "Total number of failed attempts to acquire a connection due to empty pool",
	})
	m.Registry.MustRegister(m.DBPoolEmptyAttempts)

	m.CacheHits = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: serviceName,
		Subsystem: "cache",
		Name:      "hits_total",
		Help:      "Total number of cache hits",
	})
	m.Registry.MustRegister(m.CacheHits)

	m.CacheMisses = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: serviceName,
		Subsystem: "cache",
		Name:      "misses_total",
		Help:      "Total number of cache misses",
	})
	m.Registry.MustRegister(m.CacheMisses)

	m.CacheHitRatio = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: serviceName,
		Subsystem: "cache",
		Name:      "hit_ratio",
		Help:      "Cache hit ratio (0-1)",
	})
	m.Registry.MustRegister(m.CacheHitRatio)

	m.SystemMemAlloc = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: serviceName,
		Subsystem: "system",
		Name:      "memory_alloc_bytes",
		Help:      "Memory currently allocated",
	})
	m.Registry.MustRegister(m.SystemMemAlloc)

	m.SystemMemSys = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: serviceName,
		Subsystem: "system",
		Name:      "memory_sys_bytes",
		Help:      "Memory obtained from system",
	})
	m.Registry.MustRegister(m.SystemMemSys)

	m.SystemGoroutines = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: serviceName,
		Subsystem: "system",
		Name:      "goroutines",
		Help:      "Number of goroutines",
	})
	m.Registry.MustRegister(m.SystemGoroutines)

	m.SystemGCDuration = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: serviceName,
		Subsystem: "system",
		Name:      "gc_pause_duration_seconds",
		Help:      "Duration of last GC pause",
	})
	m.Registry.MustRegister(m.SystemGCDuration)

	return m
}

func (m *Metrics) RecordHTTPRequest(method, path string, status int, duration time.Duration) {
	m.HTTPRequestsTotal.WithLabelValues(method, path, statusCodeToString(status)).Inc()
	m.HTTPRequestDuration.WithLabelValues(method, path).Observe(duration.Seconds())
}

func (m *Metrics) IncInFlight() {
	m.HTTPInFlight.Inc()
}

func (m *Metrics) DecInFlight() {
	m.HTTPInFlight.Dec()
}

func (m *Metrics) RecordDBPoolStats(stats DBPoolStats) {
	m.DBPoolConnections.WithLabelValues("acquired").Set(float64(stats.AcquiredConnections))
	m.DBPoolConnections.WithLabelValues("idle").Set(float64(stats.IdleConnections))
	m.DBPoolConnections.WithLabelValues("constructing").Set(float64(stats.ConstructingConnections))
	m.DBPoolEmptyCurrent.Set(float64(stats.EmptyAttempts))

	if stats.AcquiredConnections > 0 || stats.IdleConnections > 0 {
		ratio := float64(stats.AcquiredConnections) / float64(stats.AcquiredConnections+stats.IdleConnections)
		m.CacheHitRatio.Set(ratio)
	}
}

func (m *Metrics) RecordCacheHit() {
	m.CacheHits.Inc()
	m.updateCacheRatio()
}

func (m *Metrics) RecordCacheMiss() {
	m.CacheMisses.Inc()
	m.updateCacheRatio()
}

func (m *Metrics) updateCacheRatio() {
}

func (m *Metrics) RecordSystemMetrics() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	m.SystemMemAlloc.Set(float64(memStats.Alloc))
	m.SystemMemSys.Set(float64(memStats.Sys))
	m.SystemGoroutines.Set(float64(runtime.NumGoroutine()))
}

type DBPoolStats struct {
	AcquiredConnections   int64
	IdleConnections       int64
	ConstructingConnections int64
	EmptyAttempts         int64
	TotalAcquired         int64
}

func statusCodeToString(status int) string {
	switch {
	case status >= 200 && status < 300:
		return "2xx"
	case status >= 300 && status < 400:
		return "3xx"
	case status >= 400 && status < 500:
		return "4xx"
	case status >= 500:
		return "5xx"
	default:
		return "unknown"
	}
}

func StartDBPoolCollector(ctx context.Context, m *Metrics, getStats func() DBPoolStats, interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				stats := getStats()
				m.RecordDBPoolStats(stats)
				m.RecordSystemMetrics()
			}
		}
	}()
}