package bot

import (
	"context"
	"fmt"
	"github.com/1k-off/dev-helper-bot/internal/app"
	"github.com/1k-off/dev-helper-bot/internal/nginx"
	"github.com/1k-off/dev-helper-bot/model"
	"github.com/shomali11/slacker"
	"strconv"
	"time"
)

func Run(config *app.Config) error {
	bot := slacker.NewClient(config.SlackBotToken, config.SlackAppToken)
	config.Bot = bot

	createCommand := &slacker.CommandDefinition{
		Description: "Create a domain for provided IP.",
		Example:     "create 127.0.0.1",
		Handler: func(botCtx slacker.BotContext, request slacker.Request, response slacker.ResponseWriter) {
			ip := request.Param("IP")
			userId := botCtx.Event().User
			err := response.Reply(createHandler(config, botCtx, userId, ip), slacker.WithThreadReply(true))
			if err != nil {
				config.Log.Err(err).Msg("")
				return
			}
		},
	}

	updateCommand := &slacker.CommandDefinition{
		Description: "Update parameter for domain. Available params: expire (no value), basicauth(true|false), ip (<ip>), full-ssl(true|false).",
		Example:     "update <param> <value>",
		Handler: func(botCtx slacker.BotContext, request slacker.Request, response slacker.ResponseWriter) {
			param := request.StringParam("param", "expire")
			value := request.StringParam("value", "")
			userId := botCtx.Event().User
			err := response.Reply(updateHandler(config, botCtx, userId, param, value), slacker.WithThreadReply(true))
			if err != nil {
				config.Log.Err(err).Msg("")
				return
			}
		},
	}

	deleteCommand := &slacker.CommandDefinition{
		Description: "Delete domain assigned to you.",
		Example:     "delete",
		Handler: func(botCtx slacker.BotContext, request slacker.Request, response slacker.ResponseWriter) {
			userId := botCtx.Event().User
			err := response.Reply(deleteHandler(config, botCtx, userId), slacker.WithThreadReply(true))
			if err != nil {
				config.Log.Err(err).Msg("")
				return
			}
		},
	}

	bot.Command("create <IP>", createCommand)
	bot.Command("update <param> <value>", updateCommand)
	bot.Command("delete", deleteCommand)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	nginx.New(config.Log)
	return bot.Listen(ctx)
}

func createHandler(config *app.Config, ctx slacker.BotContext, userId, ip string) string {
	if err := nginx.CheckIfIpAllowed(config.NginxAllowedIPs, config.NginxDeniedIPs, ip); err != nil {
		config.Log.Err(err).Msg(fmt.Sprintf("[bot] ip check error. IP: %s", ip))
		return fmt.Sprintf("<@%s> [ERROR] ip check. IP: %s, error: %v", userId, ip, err)
	}
	friendlyUserName := getUserFriendlyName(ctx.Client(), userId)
	config.Log.Debug().Msg(fmt.Sprintf("[bot] requested domain creation. UserID: %s (%s), IP: %s", userId, friendlyUserName, ip))
	fqdn := transformName(friendlyUserName) + "." + config.NginxParentDomain
	deleteDate := time.Now().Add(timeStoreDomain * time.Hour * 24 * 7)
	domain := model.Domain{
		FQDN:      fqdn,
		IP:        ip,
		UserId:    userId,
		UserName:  friendlyUserName,
		CreatedAt: time.Now(),
		DeleteAt:  deleteDate,
		BasicAuth: true,
	}
	if err := nginx.Create(domain.IP, domain.FQDN, domain.BasicAuth, domain.FullSsl); err != nil {
		config.Log.Err(err).Msg(fmt.Sprintf("[bot] config creation. UserID: %s (%s), tried IP: %s.", userId, friendlyUserName, domain.IP))
		return fmt.Sprintf("<@%s> [ERROR] config creation. error: %v", userId, err)
	}
	_ = config.Store.DomainRepository().Create(&domain)
	config.Log.Info().Msg(fmt.Sprintf("Created domain %s with IP %s. Scheduled delete date: %s.", domain.FQDN, domain.IP, domain.DeleteAt))
	return fmt.Sprintf("Created domain %s with IP %s. Scheduled delete date: %s.", domain.FQDN, domain.IP, domain.DeleteAt)

}

func updateHandler(config *app.Config, ctx slacker.BotContext, userId, param, value string) string {
	d, err := config.Store.DomainRepository().Get(userId)
	friendlyUserName := getUserFriendlyName(ctx.Client(), userId)
	if err != nil {
		config.Log.Err(err).Msg(fmt.Sprintf("[bot] can't find fomain for user %s (%s)", userId, friendlyUserName))
		return "can't find domain for you"
	}

	updateExp := func() {
		d.DeleteAt = time.Now().Add(timeStoreDomain * time.Hour * 24 * 7)
	}

	switch param {
	case "":
		updateExp()
	case "expire":
		updateExp()
	case "ip":
		ip := value
		if err = nginx.CheckIfIpAllowed(config.NginxAllowedIPs, config.NginxDeniedIPs, ip); err != nil {
			config.Log.Err(err).Msg(fmt.Sprintf("[bot] ip check. IP: %s", ip))
			return fmt.Sprintf("<@%s> [ERROR] ip check. IP: %s, error: %v", userId, ip, err)
		}
		d.IP = ip
		if err = updateNginxConf(d.FQDN, d.IP, d.BasicAuth, d.FullSsl); err != nil {
			config.Log.Err(err).Msg("")
			return err.Error()
		}
	case "basicauth":
		ba, err := strconv.ParseBool(value)
		if err != nil {
			config.Log.Err(err).Msg("")
			return err.Error()
		}
		d.BasicAuth = ba
		if err := updateNginxConf(d.FQDN, d.IP, d.BasicAuth, d.FullSsl); err != nil {
			return err.Error()
		}
	case "full-ssl":
		fs, err := strconv.ParseBool(value)
		if err != nil {
			config.Log.Err(err).Msg("")
			return err.Error()
		}
		d.FullSsl = fs
		if err := updateNginxConf(d.FQDN, d.IP, d.BasicAuth, d.FullSsl); err != nil {
			return err.Error()
		}
	default:
		return "unknown parameter"
	}

	err = config.Store.DomainRepository().Update(d)
	if err != nil {
		config.Log.Err(err).Msg("")
		return err.Error()
	}
	return "updated"
}

func deleteHandler(config *app.Config, ctx slacker.BotContext, userId string) string {
	friendlyUserName := getUserFriendlyName(ctx.Client(), userId)
	config.Log.Debug().Msg(fmt.Sprintf("[bot] requested domain deletion. UserID: %s (%s)", userId, friendlyUserName))
	fqdn := transformName(friendlyUserName) + "." + config.NginxParentDomain
	config.Store.DomainRepository().DeleteByFqdn(fqdn)
	err := nginx.Delete(fqdn)
	if err != nil {
		config.Log.Err(err).Msg("")
	}
	config.Log.Info().Msg(fmt.Sprintf("[bot] deleted domain %s. UserID: %s (%s)", fqdn, userId, friendlyUserName))
	return fmt.Sprintf("Deleted domain %s", fqdn)

}
