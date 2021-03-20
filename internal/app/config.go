package app

import (
	"github.com/shomali11/slacker"
	"github.com/souladm/dev-helper-bot/internal/store"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
)

type Config struct {
	AppAdminEmail []string `yaml:"app_admin_email"`
	SlackToken         string `yaml:"slack_token"`
	SlackChannelName string `yaml:"slack_channel"`
	NginxAllowedIPs []string `yaml:"nginx_allowed_subnet"`
	NginxDeniedIPs []string `yaml:"nginx_denied_ips"`
	NginxParentDomain string `yaml:"nginx_parent_domain"`
	Store *store.Store
	Bot *slacker.Slacker
}

func newConfig() *Config {
	return &Config{
	}
}

func Configure() (*Config, error) {
	c := newConfig()
	yamlFile, err := ioutil.ReadFile("data/config.yml")
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		return nil, err
	}
	s, err := store.New()
	if err != nil {
		log.Println("[ERROR] can't connect to database.")
		panic(err)
	}
	c.Store = s
	return c, nil
}