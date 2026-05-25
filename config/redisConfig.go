package config

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type RedisConfig struct {
	RedisHost string `mapstructure:"host" yaml:"host" validate:"required"`

	RedisPassword string `mapstructure:"password" yaml:"password"`
	RedisPort     int    `mapstructure:"port" yaml:"port" validate:"required"`
	RedisDB       int    `mapstructure:"db" yaml:"db"`

	RedisPoolSize     int  `mapstructure:"pool_size" yaml:"pool_size"`
	RedisMinIdleConns int  `mapstructure:"min_idle_conns" yaml:"min_idle_conns"`
	RedisUseTLS       bool `mapstructure:"use_tls" yaml:"use_tls"`

	RedisDialTimeout  time.Duration `mapstructure:"dial_timeout" yaml:"dial_timeout"`
	RedisReadTimeout  time.Duration `mapstructure:"read_timeout" yaml:"read_timeout"`
	RedisWriteTimeout time.Duration `mapstructure:"write_timeout" yaml:"write_timeout"`

	RedisTLSCertFile string `mapstructure:"tls_cert_file" yaml:"tls_cert_file"`
	RedisTLSKeyFile  string `mapstructure:"tls_key_file" yaml:"tls_key_file"`
	RedisTLSCAFile   string `mapstructure:"tls_ca_file" yaml:"tls_ca_file"`
}

func setRedisDefaults(prefix string) {
	viper.SetDefault(prefix+".host", "localhost")
	viper.SetDefault(prefix+".port", 6379)
	viper.SetDefault(prefix+".db", 0)
	viper.SetDefault(prefix+".pool_size", 10)
	viper.SetDefault(prefix+".min_idle_conns", 2)
	viper.SetDefault(prefix+".dial_timeout", "5s")
	viper.SetDefault(prefix+".read_timeout", "3s")
	viper.SetDefault(prefix+".write_timeout", "3s")
	viper.SetDefault(prefix+".use_tls", false)
}

func registerRedisFlags(cmd *cobra.Command, prefix string) {
	hostFlag := fmt.Sprintf("%s-host", prefix)
	portFlag := fmt.Sprintf("%s-port", prefix)
	dbFlag := fmt.Sprintf("%s-db", prefix)
	poolFlag := fmt.Sprintf("%s-pool-size", prefix)
	minConnsFlag := fmt.Sprintf("%s-min-idle-conns", prefix)
	dialTimeFlag := fmt.Sprintf("%s-dial-timeout", prefix)
	readTimeFlag := fmt.Sprintf("%s-read-timeout", prefix)
	writeTimeFlag := fmt.Sprintf("%s-write-timeout", prefix)
	usetlsFlag := fmt.Sprintf("%s-use-tls", prefix)

	tlsFileCert := fmt.Sprintf("%s-tls-cert-file", prefix)
	tlsFileKey := fmt.Sprintf("%s-tls-key-file", prefix)
	tlsFileCA := fmt.Sprintf("%s-tls-ca-file", prefix)

	cmd.Flags().String(hostFlag, "", "Redis host")
	cmd.Flags().Int(portFlag, 0, "Redis port")
	cmd.Flags().Int(dbFlag, 0, "Redis DB Index")
	cmd.Flags().Int(poolFlag, 0, "Redis DB Pool Size")
	cmd.Flags().Int(minConnsFlag, 0, "Redis Min Idle Connections")
	cmd.Flags().String(dialTimeFlag, "", "Redis Dial Timeout")
	cmd.Flags().String(readTimeFlag, "", "Redis Read Timeout")
	cmd.Flags().String(writeTimeFlag, "", "Redis Write Timeout")
	cmd.Flags().Bool(usetlsFlag, false, "Weather to use TLS or not")
	cmd.Flags().String(tlsFileCert, "", "TLS Certificate File for redis")
	cmd.Flags().String(tlsFileKey, "", "TLS Key File for redis")
	cmd.Flags().String(tlsFileCA, "", "TLS CA File for redis")

	_ = viper.BindPFlag(prefix+".host", cmd.Flags().Lookup(hostFlag))
	_ = viper.BindPFlag(prefix+".port", cmd.Flags().Lookup(portFlag))
	_ = viper.BindPFlag(prefix+".db", cmd.Flags().Lookup(dbFlag))
	_ = viper.BindPFlag(prefix+".pool_size", cmd.Flags().Lookup(poolFlag))
	_ = viper.BindPFlag(prefix+".min_idle_conns", cmd.Flags().Lookup(minConnsFlag))
	_ = viper.BindPFlag(prefix+".dial_timeout", cmd.Flags().Lookup(dialTimeFlag))
	_ = viper.BindPFlag(prefix+".read_timeout", cmd.Flags().Lookup(readTimeFlag))
	_ = viper.BindPFlag(prefix+".write_timeout", cmd.Flags().Lookup(writeTimeFlag))
	_ = viper.BindPFlag(prefix+".use_tls", cmd.Flags().Lookup(usetlsFlag))
	_ = viper.BindPFlag(prefix+".tls_cer_file", cmd.Flags().Lookup(tlsFileCert))
	_ = viper.BindPFlag(prefix+".tls_key_file", cmd.Flags().Lookup(tlsFileKey))
	_ = viper.BindPFlag(prefix+".tls_ca_file", cmd.Flags().Lookup(tlsFileCA))
}

func (c *RedisConfig) validate() error {
	if c.RedisPort < 0 || c.RedisPort > 65535 {
		return fmt.Errorf("Invalid Redis Port Number: %v", c.RedisPort)
	}

	return nil
}
