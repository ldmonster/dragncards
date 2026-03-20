package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var serveCfgPath string

func init() {
	serveCmd.Flags().StringVarP(&serveCfgPath, "config", "c", "config/config.yaml.example", "path to config file")
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run the DragnCards HTTP server",
	Run: func(cmd *cobra.Command, args []string) {
		if err := RunServe(serveCfgPath); err != nil {
			fmt.Printf("server error: %v\n", err)
			os.Exit(1)
		}
	},
}
