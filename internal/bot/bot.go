package bot

import (
	"context"
	"fmt"
	"github.com/1k-off/dev-helper-bot/internal/cache"
	"github.com/1k-off/dev-helper-bot/internal/handlers"
	"github.com/rs/zerolog/log"
	"github.com/shomali11/slacker"
	"github.com/slack-go/slack"
	"strings"
)

type Config struct {
	bot            *slacker.Slacker
	SlackAuthToken string
	SlackAppToken  string
	Ctx            context.Context
	Cancel         context.CancelFunc
	AdminUserIDs   []string
	CmdHandler     *handlers.Handler
	ChannelName    string
	Cache          cache.Cache
}

var (
	messageTimeFormat string
)

func init() {
	// set time format without timezone, seconds and microseconds
	messageTimeFormat = "2006-01-02 15:04"
}

func New(authToken, appToken, channelName string, adminUserEmails []string, cmdHandler *handlers.Handler, cache cache.Cache) *Config {
	ctx, cancel := context.WithCancel(context.Background())
	bot := slacker.NewClient(
		authToken,
		appToken,
		slacker.WithoutAllFormatting(),
	)
	adminUserIds := getAdminUserIDs(bot.APIClient(), adminUserEmails)

	return &Config{
		bot:            bot,
		SlackAuthToken: authToken,
		SlackAppToken:  appToken,
		Ctx:            ctx,
		Cancel:         cancel,
		CmdHandler:     cmdHandler,
		AdminUserIDs:   adminUserIds,
		ChannelName:    channelName,
		Cache:          cache,
	}
}

func (b *Config) Run() error {
	b.defineCronJobs()
	b.defineVpnCommands()
	b.defineDomainCommands()
	return b.bot.Listen(b.Ctx)
}

func (b *Config) Stop() {
	b.Cancel()
}

