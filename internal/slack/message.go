package slack

import (
	"fmt"
	"strings"

	"github.com/nais/slack-teams-notification/internal/teams"
	slackapi "github.com/slack-go/slack"
)

func mrkdwn(format string, args ...any) *slackapi.SectionBlock {
	return slackapi.NewSectionBlock(slackapi.NewTextBlockObject("mrkdwn", fmt.Sprintf(format, args...), false, false), nil, nil)
}

func header(format string, args ...any) *slackapi.HeaderBlock {
	return slackapi.NewHeaderBlock(slackapi.NewTextBlockObject("plain_text", fmt.Sprintf(format, args...), false, false))
}

func getNotificationMessageOptions(team teams.Team, frontendURL string) []slackapi.MsgOption {
	blocks := []slackapi.Block{
		mrkdwn("👋 Hei %s!", team.Slug),
		mrkdwn("Dere er ansvarlige for å holde teamets medlemsliste oppdatert. Siden medlemskap i NAIS-team gir utvidede rettigheter til blant annet produksjonsmiljø og persondata, er det viktig å holde teamet oppdatert."),
		mrkdwn("Følgende brukere er i dag registrert som medlemmer og eiere i `%s`:", team.Slug),
	}

	memberNames := make([]string, 0)
	ownerNames := make([]string, 0)
	for _, member := range team.Members.Members {
		name := member.User.Name
		if member.IsOwner() {
			ownerNames = append(ownerNames, "- "+name)
		}
		memberNames = append(memberNames, "- "+name)
	}

	if len(memberNames) > 0 {
		blocks = append(
			blocks,
			header("Medlemmer"),
			mrkdwn("%s", strings.Join(memberNames, "\n")),
		)
	}

	if len(ownerNames) > 0 {
		blocks = append(
			blocks,
			header("Eiere"),
			mrkdwn("%s", strings.Join(ownerNames, "\n")),
		)
	}

	blocks = append(blocks, mrkdwn("Ser dette korrekt ut? Om ikke kan du administrere teamet i <%s|Console> (krever <https://docs.nais.io/explanation/naisdevice/|naisdevice>).", getTeamsURL(frontendURL, team.Slug)))

	if len(ownerNames) < 2 {
		blocks = append(blocks, mrkdwn(fmt.Sprintf("*NB!* Antall eiere for dette teamet er %d, det *bør* være minst to eiere av hvert team.", len(ownerNames))))
		if len(ownerNames) == 0 {
			blocks = append(blocks, mrkdwn("Ta kontakt med <https://nav-it.slack.com/archives/C5KUST8N6/|nais-teamet> for å få lagt inn en eier."))
		}
	}

	return []slackapi.MsgOption{
		slackapi.MsgOptionBlocks(blocks...),
		slackapi.MsgOptionText(fmt.Sprintf("Påminnelse om å holde %q-teamet oppdatert", team.Slug), false),
	}
}

func getTeamsURL(baseURL, teamSlug string) string {
	baseURL = strings.TrimSuffix(baseURL, "/")
	return fmt.Sprintf("%s/team/%s/members", baseURL, teamSlug)
}
