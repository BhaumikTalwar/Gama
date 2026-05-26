package config

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type TelemetryConfig struct {
	Enabled bool `mapstructure:"enabled" yaml:"enabled"`

	ServiceName string `mapstructure:"service_name" yaml:"service_name"`
	Environment string `mapstructure:"environment" yaml:"environment"`

	OTLPEndpoint     string  `mapstructure:"otlp_endpoint" yaml:"otlp_endpoint"`
	OTLPEndpointHTTP string  `mapstructure:"otlp_endpoint_http" yaml:"otlp_endpoint_http"`
	TraceSampleRate  float64 `mapstructure:"trace_sample_rate" yaml:"trace_sample_rate"`

	PrometheusPath string `mapstructure:"prometheus_path" yaml:"prometheus_path"`

	EnableHTTPMetrics   bool `mapstructure:"enable_http_metrics" yaml:"enable_http_metrics"`
	EnableDBMetrics     bool `mapstructure:"enable_db_metrics" yaml:"enable_db_metrics"`
	EnableCacheMetrics  bool `mapstructure:"enable_cache_metrics" yaml:"enable_cache_metrics"`
	EnableSystemMetrics bool `mapstructure:"enable_system_metrics" yaml:"enable_system_metrics"`

	ExportTimeout time.Duration `mapstructure:"export_timeout" yaml:"export_timeout"`
}

func setTelemetryDefaults(prefix string) {
	viper.SetDefault(prefix+".enabled", false)
	viper.SetDefault(prefix+".service_name", "gama-api")
	viper.SetDefault(prefix+".environment", "dev")
	viper.SetDefault(prefix+".otlp_endpoint", "localhost:4317")
	viper.SetDefault(prefix+".otlp_endpoint_http", "localhost:4318")
	viper.SetDefault(prefix+".trace_sample_rate", 1.0)
	viper.SetDefault(prefix+".prometheus_path", "/metrics")
	viper.SetDefault(prefix+".enable_http_metrics", true)
	viper.SetDefault(prefix+".enable_db_metrics", true)
	viper.SetDefault(prefix+".enable_cache_metrics", true)
	viper.SetDefault(prefix+".enable_system_metrics", true)
	viper.SetDefault(prefix+".export_timeout", "10s")
}

func registerTelemetryFlags(cmd *cobra.Command, prefix string) {
	enabledFlag := fmt.Sprintf("%s-enabled", prefix)
	serviceFlag := fmt.Sprintf("%s-service-name", prefix)
	envFlag := fmt.Sprintf("%s-environment", prefix)
	otlpFlag := fmt.Sprintf("%s-otlp-endpoint", prefix)
	otlpHTTPFlag := fmt.Sprintf("%s-otlp-endpoint-http", prefix)
	sampleRateFlag := fmt.Sprintf("%s-trace-sample-rate", prefix)
	promPathFlag := fmt.Sprintf("%s-prometheus-path", prefix)
	httpMetricsFlag := fmt.Sprintf("%s-enable-http-metrics", prefix)
	dbMetricsFlag := fmt.Sprintf("%s-enable-db-metrics", prefix)
	cacheMetricsFlag := fmt.Sprintf("%s-enable-cache-metrics", prefix)
	sysMetricsFlag := fmt.Sprintf("%s-enable-system-metrics", prefix)
	exportTimeoutFlag := fmt.Sprintf("%s-export-timeout", prefix)

	cmd.Flags().Bool(enabledFlag, false, "Enable telemetry (OTLP tracing + Prometheus metrics)")
	cmd.Flags().String(serviceFlag, "", "Service name for telemetry")
	cmd.Flags().String(envFlag, "", "Environment name (dev/prod)")
	cmd.Flags().String(otlpFlag, "", "OTLP gRPC endpoint for tracing")
	cmd.Flags().String(otlpHTTPFlag, "", "OTLP HTTP endpoint for tracing")
	cmd.Flags().Float64(sampleRateFlag, 1.0, "Trace sampling rate (0.0-1.0)")
	cmd.Flags().String(promPathFlag, "", "Prometheus metrics path")
	cmd.Flags().Bool(httpMetricsFlag, true, "Enable HTTP metrics")
	cmd.Flags().Bool(dbMetricsFlag, true, "Enable database metrics")
	cmd.Flags().Bool(cacheMetricsFlag, true, "Enable cache metrics")
	cmd.Flags().Bool(sysMetricsFlag, true, "Enable system metrics")
	cmd.Flags().String(exportTimeoutFlag, "", "Export timeout for telemetry")

	_ = viper.BindPFlag(prefix+".enabled", cmd.Flags().Lookup(enabledFlag))
	_ = viper.BindPFlag(prefix+".service_name", cmd.Flags().Lookup(serviceFlag))
	_ = viper.BindPFlag(prefix+".environment", cmd.Flags().Lookup(envFlag))
	_ = viper.BindPFlag(prefix+".otlp_endpoint", cmd.Flags().Lookup(otlpFlag))
	_ = viper.BindPFlag(prefix+".otlp_endpoint_http", cmd.Flags().Lookup(otlpHTTPFlag))
	_ = viper.BindPFlag(prefix+".trace_sample_rate", cmd.Flags().Lookup(sampleRateFlag))
	_ = viper.BindPFlag(prefix+".prometheus_path", cmd.Flags().Lookup(promPathFlag))
	_ = viper.BindPFlag(prefix+".enable_http_metrics", cmd.Flags().Lookup(httpMetricsFlag))
	_ = viper.BindPFlag(prefix+".enable_db_metrics", cmd.Flags().Lookup(dbMetricsFlag))
	_ = viper.BindPFlag(prefix+".enable_cache_metrics", cmd.Flags().Lookup(cacheMetricsFlag))
	_ = viper.BindPFlag(prefix+".enable_system_metrics", cmd.Flags().Lookup(sysMetricsFlag))
	_ = viper.BindPFlag(prefix+".export_timeout", cmd.Flags().Lookup(exportTimeoutFlag))
}

func (c *TelemetryConfig) validate() error {
	if c.TraceSampleRate < 0 || c.TraceSampleRate > 1 {
		return fmt.Errorf("trace_sample_rate must be between 0 and 1, got %f", c.TraceSampleRate)
	}
	return nil
}
