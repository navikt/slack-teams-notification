package teams

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	requestTimeout = time.Second * 10
)

type (
	apiResponse struct {
		Data apiResponseData `json:"data"`
	}

	apiResponseData struct {
		Teams []Team `json:"teams"`
	}

	Team struct {
		Slug    string   `json:"slug"`
		Members []Member `json:"members"`
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

// NewClient create a new client for the teams-backend GraphQL API
func NewClient(serverURL, apiToken string) Client {
	return Client{
		serverURL: serverURL,
		apiToken:  apiToken,
		httpClient: http.Client{
			Timeout: requestTimeout,
		},
	}
}

// GetTeams Get a list of NAIS teams from the teams backend
func (c Client) GetTeams(ctx context.Context) ([]Team, error) {
	resp, err := c.getNaisTeamsResponse(ctx)
	if err != nil {
		return nil, fmt.Errorf("request teams: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected response status code: %s", resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	bodyAsJson := &apiResponse{}
	err = json.Unmarshal(body, bodyAsJson)
	if err != nil {
		return nil, fmt.Errorf("decode JSON: %w", err)
	}

	if bodyAsJson.Data.Teams == nil {
		return nil, fmt.Errorf("unexpected JSON: %s", body)
	}

	return bodyAsJson.Data.Teams, nil
}

func (c Client) getNaisTeamsResponse(ctx context.Context) (*http.Response, error) {
	teamsQuery := `query {
		teams {
			slug
			members {
				user {
					name
					email
				}
				role
			}
		}
	}`
	payload, err := json.Marshal(map[string]string{"query": teamsQuery})
	if err != nil {
		return nil, fmt.Errorf("create request payload for teams-backend: %w", err)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.serverURL, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("create request for teams-backend: %w", err)
	}

	request.Header.Set("Authorization", "Bearer "+c.apiToken)
	request.Header.Set("Content-Type", "application/json")
	return c.httpClient.Do(request)
}

func (m Member) IsOwner() bool {
	return m.Role == "OWNER"
}
