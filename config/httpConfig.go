package config

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	DefaultHTTPHost              = "0.0.0.0"
	DefaultHTTPPort              = "8080"
	DefaultHTTPReadTimeout       = "10s"
	DefaultHTTPReadHeaderTimeout = "5s"
	DefaultHTTPWriteTimeout      = "15s"
	DefaultHTTPIdleTimeout       = "120s"
	DefaultHTTPMaxHeaderBytes    = 1 << 20 // 1MB
	DefaultHTTPShutdownTimeout   = "10s"
	DefaultHTTPKeepAlive         = true
)

type HTTPServerConfig struct {
	Host              string        `mapstructure:"host" yaml:"host"`
	Port              string        `mapstructure:"port" yaml:"port"`
	ReadTimeout       time.Duration `mapstructure:"read_timeout" yaml:"read_timeout"`
	ReadHeaderTimeout time.Duration `mapstructure:"read_header_timeout" yaml:"read_header_timeout"`
	WriteTimeout      time.Duration `mapstructure:"write_timeout" yaml:"write_timeout"`
	IdleTimeout       time.Duration `mapstructure:"idle_timeout" yaml:"idle_timeout"`

	MaxHeaderBytes  int           `mapstructure:"max_header_bytes" yaml:"max_header_bytes"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout" yaml:"shutdown_timeout"`
	KeepAlive       bool          `mapstructure:"keep_alive" yaml:"keep_alive"`
}

func setHTTPServerDefaults(prefix string) {
	viper.SetDefault(prefix+".host", DefaultHTTPHost)
	viper.SetDefault(prefix+".port", DefaultHTTPPort)
	viper.SetDefault(prefix+".read_timeout", DefaultHTTPReadTimeout)
	viper.SetDefault(prefix+".read_header_timeout", DefaultHTTPReadHeaderTimeout)
	viper.SetDefault(prefix+".write_timeout", DefaultHTTPWriteTimeout)
	viper.SetDefault(prefix+".idle_timeout", DefaultHTTPIdleTimeout)

	viper.SetDefault(prefix+".max_header_bytes", DefaultHTTPMaxHeaderBytes)
	viper.SetDefault(prefix+".shutdown_timeout", DefaultHTTPShutdownTimeout)
	viper.SetDefault(prefix+".keep_alive", DefaultHTTPKeepAlive)
}

func registerHTTPServerFlags(cmd *cobra.Command, prefix string) {
	hostFlag := fmt.Sprintf("%s-host", prefix)
	portFlag := fmt.Sprintf("%s-port", prefix)
	readTimeoutFlag := fmt.Sprintf("%s-read-timeout", prefix)
	readHeaderTimeoutFlag := fmt.Sprintf("%s-read-header-timeout", prefix)
	writeTimeoutFlag := fmt.Sprintf("%s-write-timeout", prefix)
	idleTimeoutFlag := fmt.Sprintf("%s-idle-timeout", prefix)

	maxHeaderFlag := fmt.Sprintf("%s-max-header-bytes", prefix)
	shutdownTimeoutFlag := fmt.Sprintf("%s-shutdown-timeout", prefix)
	keepAliveFlag := fmt.Sprintf("%s-keep-alive", prefix)

	cmd.Flags().String(hostFlag, "", "HTTP server Host")
	cmd.Flags().String(portFlag, "", "HTTP server port")
	cmd.Flags().String(readTimeoutFlag, "", "HTTP server read timeout")
	cmd.Flags().String(readHeaderTimeoutFlag, "", "HTTP server read header timeout")
	cmd.Flags().String(writeTimeoutFlag, "", "HTTP server write timeout")
	cmd.Flags().String(idleTimeoutFlag, "", "HTTP server idle timeout")

	cmd.Flags().Int(maxHeaderFlag, 0, "Maximum size of request headers in bytes")
	cmd.Flags().String(shutdownTimeoutFlag, "", "Graceful shutdown timeout")
	cmd.Flags().Bool(keepAliveFlag, DefaultHTTPKeepAlive, "Enable HTTP keep-alive")

	_ = viper.BindPFlag(prefix+".host", cmd.Flags().Lookup(hostFlag))
	_ = viper.BindPFlag(prefix+".port", cmd.Flags().Lookup(portFlag))
	_ = viper.BindPFlag(prefix+".read_timeout", cmd.Flags().Lookup(readTimeoutFlag))
	_ = viper.BindPFlag(prefix+".read_header_timeout", cmd.Flags().Lookup(readHeaderTimeoutFlag))
	_ = viper.BindPFlag(prefix+".write_timeout", cmd.Flags().Lookup(writeTimeoutFlag))
	_ = viper.BindPFlag(prefix+".idle_timeout", cmd.Flags().Lookup(idleTimeoutFlag))

	_ = viper.BindPFlag(prefix+".max_header_bytes", cmd.Flags().Lookup(maxHeaderFlag))
	_ = viper.BindPFlag(prefix+".shutdown_timeout", cmd.Flags().Lookup(shutdownTimeoutFlag))
	_ = viper.BindPFlag(prefix+".keep_alive", cmd.Flags().Lookup(keepAliveFlag))
}

func (c *HTTPServerConfig) validate() error {
	if c.Host == "" {
		return errors.New("http.host cannot be empty")
	}

	port, err := strconv.Atoi(c.Port)
	if err != nil {
		return fmt.Errorf("invalid http.port: %w", err)
	}

	if port < 1 || port > 65535 {
		return fmt.Errorf("invalid http.port: %d", port)
	}

	if c.ReadTimeout <= 0 {
		return errors.New("http.read_timeout must be greater than 0")
	}

	if c.ReadHeaderTimeout <= 0 {
		return errors.New("http.read_header_timeout must be greater than 0")
	}

	if c.WriteTimeout <= 0 {
		return errors.New("http.write_timeout must be greater than 0")
	}

	if c.IdleTimeout <= 0 {
		return errors.New("http.idle_timeout must be greater than 0")
	}

	if c.ShutdownTimeout <= 0 {
		return errors.New("http.shutdown_timeout must be greater than 0")
	}

	if c.MaxHeaderBytes <= 0 {
		return errors.New("http.max_header_bytes must be greater than 0")
	}

	if c.MaxHeaderBytes > 10<<20 {
		return fmt.Errorf("http.max_header_bytes too large: %d (max 10MB)", c.MaxHeaderBytes)
	}

	return nil
}
