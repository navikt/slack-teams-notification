package teams_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nais/slack-teams-notification/pkg/teams"
	"github.com/stretchr/testify/assert"
)

func TestGetTeams(t *testing.T) {
	const apiToken = "some secret token"
	ctx := context.Background()
	emptyTeamSlugsFilter := make([]string, 0)

	t.Run("empty response from server", func(t *testing.T) {
		ts := httpServerWithHandlers(t, []http.HandlerFunc{
			func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "Bearer "+apiToken, r.Header.Get("Authorization"))
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
			},
		})
		defer ts.Close()

		teamsClient := teams.NewClient(ts.URL, apiToken)
		naisTeams, err := teamsClient.GetTeams(ctx, emptyTeamSlugsFilter)
		assert.Nil(t, naisTeams)
		assert.EqualError(t, err, "decode JSON: unexpected end of JSON input")
	})

	t.Run("incorrect JSON from server", func(t *testing.T) {
		jsonResponse := `{"teams":[{"slug":"slug"}]}`
		ts := httpServerWithHandlers(t, []http.HandlerFunc{
			func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(jsonResponse))
			},
		})
		defer ts.Close()

		teamsClient := teams.NewClient(ts.URL, apiToken)
		naisTeams, err := teamsClient.GetTeams(ctx, emptyTeamSlugsFilter)
		assert.Nil(t, naisTeams)
		assert.EqualError(t, err, "unexpected JSON: "+jsonResponse)
	})

	t.Run("unexpected response status code", func(t *testing.T) {
		ts := httpServerWithHandlers(t, []http.HandlerFunc{
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(400)
			},
		})
		defer ts.Close()

		teamsClient := teams.NewClient(ts.URL, apiToken)
		naisTeams, err := teamsClient.GetTeams(ctx, emptyTeamSlugsFilter)
		assert.Nil(t, naisTeams)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unexpected response status code:")
	})

	t.Run("teams in response", func(t *testing.T) {
		ts := httpServerWithHandlers(t, []http.HandlerFunc{
			func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(`
					{
						"data": {
							"teams": [
								{
									"slug": "team1",
									"members": [
										{
											"user": {
												"name": "User Name",
												"email": "user.name@example.com"
											},
											"role": "OWNER"
										},
										{
											"user": {
												"name": "Other User Name",
												"email": "other.user.name@example.com"
											},
											"role": "MEMBER"
										}
									]
								},
								{
									"slug": "team2",
									"members": [
										{
											"user": {
												"name": "User Name",
												"email": "user.name@example.com"
											},
											"role": "MEMBER"
										},
										{
											"user": {
												"name": "Other User Name",
												"email": "other.user.name@example.com"
											},
											"role": "OWNER"
										},
										{
											"user": {
												"name": "Third User Name",
												"email": "third.user.name@example.com"
											},
											"role": "MEMBER"
										}
									]
								}
							]
						}
					}
				`))
			},
		})
		defer ts.Close()

		teamsClient := teams.NewClient(ts.URL, apiToken)
		naisTeams, err := teamsClient.GetTeams(ctx, emptyTeamSlugsFilter)
		assert.NotNil(t, naisTeams)
		assert.Len(t, naisTeams, 2)
		assert.NoError(t, err)
		assert.Equal(t, "team1", naisTeams[0].Slug)
		assert.Equal(t, "team2", naisTeams[1].Slug)

		assert.Len(t, naisTeams[0].Members, 2)
		assert.Len(t, naisTeams[1].Members, 3)
	})

	t.Run("team slugs filter is not empty", func(t *testing.T) {
		ts := httpServerWithHandlers(t, []http.HandlerFunc{
			func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(`
					{
						"data": {
							"teams": [
								{
									"slug": "team1",
									"members": []
								},
								{
									"slug": "team2",
									"members": []
								},
								{
									"slug": "team3",
									"members": []
								},
								{
									"slug": "team4",
									"members": []
								}
							]
						}
					}
				`))
			},
		})
		defer ts.Close()

		teamsClient := teams.NewClient(ts.URL, apiToken)
		naisTeams, err := teamsClient.GetTeams(ctx, []string{"team1", "team3", "team5"})
		assert.NotNil(t, naisTeams)
		assert.Len(t, naisTeams, 2)
		assert.NoError(t, err)
		assert.Equal(t, "team1", naisTeams[0].Slug)
		assert.Equal(t, "team3", naisTeams[1].Slug)
	})
}

func TestMember_IsOwner(t *testing.T) {
	assert.False(t, teams.Member{
		Role: "MEMBER",
	}.IsOwner())

	assert.True(t, teams.Member{
		Role: "OWNER",
	}.IsOwner())
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
