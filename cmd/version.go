package cmd

import (
	"fmt"

	"github.com/BhaumikTalwar/Gama/internal/buildinfo"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "To get the version of the application",
	Long:  "To get the version of the current build of the application",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("%s", buildinfo.VersionStr())
		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
