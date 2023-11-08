package slack

import (
	"context"
	"time"

	"github.com/nais/slack-teams-notification/pkg/teams"
	"github.com/sirupsen/logrus"
	slackapi "github.com/slack-go/slack"
)

type Notifier struct {
	teamsFrontendURL string
	slackAPI         *slackapi.Client
	logger           *logrus.Entry
	sleepDuration    time.Duration
}

type Option func(*Notifier)

// NewNotifier Create a new Slack notifier instance
func NewNotifier(slackApiToken, teamsFrontendURL string, options ...Option) *Notifier {
	notifier := &Notifier{
		logger:           logrus.New().WithField("component", "slack-notifier"),
		teamsFrontendURL: teamsFrontendURL,
		slackAPI:         slackapi.New(slackApiToken),
		sleepDuration:    time.Second * 1, // Slack has rather strict rate limits, sleep duration between notifications
	}

	for _, opt := range options {
		opt(notifier)
	}

	return notifier
}

// OptionSleepDuration Set a custom sleep duration
func OptionSleepDuration(d time.Duration) func(*Notifier) {
	return func(n *Notifier) {
		n.sleepDuration = d
	}
}

// OptionLogger Set a custom logger
func OptionLogger(logger *logrus.Entry) func(*Notifier) {
	return func(n *Notifier) {
		n.logger = logger
	}
}

// OptionSlackApi Set a custom Slack API client
func OptionSlackApi(client *slackapi.Client) func(*Notifier) {
	return func(n *Notifier) {
		n.slackAPI = client
	}
}

// NotifyTeams Notify all team owners on Slack that they need to keep their teams up to date
func (n *Notifier) NotifyTeams(ctx context.Context, teams []teams.Team, ownerEmailsFilter []string) error {
	for _, team := range teams {
		err := n.notifyTeam(ctx, team)
		if err != nil {
			return err
		}
	}
	return nil
}

// notifyTeam Send notifications about a team to the channel they have supplied
func (n *Notifier) notifyTeam(ctx context.Context, team teams.Team) error {
	logger := n.logger.WithField("team_slug", team.Slug)
	msgOptions := getNotificationMessageOptions(team, n.teamsFrontendURL)
	_, _, err := n.slackAPI.PostMessageContext(ctx, team.SlackChannel, msgOptions...)
	logger.Infof("'%s' notified", team.Slug)
	return err
}
