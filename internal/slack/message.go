package slack

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/nais/slack-teams-notification/internal/naisapi"
	slackapi "github.com/slack-go/slack"
)

func list(entries []string) *slackapi.RichTextBlock {
	elements := make([]slackapi.RichTextElement, len(entries))
	for i, entry := range entries {
		elements[i] = slackapi.NewRichTextSection(slackapi.NewRichTextSectionTextElement(entry, nil))
	}

	return slackapi.NewRichTextBlock(
		uuid.NewString(),
		slackapi.NewRichTextList(slackapi.RTEListBullet, 0, elements...),
	)
}

func mrkdwn(format string, args ...any) *slackapi.SectionBlock {
	return slackapi.NewSectionBlock(
		slackapi.NewTextBlockObject(
			slackapi.MarkdownType,
			fmt.Sprintf(format, args...),
			false,
			false,
		),
		nil,
		nil,
	)
}

func header(format string, args ...any) *slackapi.HeaderBlock {
	return slackapi.NewHeaderBlock(
		slackapi.NewTextBlockObject(
			slackapi.PlainTextType,
			fmt.Sprintf(format, args...),
			false,
			false,
		),
	)
}

func getNotificationMessageOptions(team naisapi.Team, frontendURL string) []slackapi.MsgOption {
	blocks := []slackapi.Block{
		mrkdwn("👋 Hei %s!", team.Slug),
		mrkdwn("Dere er ansvarlige for å holde teamets medlemsliste oppdatert. Siden medlemskap i Nais-team gir utvidede rettigheter til blant annet produksjonsmiljø og persondata, er det viktig å holde teamet oppdatert."),
		mrkdwn("Følgende brukere er i dag registrert som medlemmer og eiere i `%s`:", team.Slug),
	}

	memberNames := make([]string, 0)
	ownerNames := make([]string, 0)
	for _, member := range team.Members {
		name := member.Name
		if member.IsOwner() {
			ownerNames = append(ownerNames, name)
		}
		memberNames = append(memberNames, name)
	}

	blocks = append(blocks, header("Medlemmer"), list(memberNames))

	if len(ownerNames) > 0 {
		blocks = append(blocks, header("Eiere"), list(ownerNames))
	}

	blocks = append(
		blocks,
		mrkdwn(
			"Ser dette korrekt ut? Om ikke kan dere administrere teamet i <%s|Console>.",
			getTeamMembersAdminURL(frontendURL, team.Slug),
		),
	)

	if len(ownerNames) == 0 {
		blocks = append(blocks, mrkdwn("*NB!* Teamet har ingen eier, ta kontakt med Nais-teamet på #utviklerrommet for å få lagt inn en eier."))
	} else if len(ownerNames) < 2 {
		blocks = append(blocks, mrkdwn("*NB!* Det *bør* være minst to eiere av hvert team."))
	}

	return []slackapi.MsgOption{
		slackapi.MsgOptionBlocks(blocks...),
		slackapi.MsgOptionText(fmt.Sprintf("Påminnelse om å holde %q-teamet oppdatert", team.Slug), false),
	}
}

func getTeamMembersAdminURL(baseURL, teamSlug string) string {
	baseURL = strings.TrimSuffix(baseURL, "/")
	return fmt.Sprintf("%s/team/%s/members", baseURL, teamSlug)
}
