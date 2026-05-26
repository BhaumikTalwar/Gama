package config

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/BhaumikTalwar/Gama/utils"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.yaml.in/yaml/v3"
)

var cfg Config

func SetTestAppConfig(testCfg *AppConfig) {
	cfg.App = *testCfg
}

const (
	DefaultConfigFileName = "config"
	DefaultConfigFilePath = "$PWD/"
	DefaultConfigFileType = "yaml"
)

type DatabaseDriver string

const (
	DBSqlite   DatabaseDriver = "sqlite"
	DBPostgres DatabaseDriver = "postgres"
)

type Config struct {
	App  AppConfig        `mapstructure:"app"`
	Http HTTPServerConfig `mapstructure:"http"`

	Database  DatabaseDriver  `mapstructure:"database"`
	Logging   LoggingConfig   `mapstructure:"logging"`
	Telemetry TelemetryConfig `mapstructure:"telemetry"`

	Sqlite   SqliteConfig   `mapstructure:"sqlite"`
	Postgres PostgresConfig `mapstructure:"postgres"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Cache    CacheConfig    `mapstructure:"cache"`
	S3       S3Config       `mapstructure:"s3"`
	OTP      OTPConfig      `mapstructure:"otp"`
}

func SetDefaults() {
	setAppDefaults("app")
	setHTTPServerDefaults("http")
	setLogDefaults("logging")
	setTelemetryDefaults("telemetry")
	setPostgressDefaults("postgres")
	setSqliteDefaults("sqlite")
	setRedisDefaults("redis")
	setCacheDefaults("cache")
	setS3Defaults("s3")
	setOTPDefaults("otp")

	viper.SetDefault("database", "sqlite")
}

func RegisterServeFlags(cmd *cobra.Command) {
	registerAppFlags(cmd, "app")
	registerHTTPServerFlags(cmd, "http")
	registerLogFlags(cmd, "logging")
	registerTelemetryFlags(cmd, "telemetry")
	registerSqliteFlags(cmd, "sqlite")
	registerPostgresFlags(cmd, "postgres")
	registerRedisFlags(cmd, "redis")
	registerCacheFlags(cmd, "cache")
	registerS3Flags(cmd, "s3")
	registerOTPFlags(cmd, "otp")

	cmd.Flags().String("database", "", "Database driver to use (sqlite|postgres)")
	_ = viper.BindPFlag("database", cmd.Flags().Lookup("database"))
}

func LoadConfig(cfgFile string) error {
	if err := godotenv.Load(); err != nil {
		log.Print(err)
	}

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath("")
		viper.AddConfigPath(DefaultConfigFilePath)
		viper.SetConfigName(DefaultConfigFileName)
		viper.SetConfigType(DefaultConfigFileType)
	}

	viper.SetEnvPrefix("APP")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("Config file was found but unreadable: %w", err)
		}

		log.Println("No config file found, using defaults and environment variables")
	}

	err := viper.Unmarshal(&cfg)
	if err != nil {
		return fmt.Errorf("Unable to decode into struct, %w", err)
	}

	err = cfg.Validate()
	if err != nil {
		return fmt.Errorf("Config Validation Issue: %w", err)
	}

	return nil
}

func (c *Config) Validate() error {
	if err := c.App.validate(); err != nil {
		return fmt.Errorf("Error in App Config: %w", err)
	}

	if err := c.Http.validate(); err != nil {
		return fmt.Errorf("Error in HTTP Config: %w", err)
	}

	if err := c.Telemetry.validate(); err != nil {
		return fmt.Errorf("Error in Telemetry Config: %w", err)
	}

	if err := c.Postgres.validate(); err != nil {
		return fmt.Errorf("Error in Postgres Config: %w", err)
	}

	if err := c.Sqlite.validate(); err != nil {
		return fmt.Errorf("Error in Sqlite Config: %w", err)
	}

	if err := c.Redis.validate(); err != nil {
		return fmt.Errorf("Error in Redis Config: %w", err)
	}

	if err := c.Logging.validate(); err != nil {
		return fmt.Errorf("Error in Logging Config: %w", err)
	}

	if err := c.Cache.validate(); err != nil {
		return fmt.Errorf("Error in Cache Config: %w", err)
	}

	if err := c.S3.validate(); err != nil {
		return fmt.Errorf("Error in S3 Config: %w", err)
	}

	if err := c.OTP.validate(); err != nil {
		return fmt.Errorf("Error in OTP Config: %w", err)
	}

	switch c.Database {
	case DBSqlite:
		if c.Sqlite.Path == "" {
			return errors.New("Sqlite selected but sqlite.path is empty")
		}
	case DBPostgres:
		if c.Postgres.Host == "" || c.Postgres.Database == "" {
			return errors.New("Postgres selected but postgres config is incomplete")
		}
	default:
		return fmt.Errorf("Unknown database driver: %s", c.Database)
	}

	return nil
}

func ExampleConfig() error {
	SetDefaults()

	var c Config

	err := viper.Unmarshal(&c)
	if err != nil {
		return fmt.Errorf("UnMarshalling Error Unable to decode into struct, %w", err)
	}

	c.App.AppName = "Gama"
	c.App.AppSecret = utils.GetRandomKeyB64(32)
	c.App.JWTKey = utils.GetRandomKeyB64(32)
	c.App.AESKey = utils.GetRandomKeyB64(32)

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("Cant Marshal Into Yaml: %w", err)
	}

	configFilePath := DefaultConfigFileName + "." + DefaultConfigFileType
	if utils.FileExistsInPath(configFilePath) {
		return fmt.Errorf("Config File already Exist at: %s", configFilePath)
	}

	err = os.WriteFile(configFilePath, data, 0o644)
	if err != nil {
		return err
	}

	return nil
}

func VerifyConfig(cfgFile string) error {
	if err := godotenv.Load(); err != nil {
		log.Print(err)
	}

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath("")
		viper.AddConfigPath(DefaultConfigFilePath)
		viper.SetConfigName(DefaultConfigFileName)
		viper.SetConfigType(DefaultConfigFileType)
	}

	viper.SetEnvPrefix("APP")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("Config file was not found or unreadable: %w", err)
	}

	var c Config
	err := viper.Unmarshal(&c)
	if err != nil {
		return fmt.Errorf("Unable to decode into struct, %w", err)
	}

	err = c.Validate()
	if err != nil {
		return fmt.Errorf("Config Validation Issue: %w", err)
	}

	return nil
}

func GetAppConfig() *AppConfig {
	return &cfg.App
}

func GetHttpServerConfig() *HTTPServerConfig {
	return &cfg.Http
}

func GetRedisConfig() *RedisConfig {
	return &cfg.Redis
}

func GetPostgresConfig() *PostgresConfig {
	return &cfg.Postgres
}

func GetLogConfig() *LoggingConfig {
	return &cfg.Logging
}

func GetSqliteConfig() *SqliteConfig {
	return &cfg.Sqlite
}

func GetCacheConfig() *CacheConfig {
	return &cfg.Cache
}

func GetS3Config() *S3Config {
	return &cfg.S3
}

func GetOTPConfig() *OTPConfig {
	return &cfg.OTP
}

func GetTelemetryConfig() *TelemetryConfig {
	return &cfg.Telemetry
}

func GetDatabaseDriver() *DatabaseDriver {
	return &cfg.Database
}

func DumpConfig() string {
	tmp := cfg

	redacttext := "<REDACTED>"
	tmp.App.AppSecret = redacttext
	tmp.App.JWTKey = redacttext
	tmp.App.AESKey = redacttext
	tmp.Postgres.Password = redacttext
	tmp.Redis.RedisPassword = redacttext
	tmp.OTP.APIKey = redacttext
	tmp.S3.AccessKey = redacttext
	tmp.S3.SecretKey = redacttext

	return fmt.Sprintf("%+v\n", tmp)
}
