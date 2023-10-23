package main

import (
	"github.com/1k-off/dev-helper-bot/internal/bot"
	"github.com/1k-off/dev-helper-bot/internal/cache"
	"github.com/1k-off/dev-helper-bot/internal/config"
	"github.com/1k-off/dev-helper-bot/internal/handlers"
	"github.com/1k-off/dev-helper-bot/internal/store/mongostore"
	"github.com/1k-off/dev-helper-bot/pkg/pritunl"
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
	pritunlEUClient := pritunl.New(cfg.Pritunl.HostEU, cfg.Pritunl.TokenEU, cfg.Pritunl.SecretEU, cfg.Pritunl.OrganizationEU)
	store := mongostore.New(cfg.App.DatasourceConnectionString)
	messageTemplates := map[string]string{
		"vpnWelcomeMessage": cfg.Pritunl.WelcomeMessage,
	}
	handler := handlers.New(pritunlClient, pritunlEUClient, cfg.Webserver, store, cfg.Timezone, messageTemplates)
	c, err := cache.New("./data/cache")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create cache")
	}

	slackBot := bot.New(cfg.Slack.AuthToken, cfg.Slack.AppToken, cfg.Slack.Channel, cfg.App.AdminEmails, handler, c)

	stopCh := make(chan os.Signal)
	signal.Notify(stopCh, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-stopCh
		exitCode := 0
		err := store.Close()
		if err != nil {
			log.Err(err).Msg("failed to close store")
			exitCode = 1
		}
		err = c.Close()
		if err != nil {
			log.Err(err).Msg("failed to close cache")
			exitCode = 1
		}
		slackBot.Stop()
		log.Info().Msg("Application stopped")
		os.Exit(exitCode)
	}()

	err = slackBot.Run()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to run bot bot")
	}
}
