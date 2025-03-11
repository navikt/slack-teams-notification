package naisapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"net/http"
	"net/url"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/nais/slack-teams-notification/internal/httputils"
	"github.com/sirupsen/logrus"
)

const (
	requestTimeout = time.Second * 10
)

type PaginatedGraphQLResponse struct {
	Data struct {
		Teams struct {
			PageInfo struct {
				TotalCount  int    `json:"totalCount"`
				HasNextPage bool   `json:"hasNextPage"`
				EndCursor   string `json:"endCursor"`
			} `json:"pageInfo"`
			Nodes []struct {
				Slug         string `json:"slug"`
				SlackChannel string `json:"slackChannel"`
				Members      struct {
					PageInfo struct {
						TotalCount  int    `json:"totalCount"`
						HasNextPage bool   `json:"hasNextPage"`
						EndCursor   string `json:"endCursor"`
					} `json:"pageInfo"`
					Nodes []struct {
						User struct {
							Name  string `json:"name"`
							Email string `json:"email"`
						} `json:"user"`
						Role string `json:"role"`
					}
				} `json:"members"`
			} `json:"nodes"`
		} `json:"teams"`
	} `json:"data"`
}

type Team struct {
	Slug         string
	SlackChannel string
	Members      []Member
}

type Member struct {
	Name  string
	Email string
	Role  string
}

type NaisTeam struct {
	Slug         string
	SlackChannel string
}

type Client struct {
	endpoint string
	apiToken string
	log      logrus.FieldLogger
}

func NewClient(endpoint, apiToken string, log logrus.FieldLogger) *Client {
	return &Client{
		endpoint: endpoint,
		apiToken: apiToken,
		log:      log,
	}
}

func (c *Client) GetTeams(ctx context.Context, teamSlugsFilter []string) ([]Team, error) {
	query := `query getTeamsAndMembers {
  		teams(first:100 after:%q) {
			pageInfo {
				totalCount
				hasNextPage
				endCursor
			}
  			nodes {
      			slug
				slackChannel
      			members(first:100 after:%q) {
					pageInfo {
						totalCount
						hasNextPage
						endCursor
					}
        			nodes {
						user {
							name
							email
						}
          				role
        			}
      			}
    		}
  		}
	}`

	allTeams := make(map[string]Team)
	teamsCursor, membersCursor := "", ""
	teamsHasNextPage := true
	resp := &PaginatedGraphQLResponse{}

	c.log.Debugf("start fetching teams and members from Nais API")
	for teamsHasNextPage {
	fetch:
		err := func() error {
			responseBody, err := gqlRequest(
				ctx,
				c.endpoint,
				fmt.Sprintf(`{"query": %q}`, fmt.Sprintf(query, teamsCursor, membersCursor)),
				http.Header{
					"User-Agent":    {httputils.UserAgent},
					"Content-Type":  {"application/json"},
					"Authorization": {"Bearer " + c.apiToken},
				},
			)
			if err != nil {
				return err
			}
			defer func() {
				if err := responseBody.Close(); err != nil {
					c.log.WithError(err).Errorf("failed to close response body")
				}
			}()
			return json.NewDecoder(responseBody).Decode(resp)
		}()
		if err != nil {
			return nil, err
		}

		for _, teamNode := range resp.Data.Teams.Nodes {
			team, exists := allTeams[teamNode.Slug]
			if !exists {
				team = Team{
					Slug:         teamNode.Slug,
					SlackChannel: teamNode.SlackChannel,
					Members:      make([]Member, 0),
				}
			}

			for _, memberNode := range teamNode.Members.Nodes {
				team.Members = append(team.Members, Member{
					Name:  memberNode.User.Name,
					Email: memberNode.User.Email,
					Role:  memberNode.Role,
				})
			}
			allTeams[teamNode.Slug] = team

			if teamNode.Members.PageInfo.HasNextPage {
				c.log.WithField("team_slug", teamNode.Slug).Debugf("team has more members, fetching next page")
				membersCursor = teamNode.Members.PageInfo.EndCursor
				goto fetch
			}
		}

		membersCursor = ""
		teamsCursor = resp.Data.Teams.PageInfo.EndCursor
		teamsHasNextPage = resp.Data.Teams.PageInfo.HasNextPage

		c.log.WithFields(logrus.Fields{
			"total_count":   resp.Data.Teams.PageInfo.TotalCount,
			"has_next_page": teamsHasNextPage,
		}).Debugf("fetched page of teams")
	}

	c.log.Debugf("done fetching Nais teams")

	if len(teamSlugsFilter) == 0 {
		c.log.Debugf("no filter specified, return all teams")
		return slices.SortedStableFunc(maps.Values(allTeams), func(a Team, b Team) int {
			return strings.Compare(a.Slug, b.Slug)
		}), nil
	}

	filteredTeams := make([]Team, 0)
	c.log.Debugf("filter teams: %q", strings.Join(teamSlugsFilter, ", "))
	for _, team := range allTeams {
		for _, includeTeam := range teamSlugsFilter {
			if team.Slug == includeTeam {
				filteredTeams = append(filteredTeams, team)
			}
		}
	}
	sort.SliceStable(filteredTeams, func(a, b int) bool {
		return filteredTeams[a].Slug < filteredTeams[b].Slug
	})
	return filteredTeams, nil
}

func (m Member) IsOwner() bool {
	return m.Role == "OWNER"
}

func gqlRequest(ctx context.Context, rawUrl, body string, headers http.Header) (io.ReadCloser, error) {
	u, err := url.Parse(rawUrl)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", u.String(), bytes.NewBuffer([]byte(body)))
	if err != nil {
		return nil, err
	}
	req.Header = headers
	client := http.Client{
		Timeout: requestTimeout,
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected HTTP status code %d from %q: %v", res.StatusCode, rawUrl, res)
	}
	return res.Body, nil
}
