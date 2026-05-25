package config

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type S3Config struct {
	Endpoint        string `mapstructure:"endpoint" yaml:"endpoint" validate:"required"`
	Region          string `mapstructure:"region" yaml:"region" validate:"required"`
	AccessKey       string `mapstructure:"access_key" yaml:"access_key" validate:"required"`
	SecretKey       string `mapstructure:"secret_key" yaml:"secret_key" validate:"required"`
	UseSSL          bool   `mapstructure:"use_ssl" yaml:"use_ssl"`
	PresignedURLTTL int    `mapstructure:"presigned_url_ttl" yaml:"presigned_url_ttl"`
	PublicBucket    string `mapstructure:"public_bucket" yaml:"public_bucket" validate:"required"`
	PrivateBucket   string `mapstructure:"private_bucket" yaml:"private_bucket" validate:"required"`
	PublicBaseURL   string `mapstructure:"public_base_url" yaml:"public_base_url"`
}

func setS3Defaults(prefix string) {
	viper.SetDefault(prefix+".endpoint", "http://localhost:9000")
	viper.SetDefault(prefix+".region", "us-east-1")
	viper.SetDefault(prefix+".use_ssl", false)
	viper.SetDefault(prefix+".presigned_url_ttl", 3600)
	viper.SetDefault(prefix+".public_bucket", "gama-public")
	viper.SetDefault(prefix+".private_bucket", "gama-private")
	viper.SetDefault(prefix+".public_base_url", "")
}

func registerS3Flags(cmd *cobra.Command, prefix string) {
	endpointFlag := fmt.Sprintf("%s-endpoint", prefix)
	regionFlag := fmt.Sprintf("%s-region", prefix)
	accessKeyFlag := fmt.Sprintf("%s-access-key", prefix)
	secretKeyFlag := fmt.Sprintf("%s-secret-key", prefix)
	useSSLFlag := fmt.Sprintf("%s-use-ssl", prefix)
	ttlFlag := fmt.Sprintf("%s-presigned-url-ttl", prefix)
	publicBucketFlag := fmt.Sprintf("%s-public-bucket", prefix)
	privateBucketFlag := fmt.Sprintf("%s-private-bucket", prefix)
	publicBaseURLFlag := fmt.Sprintf("%s-public-base-url", prefix)

	cmd.Flags().String(endpointFlag, "", "S3 endpoint URL")
	cmd.Flags().String(regionFlag, "", "S3 region")
	cmd.Flags().String(accessKeyFlag, "", "S3 access key")
	cmd.Flags().String(secretKeyFlag, "", "S3 secret key")
	cmd.Flags().Bool(useSSLFlag, false, "Use SSL for S3")
	cmd.Flags().Int(ttlFlag, 0, "Presigned URL TTL in seconds")
	cmd.Flags().String(publicBucketFlag, "", "Public bucket name (product images, banners)")
	cmd.Flags().String(privateBucketFlag, "", "Private bucket name (invoices, sensitive data)")
	cmd.Flags().String(publicBaseURLFlag, "", "Public base URL for CDN")

	_ = viper.BindPFlag(prefix+".endpoint", cmd.Flags().Lookup(endpointFlag))
	_ = viper.BindPFlag(prefix+".region", cmd.Flags().Lookup(regionFlag))
	_ = viper.BindPFlag(prefix+".access_key", cmd.Flags().Lookup(accessKeyFlag))
	_ = viper.BindPFlag(prefix+".secret_key", cmd.Flags().Lookup(secretKeyFlag))
	_ = viper.BindPFlag(prefix+".use_ssl", cmd.Flags().Lookup(useSSLFlag))
	_ = viper.BindPFlag(prefix+".presigned_url_ttl", cmd.Flags().Lookup(ttlFlag))
	_ = viper.BindPFlag(prefix+".public_bucket", cmd.Flags().Lookup(publicBucketFlag))
	_ = viper.BindPFlag(prefix+".private_bucket", cmd.Flags().Lookup(privateBucketFlag))
	_ = viper.BindPFlag(prefix+".public_base_url", cmd.Flags().Lookup(publicBaseURLFlag))
}

func (c *S3Config) validate() error {
	if c.Endpoint == "" {
		return fmt.Errorf("S3 endpoint is required")
	}
	if c.Region == "" {
		return fmt.Errorf("S3 region is required")
	}
	if c.AccessKey == "" {
		return fmt.Errorf("S3 access key is required")
	}
	if c.SecretKey == "" {
		return fmt.Errorf("S3 secret key is required")
	}
	if c.PublicBucket == "" {
		return fmt.Errorf("S3 public bucket is required")
	}
	if c.PrivateBucket == "" {
		return fmt.Errorf("S3 private bucket is required")
	}
	return nil
}
