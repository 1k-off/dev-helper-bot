package app

import (
	"fmt"
	"github.com/1k-off/dev-helper-bot/internal/store"
	"github.com/1k-off/dev-helper-bot/internal/store/mongostore"
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog"
	"github.com/shomali11/slacker"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
)

type Config struct {
	AppAdminEmail              []string `yaml:"app_admin_email"`
	SlackBotToken              string   `yaml:"slack_bot_token"`
	SlackAppToken              string   `yaml:"slack_app_token"`
	SlackChannelName           string   `yaml:"slack_channel"`
	PritunlHost                string   `yaml:"pritunl_host"`
	PritunlToken               string   `yaml:"pritunl_token"`
	PritunlSecret              string   `yaml:"pritunl_secret"`
	PritunlOrganization        string   `yaml:"pritunl_organization"`
	NginxAllowedIPs            []string `yaml:"nginx_allowed_subnet"`
	NginxDeniedIPs             []string `yaml:"nginx_denied_ips"`
	NginxParentDomain          string   `yaml:"nginx_parent_domain"`
	Store                      store.Data
	Bot                        *slacker.Slacker
	Log                        zerolog.Logger
	LogLevel                   string `yaml:"log_level" validate:"eq=debug|eq=info|eq=error"`
	DatasourceConnectionString string `yaml:"datasource_connection_string" validate:"required"`
}

var (
	logger zerolog.Logger
)

func newConfig() *Config {
	logger = zerolog.New(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "2006-01-02 15:04:05",
	}).With().Timestamp().Logger()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	return &Config{
		Log:      logger,
		LogLevel: LogLevelInfo,
	}
}

// Startup is a function to validate config file, apply configuration parameters, connect to database, etc
func Startup(configPath string) *Config {
	app := Configure(configPath)
	app.Validate()
	app.Log = logger
	app.setLogLevel(app.LogLevel)

	app.Store = mongostore.New(app.DatasourceConnectionString, app.Log)
	return app
}

// Configure reads config file
func Configure(configPath string) *Config {
	app := newConfig()
	// Parse configuration from yaml file
	configFile, err := ioutil.ReadFile(configPath)
	if err != nil {
		app.Log.Fatal().Err(err).Msg("")
	}
	err = yaml.Unmarshal(configFile, app)
	if err != nil {
		app.Log.Fatal().Err(err).Msg("")
	}
	return app
}

// setLogLevel setting up global log level
func (app *Config) setLogLevel(level string) {
	switch level {
	case LogLevelDebug:
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case LogLevelInfo:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case LogLevelError:
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
	app.Log.Info().Msg(fmt.Sprintf("Using log level: %s", level))
}

// Validate is a function to validate configuration parameters
func (app *Config) Validate() {
	validate := validator.New()
	err := validate.Struct(app)
	if err != nil {
		validationErrors := err.(validator.ValidationErrors)
		if len(validationErrors) != 0 {
			for _, e := range validationErrors {
				app.Log.Error().Msg(fmt.Sprintf("%s", e))
			}
		}
		app.Log.Fatal().Msg("Config validation failed")
	}
}

// Stop gracefully finishes application work
func (app *Config) Stop() {
	app.Store.Close()
}
