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
}

// NewNotifier Create a new Slack notifier instance
func NewNotifier(teamsFrontendURL, slackAPIToken string) Notifier {
	return Notifier{
		teamsFrontendURL: teamsFrontendURL,
		slackAPI:         slackapi.New(slackAPIToken),
	}
}

// NotifyTeams Notify all team owners on Slack that they need to keep their teams up to date with regards to
// owners/members
func (s Notifier) NotifyTeams(ctx context.Context, teams []teams.Team) error {
	for _, team := range teams {
		err := s.notifyTeam(ctx, team)
		if err != nil {
			return err
		}
	}
	return nil
}

// notifyTeam Send notifications about a team to the owners of the team
func (s Notifier) notifyTeam(ctx context.Context, team teams.Team) error {
	logger := logrus.WithField("team_slug", team.Slug)

	owners := teamOwners(team)
	if len(owners) == 0 {
		logger.Infof("team does not have any owners, unable to notify")
		return nil
	}

	for _, owner := range owners {
		email := owner.Email
		logger = logger.WithField("user_email", email)

		slackUser, err := s.slackAPI.GetUserByEmailContext(ctx, email)
		if err != nil {
			logger.WithError(err).Errorf("unable to lookup Slack user")
			continue
		}

		logger = logger.WithFields(logrus.Fields{
			"user_slack_id":   slackUser.ID,
			"user_slack_name": slackUser.RealName,
		})

		err = s.notifyOwner(ctx, team, owner, slackUser.ID)
		if err != nil {
			logger.WithError(err).Errorf("unable to notify Slack user")
			continue
		}

		// Slack has rather strict rate limits, so we'll sleep after each message
		time.Sleep(time.Second * 1)
	}

	return nil
}

// notifyOwner Send a notification to a specific owner regarding a team
func (s Notifier) notifyOwner(ctx context.Context, team teams.Team, owner teams.User, slackUserID string) error {
	msgOptions := getNotificationMessageOptions(team, owner.Name, s.teamsFrontendURL)
	_, _, err := s.slackAPI.PostMessageContext(ctx, slackUserID, msgOptions...)
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
