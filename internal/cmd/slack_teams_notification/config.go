package slack_teams_notification

import (
	"context"
	"fmt"

	"github.com/sethvargo/go-envconfig"
)

type LogConfig struct {
	// Format is the log format.
	Format string `env:"LOG_FORMAT,default=json"`

	// Level is the logging level.
	Level string `env:"LOG_LEVEL,default=info"`
}

type SlackConfig struct {
	// ApiToken is the API token used with the Slack API.
	ApiToken string `env:"SLACK_API_TOKEN,required"`
}

type NaisApiConfig struct {
	// ApiToken is the API token used with the Nais API.
	ApiToken string `env:"NAIS_API_TOKEN,required"`

	// Endpoint is the URL to the GraphQL API.
	Endpoint string `env:"NAIS_API_ENDPOINT,default=https://console.nav.cloud.nais.io/graphql"`

	// ConsoleUrl is the URL to the root of the Console frontend. Used for links in the notification message sent to the
	// owners of the teams.
	ConsoleUrl string `env:"CONSOLE_URL,default=https://console.nav.cloud.nais.io/"`

	// TeamsFilter is a list that can be supplied to only send a message to the teams included in the filter.
	TeamsFilter []string `env:"TEAMS_FILTER"`
}

type config struct {
	Log     *LogConfig
	Slack   *SlackConfig
	NaisApi *NaisApiConfig
}

func newConfig(ctx context.Context) (*config, error) {
	cfg := &config{}
	if err := envconfig.Process(ctx, cfg); err != nil {
		return nil, err
	}

	if err := validateConfig(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func validateConfig(cfg *config) error {
	if cfg.Slack.ApiToken == "" {
		return fmt.Errorf("missing Slack API token")
	}

	if cfg.NaisApi.ApiToken == "" {
		return fmt.Errorf("missing Nais API token")
	}

	return nil
}
