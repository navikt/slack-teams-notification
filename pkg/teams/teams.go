package teams

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	requestTimeout = time.Second * 10
)

type (
	apiResponse struct {
		Data struct {
			Teams apiResponseData `json:"teams"`
		} `json:"data"`
	}

	apiResponseData struct {
		Teams    []Team   `json:"nodes"`
		PageInfo PageInfo `json:"pageInfo"`
	}

	PageInfo struct {
		TotalCount int  `json:"totalCount"`
		HasNext    bool `json:"hasNextPage"`
	}

	Team struct {
		Slug         string  `json:"slug"`
		SlackChannel string  `json:"slackChannel"`
		Members      Members `json:"members"`
	}

	Members struct {
		Members  []Member `json:"nodes"`
		PageInfo PageInfo `json:"pageInfo"`
	}

	Member struct {
		User User   `json:"user"`
		Role string `json:"role"`
	}

	User struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	Client struct {
		serverURL  string
		apiToken   string
		httpClient http.Client
	}
)

var httpClient = http.Client{}

func NewClient(serverURL, apiToken string) *Client {
	return &Client{
		serverURL: serverURL,
		apiToken:  apiToken,
		httpClient: http.Client{
			Timeout: requestTimeout,
		},
	}
}

func (c *Client) GetTeams(teamSlugsFilter []string) ([]Team, error) {
	var teams []Team
	hasNext := true
	teamsOffset := 0
	pageSize := 100
	for hasNext {
		logrus.Infof("fetching teams from offset %d to %d", teamsOffset, teamsOffset+pageSize)
		response, err := c.requestPage(teamsOffset, pageSize)
		if err != nil {
			return nil, fmt.Errorf("performing request: %w", err)
		}
		teams = append(teams, response.Data.Teams.Teams...)
		teamsOffset += response.Data.Teams.PageInfo.TotalCount
		hasNext = response.Data.Teams.PageInfo.HasNext
	}

	if len(teamSlugsFilter) == 0 {
		return teams, nil
	}

	filteredTeams := make([]Team, 0)
	for _, team := range teams {
		for _, includeTeam := range teamSlugsFilter {
			if team.Slug == includeTeam {
				filteredTeams = append(filteredTeams, team)
			}
		}
	}

	return filteredTeams, nil
}

func (c *Client) requestPage(teamsOffset, teamsLimit int) (apiResponse, error) {
	queryString := fmt.Sprintf(`"query TeamsAndMembers(
      $teamsOffset: Int!
      $teamsLimit: Int!
      $membersOffset: Int!
      $membersLimit: Int!
    ) {
      teams(offset: $teamsOffset, limit: $teamsLimit) {
        pageInfo {
          hasNextPage
          totalCount
        }
        nodes {
          slug
          members(offset: $membersOffset, limit: $membersLimit) {
            pageInfo {
              hasNextPage
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
    }",
    "variables": {
      "teamsOffset": %d, 
      "teamsLimit": %d,
      "membersOffset": %d,
      "membersLimit": %d
    }`, teamsOffset, teamsLimit, 0, 100)
	requestBody := strings.ReplaceAll(
		fmt.Sprintf(`{"query": %s}`, queryString), "\n", " ",
	)
	response, err := c.PerformGQLRequest([]byte(requestBody))
	if err != nil {
		return apiResponse{}, fmt.Errorf("http: %w", err)
	}
	var deserialized apiResponse
	err = json.Unmarshal(response, &deserialized)
	if err != nil {
		return apiResponse{}, fmt.Errorf("unmarshaling response body: %w", err)
	}
	return deserialized, nil
}

func (c *Client) PerformGQLRequest(body []byte) ([]byte, error) {
	req, err := http.NewRequest("POST", c.serverURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiToken))
	req.Header.Set("Content-Type", "application/json")
	res, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("got a %d from %s: %v", res.StatusCode, c.serverURL, res)
	}
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return resBody, nil
}

func (m Member) IsOwner() bool {
	return m.Role == "OWNER"
}
