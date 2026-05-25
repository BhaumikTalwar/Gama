package config

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type PostgresConfig struct {
	Host     string `mapstructure:"host" yaml:"host" validate:"required"`
	Port     int    `mapstructure:"port" yaml:"port" validate:"required"`
	User     string `mapstructure:"user" yaml:"user" validate:"required"`
	Password string `mapstructure:"password" yaml:"password" validate:"required"`
	Database string `mapstructure:"database" yaml:"database" validate:"required"`

	SSLMode     string `mapstructure:"ssl_mode" yaml:"ssl_mode"`
	SSLCertPath string `mapstructure:"ssl_cert_path" yaml:"ssl_cert_path"`
	SSLKeyPath  string `mapstructure:"ssl_key_path" yaml:"ssl_key_path"`
	SSLCAPath   string `mapstructure:"ssl_ca_path" yaml:"ssl_ca_path"`

	MaxConns int `mapstructure:"max_conns" yaml:"max_conns"`
	MinConns int `mapstructure:"min_conns" yaml:"min_conns"`

	MaxConnIdleTime time.Duration `mapstructure:"max_conn_idle_time" yaml:"max_conn_idle_time"`
	AcquireTimeout  time.Duration `mapstructure:"acquire_timeout" yaml:"acquire_timeout"`

	ConnectTimeout   time.Duration `mapstructure:"connect_timeout" yaml:"connect_timeout"`
	StatementTimeout time.Duration `mapstructure:"statement_timeout" yaml:"statement_timeout"`
	IdleTxnTimeout   time.Duration `mapstructure:"idle_txn_timeout" yaml:"idle_txn_timeout"`

	KeepAliveInterval  time.Duration `mapstructure:"keep_alive_interval" yaml:"keep_alive_interval"`
	SlowQueryThreshold time.Duration `mapstructure:"slow_query_threshold" yaml:"slow_query_threshold"`
	KeepAlive          bool          `mapstructure:"keep_alive" yaml:"keep_alive"`
}

func setPostgressDefaults(prefix string) {
	viper.SetDefault(prefix+".host", "localhost")
	viper.SetDefault(prefix+".port", 5432)
	viper.SetDefault(prefix+".ssl_mode", "disable")

	viper.SetDefault(prefix+".max_conns", 10)
	viper.SetDefault(prefix+".min_conns", 2)
	viper.SetDefault(prefix+".max_conn_idle_time", "15m")
	viper.SetDefault(prefix+".acquire_timeout", "30s")

	viper.SetDefault(prefix+".connect_timeout", "10s")
	viper.SetDefault(prefix+".statement_timeout", "0")
	viper.SetDefault(prefix+".idle_txn_timeout", "1m")

	viper.SetDefault(prefix+".keep_alive", true)
	viper.SetDefault(prefix+".keep_alive_interval", "1m")
	viper.SetDefault(prefix+".slow_query_threshold", "1m")
}

