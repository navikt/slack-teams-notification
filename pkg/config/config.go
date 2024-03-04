package config

import (
	"github.com/kelseyhightower/envconfig"
)

type (
	Config struct {
		Log   *log
		Slack *slackConfig
		Teams *teamsConfig

		// TeamsFilter Specify a list of team slugs that should be included when sending notifications. All other teams
		// will be ignored when set.
		TeamsFilter []string `envconfig:"TEAMS_FILTER"`
	}

	log struct {
		// Format Log format
		Format string `envconfig:"LOG_FORMAT" default:"text"`

		// Level Log level
		Level string `envconfig:"LOG_LEVEL" default:"DEBUG"`
	}

	slackConfig struct {
		// APIToken API token used with the Slack API
		APIToken string `envconfig:"SLACK_API_TOKEN"`
	}

	teamsConfig struct {
		// APIToken API token used with the teams-backend GraphQL API.
		APIToken string `envconfig:"NAIS_API_TOKEN"`

		// BackendURL URL to the query endpoint for the teams-backend service.
		BackendURL string `envconfig:"NAIS_API_URL" default:"http://localhost:3000/query"`

		// FrontendURL URL to the root of the teams-frontend service. Used for links in the notification message sent to
		// the owners of the teams.
		FrontendURL string `envconfig:"CONSOLE_FRONTEND_URL" default:"http://localhost:3001/"`
	}
)

func New() (*Config, error) {
	cfg := &Config{}
	err := envconfig.Process("", cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
