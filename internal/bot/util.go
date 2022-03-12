package bot

import (
	"github.com/1k-off/dev-helper-bot/internal/nginx"
	"github.com/marstr/guid"
	"github.com/slack-go/slack"
	"regexp"
	"strings"
)

func getUserFriendlyName(c *slack.Client, userId string) string {
	user, err := c.GetUserInfo(userId)
	if err != nil || len(user.Profile.DisplayName) == 0 {
		return userWithEmptyDisplayName
	}
	return user.Profile.DisplayName
}

func getRandomString() string {
	return guid.NewGUID().String()
}

func transformName(name string) string {
	if len(name) == 0 || name == userWithEmptyDisplayName {
		return getRandomString()
	}
	nameArr := strings.FieldsFunc(name, split)
	firstName := nameArr[0]
	lastName := nameArr[1]
	allowedSymbols, _ := regexp.Compile("[^A-Za-z0-9]+")
	subdomain := string(firstName[0]) + "-" + lastName
	return allowedSymbols.ReplaceAllString(strings.ToLower(subdomain), "-")
}

func split(r rune) bool {
	return r == ' ' || r == '.'
}

func updateNginxConf(domain, ip string, basicauth, fullSsl bool) error {
	err := nginx.Delete(domain)
	if err != nil {
		return err
	}
	if err = nginx.Create(ip, domain, basicauth, fullSsl); err != nil {
		return err
	}
	return nil
}