func registerPostgresFlags(cmd *cobra.Command, prefix string) {
	hostFlag := fmt.Sprintf("%s-host", prefix)
	portFlag := fmt.Sprintf("%s-port", prefix)
	userFlag := fmt.Sprintf("%s-user", prefix)
	passFlag := fmt.Sprintf("%s-password", prefix)
	dbFlag := fmt.Sprintf("%s-database", prefix)

	sslModeFlag := fmt.Sprintf("%s-ssl-mode", prefix)
	sslCertFlag := fmt.Sprintf("%s-ssl-cert-path", prefix)
	sslKeyFlag := fmt.Sprintf("%s-ssl-key-path", prefix)
	sslCAFlag := fmt.Sprintf("%s-ssl-ca-path", prefix)

	poolMaxFlag := fmt.Sprintf("%s-pool-max", prefix)
	poolMinFlag := fmt.Sprintf("%s-pool-min", prefix)
	poolIdleFlag := fmt.Sprintf("%s-pool-idle", prefix)
	poolAcqFlag := fmt.Sprintf("%s-pool-acquire", prefix)

	connTimeFlag := fmt.Sprintf("%s-connect-timeout", prefix)
	stmtTimeFlag := fmt.Sprintf("%s-statement-timeout", prefix)
	idleTxnFlag := fmt.Sprintf("%s-idle-txn-timeout", prefix)

	kaIntFlag := fmt.Sprintf("%s-keep-alive-interval", prefix)
	slowQFlag := fmt.Sprintf("%s-slow-query-threshold", prefix)
	kaBoolFlag := fmt.Sprintf("%s-keep-alive", prefix)

	cmd.Flags().String(hostFlag, "", "Postgres Host")
	cmd.Flags().Int(portFlag, 0, "Postgres Port")
	cmd.Flags().String(userFlag, "", "Postgres User")
	cmd.Flags().String(passFlag, "", "Postgres Password")
	cmd.Flags().String(dbFlag, "", "Postgres Database Name")

	cmd.Flags().String(sslModeFlag, "", "SSL Mode (disable, require, verify-full)")
	cmd.Flags().String(sslCertFlag, "", "Path to SSL Cert")
	cmd.Flags().String(sslKeyFlag, "", "Path to SSL Key")
	cmd.Flags().String(sslCAFlag, "", "Path to SSL CA")

	cmd.Flags().Int(poolMaxFlag, 0, "Max connections in pool")
	cmd.Flags().Int(poolMinFlag, 0, "Min connections in pool")
	cmd.Flags().String(poolIdleFlag, "", "Max idle time for connection")
	cmd.Flags().String(poolAcqFlag, "", "Max wait time to acquire connection")

	cmd.Flags().String(connTimeFlag, "", "Database connection timeout")
	cmd.Flags().String(stmtTimeFlag, "", "Database statement timeout")
	cmd.Flags().String(idleTxnFlag, "", "Idle transaction timeout")

	cmd.Flags().String(kaIntFlag, "", "Keep Alive Interval")
	cmd.Flags().String(slowQFlag, "", "Slow Query Threshold log")
	cmd.Flags().Bool(kaBoolFlag, false, "Enable Keep Alive")

	_ = viper.BindPFlag(prefix+".host", cmd.Flags().Lookup(hostFlag))
	_ = viper.BindPFlag(prefix+".port", cmd.Flags().Lookup(portFlag))
	_ = viper.BindPFlag(prefix+".user", cmd.Flags().Lookup(userFlag))
	_ = viper.BindPFlag(prefix+".password", cmd.Flags().Lookup(passFlag))
	_ = viper.BindPFlag(prefix+".database", cmd.Flags().Lookup(dbFlag))

	_ = viper.BindPFlag(prefix+".ssl_mode", cmd.Flags().Lookup(sslModeFlag))
	_ = viper.BindPFlag(prefix+".ssl_cert_path", cmd.Flags().Lookup(sslCertFlag))
	_ = viper.BindPFlag(prefix+".ssl_key_path", cmd.Flags().Lookup(sslKeyFlag))
	_ = viper.BindPFlag(prefix+".ssl_ca_path", cmd.Flags().Lookup(sslCAFlag))

	_ = viper.BindPFlag(prefix+".pool_max", cmd.Flags().Lookup(poolMaxFlag))
	_ = viper.BindPFlag(prefix+".pool_min", cmd.Flags().Lookup(poolMinFlag))
	_ = viper.BindPFlag(prefix+".pool_idle", cmd.Flags().Lookup(poolIdleFlag))
	_ = viper.BindPFlag(prefix+".pool_acquire", cmd.Flags().Lookup(poolAcqFlag))

	_ = viper.BindPFlag(prefix+".connect_timeout", cmd.Flags().Lookup(connTimeFlag))
	_ = viper.BindPFlag(prefix+".statement_timeout", cmd.Flags().Lookup(stmtTimeFlag))
	_ = viper.BindPFlag(prefix+".idle_txn_timeout", cmd.Flags().Lookup(idleTxnFlag))

	_ = viper.BindPFlag(prefix+".keep_alive_interval", cmd.Flags().Lookup(kaIntFlag))
	_ = viper.BindPFlag(prefix+".slow_query_threshold", cmd.Flags().Lookup(slowQFlag))
	_ = viper.BindPFlag(prefix+".keep_alive", cmd.Flags().Lookup(kaBoolFlag))
}

func (c *PostgresConfig) validate() error {
	if c.Port < 0 || c.Port > 65535 {
		return fmt.Errorf("Invalid Postgres Port Number: %v", c.Port)
	}

	return nil
}
