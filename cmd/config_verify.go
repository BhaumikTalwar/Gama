package cmd

import (
	"fmt"

	"github.com/BhaumikTalwar/Gama/config"
	"github.com/spf13/cobra"
)

var configVerifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "To Verify the config",
	Long:  "To verify the config present at default location or explicitly passed config at path",

	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.VerifyConfig(cfgFile); err != nil {
			return fmt.Errorf("Config Verification failed: %w", err)
		}

		fmt.Printf("Successfully Verified Config")
		return nil
	},
}

func init() {
	configCmd.AddCommand(configVerifyCmd)
}
