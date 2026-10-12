package github

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/blastjax/golang-redis/internal/config"
)

const (
	// User-Agent header
	userAgent = "maya-golang-cli/1.0"

	// GitHub API specs
	maxUsersPerPage = 100 // GitHub API per_page limit for the users endpoint
)

// ErrNotFound is returned when the GitHub API responds with 404 Not Found
var ErrNotFound = errors.New("github: not found")

// Client represents a GitHub API client
type Client struct {
	httpClient *http.Client
	config     *config.GitHubConfig
}

// NewClient creates a new GitHub API client
func NewClient(cfg *config.GitHubConfig) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		config: cfg,
	}
}

// FetchUsers fetches one page of users with an ID greater than since
func (c *Client) FetchUsers(ctx context.Context, perPage, since int) ([]User, error) {
	// Validate and adjust perPage according to GitHub API specs
	if perPage <= 0 || perPage > maxUsersPerPage {
		perPage = maxUsersPerPage
	}

	q := url.Values{}
	q.Set("per_page", strconv.Itoa(perPage))
	if since > 0 {
		q.Set("since", strconv.Itoa(since))
	}

	var users []User
	if err := c.get(ctx, "/users?"+q.Encode(), &users); err != nil {
		return nil, err
	}

	return users, nil
}

// FetchAllUsers fetches up to totalCount users with an ID greater than since.
// The users endpoint is paginated by cursor (each page starts after the previous
// page's last user ID), so pages have to be fetched one after another.
func (c *Client) FetchAllUsers(ctx context.Context, totalCount, since int) ([]User, error) {
	var allUsers []User

	for len(allUsers) < totalCount {
		perPage := min(totalCount-len(allUsers), maxUsersPerPage)

		users, err := c.FetchUsers(ctx, perPage, since)
		if err != nil {
			return allUsers, err
		}

		allUsers = append(allUsers, users...)

		// A short page means there are no more users
		if len(users) < perPage {
			break
		}

		// The next page starts after the last user ID of this one
		since = users[len(users)-1].ID
	}

	return allUsers, nil
}

// get sends a GET request for path and decodes the JSON response into out
func (c *Client) get(ctx context.Context, path string, out any) error {
	// Create the HTTP request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.config.BaseURL+path, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set required headers
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", c.config.APIVersion)
	req.Header.Set("User-Agent", userAgent)

	// Add GitHub token if available
	if c.config.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.config.Token)
	}

	// Execute the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Handle different status codes
	switch resp.StatusCode {
	case http.StatusOK:
		if err := json.Unmarshal(body, out); err != nil {
			return fmt.Errorf("failed to parse JSON response: %w", err)
		}
		return nil

	case http.StatusNotFound:
		return ErrNotFound

	default:
		var apiErr APIError
		if err := json.Unmarshal(body, &apiErr); err == nil {
			return fmt.Errorf("GitHub API error (%d): %s", resp.StatusCode, apiErr.Message)
		}
		return fmt.Errorf("GitHub API error (%d): %s", resp.StatusCode, string(body))
	}
}
