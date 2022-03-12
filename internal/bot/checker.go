package bot

import (
	"fmt"
	"github.com/1k-off/dev-helper-bot/internal/app"
	"github.com/1k-off/dev-helper-bot/internal/nginx"
	"github.com/slack-go/slack"
)

func SendDeleteReminders(config *app.Config) {
	domains, err := config.Store.DomainRepository().GetAllRecordsToDeleteInDays(1)
	if err != nil {
		config.Log.Error().Err(err).Msg(fmt.Sprintf("[scheduler] error while searching domains to delete in a day."))
	}
	if len(domains) != 0 {
		for _, d := range domains {
			_, _, err = config.Bot.Client().PostMessage(
				config.SlackChannelName,
				slack.MsgOptionText(fmt.Sprintf("<@%s>", d.UserId)+fmt.Sprintf("Your domain %s scheduled to delete at %s.", d.FQDN, d.DeleteAt), false),
				slack.MsgOptionAsUser(true),
			)
			if err != nil {
				config.Log.Error().Err(err).Msg("")
			}
			_, _, err = config.Bot.Client().PostMessage(
				d.UserId,
				slack.MsgOptionText(fmt.Sprintf("Your domain %s scheduled to delete at %s.", d.FQDN, d.DeleteAt), false),
				slack.MsgOptionAsUser(true),
			)
			if err != nil {
				config.Log.Error().Err(err).Msg("")
			}
		}
	}
}

func DeleteExpiredDomains(config *app.Config) {
	domains, err := config.Store.DomainRepository().GetAllRecordsToDeleteInDays(0)
	if err != nil {
		config.Log.Error().Err(err).Msg("[scheduler] searching domains to delete in a day")
	}

	for _, d := range domains {
		err = config.Store.DomainRepository().DeleteByFqdn(d.FQDN)
		if err != nil {
			config.Log.Error().Err(err).Msg("")
			return
		}
		err = nginx.Delete(d.FQDN)
		if err != nil {
			config.Log.Error().Err(err).Msg("")
		}

		_, _, err = config.Bot.Client().PostMessage(
			config.SlackChannelName,
			slack.MsgOptionText(fmt.Sprintf("<@%s>", d.UserId)+fmt.Sprintf("Your domain %s deleted.", d.FQDN), false),
			slack.MsgOptionAsUser(true),
		)
		if err != nil {
			config.Log.Error().Err(err).Msg("")
		}
		_, _, err = config.Bot.Client().PostMessage(
			d.UserId,
			slack.MsgOptionText(fmt.Sprintf("Your domain %s deleted.", d.FQDN), false),
			slack.MsgOptionAsUser(true),
		)
		if err != nil {
			config.Log.Error().Err(err).Msg("")
		}
	}
}
