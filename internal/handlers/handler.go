package handlers

import (
	"github.com/1k-off/dev-helper-bot/internal/config"
	"github.com/1k-off/dev-helper-bot/internal/store"
	"github.com/1k-off/dev-helper-bot/pkg/pritunl"
	"time"
)

type Handler struct {
	PritunlClient    *pritunl.Client
	MessageTemplates map[string]string
	Webserver        config.Webserver
	Store            store.Store
	Timezone         *time.Location
}

func New(c *pritunl.Client, wc config.Webserver, s store.Store, timezone *time.Location, msgTemplates map[string]string) *Handler {
	return &Handler{
		PritunlClient:    c,
		MessageTemplates: msgTemplates,
		Webserver:        wc,
		Store:            s,
		Timezone:         timezone,
	}
}
