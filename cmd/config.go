package cmd

import (
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configuration related commands",
	Long:  "Inspect, validate, and generate configuration files",
}

func init() {
	rootCmd.AddCommand(configCmd)
}
