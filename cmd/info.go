package cmd

import (
	"fmt"

	"github.com/BhaumikTalwar/Gama/internal/buildinfo"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "To get the Build Info for the server application",
	Long:  "To get the detailed build info the version, commit and the build time of the application",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("%s", buildinfo.BuildInfo())
		return nil
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)
}
