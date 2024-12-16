package slack

import (
	"context"
	"time"

	"github.com/nais/slack-teams-notification/internal/teams"
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

// OptionLogger Set a custom logger
func OptionLogger(logger *logrus.Entry) func(*Notifier) {
	return func(n *Notifier) {
		n.logger = logger
	}
}

// NotifyTeams Notify all teams on Slack that they need to keep their teams up to date
func (n *Notifier) NotifyTeams(ctx context.Context, teams []teams.Team) {
	for _, team := range teams {
		err := n.notifyTeam(ctx, team)
		if err != nil {
			n.logger.Errorf("posting msg to channel '%s': %v", team.SlackChannel, err)
		}
	}
}

func (n *Notifier) notifyTeam(ctx context.Context, team teams.Team) error {
	logger := n.logger.WithField("team_slug", team.Slug)
	msgOptions := getNotificationMessageOptions(team, n.teamsFrontendURL)
	var recipients []string
	owners := ownersOf(team)
	for _, member := range owners {
		slackUser, err := n.slackAPI.GetUserByEmailContext(ctx, member.Email)
		if err != nil {
			return err
		}
		recipients = append(recipients, slackUser.ID)
	}
	if len(recipients) == 0 {
		recipients = append(recipients, team.SlackChannel)
	}
	for _, recipient := range recipients {
		time.Sleep(time.Second)
		_, _, err := n.slackAPI.PostMessageContext(ctx, recipient, msgOptions...)
		if err != nil {
			logger.Errorf("unable to post message to %s: %v", recipient, err)
		} else {
			logger.Infof("'%s' notified", team.Slug)
		}
	}

	return nil
}

func ownersOf(team teams.Team) []teams.User {
	owners := make([]teams.User, 0)
	for _, member := range team.Members.Members {
		if member.IsOwner() {
			owners = append(owners, member.User)
		}
	}
	if len(owners) == 0 {
		logrus.Infof("couldn't decide who owns team '%s'", team.Slug)
	}

	return owners
}
