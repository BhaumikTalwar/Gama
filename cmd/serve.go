package cmd

import (
	"github.com/BhaumikTalwar/Gama/config"
	"github.com/BhaumikTalwar/Gama/internal/server"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "server",
	Short: "To start the server application",
	Long: `The command to start the server application and listen to the specified port and respond accoding
to the setep handler and application logic`,

	Run: func(cmd *cobra.Command, args []string) {
		server.RunServer()
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
	config.RegisterServeFlags(serveCmd)
}
