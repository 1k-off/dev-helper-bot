package bot

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/shomali11/slacker"
	"github.com/slack-go/slack"
	"os"
	"strconv"
)

func (b *Config) defineCronJobs() {
	// Seconds: The second the cron job should run on (0-59)
	// Minutes: The minute the cron job should run on (0-59)
	// Hours: The hour the cron job should run on (0-23)
	// Day of month: The day of the month the cron job should run on (1-31)
	// Month: The month the cron job should run on (1-12)
	// Day of week: The day of the week the cron job should run on (0-7, where both 0 and 7 represent Sunday)

	cronValue := "0 9 * * * *"
	// get os env
	debugModeString := os.Getenv("OOOPS_DEBUG")
	debugMode, err := strconv.ParseBool(debugModeString)
	if err != nil {
		debugMode = false
		log.Err(err).Msg("Error parsing OOOPS_DEBUG env variable")
	}
	if debugMode {
		cronValue = "0 * * * * *"
	}

	b.bot.Job(cronValue, &slacker.JobDefinition{
		Description: "Reminder about expiring soon domains",
		Handler: func(jobCtx slacker.JobContext) {
			client := jobCtx.APIClient()
			domains, err := b.CmdHandler.DomainGetExpirationSoon()
			if err != nil {
				log.Err(err).Msg("Error getting domains")
				return
			}
			if len(domains) != 0 {
				for _, d := range domains {
					// check if user already notified
					notified, err := b.isNotified(d.UserId)
					if err != nil {
						log.Err(err).Msg("Error getting notified flag")
					}
					if notified {
						continue
					}
					// send to channel
					_, _, err = client.PostMessage(
						b.ChannelName,
						slack.MsgOptionText(fmt.Sprintf("<@%s>", d.UserId)+fmt.Sprintf("Your domain %s scheduled to delete at %s.", d.FQDN, d.DeleteAt.In(b.CmdHandler.Timezone).Format(messageTimeFormat)), false),
						slack.MsgOptionAsUser(true),
					)
					if err != nil {
						log.Error().Err(err).Msgf("ID: %s, domain: %s", d.UserId, d.FQDN)
					}
					// send to user
					_, _, err = client.PostMessage(
						d.UserId,
						slack.MsgOptionText(fmt.Sprintf("Your domain %s scheduled to delete at %s.", d.FQDN, d.DeleteAt.In(b.CmdHandler.Timezone).Format(messageTimeFormat)), false),
						slack.MsgOptionAsUser(true),
					)
					if err != nil {
						log.Error().Err(err).Msgf("ID: %s, domain: %s", d.UserId, d.FQDN)
					}
					// set notified flag
					err = b.setNotified(d.UserId)
					if err != nil {
						log.Error().Err(err).Msgf("Failed to set notified flag", d.UserId, d.FQDN)
					}
				}
			}
		},
	})

	b.bot.Job(cronValue, &slacker.JobDefinition{
		Description: "Notification about deleted domains",
		Handler: func(jobCtx slacker.JobContext) {
			client := jobCtx.APIClient()
			domains, err := b.CmdHandler.DomainGetExpired()
			if err != nil {
				log.Err(err).Msg("Error getting domains")
				return
			}
			if len(domains) != 0 {
				err := b.CmdHandler.DomainDeleteExpired()
				if err != nil {
					log.Err(err).Msg("Error deleting domains")
					// post errors to channel
					adminsMention := ""
					for _, admin := range b.AdminUserIDs {
						adminsMention += fmt.Sprintf("<@%s> ", admin)
					}
					_, _, err := client.PostMessage(
						b.ChannelName,
						slack.MsgOptionText(adminsMention+fmt.Sprintf("%v.", err), false),
						slack.MsgOptionAsUser(true),
					)
					if err != nil {
						log.Error().Err(err).Msgf("")
					}
					return
				}
				for _, d := range domains {
					_, _, err := client.PostMessage(
						b.ChannelName,
						slack.MsgOptionText(fmt.Sprintf("<@%s>", d.UserId)+fmt.Sprintf("Your domain %s deleted.", d.FQDN), false),
						slack.MsgOptionAsUser(true),
					)
					if err != nil {
						log.Error().Err(err).Msgf("ID: %s, domain: %s", d.UserId, d.FQDN)
					}
					// send to user
					_, _, err = client.PostMessage(
						d.UserId,
						slack.MsgOptionText(fmt.Sprintf("Your domain %s deleted.", d.FQDN), false),
						slack.MsgOptionAsUser(true),
					)
					if err != nil {
						log.Error().Err(err).Msgf("ID: %s, domain: %s", d.UserId, d.FQDN)
					}
					// delete notified flag
					err = b.clearNotified(d.UserId)
					if err != nil {
						log.Error().Err(err).Msgf("Failed to clear notified flag", d.UserId, d.FQDN)
					}
				}
			}
		},
	})
}
