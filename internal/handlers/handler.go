package handlers

import (
	"devhelper/internal/config"
	"devhelper/internal/store"
	"devhelper/pkg/pritunl"
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
