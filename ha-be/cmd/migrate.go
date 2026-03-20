package cmd

import (
	"fmt"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/ldmonster/dragncards/ha-be/config"
	"github.com/spf13/cobra"
)

var migrateCfgPath string
var migrationSourcePath string

func init() {
	migrateCmd.Flags().StringVarP(&migrateCfgPath, "config", "c", "config/config.yaml.example", "path to config file")
	migrateCmd.Flags().StringVarP(&migrationSourcePath, "path", "p", "migrations", "path to migrations directory")
}

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run database migrations",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadConfig(migrateCfgPath)
		if err != nil {
			fmt.Printf("load config: %v\n", err)
			os.Exit(1)
		}

		if cfg.Database.URL == "" {
			fmt.Fprintf(os.Stderr, "database URL is not configured\n")
			os.Exit(1)
		}

		m, err := migrate.New("file://"+migrationSourcePath, cfg.Database.URL)
		if err != nil {
			fmt.Printf("failed to initialize migrations: %v\n", err)
			os.Exit(1)
		}

		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			fmt.Printf("migrations failed: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("migrations applied")
	},
}
