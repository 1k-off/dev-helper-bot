package bot

import (
	"context"
	"fmt"
	"github.com/1k-off/dev-helper-bot/internal/cache"
	"github.com/1k-off/dev-helper-bot/internal/handlers"
	"github.com/rs/zerolog/log"
	"github.com/shomali11/slacker"
	"github.com/slack-go/slack"
	"strconv"
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
	b.defineDomainCronJobs()
	b.defineVpnEUCronJobs()
	b.defineVpnCommands()
	b.defineDomainCommands()
	b.defineVpnEUCommands()
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
			var userId string
			login := request.Param("login")
			email := request.Param("email")
			if email == "" {
				if login != "" {
					if strings.HasPrefix(login, "<@") {
						id := strings.TrimSuffix(strings.TrimPrefix(login, "<@"), ">")
						userId = id
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
				var err error
				userId, err = getSlackUserIdByEmail(botCtx.APIClient(), email)
				if err != nil {
					log.Err(err).Msgf("Error getting slack user id by email. Request: %v, user: %v", botCtx.Event().Text, botCtx.Event().UserID)
					return
				}
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
				err = sendVPNWelcomeMessage(botCtx.APIClient(), userId, b.CmdHandler.GetVpnWelcomeMessage())
				if err != nil {
					log.Err(err).Msgf("Error sending vpn welcome message. Request: %v, user: %v", botCtx.Event().Text, userId)
					return
				}
				err = response.Reply(fmt.Sprintf("Welcome message to %s is sent.", userId), slacker.WithThreadReply(true))
				if err != nil {
					log.Err(err).Msgf("Error sending reply. Request: %v, user: %v", botCtx.Event().Text, botCtx.Event().UserID)
					return
				}
			}
			return
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

	sendConfig := &slacker.CommandDefinition{
		Description: "[ADMIN] Send VPN config to a user.",
		Examples:    []string{"vpn send <email>", "vpn send @user"},
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

			client := botCtx.APIClient()

			var id string

			if strings.HasPrefix(email, "<@") {
				id = strings.TrimSuffix(strings.TrimPrefix(email, "<@"), ">")
				email = getUserEmail(botCtx.APIClient(), id)
			} else {
				email = extractEmail(email)
				id = getUserIdByEmail(client, email)
			}

			result, err := b.CmdHandler.VpnGetConfigUrl(email)
			if err != nil {
				log.Err(err).Msgf("Error getting vpn config. Request: %v, user: %v", botCtx.Event().Text, id)
				replyErr := response.Reply("Error getting vpn config. Please contact admin.", slacker.WithThreadReply(true))
				if replyErr != nil {
					log.Err(replyErr).Msgf("Error sending reply. Request: %v, user: %v", botCtx.Event().Text, id)
				}
				return
			}
			_, _, err = client.PostMessage(id, slack.MsgOptionText(result, false))
			if err != nil {
				log.Err(err).Msgf("Error sending direct message. Request: %v, user: %v", botCtx.Event().Text, id)
				return
			}

			err = response.Reply(fmt.Sprintf("Just sent the link to user <@%s>.", id), slacker.WithThreadReply(true))
			if err != nil {
				log.Err(err).Msgf("Error sending reply. Request: %v, user: %v", botCtx.Event().Text, id)
				return
			}
		},
	}

	b.bot.Command("vpn get", getConfig)
	b.bot.Command("vpn create <login> <email>", createProfile)
	b.bot.Command("vpn delete <email>", deleteProfile)
	b.bot.Command("vpn send <email>", sendConfig)
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
		Description: "Update parameter for domain. Available params: expire (no value), basic-auth(true|false), ip (<ip>), full-ssl(true|false), port <port>.",
		Examples:    []string{"update <param> <value>", "update expire", "update basic-auth true", "update ip 127.0.0.1", "update port 3000", "update full-ssl true"},
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

	b.bot.Command("domain create <IP>", createCommand)
	b.bot.Command("domain update <param> <value>", updateCommand)
	b.bot.Command("domain delete", deleteCommand)
}

func (b *Config) defineVpnEUCommands() {
	getConfig := &slacker.CommandDefinition{
		Description: "Get your personal EU vpn config for X hours. Possible values: 1,2,4. Without params creates account for 1 hour.",
		Examples:    []string{"vpn eu get <hours> <duration>", "vpn eu get", "vpn eu get hours 1"},
		Handler: func(botCtx slacker.BotContext, request slacker.Request, response slacker.ResponseWriter) {
			param := request.StringParam("param", "hours")
			value, err := strconv.Atoi(request.StringParam("value", "1"))
			if err != nil {
				log.Err(err).Msgf("Error converting string to int. Request: %v, user: %v", botCtx.Event().Text, botCtx.Event().UserID)
				replyErr := response.Reply("Error converting string to int.", slacker.WithThreadReply(true))
				if replyErr != nil {
					log.Err(replyErr).Msgf("Error sending reply. Request: %v, user: %v", botCtx.Event().Text, botCtx.Event().UserID)
				}
				return
			}
			switch param {
			case "hours":
				userId := botCtx.Event().UserID
				email := getUserEmail(botCtx.APIClient(), userId)
				id := strings.TrimSuffix(strings.TrimPrefix(userId, "<@"), ">")
				userName := getUserFriendlyName(botCtx.APIClient(), id)
				result, err := b.CmdHandler.VpnGetEUConfigUrl(userName, email, id, value)
				if err != nil {
					log.Err(err).Msgf("Error getting EU vpn config url. Request: %v, user: %v", botCtx.Event().Text, userId)
					replyErr := response.Reply(fmt.Sprintf("Error getting EU vpn config url. %s", err.Error()), slacker.WithThreadReply(true))
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

				err = response.Reply("Just sent instructions in a private message.", slacker.WithThreadReply(true))
				if err != nil {
					log.Err(err).Msgf("Error sending reply. Request: %v, user: %v", botCtx.Event().Text, userId)
					return
				}
			default:
				err = response.Reply("Unsupported parameter. Possible values: `hours`.", slacker.WithThreadReply(true))
				if err != nil {
					log.Err(err).Msgf("Error sending reply. Request: %v, user: %v", botCtx.Event().Text, botCtx.Event().UserID)
					return
				}
			}
		},
	}

	b.bot.Command("vpn eu get <param> <value>", getConfig)
}
