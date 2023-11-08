package slack_test

import (
	"context"
	"testing"

	"github.com/nais/slack-teams-notification/pkg/slack"
	"github.com/nais/slack-teams-notification/pkg/teams"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
)

func TestNotifier_NotifyTeams(t *testing.T) {
	const frontendURL = "http://localhost:3001/"
	const token = "token"
	ctx := context.Background()
	logger, _ := test.NewNullLogger()
	log := logrus.NewEntry(logger)

	t.Run("no teams", func(t *testing.T) {
		notifier := slack.NewNotifier(token, frontendURL, slack.OptionLogger(log))
		assert.NoError(t, notifier.NotifyTeams(ctx, []teams.Team{}))
	})
}
