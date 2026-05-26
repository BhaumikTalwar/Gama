package cmd

import (
	"fmt"
	"os"

	"github.com/BhaumikTalwar/Gama/config"
	"github.com/spf13/cobra"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "Gama: A Golang Backend Template based on Gin",
	Short: "Gama is a Golang Backend Template build for rapid development and deployment",
	Long: `Gama: is a Go Lang backend Template based on Gin framework and provides several pre built options
Core Stack Options: 
	- Sqlite (Default DB)
	- PostgresSql 
	- Redis 
	- In Memory Cache`,

	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if cmd.Parent() != nil && cmd.Parent().Name() == "config" {
			return nil
		}

		config.SetDefaults()
		return config.LoadConfig(cfgFile)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $PWD/config.yaml)")
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