func (b *Config) defineVpnCommands() {
	getConfig := &slacker.CommandDefinition{
		Description: "Get your personal vpn config",
		Examples:    []string{"vpn get"},
		Handler: func(botCtx slacker.BotContext, request slacker.Request, response slacker.ResponseWriter) {
			userId := botCtx.Event().UserID
			email := getUserEmail(botCtx.APIClient(), userId)
			result, err := b.CmdHandler.VpnGetConfigUrl(email)
			if err != nil {
				log.Err(err).Msgf("Error getting vpn config url. Request: %v, user: %v", botCtx.Event().Text, userId)
				replyErr := response.Reply("Error getting vpn config url. Please contact admin.", slacker.WithThreadReply(true))
				if replyErr != nil {
					log.Err(replyErr).Msgf("Error sending reply. Request: %v, user: %v", botCtx.Event().Text, userId)
				}
				return
			}
			client := botCtx.APIClient()
			_, _, err = client.PostMessage(userId, slack.MsgOptionText(result, false))
			if err != nil {
				log.Err(err).Msgf("Error sending direct message. Request: %v, user: %v", botCtx.Event().Text, userId)
				return
			}

			err = response.Reply("Just sent the link in a private message.", slacker.WithThreadReply(true))
			if err != nil {
				log.Err(err).Msgf("Error sending reply. Request: %v, user: %v", botCtx.Event().Text, userId)
				return
			}
		},
	}

	createProfile := &slacker.CommandDefinition{
		Description: "[ADMIN] Create a new vpn profile.",
		Examples:    []string{"vpn create <login> <email>", "vpn create @user"},
		AuthorizationFunc: func(botCtx slacker.BotContext, request slacker.Request) bool {
			return contains(b.AdminUserIDs, botCtx.Event().UserID)
		},
		Handler: func(botCtx slacker.BotContext, request slacker.Request, response slacker.ResponseWriter) {
			login := request.Param("login")
			email := request.Param("email")
			if email == "" {
				if login != "" {
					if strings.HasPrefix(login, "<@") {
						id := strings.TrimSuffix(strings.TrimPrefix(login, "<@"), ">")
						email = getUserEmail(botCtx.APIClient(), id)
						login = getUserFriendlyName(botCtx.APIClient(), id)
					} else {
						err := response.Reply("Not enough arguments", slacker.WithThreadReply(true))
						if err != nil {
							log.Err(err).Msgf("Error sending reply. Request: %v, user: %v", botCtx.Event().Text, botCtx.Event().UserID)
							return
						}
						return
					}
				} else {
					err := response.Reply("Not enough arguments", slacker.WithThreadReply(true))
					if err != nil {
						log.Err(err).Msgf("Error sending reply. Request: %v, user: %v", botCtx.Event().Text, botCtx.Event().UserID)
						return
					}
					return
				}
			} else {
				email = extractEmail(email)
			}
			err := b.CmdHandler.VpnCreateUser(login, email)
			if err != nil {
				log.Err(err).Msgf("Error creating vpn profile. Request: %v, user: %v", botCtx.Event().Text, botCtx.Event().UserID)
				replyErr := response.Reply(fmt.Sprintf("Error creating vpn profile. %v", err), slacker.WithThreadReply(true))
				if replyErr != nil {
					log.Err(replyErr).Msgf("Error sending reply. Request: %v, user: %v", botCtx.Event().Text, botCtx.Event().UserID)
				}
			} else {
				err = response.Reply("User created.", slacker.WithThreadReply(true))
				if err != nil {
					log.Err(err).Msgf("Error sending reply. Request: %v, user: %v", botCtx.Event().Text, botCtx.Event().UserID)
					return
				}
			}
		},
	}

	deleteProfile := &slacker.CommandDefinition{
		Description: "[ADMIN] Delete a vpn profile.",
		Examples:    []string{"vpn delete <email>", "vpn delete @user"},
		AuthorizationFunc: func(botCtx slacker.BotContext, request slacker.Request) bool {
			return contains(b.AdminUserIDs, botCtx.Event().UserID)
		},
		Handler: func(botCtx slacker.BotContext, request slacker.Request, response slacker.ResponseWriter) {
			email := request.Param("email")
			if email == "" {
				err := response.Reply("Not enough arguments", slacker.WithThreadReply(true))
				if err != nil {
					log.Err(err).Msgf("Error sending reply. Request: %v, user: %v", botCtx.Event().Text, botCtx.Event().UserID)
					return
				}
				return
			}
			if strings.HasPrefix(email, "<@") {
				id := strings.TrimSuffix(strings.TrimPrefix(email, "<@"), ">")
				email = getUserEmail(botCtx.APIClient(), id)
			} else {
				email = extractEmail(email)
			}
			err := b.CmdHandler.VpnDeleteUser(email)
			if err != nil {
				log.Err(err).Msgf("Error deleting vpn profile. Request: %v, user: %v", botCtx.Event().Text, botCtx.Event().UserID)
				replyErr := response.Reply(fmt.Sprintf("Error deleting vpn profile. %v", err), slacker.WithThreadReply(true))
				if replyErr != nil {
					log.Err(replyErr).Msgf("Error sending reply. Request: %v, user: %v", botCtx.Event().Text, botCtx.Event().UserID)
				}
			} else {
				replyErr := response.Reply("Deleted vpn profile.", slacker.WithThreadReply(true))
				if replyErr != nil {
					log.Err(replyErr).Msgf("Error sending reply. Request: %v, user: %v", botCtx.Event().Text, botCtx.Event().UserID)
				}
			}
		},
	}
	b.bot.Command("vpn get", getConfig)
	b.bot.Command("vpn create <login> <email>", createProfile)
	b.bot.Command("vpn delete <email>", deleteProfile)
}

