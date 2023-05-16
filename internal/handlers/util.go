package handlers

import (
	"github.com/marstr/guid"
	"regexp"
	"strings"
)

func transformName(name string) string {
	if len(name) == 0 {
		return getRandomString()
	}
	nameArr := strings.FieldsFunc(name, split)
	firstName := nameArr[0]
	lastName := nameArr[1]
	allowedSymbols, _ := regexp.Compile("[^A-Za-z0-9]+")
	subdomain := string(firstName[0]) + "-" + lastName
	return allowedSymbols.ReplaceAllString(strings.ToLower(subdomain), "-")
}

func getRandomString() string {
	return guid.NewGUID().String()
}

func split(r rune) bool {
	return r == ' ' || r == '.'
}
