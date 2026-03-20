package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ldmonster/dragncards/ha-be/config"
	"github.com/ldmonster/dragncards/ha-be/internal/application/deck"
	"github.com/ldmonster/dragncards/ha-be/internal/application/game"
	"github.com/ldmonster/dragncards/ha-be/internal/application/replay"
	"github.com/ldmonster/dragncards/ha-be/internal/domain/alert"
	deckDomain "github.com/ldmonster/dragncards/ha-be/internal/domain/deck"
	"github.com/ldmonster/dragncards/ha-be/internal/domain/identity"
	"github.com/ldmonster/dragncards/ha-be/internal/domain/lfg"
	"github.com/ldmonster/dragncards/ha-be/internal/domain/plugin"
	replayDomain "github.com/ldmonster/dragncards/ha-be/internal/domain/replay"
	"github.com/ldmonster/dragncards/ha-be/internal/domain/room"
	"github.com/ldmonster/dragncards/ha-be/internal/infrastructure/email"
	"github.com/ldmonster/dragncards/ha-be/internal/infrastructure/persistence"
	httpapi "github.com/ldmonster/dragncards/ha-be/internal/interfaces/http"
	wsapi "github.com/ldmonster/dragncards/ha-be/internal/interfaces/ws"
	"github.com/ldmonster/dragncards/ha-be/internal/platform/auth"
	"github.com/ldmonster/dragncards/ha-be/internal/platform/database"
	"github.com/ldmonster/dragncards/ha-be/internal/platform/logger"
)

func main() {
	cfgPath := flag.String("config", "config/config.yaml.example", "path to config file")
	flag.Parse()

	cfg, err := config.LoadConfig(*cfgPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	logr := logger.New(cfg.Log.Level)
	logr.Info("starting server", "port", cfg.Server.Port)

	db, dbErr := database.New(cfg.Database.URL)
	if dbErr != nil {
		logr.Error("db connection failed, using in-memory store", "error", dbErr)
	}

	if db != nil {
		if err := db.AutoMigrate(&identity.User{}, &room.Room{}, &room.RoomAction{}, &plugin.Plugin{}, &plugin.CustomCard{}, &lfg.LfgPost{}, &alert.Alert{}); err != nil {
			logr.Error("auto migrate failed", "error", err)
		}
	}

	var userRepo identity.UserRepository
	var roomRepo room.RoomRepository
	var pluginRepo plugin.PluginRepository
	var deckRepo deckDomain.DeckRepository
	var replayRepo replayDomain.ReplayRepository
	var lfgRepo lfg.LfgRepository
	var alertRepo alert.AlertRepository

	if db != nil {
		userRepo = persistence.NewGormUserRepository(db)
		roomRepo = persistence.NewGormRoomRepository(db)
		pluginRepo = persistence.NewGormPluginRepository(db)
		deckRepo = persistence.NewGormDeckRepository(db)
		replayRepo = persistence.NewGormReplayRepository(db)
		lfgRepo = persistence.NewGormLfgRepository(db)
		alertRepo = persistence.NewGormAlertRepository(db)
	} else {
		userRepo = persistence.NewInMemoryUserRepository()
		roomRepo = persistence.NewInMemoryRoomRepository()
		pluginRepo = persistence.NewInMemoryPluginRepository()
		deckRepo = persistence.NewInMemoryDeckRepository()
		replayRepo = persistence.NewInMemoryReplayRepository()
		lfgRepo = persistence.NewInMemoryLfgRepository()
		alertRepo = persistence.NewInMemoryAlertRepository()
	}

	identitySvc := identity.NewService(userRepo)
	roomSvc := room.NewService(roomRepo)
	pluginSvc := plugin.NewService(pluginRepo)
	deckSvc := deck.NewDeckService(deckRepo)
	replaySvc := replay.NewReplayService(replayRepo)
	lfgSvc := lfg.NewService(lfgRepo)
	alertSvc := alert.NewService(alertRepo)
	gameRegistry := game.NewRoomRegistry()
	gameSvc := game.NewGameService(roomSvc, gameRegistry, nil)

	if cfg.Auth.JWTSecret != "" {
		auth.SetSecret(cfg.Auth.JWTSecret)
	}

	mailer := email.NewSMTPMailer("no-reply@dragncards.com", cfg.Email.SMTPHost, cfg.Email.SMTPPort, cfg.Email.SMTPUsername, cfg.Email.SMTPPassword)
	identitySvc.SetTokenTTL(time.Duration(cfg.Auth.AccessLifetime)*time.Minute, time.Duration(cfg.Auth.RefreshLifetime)*time.Hour)
	apiHandler := httpapi.NewAPIHandler(identitySvc, roomSvc, pluginSvc, gameSvc, deckSvc, replaySvc, lfgSvc, alertSvc, mailer, cfg.Recaptcha.SecretKey, time.Duration(cfg.Auth.AccessLifetime)*time.Minute, time.Duration(cfg.Auth.RefreshLifetime)*time.Hour)
	mux := http.NewServeMux()
	hub := wsapi.NewHub()
	mux.Handle("/be/socket", wsapi.NewWSHandler(hub, roomSvc, lfgSvc, gameSvc, replaySvc))
	mux.Handle("/be/", httpapi.NewRouter(apiHandler))

	srv := &http.Server{Addr: fmt.Sprintf(":%d", cfg.Server.Port), Handler: mux}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logr.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logr.Error("graceful shutdown failed", "error", err)
	}
}
