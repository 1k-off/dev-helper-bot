package bot

import (
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
	re := regexp.MustCompile(`<mailto:(.*?)\|(.*?)>`)
	matches := re.FindStringSubmatch(text)
	if len(matches) < 3 {
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
