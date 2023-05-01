package handlers

import (
	"github.com/1k-off/dev-helper-bot/internal/config"
	"github.com/1k-off/dev-helper-bot/internal/store"
	"github.com/1k-off/dev-helper-bot/pkg/pritunl"
)

type Handler struct {
	PritunlClient *pritunl.Client
	NginxConfig   config.Nginx
	Store         store.Store
}

func New(c *pritunl.Client, nc config.Nginx, s store.Store) *Handler {
	return &Handler{
		PritunlClient: c,
		NginxConfig:   nc,
		Store:         s,
	}
}
