package cmd

import (
	"fmt"

	"github.com/BhaumikTalwar/Gama/config"
	"github.com/spf13/cobra"
)

var srvDump = &cobra.Command{
	Use:   "dump",
	Short: "To Dump the current read config with teh env vars loaded",
	Long:  "To dump the config it will redact the sensitive passwords",

	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("%s", config.DumpConfig())
		return nil
	},
}

func init() {
	serveCmd.AddCommand(srvDump)
}
