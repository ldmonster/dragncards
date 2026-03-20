package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run the DragnCards HTTP server",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Server startup path is not implemented in offline legacy mode. Use go run main.go")
	},
}
