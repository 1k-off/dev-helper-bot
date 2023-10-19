package caddy_svc

import (
	"fmt"
	"github.com/1k-off/dev-helper-bot/internal/entities"
	"github.com/1k-off/dev-helper-bot/internal/webserver"
	"github.com/rs/zerolog/log"
	"html/template"
	"os"
	"os/exec"
)

const (
	configBasePath = "./caddy"
)

type Webserver struct{}

func New() *Webserver {
	return &Webserver{}
}

func init() {
	if _, err := os.Stat(configBasePath); os.IsNotExist(err) {
		err := os.Mkdir(configBasePath, os.ModeDir)
		if err != nil {
			log.Fatal().Err(err).Msg("Error creating config directory")
			return
		}
	}
}

func (w *Webserver) Create(c *entities.Domain) error {
	ba := "Restricted"
	if !c.BasicAuth {
		ba = "off"
	}

	scheme := webserver.SchemeHttp
	if c.FullSsl {
		scheme = webserver.SchemeHttps
		if c.Port == "" || c.Port == "80" {
			c.Port = "443"
		}
	} else {
		if c.Port == "" || c.Port == "443" {
			c.Port = "80"
		}
	}

	configData := map[string]string{
		"ip":        c.IP,
		"domain":    c.FQDN,
		"basicauth": ba,
		"scheme":    scheme,
		"port":      c.Port,
	}
	if _, err := os.Stat(configBasePath + "/" + c.FQDN); os.IsNotExist(err) {
		t, err := template.ParseFiles("config/caddy.conf.tpl")
		if err != nil {
			return err
		}
		f, err := os.Create(configBasePath + "/" + c.FQDN)
		if err != nil {
			return err
		}
		err = t.Execute(f, configData)
		if err != nil {
			return err
		}
		err = f.Close()
		if err != nil {
			return err
		}
		if !webserver.Debug {
			if err := reloadServer(); err != nil {
				return err
			}
		}
		log.Info().Msg(fmt.Sprintf("[caddy] created config. ClientIP: %v, Domain: %s.", c.IP, c.FQDN))
		return nil
	} else {
		return webserver.ErrConfigExists
	}

}

func (w *Webserver) Delete(domain string) error {
	if err := os.Remove(configBasePath + "/" + domain); err != nil {
		return err
	}
	if !webserver.Debug {
		if err := reloadServer(); err != nil {
			return err
		}
	}
	log.Info().Msg(fmt.Sprintf("[caddy] deleted config. Domain: %s.", domain))
	return nil
}

func reloadServer() error {
	exe := "caddy"
	cmd := exec.Command(exe, "validate", "--config", "/etc/caddy/Caddyfile")
	_, err := cmd.Output()
	if err != nil {
		return err
	}
	cmd = exec.Command(exe, "reload", "--config", "/etc/caddy/Caddyfile")
	_, err = cmd.Output()
	if err != nil {
		return err
	}
	return nil
}
