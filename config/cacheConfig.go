package config

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	DefaultCacheSizeMB   = 400
	DefaultCacheLocalTTL = 5
)

type CacheConfig struct {
	CacheNamespace string `mapstructure:"namespace" yaml:"namespace"`

	EnableLocalCache bool  `mapstructure:"enable_local" yaml:"enable_local"`
	LocalCacheSize   int64 `mapstructure:"local_size" yaml:"local_size"`
	LocalCacheTTL    int64 `mapstructure:"local_ttl" yaml:"local_ttl"`

	Codec string `mapstructure:"codec" yaml:"codec"`
}

func setCacheDefaults(prefix string) {
	viper.SetDefault(prefix+".namespace", "app_cache")
	viper.SetDefault(prefix+".enable_local", true)
	viper.SetDefault(prefix+".local_size", DefaultCacheSizeMB)
	viper.SetDefault(prefix+".local_ttl", DefaultCacheLocalTTL)
	viper.SetDefault(prefix+".codec", "json")
}

func registerCacheFlags(cmd *cobra.Command, prefix string) {
	nsFlag := fmt.Sprintf("%s-namespace", prefix)
	enableFlag := fmt.Sprintf("%s-enable-local", prefix)
	sizeFlag := fmt.Sprintf("%s-local-size", prefix)
	ttlFlag := fmt.Sprintf("%s-local-ttl", prefix)
	codecFlag := fmt.Sprintf("%s-codec", prefix)

	cmd.Flags().String(nsFlag, "", "Cache namespace prefix")
	cmd.Flags().Bool(enableFlag, true, "Enable local in-memory cache")
	cmd.Flags().Int64(sizeFlag, 0, "Local cache size in MB")
	cmd.Flags().Int64(ttlFlag, 0, "Local cache ttl in Minutes")
	cmd.Flags().String(codecFlag, "", "Cache codec (json or msgpack)")

	_ = viper.BindPFlag(prefix+".namespace", cmd.Flags().Lookup(nsFlag))
	_ = viper.BindPFlag(prefix+".enable_local", cmd.Flags().Lookup(enableFlag))
	_ = viper.BindPFlag(prefix+".local_size", cmd.Flags().Lookup(sizeFlag))
	_ = viper.BindPFlag(prefix+".local_ttl", cmd.Flags().Lookup(ttlFlag))
	_ = viper.BindPFlag(prefix+".codec", cmd.Flags().Lookup(codecFlag))
}

func (c *CacheConfig) validate() error {
	if strings.TrimSpace(c.CacheNamespace) == "" {
		return errors.New("cache.namespace must not be empty")
	}

	if c.EnableLocalCache {
		if c.LocalCacheSize <= 0 {
			return errors.New("cache.local_size must be greater than 0 when local cache is enabled")
		}

		if c.LocalCacheSize < 32 {
			return errors.New("cache.local_size too small (minimum 32MB)")
		}

		if c.LocalCacheTTL <= 0 {
			return errors.New("cache.local_ttl too small (minimum 1min)")
		}
	}

	codec := strings.ToLower(c.Codec)
	if codec != "json" && codec != "msgpack" {
		return fmt.Errorf("invalid cache.codec: %s (allowed: json, msgpack)", c.Codec)
	}

	return nil
}
