package slack

import (
	"context"
	"strings"
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
		logger:           logrus.New().WithField("component", "notifier"),
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

// NotifyTeams Notify all team owners on Slack that they need to keep their teams up to date with regards to
// owners/members
func (n *Notifier) NotifyTeams(ctx context.Context, teams []teams.Team, ownerEmailsFilter []string) error {
	for _, team := range teams {
		err := n.notifyTeam(ctx, team, ownerEmailsFilter)
		if err != nil {
			return err
		}
	}
	return nil
}

// notifyTeam Send notifications about a team to the owners of the team
func (n *Notifier) notifyTeam(ctx context.Context, team teams.Team, ownerEmailsFilter []string) error {
	logger := n.logger.WithField("team_slug", team.Slug)

	owners := teamOwners(team)
	if len(owners) == 0 {
		logger.Infof("team does not have any owners, unable to notify")
		return nil
	}

	for _, owner := range owners {
		email := owner.Email

		if !ownerShouldReceiveNotification(email, ownerEmailsFilter) {
			continue
		}

		logger = logger.WithField("user_email", email)

		slackUser, err := n.slackAPI.GetUserByEmailContext(ctx, email)
		if err != nil {
			logger.WithError(err).Errorf("unable to lookup Slack user")
			continue
		}

		logger = logger.WithFields(logrus.Fields{
			"user_slack_id":   slackUser.ID,
			"user_slack_name": slackUser.RealName,
		})

		err = n.notifyOwner(ctx, team, owner, slackUser.ID)
		if err != nil {
			logger.WithError(err).Errorf("unable to notify Slack user")
			continue
		}

		time.Sleep(n.sleepDuration)
	}

	return nil
}

// notifyOwner Send a notification to a specific owner regarding a team
func (n *Notifier) notifyOwner(ctx context.Context, team teams.Team, owner teams.User, slackUserID string) error {
	msgOptions := getNotificationMessageOptions(team, owner.Name, n.teamsFrontendURL)
	_, _, err := n.slackAPI.PostMessageContext(ctx, slackUserID, msgOptions...)
	return err
}

// teamOwners return a list of team owners for a team
func teamOwners(team teams.Team) []teams.User {
	owners := make([]teams.User, 0)
	for _, member := range team.Members {
		if member.IsOwner() {
			owners = append(owners, member.User)
		}
	}

	return owners
}

// ownerShouldReceiveNotification check if an email should receive a notification or not
func ownerShouldReceiveNotification(email string, ownerEmailsFilter []string) bool {
	if len(ownerEmailsFilter) == 0 {
		return true
	}

	for _, filter := range ownerEmailsFilter {
		if strings.EqualFold(email, filter) {
			return true
		}
	}

	return false
}
