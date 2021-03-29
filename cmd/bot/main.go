package main

import (
	"github.com/robfig/cron/v3"
	"github.com/souladm/dev-helper-bot/internal/app"
	"github.com/souladm/dev-helper-bot/internal/bot"
	"log"
)

func main() {
	config, err := app.Configure()
	if err != nil {
		log.Fatal(err)
	}

	c := cron.New(cron.WithLogger(cron.VerbosePrintfLogger(log.Default())))
	c.AddFunc("00 09 */1 * *", func() {bot.SendDeleteReminders(config)})
	c.AddFunc("00 09 */1 * *", func() {bot.DeleteExpiredDomains(config)})
	if config.Debug {
		c.AddFunc("*/1 * * * *", func() {bot.SendDeleteReminders(config)})
		c.AddFunc("*/1 * * * *", func() {bot.DeleteExpiredDomains(config)})
	}
	c.Start()

	if err := bot.Run(config); err != nil {
		log.Fatal(err)
	}
}