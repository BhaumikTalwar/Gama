package config

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type AppConfig struct {
	AppName       string   `mapstructure:"app_name" yaml:"app_name"`
	Env           string   `mapstructure:"env" yaml:"env"`
	CorsAddresses []string `mapstructure:"cors_addresses" yaml:"cors_addresses"`

	AppSecret string `mapstructure:"app_secret" yaml:"app_secret"`
	JWTKey    string `mapstructure:"jwt_key" yaml:"jwt_key"`
	AESKey    string `mapstructure:"aes_key" yaml:"aes_key"`

	MFASmsEnabled bool `mapstructure:"mfa_sms" yaml:"mfa_sms"`

	AccessTokenDuration  time.Duration `mapstructure:"access_token_duration" yaml:"access_token_duration"`
	RefreshTokenDuration time.Duration `mapstructure:"refresh_token_duration" yaml:"refresh_token_duration"`

	RefreshRotationThreshold time.Duration `mapstructure:"refresh_rotation_threshold" yaml:"refresh_rotation_threshold"`
}

func setAppDefaults(prefix string) {
	viper.SetDefault(prefix+".env", "dev")
	viper.SetDefault(prefix+".app_secret", "")
	viper.SetDefault(prefix+".cors_addresses", []string{
		"http://localhost:3000",
		"http://127.0.0.1:3000",
		"http://localhost:5173",
		"http://127.0.0.1:5173",
	})
	viper.SetDefault(prefix+".jwt_key", "")
	viper.SetDefault(prefix+".mfa_sms", false)
	viper.SetDefault(prefix+".aes_key", "")
	viper.SetDefault(prefix+".access_token_duration", "15m")
	viper.SetDefault(prefix+".refresh_token_duration", "24h")
	viper.SetDefault(prefix+".refresh_rotation_threshold", "1h")
}

func registerAppFlags(cmd *cobra.Command, prefix string) {
	envFlag := fmt.Sprintf("%s-env", prefix)
	corsFlag := fmt.Sprintf("%s-cors-address", prefix)
	secFlag := fmt.Sprintf("%s-secret", prefix)
	jwtFlag := fmt.Sprintf("%s-jwt", prefix)
	mfaSmsFlag := fmt.Sprintf("%s-mfa-sms", prefix)
	aesFlag := fmt.Sprintf("%s-aes", prefix)
	accessDurFlag := fmt.Sprintf("%s-access-token-duration", prefix)
	refreshDurFlag := fmt.Sprintf("%s-refresh-token-duration", prefix)
	refreshRotFlag := fmt.Sprintf("%s-refresh-rotation-threshold", prefix)

	cmd.Flags().String(envFlag, "", "Environment (dev/prod)")
	cmd.Flags().String(corsFlag, "", "Cors Addresses")
	cmd.Flags().String(secFlag, "", "App Secret 32 bytes long")
	cmd.Flags().String(jwtFlag, "", "JWT Key to be used 32 bytes long")
	cmd.Flags().String(mfaSmsFlag, "", "Weather to Enable sms based MFA as an opton")
	cmd.Flags().String(aesFlag, "", "AES encryption key to be used for Encryption")
	cmd.Flags().String(accessDurFlag, "", "Access Token Duration For JWT Token")
	cmd.Flags().String(refreshDurFlag, "", "Refresh Token Duration")
	cmd.Flags().String(refreshRotFlag, "", "Refresh token rotation threshold duration")

	_ = viper.BindPFlag(prefix+".env", cmd.Flags().Lookup(envFlag))
	_ = viper.BindPFlag(prefix+".cors_addresses", cmd.Flags().Lookup(corsFlag))
	_ = viper.BindPFlag(prefix+".app_secret", cmd.Flags().Lookup(secFlag))
	_ = viper.BindPFlag(prefix+".jwt_key", cmd.Flags().Lookup(jwtFlag))
	_ = viper.BindPFlag(prefix+".mfa_sms", cmd.Flags().Lookup(mfaSmsFlag))
	_ = viper.BindPFlag(prefix+".aes_key", cmd.Flags().Lookup(aesFlag))
	_ = viper.BindPFlag(prefix+".access_token_duration", cmd.Flags().Lookup(accessDurFlag))
	_ = viper.BindPFlag(prefix+".refresh_token_duration", cmd.Flags().Lookup(refreshDurFlag))
	_ = viper.BindPFlag(prefix+".refresh_rotation_threshold", cmd.Flags().Lookup(refreshRotFlag))
}

func (c *AppConfig) validate() error {
	if !(strings.EqualFold(c.Env, "prod") || strings.EqualFold(c.Env, "dev")) {
		return fmt.Errorf("Invalid Env Setting:- %s", c.Env)
	}

	if c.AppSecret == "" {
		return errors.New("app.app_secret must be set")
	}

	if len(c.CorsAddresses) == 0 {
		return errors.New("Cors Address Needs to be set")
	}

	if len(c.AppSecret) < 32 {
		return errors.New("App Secret Key Should be atleast 32 bytes long")
	}

	if c.JWTKey == "" {
		return errors.New("app.jwt_key must be set")
	}

	if len(c.JWTKey) < 32 {
		return errors.New("JWT Secret Key Should be atleast 32 bytes long")
	}

	if c.AESKey == "" {
		return errors.New("app.aes_key must be set")
	}

	if len(c.AESKey) < 32 {
		return errors.New("AES Secret Key Should be atleast 32 bytes long")
	}

	if c.RefreshRotationThreshold <= 0 {
		return errors.New("app.refresh_rotation_threshold must be greater than 0")
	}

	if c.RefreshRotationThreshold >= c.RefreshTokenDuration {
		return errors.New("app.refresh_rotation_threshold must be less than refresh_token_duration")
	}

	return nil
}
