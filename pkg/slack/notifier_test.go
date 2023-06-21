package slack_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nais/slack-teams-notification/pkg/slack"
	"github.com/nais/slack-teams-notification/pkg/teams"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	slackapi "github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
)

func TestNotifier_NotifyTeams(t *testing.T) {
	const frontendURL = "http://localhost:3001/"
	const token = "token"
	ctx := context.Background()
	logger, _ := test.NewNullLogger()
	log := logrus.NewEntry(logger)

	t.Run("no teams", func(t *testing.T) {
		notifier := slack.NewNotifier(token, frontendURL, slack.OptionLogger(log))
		assert.NoError(t, notifier.NotifyTeams(ctx, []teams.Team{}, []string{}))
	})

	t.Run("teams, no owner emails filter", func(t *testing.T) {
		ts := httpServerWithHandlers(t, []http.HandlerFunc{
			// owner for team 1 does not exist
			func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(`{
					"ok": false,
					"error": "users_not_found"
				}`))
			},
			// owner for team 2
			func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(`{
					"ok": true,
					"user": {"id": "U123", "real_name": "Some Real Name"}
				}`))
			},
			// message for owner for team 2
			func(w http.ResponseWriter, r *http.Request) {
				body, _ := io.ReadAll(r.Body)
				assert.Contains(t, string(body), "Hei+Team+2+Owner+1")
				assert.Contains(t, string(body), "channel=U123")
				w.Write([]byte(`{"ok": true}`))
			},
			// owner for team 3
			func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(`{
					"ok": true,
					"user": {"id": "U456", "real_name": "Some Other Real Name"}
				}`))
			},
			// message for owner for team 3
			func(w http.ResponseWriter, r *http.Request) {
				body, _ := io.ReadAll(r.Body)
				assert.Contains(t, string(body), "Hei+Team+3+Owner+1")
				assert.Contains(t, string(body), "channel=U456")
				w.Write([]byte(`{"ok": true}`))
			},
		})
		defer ts.Close()

		naisTeams := []teams.Team{
			{
				Slug: "team1",
				Members: []teams.Member{
					{
						User: teams.User{
							Name:  "Team 1 Member 1",
							Email: "team1member1@example.com",
						},
						Role: "MEMBER",
					},
					{
						User: teams.User{
							Name:  "Team 1 Member 2",
							Email: "team1member2@example.com",
						},
						Role: "MEMBER",
					},
					{
						User: teams.User{
							Name:  "Team 1 Owner 1",
							Email: "team1owner1@example.com",
						},
						Role: "OWNER",
					},
				},
			},
			{
				Slug: "team2",
				Members: []teams.Member{
					{
						User: teams.User{
							Name:  "Team 2 Member 1",
							Email: "team2member1@example.com",
						},
						Role: "MEMBER",
					},
					{
						User: teams.User{
							Name:  "Team 2 Member 2",
							Email: "team2member2@example.com",
						},
						Role: "MEMBER",
					},
					{
						User: teams.User{
							Name:  "Team 2 Owner 1",
							Email: "team2owner1@example.com",
						},
						Role: "OWNER",
					},
				},
			},
			{
				Slug: "team3",
				Members: []teams.Member{
					{
						User: teams.User{
							Name:  "Team 3 Member 1",
							Email: "team3member1@example.com",
						},
						Role: "MEMBER",
					},
					{
						User: teams.User{
							Name:  "Team 3 Owner 1",
							Email: "team3owner1@example.com",
						},
						Role: "OWNER",
					},
				},
			},
		}
		err := slack.
			NewNotifier(
				token,
				frontendURL,
				slack.OptionSlackApi(slackapi.New(token, slackapi.OptionAPIURL(ts.URL+"/"))),
				slack.OptionLogger(log),
				slack.OptionSleepDuration(0),
			).
			NotifyTeams(ctx, naisTeams, []string{})
		assert.NoError(t, err)
	})

	t.Run("teams, with owner emails filter", func(t *testing.T) {
		ts := httpServerWithHandlers(t, []http.HandlerFunc{
			func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(`{
					"ok": true,
					"user": {"id": "U123456", "real_name": "Some Name"}
				}`))
			},
			func(w http.ResponseWriter, r *http.Request) {
				body, _ := io.ReadAll(r.Body)
				assert.Contains(t, string(body), "Hei+Team+2+Owner+1")
				assert.Contains(t, string(body), "channel=U123456")
				w.Write([]byte(`{"ok": true}`))
			},
		})
		defer ts.Close()

		naisTeams := []teams.Team{
			{
				Slug: "team1",
				Members: []teams.Member{
					{
						User: teams.User{
							Name:  "Team 1 Owner 1",
							Email: "team1owner1@example.com",
						},
						Role: "OWNER",
					},
				},
			},
			{
				Slug: "team2",
				Members: []teams.Member{
					{
						User: teams.User{
							Name:  "Team 2 Owner 1",
							Email: "team2owner1@example.com",
						},
						Role: "OWNER",
					},
				},
			},
			{
				Slug: "team3",
				Members: []teams.Member{
					{
						User: teams.User{
							Name:  "Team 3 Owner 1",
							Email: "team3owner1@example.com",
						},
						Role: "OWNER",
					},
				},
			},
		}
		err := slack.
			NewNotifier(
				token,
				frontendURL,
				slack.OptionSlackApi(slackapi.New(token, slackapi.OptionAPIURL(ts.URL+"/"))),
				slack.OptionLogger(log),
				slack.OptionSleepDuration(0),
			).
			NotifyTeams(ctx, naisTeams, []string{"team2owner1@example.com"})
		assert.NoError(t, err)
	})
}

func httpServerWithHandlers(t *testing.T, handlers []http.HandlerFunc) *httptest.Server {
	idx := 0
	t.Cleanup(func() {
		if len(handlers) != idx {
			assert.Fail(t, "Not all handlers have been executed", "%d handler(s) was added to the test server, %d was executed", len(handlers), idx)
		}
	})
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(handlers) < idx+1 {
			t.Fatalf("unexpected request, add missing handler func: %v", r)
		}
		handlers[idx](w, r)
		idx += 1
	}))
}
