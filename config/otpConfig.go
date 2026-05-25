package config

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type OTPConfig struct {
	APIKey string `mapstructure:"api_key" yaml:"api_key" validate:"required"`
}

func setOTPDefaults(prefix string) {
	viper.SetDefault(prefix+".api_key", "")
}

func registerOTPFlags(cmd *cobra.Command, prefix string) {
	apiKeyFlag := fmt.Sprintf("%s-api-key", prefix)

	cmd.Flags().String(apiKeyFlag, "", "2Factor.in API Key for OTP SMS")

	_ = viper.BindPFlag(prefix+".api_key", cmd.Flags().Lookup(apiKeyFlag))
}

func (c *OTPConfig) validate() error {
	if c.APIKey == "" {
		return fmt.Errorf("OTP API key is required")
	}
	return nil
}