func (b *Config) defineDomainCommands() {
	createCommand := &slacker.CommandDefinition{
		Description: "Create a domain for provided IP.",
		Examples:    []string{"create 127.0.0.1"},
		Handler: func(botCtx slacker.BotContext, request slacker.Request, response slacker.ResponseWriter) {
			ip := request.Param("IP")
			userId := botCtx.Event().UserID
			id := strings.TrimSuffix(strings.TrimPrefix(userId, "<@"), ">")
			userName := getUserFriendlyName(botCtx.APIClient(), id)
			d, err := b.CmdHandler.DomainCreate(id, userName, ip)
			if err != nil {
				log.Err(err).Msgf("Error creating domain. Request: %v, user: %v", botCtx.Event().Text, botCtx.Event().UserID)
				replyErr := response.Reply(fmt.Sprintf("Error creating domain. %v", err), slacker.WithThreadReply(true))
				if replyErr != nil {
					log.Err(replyErr).Msgf("Error sending reply. Request: %v, user: %v", botCtx.Event().Text, botCtx.Event().UserID)
				}
				return
			}
			err = response.Reply(
				fmt.Sprintf("Created domain %s with IP %s. Scheduled delete date: %s.", d.FQDN, d.IP, d.DeleteAt.In(b.CmdHandler.Timezone).Format(messageTimeFormat)),
				slacker.WithThreadReply(true),
			)
			if err != nil {
				log.Err(err).Msgf("Error sending reply. Request: %v, user: %v", botCtx.Event().Text, botCtx.Event().UserID)
				return
			}
		},
	}

	updateCommand := &slacker.CommandDefinition{
		Description: "Update parameter for domain. Available params: expire (no value), basic-auth(true|false), ip (<ip>), full-ssl(true|false).",
		Examples:    []string{"update <param> <value>", "update expire", "update basic-auth true", "update ip 127.0.0.1"},
		Handler: func(botCtx slacker.BotContext, request slacker.Request, response slacker.ResponseWriter) {
			param := request.StringParam("param", "expire")
			value := request.StringParam("value", "")
			userId := botCtx.Event().UserID
			result, err := b.CmdHandler.DomainUpdate(userId, param, value)
			if err != nil {
				log.Err(err).Msgf("Error updating domain. Request: %v, user: %v", botCtx.Event().Text, botCtx.Event().UserID)
				replyErr := response.Reply(fmt.Sprintf("Error updating domain. %v", err), slacker.WithThreadReply(true))
				if replyErr != nil {
					log.Err(replyErr).Msgf("Error sending reply. Request: %v, user: %v", botCtx.Event().Text, botCtx.Event().UserID)
				}
				return
			}
			err = response.Reply(result, slacker.WithThreadReply(true))
			if err != nil {
				log.Err(err).Msgf("Error sending reply. Request: %v, user: %v", botCtx.Event().Text, botCtx.Event().UserID)
				return
			}
		},
	}

	deleteCommand := &slacker.CommandDefinition{
		Description: "Delete domain assigned to you.",
		Examples:    []string{"delete"},
		Handler: func(botCtx slacker.BotContext, request slacker.Request, response slacker.ResponseWriter) {
			userId := botCtx.Event().UserID
			result, err := b.CmdHandler.DomainDelete(userId)
			if err != nil {
				log.Err(err).Msgf("Error deleting domain. Request: %v, user: %v", botCtx.Event().Text, botCtx.Event().UserID)
				replyErr := response.Reply(fmt.Sprintf("Error deleting domain. %v", err), slacker.WithThreadReply(true))
				if replyErr != nil {
					log.Err(replyErr).Msgf("Error sending reply. Request: %v, user: %v", botCtx.Event().Text, botCtx.Event().UserID)
				}
				return
			}
			err = response.Reply(result, slacker.WithThreadReply(true))
			if err != nil {
				log.Err(err).Msgf("Error sending reply. Request: %v, user: %v", botCtx.Event().Text, botCtx.Event().UserID)
				return
			}
		},
	}

	b.bot.Command("create <IP>", createCommand)
	b.bot.Command("domain create <IP>", createCommand)
	b.bot.Command("update <param> <value>", updateCommand)
	b.bot.Command("domain update <param> <value>", updateCommand)
	b.bot.Command("delete", deleteCommand)
	b.bot.Command("domain delete", deleteCommand)
}
