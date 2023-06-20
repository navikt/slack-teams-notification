package config

import (
	"github.com/kelseyhightower/envconfig"
)

type (
	Config struct {
		Log   *log
		Slack *slackConfig
		Teams *teamsConfig
	}

	log struct {
		Format string `envconfig:"LOG_FORMAT" default:"text"`
		Level  string `envconfig:"LOG_LEVEL" default:"DEBUG"`
	}

	slackConfig struct {
		APIToken string `envconfig:"SLACK_API_TOKEN"`
	}

	teamsConfig struct {
		APIToken    string `envconfig:"TEAMS_BACKEND_API_TOKEN"`
		BackendURL  string `envconfig:"TEAMS_BACKEND_URL" default:"http://localhost:3000/query"`
		FrontendURL string `envconfig:"TEAMS_FRONTEND_URL" default:"http://localhost:3001/"`
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
