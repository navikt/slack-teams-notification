package slack

import (
	"fmt"
	"strings"

	"github.com/nais/slack-teams-notification/pkg/teams"
	slackapi "github.com/slack-go/slack"
)

func mrkdwn(format string, args ...any) *slackapi.SectionBlock {
	return slackapi.NewSectionBlock(slackapi.NewTextBlockObject("mrkdwn", fmt.Sprintf(format, args...), false, false), nil, nil)
}

func header(format string, args ...any) *slackapi.HeaderBlock {
	return slackapi.NewHeaderBlock(slackapi.NewTextBlockObject("plain_text", fmt.Sprintf(format, args...), false, false))
}

func getNotificationMessageOptions(team teams.Team, recipientName, frontendURL string) []slackapi.MsgOption {
	blocks := []slackapi.Block{
		mrkdwn("Hei %s,", recipientName),
		mrkdwn("du står som *eier* av NAIS teamet `%s`, og er derfor ansvarlig for å holde teamet oppdatert med tanke på hvem som er medlem. Fordi medlemsskap i NAIS teams gir utvidede rettigheter til blant annet produksjonsmiljø og persondata er det viktig å holde teamene oppdatert.", team.Slug),
		mrkdwn("Følgende brukere ligger i dag inne som medlemmer / eiere i `%s`:", team.Slug),
	}

	memberNames := make([]string, 0)
	ownerNames := make([]string, 0)
	for _, member := range team.Members {
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

	blocks = append(blocks, mrkdwn("Ser dette korrekt ut? Om ikke kan du administrere teamet i <%s|NAIS teams>. Alle eierne av teamet vil motta denne meldingen.", getTeamsURL(frontendURL, team.Slug)))

	if len(ownerNames) == 1 {
		blocks = append(blocks, mrkdwn("*NB!* Du står oppført som den eneste eieren av dette teamet, det *bør* være minst to eiere av hvert team."))
	}

	return []slackapi.MsgOption{
		slackapi.MsgOptionBlocks(blocks...),
		slackapi.MsgOptionText(fmt.Sprintf("Påminnelse om å holde %q-teamet oppdatert", team.Slug), false),
	}
}

func getTeamsURL(baseURL, teamSlug string) string {
	baseURL = strings.TrimSuffix(baseURL, "/")
	return fmt.Sprintf("%s/teams/%s", baseURL, teamSlug)
}
