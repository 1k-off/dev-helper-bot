package slack

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
