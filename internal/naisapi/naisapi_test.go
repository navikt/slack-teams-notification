package naisapi_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nais/slack-teams-notification/internal/naisapi"
	logrustest "github.com/sirupsen/logrus/hooks/test"
)

func TestGetTeams(t *testing.T) {
	ctx := context.Background()
	const apiToken = "some secret token"
	emptyTeamSlugsFilter := make([]string, 0)
	log, _ := logrustest.NewNullLogger()

	t.Run("empty response from server", func(t *testing.T) {
		ts := httpServerWithHandlers(t, []http.HandlerFunc{
			func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("Authorization") != "Bearer "+apiToken {
					t.Errorf("unexpected authorization header: %v", r.Header.Get("Authorization"))
				}

				if r.Header.Get("Content-Type") != "application/json" {
					t.Errorf("unexpected content type: %v", r.Header.Get("Content-Type"))
				}
			},
		})
		defer ts.Close()

		apiClient := naisapi.NewClient(ts.URL, apiToken, log)
		teams, err := apiClient.GetTeams(ctx, emptyTeamSlugsFilter)

		if err == nil {
			t.Fatalf("expected error, got nil")
		} else if !strings.Contains(err.Error(), "EOF") {
			t.Errorf("unexpected response: %v", err)
		} else if teams != nil {
			t.Errorf("expected nil teams, got: %v", teams)
		}
	})

	t.Run("no hits yield empty array", func(t *testing.T) {
		jsonResponse := `{"teams":[{"slug":"slug"}]}`
		ts := httpServerWithHandlers(t, []http.HandlerFunc{
			func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte(jsonResponse))
			},
		})
		defer ts.Close()

		teamsClient := naisapi.NewClient(ts.URL, apiToken, log)
		naisTeams, err := teamsClient.GetTeams(ctx, emptyTeamSlugsFilter)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		} else if len(naisTeams) != 0 {
			t.Errorf("expected empty array, got: %v", naisTeams)
		}
	})

	t.Run("unexpected response status code", func(t *testing.T) {
		ts := httpServerWithHandlers(t, []http.HandlerFunc{
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(400)
			},
		})
		defer ts.Close()

		teamsClient := naisapi.NewClient(ts.URL, apiToken, log)
		naisTeams, err := teamsClient.GetTeams(ctx, emptyTeamSlugsFilter)
		if naisTeams != nil {
			t.Errorf("expected nil teams, got: %v", naisTeams)
		} else if err == nil {
			t.Fatalf("expected error, got nil")
		} else if !strings.Contains(err.Error(), "HTTP status code 400") {
			t.Errorf("unexpected response: %v", err)
		}
	})

	t.Run("teams in response", func(t *testing.T) {
		ts := httpServerWithHandlers(t, []http.HandlerFunc{
			func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte(`{
					"data": {
						"teams": {
							"pageInfo": {
								"totalCount": 2,
								"hasNextPage": false,
								"endCursor": ""
							},
							"nodes": [
								{
									"slug": "team1",
									"members": {
										"pageInfo": {
											"totalCount": 2,
											"hasNextPage": false,
											"endCursor": ""
										},
										"nodes": [
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
									}
								},
								{
									"slug": "team2",
									"members": {
										"pageInfo": {
											"totalCount": 3,
											"hasNextPage": false,
											"endCursor": ""
										},
										"nodes": [
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
								}
							]
						}
					}
				}`))
			},
		})
		defer ts.Close()

		teamsClient := naisapi.NewClient(ts.URL, apiToken, log)
		naisTeams, err := teamsClient.GetTeams(ctx, emptyTeamSlugsFilter)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		} else if len(naisTeams) != 2 {
			t.Fatalf("expected 2 teams, got: %v", naisTeams)
		}

		if naisTeams[0].Slug != "team1" {
			t.Errorf("expected team1, got: %v", naisTeams[0].Slug)
		}

		if len(naisTeams[0].Members) != 2 {
			t.Fatalf("expected 2 members, got: %v", naisTeams[0].Members)
		}

		if naisTeams[1].Slug != "team2" {
			t.Errorf("expected team2, got: %v", naisTeams[1].Slug)
		}

		if len(naisTeams[1].Members) != 3 {
			t.Fatalf("expected 3 members, got: %v", naisTeams[1].Members)
		}
	})

	t.Run("team slugs filter is not empty", func(t *testing.T) {
		ts := httpServerWithHandlers(t, []http.HandlerFunc{
			func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte(`{
					"data": {
						"teams": {
							"pageInfo": {
								"totalCount": 4,
								"hasNextPage": false,
								"endCursor": ""
							},
							"nodes": [
								{
									"slug": "team1",
									"members": {
										"pageInfo": {
											"totalCount": 0,
											"hasNextPage": false,
											"endCursor": ""
										},
										"nodes": []
									}
								},
								{
									"slug": "team2",
									"members": {
										"pageInfo": {
											"totalCount": 0,
											"hasNextPage": false,
											"endCursor": ""
										},
										"nodes": []
									}
								},
								{
									"slug": "team3",
									"members": {
										"pageInfo": {
											"totalCount": 0,
											"hasNextPage": false,
											"endCursor": ""
										},
										"nodes": []
									}
								},
								{
									"slug": "team4",
									"members": {
										"pageInfo": {
											"totalCount": 0,
											"hasNextPage": false,
											"endCursor": ""
										},
										"nodes": []
									}
								}
							]
						}
					}
				}`))
			},
		})
		defer ts.Close()

		teamsClient := naisapi.NewClient(ts.URL, apiToken, log)
		naisTeams, err := teamsClient.GetTeams(ctx, []string{"team1", "team3", "team5"})
		if naisTeams == nil {
			t.Fatalf("expected teams, got nil")
		} else if len(naisTeams) != 2 {
			t.Fatalf("expected 2 teams, got: %v", naisTeams)
		} else if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if naisTeams[0].Slug != "team1" {
			t.Errorf("expected team1, got: %v", naisTeams[0].Slug)
		}

		if naisTeams[1].Slug != "team3" {
			t.Errorf("expected team3, got: %v", naisTeams[1].Slug)
		}
	})
}

func TestMember_IsOwner(t *testing.T) {
	member := naisapi.Member{Role: "MEMBER"}
	if member.IsOwner() != false {
		t.Errorf("member should not be owner: %+v", member)
	}

	owner := naisapi.Member{Role: "OWNER"}
	if owner.IsOwner() != true {
		t.Errorf("member should be owner: %+v", owner)
	}
}

func httpServerWithHandlers(t *testing.T, handlers []http.HandlerFunc) *httptest.Server {
	idx := 0
	t.Cleanup(func() {
		if len(handlers) != idx {
			t.Fatalf("Not all handlers have been executed, %d handler(s) was added to the test server, %d was executed", len(handlers), idx)
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
