package bot

import (
	"fmt"
	"github.com/slack-go/slack"
	"github.com/souladm/dev-helper-bot/internal/app"
	"github.com/souladm/dev-helper-bot/internal/nginx"
	"log"
)

func SendDeleteReminders(config *app.Config) {
	domains, err := config.Store.Domain().GetAllRecordsToDeleteInDays(1)
	if err != nil {
		log.Println(fmt.Sprintf("[ERROR] searching domains to delete in a day. %v", err))
	}
	if len(domains) != 0 {
		for _, d := range domains {
			_, _, err := config.Bot.Client().PostMessage(
				config.SlackChannelName,
				slack.MsgOptionText(fmt.Sprintf("<@%s>", d.UserId) + fmt.Sprintf("Your domain %s scheduled to delete at %s.", d.FQDN, d.DeleteAt), false),
				slack.MsgOptionAsUser(true),
			)
			if err != nil {
				log.Println(err)
			}
			_, _, err = config.Bot.Client().PostMessage(
				d.UserId,
				slack.MsgOptionText(fmt.Sprintf("Your domain %s scheduled to delete at %s.", d.FQDN, d.DeleteAt), false),
				slack.MsgOptionAsUser(true),
			)
			if err != nil {
				log.Println(err)
			}
		}
	}
}

func DeleteExpiredDomains(config *app.Config) {
	domains, err := config.Store.Domain().GetAllRecordsToDeleteInDays(0)
	if err != nil {
		log.Println(fmt.Sprintf("[ERROR] searching domains to delete in a day. %v", err))
	}

	for _, d := range domains {
		config.Store.Domain().DeleteByFqdn(d.FQDN)
		err := nginx.Delete(d.FQDN)
		if err != nil {
			log.Println(err)
		}

		_, _, err = config.Bot.Client().PostMessage(
			config.SlackChannelName,
			slack.MsgOptionText(fmt.Sprintf("<@%s>", d.UserId) + fmt.Sprintf("Your domain %s deleted.", d.FQDN), false),
			slack.MsgOptionAsUser(true),
		)
		if err != nil {
			log.Println(err)
		}
		_, _, err = config.Bot.Client().PostMessage(
			d.UserId,
			slack.MsgOptionText(fmt.Sprintf("Your domain %s deleted.", d.FQDN), false),
			slack.MsgOptionAsUser(true),
		)
		if err != nil {
			log.Println(err)
		}
	}

}
