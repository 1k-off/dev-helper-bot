package main

import (
	"devhelper/internal/config"
	"devhelper/internal/handlers"
	"devhelper/internal/slack"
	"devhelper/internal/store/mongostore"
	"devhelper/pkg/pritunl"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"os/signal"
	"syscall"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.ErrorLevel)
}

func main() {
	cfg, err := config.Load("./config/")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}

	pritunlClient := pritunl.New(cfg.Pritunl.Host, cfg.Pritunl.Token, cfg.Pritunl.Secret, cfg.Pritunl.Organization)
	store := mongostore.New(cfg.App.DatasourceConnectionString)
	handler := handlers.New(pritunlClient, cfg.Nginx, store)
	slackBot := slack.New(cfg.Slack.AuthToken, cfg.Slack.AppToken, cfg.App.AdminEmails, handler)

	stopCh := make(chan os.Signal)
	signal.Notify(stopCh, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-stopCh
		err := store.Close()
		if err != nil {
			return
		}
		slackBot.Stop()
		log.Info().Msg("Application stopped")
		os.Exit(0)
	}()

	err = slackBot.Run()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to run slack bot")
	}
}
