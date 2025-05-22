package main

import (
	"context"

	"github.com/nais/slack-teams-notification/internal/cmd/slack_teams_notification"
)

func main() {
	slack_teams_notification.Run(context.Background())
}
