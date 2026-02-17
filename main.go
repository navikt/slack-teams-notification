package main

import (
	"context"

	slackteamsnotification "github.com/nais/slack-teams-notification/internal/cmd/slack_teams_notification"
)

func main() {
	slackteamsnotification.Run(context.Background())
}
