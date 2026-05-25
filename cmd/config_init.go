package cmd

import (
	"github.com/BhaumikTalwar/Gama/config"
	"github.com/spf13/cobra"
)

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "To Initialize a basic config File",
	Long:  "To create a basic default populated config file at $(PWD)/config.yaml",

	RunE: func(cmd *cobra.Command, args []string) error {
		return config.ExampleConfig()
	},
}

func init() {
	configCmd.AddCommand(configInitCmd)
}
