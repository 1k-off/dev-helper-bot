package bot

import (
	"context"
	"fmt"
	"github.com/shomali11/slacker"
	"github.com/souladm/dev-helper-bot/internal/app"
	"github.com/souladm/dev-helper-bot/internal/nginx"
	"github.com/souladm/dev-helper-bot/model"
	"log"
	"strconv"
	"time"
)

func Run(config *app.Config) error {
	bot := slacker.NewClient(config.SlackToken)
	config.Bot = bot

	createCommand := &slacker.CommandDefinition{
		Description: "Create a domain for provided IP.",
		Example:     "create 127.0.0.1",
		Handler: func(botCtx slacker.BotContext, request slacker.Request, response slacker.ResponseWriter) {
			ip := request.Param("IP")
			userId := botCtx.Event().User
			msg := createHandler(config, botCtx, userId, ip)
			response.Reply(msg)
		},
	}

	updateCommand := &slacker.CommandDefinition{
		Description: "Update parameter for domain. Available params: expire (no value), basicauth(true|false), ip (<ip>).",
		Example:     "update <param> <value>",
		Handler: func(botCtx slacker.BotContext, request slacker.Request, response slacker.ResponseWriter) {
			param := request.StringParam("param", "expire")
			value := request.StringParam("value", "")
			userId := botCtx.Event().User
			response.Reply(updateHandler(config, botCtx, userId, param, value), slacker.WithThreadReply(true))
		},
	}

	bot.Command("create <IP>", createCommand)
	bot.Command("update <param> <value>", updateCommand)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	return bot.Listen(ctx)
}

func createHandler(config *app.Config, ctx slacker.BotContext, userId, ip string) string {
	if err := nginx.CheckIfIpAllowed(config.NginxAllowedIPs, config.NginxDeniedIPs, ip); err != nil {
		log.Println(fmt.Sprintf("[ERROR] ip check. IP: %s, error: %v", ip, err))
		return fmt.Sprintf("<@%s> [ERROR] ip check. IP: %s, error: %v", userId, ip, err)
	}
	friendlyUserName := getUserFriendlyName(ctx.Client(), userId)
	log.Println(fmt.Sprintf("[INFO] Requested domain creation. UserID: %s (%s), IP: %s", userId, friendlyUserName, ip))
	fqdn := transformName(friendlyUserName) + "." + config.NginxParentDomain
	deleteDate := time.Now().Add(timeStoreDomain * time.Hour * 24 * 7).Format("2006-01-02")
	domain := model.Domain{
		FQDN: fqdn,
		IP: ip,
		UserId: userId,
		UserName: friendlyUserName,
		CreatedAt: time.Now().Format("2006-01-02 15:04:05"),
		DeleteAt: deleteDate,
		BasicAuth: true,
	}
	if err := nginx.Create(domain.IP, domain.FQDN, domain.BasicAuth); err != nil {
		log.Println(fmt.Sprintf("[ERROR] config creation. UserID: %s (%s), tried IP: %s, error: %v", userId, friendlyUserName, domain.IP, err))
		return fmt.Sprintf("<@%s> [ERROR] config creation. error: %v", userId, err)
	}
	_ = config.Store.Domain().Create(&domain)
	return fmt.Sprintf("Created domain %s with IP %s. Scheduled delete date: %s.", domain.FQDN, domain.IP, domain.DeleteAt)

}

func updateHandler(config *app.Config, ctx slacker.BotContext, userId, param, value string) string {
	d, err := config.Store.Domain().Get(userId)
	if err != nil {
		return "can't find domain for you"
	}

	updateExp := func(){
		d.DeleteAt = time.Now().Add(timeStoreDomain * time.Hour * 24 * 7).Format("2006-01-02")
	}

	switch param {
	case "":
		updateExp()
	case "expire":
		updateExp()
	case "ip":
		ip := value
		if err := nginx.CheckIfIpAllowed(config.NginxAllowedIPs, config.NginxDeniedIPs, ip); err != nil {
			log.Println(fmt.Sprintf("[ERROR] ip check. IP: %s, error: %v", ip, err))
			return fmt.Sprintf("<@%s> [ERROR] ip check. IP: %s, error: %v", userId, ip, err)
		}
		d.IP = ip
		if err := updateNginxConf(d.FQDN, d.IP, d.BasicAuth); err != nil {
			return err.Error()
		}
	case "basicauth":
		ba, err:= strconv.ParseBool(value)
		if err != nil {
			return err.Error()
		}
		d.BasicAuth = ba
		if err := updateNginxConf(d.FQDN, d.IP, d.BasicAuth); err != nil {
			return err.Error()
		}
	default:
		return "unknown parameter"
	}

	err = config.Store.Domain().Update(d)
	if err != nil {
		return err.Error()
	}
	return "updated"
}