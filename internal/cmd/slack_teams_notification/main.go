package slack_teams_notification

import (
	"context"
	"fmt"
	"os"

	"github.com/nais/slack-teams-notification/internal/naisapi"
	"github.com/nais/slack-teams-notification/internal/slack"
	"github.com/sirupsen/logrus"
)

const (
	exitCodeSuccess = iota
	exitCodeEnvFileError
	exitCodeConfigError
	exitCodeLoggerError
	exitCodeRunError
)

func Run(ctx context.Context) {
	log := logrus.StandardLogger()
	log.SetFormatter(&logrus.JSONFormatter{})

	if err := loadEnvFile(log); err != nil {
		log.WithError(err).Errorf("error loading .env file")
		os.Exit(exitCodeEnvFileError)
	}

	cfg, err := newConfig(ctx)
	if err != nil {
		log.WithError(err).Errorf("error when loading config")
		os.Exit(exitCodeConfigError)
	}

	appLogger, err := newLogger(cfg.Log.Format, cfg.Log.Level)
	if err != nil {
		log.WithError(err).Errorf("creating application logger")
		os.Exit(exitCodeLoggerError)
	}

	if err := run(ctx, cfg, appLogger); err != nil {
		appLogger.WithError(err).Errorf("error in run()")
		os.Exit(exitCodeRunError)
	}

	os.Exit(exitCodeSuccess)
}

func run(ctx context.Context, cfg *config, log logrus.FieldLogger) error {
	naisTeams, err := naisapi.
		NewClient(cfg.NaisApi.Endpoint, cfg.NaisApi.ApiToken, log.WithField("component", "nais-api-client")).
		GetTeams(ctx, cfg.NaisApi.TeamsFilter)
	if err != nil {
		return err
	}

	if len(naisTeams) == 0 {
		return fmt.Errorf("no Nais teams returned from the API, this is most likely an error")
	}

	slack.
		NewNotifier(cfg.Slack.ApiToken, cfg.NaisApi.ConsoleUrl, log.WithField("component", "slack-notifier")).
		NotifyTeams(ctx, naisTeams)

	return nil
}
