package slack

import (
	"context"
	"time"

	"github.com/nais/slack-teams-notification/internal/naisapi"
	"github.com/sirupsen/logrus"
	slackapi "github.com/slack-go/slack"
)

type Notifier struct {
	consoleFrontendURL string
	slackApi           *slackapi.Client
	log                logrus.FieldLogger
}

// NewNotifier Create a new Slack notifier instance
func NewNotifier(slackApiToken, consoleFrontendURL string, log logrus.FieldLogger) *Notifier {
	return &Notifier{
		log:                log,
		consoleFrontendURL: consoleFrontendURL,
		slackApi:           slackapi.New(slackApiToken),
	}
}

// NotifyTeams Notify all teams on Slack that they need to keep their teams up to date
func (n *Notifier) NotifyTeams(ctx context.Context, teams []naisapi.Team) {
	for _, team := range teams {
		if err := n.notifyTeam(ctx, team); err != nil {
			n.log.
				WithError(err).
				WithField("slack_channel", team.SlackChannel).
				Errorf("posting message to Slack")
		}
	}
}

func (n *Notifier) notifyTeam(ctx context.Context, team naisapi.Team) error {
	msgOptions := getNotificationMessageOptions(team, n.consoleFrontendURL)
	var recipients []string
	owners := n.ownersOf(team)
	for _, member := range owners {
		slackUser, err := n.slackApi.GetUserByEmailContext(ctx, member.Email)
		if err != nil {
			return err
		}
		recipients = append(recipients, slackUser.ID)
	}
	if len(recipients) == 0 {
		recipients = append(recipients, team.SlackChannel)
	}
	for _, recipient := range recipients {
		log := n.log.WithFields(logrus.Fields{
			"team_slug":    team.Slug,
			"recipient_id": recipient,
		})
		_, _, err := n.slackApi.PostMessageContext(ctx, recipient, msgOptions...)
		if err != nil {
			log.WithError(err).Errorf("post message to Slack")
		} else {
			log.Infof("notification sent")
		}

		time.Sleep(time.Second) // Sleep due to strict rate limiting
	}

	return nil
}

func (n *Notifier) ownersOf(team naisapi.Team) []naisapi.Member {
	owners := make([]naisapi.Member, 0)
	for _, member := range team.Members {
		if member.IsOwner() {
			owners = append(owners, member)
		}
	}
	if len(owners) == 0 {
		n.log.WithField("team_slug", team.Slug).Infof("unable to find team owner")
	}

	return owners
}
