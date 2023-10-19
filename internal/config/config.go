package config

import (
	"fmt"
	"github.com/1k-off/dev-helper-bot/internal/webserver"
	caddy_svc "github.com/1k-off/dev-helper-bot/internal/webserver/caddy-svc"
	"github.com/1k-off/dev-helper-bot/internal/webserver/nginx"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"time"
)

type Config struct {
	App       App       `mapstructure:"app"`
	Pritunl   Pritunl   `mapstructure:"pritunl"`
	Webserver Webserver `mapstructure:"webserver"`
	Slack     Slack     `mapstructure:"slack"`
	Timezone  *time.Location
}

type App struct {
	AdminEmails                []string `mapstructure:"admin_emails"`
	DatasourceConnectionString string   `mapstructure:"datasource_connection_string"`
	LogLevel                   string   `mapstructure:"log_level"`
	Timezone                   string   `mapstructure:"timezone"`
}

type Pritunl struct {
	Host           string `mapstructure:"host"`
	Token          string `mapstructure:"token"`
	Secret         string `mapstructure:"secret"`
	Organization   string `mapstructure:"organization"`
	WelcomeMessage string `mapstructure:"welcome_message"`
	HostEU         string `mapstructure:"host_eu"`
	TokenEU        string `mapstructure:"token_eu"`
	SecretEU       string `mapstructure:"secret_eu"`
	OrganizationEU string `mapstructure:"organization_eu"`
}

type Webserver struct {
	ParentDomain   string              `mapstructure:"parent_domain"`
	AllowedSubnets []string            `mapstructure:"allowed_subnets"`
	DeniedIPs      []string            `mapstructure:"denied_ips"`
	Nginx          bool                `mapstructure:"nginx"`
	Caddy          bool                `mapstructure:"caddy"`
	Service        webserver.Webserver `mapstructure:"-"`
}

type Slack struct {
	AuthToken string `mapstructure:"auth_token"`
	AppToken  string `mapstructure:"app_token"`
	Channel   string `mapstructure:"channel"`
}

func newDefaultConfig() *Config {
	return &Config{
		App: App{
			LogLevel: "info",
			Timezone: "Europe/Kyiv",
		},
	}
}

func Load(path string) (*Config, error) {
	cfg := newDefaultConfig()
	viper.AddConfigPath(path)
	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		log.Debug().Msgf("failed to read config: %s", err)
		return nil, err
	}
	if err := viper.Unmarshal(cfg); err != nil {
		log.Debug().Msgf("failed to unmarshal config: %s", err)
		return nil, err
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	setupLogger(cfg.App.LogLevel)
	tz, err := time.LoadLocation(cfg.App.Timezone)
	if err != nil {
		log.Err(err).Msgf("failed to load timezone: %s", cfg.App.Timezone)
		tz, _ = time.LoadLocation("Europe/Kyiv")
	}
	cfg.Timezone = tz
	if cfg.Webserver.Nginx {
		cfg.Webserver.Service = nginx.New()
	} else if cfg.Webserver.Caddy {
		cfg.Webserver.Service = caddy_svc.New()
	}
	return cfg, nil
}

func (c *Config) Validate() error {
	if err := validateLogLevel(c.App.LogLevel); err != nil {
		log.Debug().Msgf("failed to validate log level: %s", err)
		return err
	}
	if !c.Webserver.Nginx && !c.Webserver.Caddy {
		return fmt.Errorf("[config] both nginx and caddy are disabled")
	}
	if c.Webserver.Nginx && c.Webserver.Caddy {
		return fmt.Errorf("[config] both nginx and caddy are enabled")
	}
	return nil
}
