package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/ldmonster/dragncards/ha-be/config"
	pluginapp "github.com/ldmonster/dragncards/ha-be/internal/application/plugin"
	"github.com/ldmonster/dragncards/ha-be/internal/domain/alert"
	"github.com/ldmonster/dragncards/ha-be/internal/domain/deck"
	"github.com/ldmonster/dragncards/ha-be/internal/domain/lfg"
	plugDomain "github.com/ldmonster/dragncards/ha-be/internal/domain/plugin"
	"github.com/ldmonster/dragncards/ha-be/internal/domain/replay"
	"github.com/ldmonster/dragncards/ha-be/internal/domain/room"
	"github.com/ldmonster/dragncards/ha-be/internal/domain/settings"
	"github.com/ldmonster/dragncards/ha-be/internal/infrastructure/persistence"
	"github.com/ldmonster/dragncards/ha-be/internal/platform/database"
)

var updatePluginCfgPath string
var updatePluginRepoURL string

func init() {
	updatePluginCmd.Flags().StringVarP(&updatePluginCfgPath, "config", "c", "config/config.yaml.example", "path to config file")
	updatePluginCmd.Flags().StringVarP(&updatePluginRepoURL, "repo-url", "r", "", "plugin repo URL to sync")
	rootCmd.AddCommand(updatePluginCmd)
}

var updatePluginCmd = &cobra.Command{
	Use:   "plugin-update",
	Short: "Sync plugin repository into database",
	Run: func(cmd *cobra.Command, args []string) {
		if updatePluginRepoURL == "" {
			fmt.Println("repo-url is required")
			os.Exit(1)
		}

		cfg, err := config.LoadConfig(updatePluginCfgPath)
		if err != nil {
			fmt.Printf("load config: %v\n", err)
			os.Exit(1)
		}

		db, dbErr := database.New(cfg.Database.URL)
		if dbErr != nil {
			fmt.Printf("db connection failed: %v (using in-memory repository only)\n", dbErr)
		}

		var pluginRepo plugDomain.PluginRepository
		if db != nil {
			if err := db.AutoMigrate(&room.Room{}, &plugDomain.Plugin{}, &plugDomain.CustomCard{}, &plugDomain.UserPluginPermission{}, &lfg.LfgPost{}, &alert.Alert{}, &deck.Deck{}, &replay.Replay{}, &settings.Setting{}); err != nil {
				fmt.Printf("auto migrate failed: %v\n", err)
			}
			pluginRepo = persistence.NewGormPluginRepository(db)
		} else {
			pluginRepo = persistence.NewInMemoryPluginRepository()
		}

		pluginSvc := pluginapp.NewService(pluginRepo)
		plugins, err := pluginSvc.SyncRepository(updatePluginRepoURL)
		if err != nil {
			fmt.Printf("failed to sync plugin repo: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("synced %d plugins\n", len(plugins))
	},
}
