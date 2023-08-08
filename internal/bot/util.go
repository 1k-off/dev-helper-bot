package bot

import (
	"github.com/rs/zerolog/log"
	"github.com/slack-go/slack"
	"regexp"
	"strings"
)

func contains(list []string, element string) bool {
	for _, value := range list {
		if value == element {
			return true
		}
	}
	return false
}

func extractEmail(text string) string {
	re := regexp.MustCompile(`<mailto:(.*?)\|.*?(?:>|$)`)
	matches := re.FindStringSubmatch(text)
	if len(matches) < 2 {
		return ""
	}
	email := strings.TrimSpace(matches[1])
	return email
}

// setNotified sets the notified flag for a user in the local cache
func (b *Config) setNotified(userId string) error {
	return b.Cache.Set(cacheNamespaceNotified, userId, "true")
}

// isNotified checks if a user has been notified
func (b *Config) isNotified(userId string) (bool, error) {
	return b.Cache.Has(cacheNamespaceNotified, userId)
}

// clearNotified clears the notified flag for a user in the local cache
func (b *Config) clearNotified(userId string) error {
	return b.Cache.Delete(cacheNamespaceNotified, userId)
}

// sendVpnWelcomeMessage sends a welcome message if it is set to a user when creating new config
func sendVPNWelcomeMessage(client *slack.Client, userId, message string) error {
	if message == "" {
		log.Info().Msg("No VPN welcome message set")
		return nil
	}
	_, _, err := client.PostMessage(userId, slack.MsgOptionText(message, false))
	if err != nil {
		return err
	}
	return nil
}

// getSlackUserIdByEmail gets the Slack user ID for a given email address
func getSlackUserIdByEmail(client *slack.Client, email string) (string, error) {
	users, err := client.GetUsers()
	if err != nil {
		return "", err
	}
	for _, user := range users {
		if user.Profile.Email == email {
			return user.ID, nil
		}
	}
	return "", nil
}
