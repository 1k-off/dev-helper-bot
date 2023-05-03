package bot

import (
	"github.com/rs/zerolog/log"
	"github.com/slack-go/slack"
)

func getAdminUserIDs(c *slack.Client, emails []string) []string {
	var adminUserIDs []string
	for _, email := range emails {
		adminUserIDs = append(adminUserIDs, getUserIdByEmail(c, email))
	}
	return adminUserIDs
}

func getUserEmail(c *slack.Client, userId string) string {
	user, err := c.GetUserInfo(userId)
	if err != nil || len(user.Profile.Email) == 0 {
		log.Error().Err(err).Msgf("ID: %s", userId)
		return ""
	}
	return user.Profile.Email
}

func getUserIdByEmail(c *slack.Client, email string) string {
	users, err := c.GetUsers()
	if err != nil {
		log.Error().Err(err).Msgf("Email: %s", email)
		return ""
	}
	for _, user := range users {
		if user.Profile.Email == email {
			return user.ID
		}
	}
	log.Error().Msgf("User not found. Email: %s", email)
	return ""
}

func getUserFriendlyName(c *slack.Client, userId string) string {
	user, err := c.GetUserInfo(userId)
	if err != nil {
		log.Error().Err(err).Msgf("ID: %s", userId)
		return userId
	}
	return user.Profile.DisplayName
}
