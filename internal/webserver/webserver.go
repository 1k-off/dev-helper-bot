package webserver

import (
	"fmt"
	"github.com/1k-off/dev-helper-bot/internal/entities"
	caddy_svc "github.com/1k-off/dev-helper-bot/internal/webserver/caddy-svc"
	"github.com/1k-off/dev-helper-bot/internal/webserver/nginx"
	"github.com/rs/zerolog/log"
	"html/template"
	"net"
	"os"
	"strconv"
)

type IP struct {
	Allowed []*net.IPNet
	Denied  []net.IP
}

var (
	privateIPBlocks []*net.IPNet
	sysIP           IP
	Debug           bool
)

type Webserver interface {
	Create(c *entities.Domain) error
	Delete(domain string) error
}

type Server struct {
	kind         string
	templatePath string
}

func init() {
	debugModeString := os.Getenv("OOOPS_DEBUG")
	d, err := strconv.ParseBool(debugModeString)
	if err != nil {
		Debug = false
		log.Err(err).Msg("Error parsing OOOPS_DEBUG env variable")
	}
	Debug = d
	for _, cidr := range []string{
		"127.0.0.0/8",    // IPv4 loopback
		"10.0.0.0/8",     // RFC1918
		"172.16.0.0/12",  // RFC1918
		"192.168.0.0/16", // RFC1918
		"169.254.0.0/16", // RFC3927 link-local
		"::1/128",        // IPv6 loopback
		"fe80::/10",      // IPv6 link-local
		"fc00::/7",       // IPv6 unique local addr
	} {
		_, block, err := net.ParseCIDR(cidr)
		if err != nil {
			panic(fmt.Errorf("parse error on %q: %v", cidr, err))
		}
		privateIPBlocks = append(privateIPBlocks, block)
	}
}

func New(s string) *Server {
	if _, err := os.Stat(configBasePath + s); os.IsNotExist(err) {
		err := os.Mkdir(configBasePath+s, os.ModeDir)
		if err != nil {
			log.Fatal().Err(err).Msg("Error creating config directory")
			return nil
		}
	}
	return &Server{
		kind:         s,
		templatePath: "./config/" + s + ".conf.tpl",
	}
}

func (s *Server) Create(c *entities.Domain) error {
	ba := "Restricted"
	if !c.BasicAuth {
		ba = "off"
	}

	scheme := SchemeHttp
	if c.FullSsl {
		scheme = SchemeHttps
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
	if _, err := os.Stat(configBasePath + s.kind + "/" + c.FQDN); os.IsNotExist(err) {
		t, err := template.ParseFiles(s.templatePath)
		if err != nil {
			return err
		}
		f, err := os.Create(configBasePath + s.kind + "/" + c.FQDN)
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
		if !Debug {
			if err := s.reload(); err != nil {
				return err
			}
		}
		log.Info().Msg(fmt.Sprintf("[%s] created config. ClientIP: %v, Domain: %s.", s.kind, c.IP, c.FQDN))
		return nil
	} else {
		return ErrConfigExists
	}
}

func (s *Server) Delete(domain string) error {
	if err := os.Remove(configBasePath + s.kind + "/" + domain); err != nil {
		return err
	}
	if !Debug {
		if err := s.reload(); err != nil {
			return err
		}
	}
	log.Info().Msg(fmt.Sprintf("[%s] deleted config. Domain: %s.", s.kind, domain))
	return nil
}

func (s *Server) reload() error {
	switch s.kind {
	case ServerCaddy:
		return caddy_svc.Reload()
	case ServerNginx:
		return nginx.Reload()
	}
	return nil
}
