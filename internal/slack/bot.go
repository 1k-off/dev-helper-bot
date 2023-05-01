package slack

import (
	"context"
	"fmt"
	"github.com/1k-off/dev-helper-bot/internal/handlers"
	"github.com/rs/zerolog/log"
	"github.com/shomali11/slacker"
	"github.com/slack-go/slack"
	"strings"
)

type BotConfig struct {
	bot *slacker.Slacker
	//slackClient    *slack.Client
	SlackAuthToken string
	SlackAppToken  string
	Ctx            context.Context
	Cancel         context.CancelFunc
	AdminUserIDs   []string
	CmdHandler     *handlers.Handler
}

func New(authToken, appToken string, adminUserEmails []string, cmdHandler *handlers.Handler) *BotConfig {
	ctx, cancel := context.WithCancel(context.Background())
	//socketClient := socketmode.New(
	//	slackClient,
	//	socketmode.OptionDebug(false),
	//	socketmode.OptionLog(log.New(os.Stdout, "socketmode: ", log.Lshortfile|log.LstdFlags)),
	//)
	bot := slacker.NewClient(authToken, appToken)
	adminUserIds := getAdminUserIDs(bot.APIClient(), adminUserEmails)

	return &BotConfig{
		bot:            bot,
		SlackAuthToken: authToken,
		SlackAppToken:  appToken,
		Ctx:            ctx,
		Cancel:         cancel,
		CmdHandler:     cmdHandler,
		AdminUserIDs:   adminUserIds,
	}
}

func (b *BotConfig) Run() error {
	b.defineVpnCommands()
	b.defineDomainCommands()
	return b.bot.Listen(b.Ctx)
}

func (b *BotConfig) Stop() {
	b.Cancel()
}

func (b *BotConfig) defineVpnCommands() {
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

func (b *BotConfig) defineDomainCommands() {
	createConfig := &slacker.CommandDefinition{
		Description: "Create a domain for provided IP.",
		Examples:    []string{"create 127.0.0.1"},
		Handler: func(botCtx slacker.BotContext, request slacker.Request, response slacker.ResponseWriter) {
			ip := request.Param("IP")
			userId := botCtx.Event().UserID
			id := strings.TrimSuffix(strings.TrimPrefix(userId, "<@"), ">")
			userName := getUserFriendlyName(botCtx.APIClient(), id)
			result, err := b.CmdHandler.DomainCreate(id, userName, ip)
			if err != nil {
				log.Err(err).Msgf("Error creating domain. Request: %v, user: %v", botCtx.Event().Text, botCtx.Event().UserID)
				replyErr := response.Reply(fmt.Sprintf("Error creating domain. %v", err), slacker.WithThreadReply(true))
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

	updateConfig := &slacker.CommandDefinition{
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

	b.bot.Command("create <IP>", createConfig)
	b.bot.Command("domain create <IP>", createConfig)
	b.bot.Command("update <param> <value>", updateConfig)
	b.bot.Command("domain update <param> <value>", updateConfig)
	b.bot.Command("delete", deleteCommand)
	b.bot.Command("domain delete", deleteCommand)
}
