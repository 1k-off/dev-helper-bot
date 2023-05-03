package handlers

import (
	"github.com/1k-off/dev-helper-bot/internal/config"
	"github.com/1k-off/dev-helper-bot/internal/store"
	"github.com/1k-off/dev-helper-bot/pkg/pritunl"
	"time"
)

type Handler struct {
	PritunlClient *pritunl.Client
	NginxConfig   config.Nginx
	Store         store.Store
	Timezone      *time.Location
}

func New(c *pritunl.Client, nc config.Nginx, s store.Store, timezone *time.Location) *Handler {
	return &Handler{
		PritunlClient: c,
		NginxConfig:   nc,
		Store:         s,
		Timezone:      timezone,
	}
}
