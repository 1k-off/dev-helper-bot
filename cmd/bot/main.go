package main

import (
	"github.com/1k-off/dev-helper-bot/internal/app"
	"github.com/1k-off/dev-helper-bot/internal/bot"
	"github.com/robfig/cron/v3"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	a := app.Startup("data/config.yml")

	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-ch
		a.Stop()
		os.Exit(1)
	}()

	c := cron.New(cron.WithLogger(cron.VerbosePrintfLogger(log.Default())))
	c.AddFunc("00 09 */1 * *", func() { bot.SendDeleteReminders(a) })
	c.AddFunc("00 09 */1 * *", func() { bot.DeleteExpiredDomains(a) })
	if a.LogLevel == app.LogLevelDebug {
		c.AddFunc("*/1 * * * *", func() { bot.SendDeleteReminders(a) })
		c.AddFunc("*/1 * * * *", func() { bot.DeleteExpiredDomains(a) })
	}
	c.Start()

	if err := bot.Run(a); err != nil {
		a.Log.Fatal().Err(err).Msg("")
	}

}
