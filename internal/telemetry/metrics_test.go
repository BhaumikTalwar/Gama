package telemetry

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMetrics_AllRegistered(t *testing.T) {
	m := NewMetrics("testapp")
	require.NotNil(t, m.Registry)

	_, err := m.Registry.Gather()
	require.NoError(t, err)
}

func TestNewMetrics_RegistryContainsExpectedMetrics(t *testing.T) {
	m := NewMetrics("testapp")

	m.RecordHTTPRequest("GET", "/", 200, time.Second)
	m.RecordDBPoolStats(DBPoolStats{AcquiredConnections: 1})

	families, err := m.Registry.Gather()
	require.NoError(t, err)

	names := make(map[string]bool)
	for _, f := range families {
		names[f.GetName()] = true
	}

	assert.True(t, names["testapp_http_requests_total"])
	assert.True(t, names["testapp_http_request_duration_seconds"])
	assert.True(t, names["testapp_http_requests_in_flight"])
	assert.True(t, names["testapp_db_pool_connections"])
	assert.True(t, names["testapp_db_pool_acquires_total"])
	assert.True(t, names["testapp_db_pool_acquire_duration_seconds"])
	assert.True(t, names["testapp_db_pool_empty_attempts_total"])
}

func TestNewMetrics_InvalidServiceNamePanics(t *testing.T) {
	assert.Panics(t, func() {
		NewMetrics("test-app")
	})
}

func TestRecordHTTPRequest_SetsLabels(t *testing.T) {
	m := NewMetrics("testapp")
	m.RecordHTTPRequest("GET", "/api/test", 200, 100*time.Millisecond)

	families, _ := m.Registry.Gather()
	for _, f := range families {
		if f.GetName() == "testapp_http_requests_total" {
			for _, m := range f.GetMetric() {
				assert.Equal(t, "GET", m.GetLabel()[0].GetValue())
				assert.Equal(t, "/api/test", m.GetLabel()[1].GetValue())
				assert.Equal(t, "2xx", m.GetLabel()[2].GetValue())
			}
		}
	}
}

func TestRecordHTTPRequest_StatusRanges(t *testing.T) {
	m := NewMetrics("testapp")

	tests := []struct {
		status int
		expect string
	}{
		{200, "2xx"},
		{301, "3xx"},
		{404, "4xx"},
		{500, "5xx"},
		{199, "unknown"},
	}

	for _, tt := range tests {
		m.RecordHTTPRequest("GET", "/test", tt.status, time.Second)
	}

	families, _ := m.Registry.Gather()
	for _, f := range families {
		if f.GetName() == "testapp_http_requests_total" {
			assert.Len(t, f.GetMetric(), len(tests))
		}
	}
}

func TestIncDecInFlight(t *testing.T) {
	m := NewMetrics("testapp")

	m.IncInFlight()
	m.IncInFlight()
	m.DecInFlight()

	families, _ := m.Registry.Gather()
	for _, f := range families {
		if f.GetName() == "testapp_http_requests_in_flight" {
			assert.Equal(t, float64(1), f.GetMetric()[0].GetGauge().GetValue())
		}
	}
}

func TestRecordDBPoolStats_FirstCall(t *testing.T) {
	m := NewMetrics("testapp")
	stats := DBPoolStats{
		AcquiredConnections: 5, IdleConnections: 3, ConstructingConnections: 1,
		EmptyAttempts: 2, TotalAcquired: 10, TotalAcquireDuration: time.Second,
	}
	m.RecordDBPoolStats(stats)

	assert.Equal(t, int64(10), m.prevAcquireCount)
	assert.Equal(t, time.Second, m.prevAcquireDuration)
}

func TestRecordDBPoolStats_DeltaAcquires(t *testing.T) {
	m := NewMetrics("testapp")

	m.RecordDBPoolStats(DBPoolStats{TotalAcquired: 10, TotalAcquireDuration: time.Second, EmptyAttempts: 2})
	m.RecordDBPoolStats(DBPoolStats{TotalAcquired: 15, TotalAcquireDuration: 3 * time.Second, EmptyAttempts: 5})

	assert.Equal(t, int64(15), m.prevAcquireCount)
	assert.Equal(t, int64(5), m.prevEmptyAcquireCount)
}

func TestRecordDBPoolStats_NoDelta(t *testing.T) {
	m := NewMetrics("testapp")
	m.RecordDBPoolStats(DBPoolStats{TotalAcquired: 10, EmptyAttempts: 2})
	m.RecordDBPoolStats(DBPoolStats{TotalAcquired: 10, EmptyAttempts: 2})

	assert.Equal(t, int64(10), m.prevAcquireCount)
	assert.Equal(t, int64(2), m.prevEmptyAcquireCount)
}

func TestRecordSystemMetrics_SetsNonZero(t *testing.T) {
	m := NewMetrics("testapp")
	m.RecordSystemMetrics()

	families, _ := m.Registry.Gather()
	for _, f := range families {
		val := f.GetMetric()[0].GetGauge().GetValue()
		switch f.GetName() {
		case "testapp_system_memory_alloc_bytes",
			"testapp_system_memory_sys_bytes",
			"testapp_system_goroutines":
			assert.Greater(t, val, 0.0)
		}
	}
}

func TestStartDBPoolCollector_CallsOnTick(t *testing.T) {
	m := NewMetrics("testapp")
	ctx, cancel := context.WithCancel(context.Background())

	var callCount int
	StartDBPoolCollector(ctx, m, func() DBPoolStats {
		callCount++
		return DBPoolStats{TotalAcquired: 1}
	}, 10*time.Millisecond)

	time.Sleep(25 * time.Millisecond)
	cancel()
	time.Sleep(10 * time.Millisecond)

	assert.GreaterOrEqual(t, callCount, 1)
}

func TestStartDBPoolCollector_StopsOnCancel(t *testing.T) {
	m := NewMetrics("testapp")
	ctx, cancel := context.WithCancel(context.Background())

	var callCount int
	StartDBPoolCollector(ctx, m, func() DBPoolStats {
		callCount++
		return DBPoolStats{}
	}, time.Hour)

	time.Sleep(10 * time.Millisecond)
	cancel()
	time.Sleep(10 * time.Millisecond)

	before := callCount
	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, before, callCount)
}

func TestStatusCodeToString(t *testing.T) {
	tests := []struct {
		code int
		want string
	}{
		{200, "2xx"},
		{299, "2xx"},
		{300, "3xx"},
		{399, "3xx"},
		{400, "4xx"},
		{499, "4xx"},
		{500, "5xx"},
		{599, "5xx"},
		{199, "unknown"},
		{0, "unknown"},
		{-1, "unknown"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, statusCodeToString(tt.code), "statusCodeToString(%d)", tt.code)
	}
}

func TestNewMetrics_DefaultServiceName(t *testing.T) {
	m := NewMetrics("")
	require.NotNil(t, m)
	require.NotNil(t, m.Registry)
}
