package config

import (
	"github.com/rs/zerolog"
	"strings"
)

func setupLogger(logLevel string) {
	level, _ := zerolog.ParseLevel(strings.ToLower(logLevel))
	zerolog.SetGlobalLevel(level)
}

func validateLogLevel(logLevel string) error {
	_, err := zerolog.ParseLevel(strings.ToLower(logLevel))
	if err != nil {
		return err
	}
	return nil
}
