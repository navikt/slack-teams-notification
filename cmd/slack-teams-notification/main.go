package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/nais/slack-teams-notification/internal/config"
	"github.com/nais/slack-teams-notification/internal/slack"
	"github.com/nais/slack-teams-notification/internal/teams"
	"github.com/sirupsen/logrus"
)

func main() {
	err := run()
	if err != nil {
		logrus.WithError(err).Errorf("fatal")
		os.Exit(1)
	}
}

func run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := config.New()
	if err != nil {
		return err
	}

	err = setupLogging(cfg.Log.Format, cfg.Log.Level)
	if err != nil {
		return err
	}

	naisTeams, err := teams.
		NewClient(cfg.Teams.BackendURL, cfg.Teams.APIToken).
		GetTeams(cfg.TeamsFilter)
	if err != nil {
		return err
	}

	if len(naisTeams) == 0 {
		logrus.Errorf("No NAIS teams returned from the server")
		return nil
	}

	slack.NewNotifier(cfg.Slack.APIToken, cfg.Teams.FrontendURL).NotifyTeams(ctx, naisTeams)

	return nil
}

func setupLogging(format, level string) error {
	switch format {
	case "json":
		logrus.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339Nano,
		})
	case "text":
		logrus.SetFormatter(&logrus.TextFormatter{
			TimestampFormat: time.RFC3339Nano,
		})
	default:
		return fmt.Errorf("invalid log format: %s", format)
	}

	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		return err
	}

	logrus.SetLevel(logLevel)
	return nil
}
