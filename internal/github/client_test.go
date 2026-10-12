package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/blastjax/golang-redis/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	cfg := &config.GitHubConfig{
		BaseURL:    "https://api.github.com",
		APIVersion: "2022-11-28",
		Token:      "test-token",
	}

	client := NewClient(cfg)

	assert.NotNil(t, client)
	assert.Equal(t, cfg, client.config)
	assert.NotNil(t, client.httpClient)
	assert.Equal(t, 30*time.Second, client.httpClient.Timeout)
}

func TestClient_FetchUsers(t *testing.T) {
	tests := []struct {
		name            string
		perPage         int
		since           int
		mockResponse    []User
		mockStatusCode  int
		mockError       *APIError
		expectedPerPage string
		expectedError   string
		expectedCount   int
	}{
		{
			name:    "successful fetch with valid parameters",
			perPage: 10,
			since:   5,
			mockResponse: []User{
				{ID: 6, Login: "user6", Type: "User"},
				{ID: 7, Login: "user7", Type: "User"},
			},
			mockStatusCode:  http.StatusOK,
			expectedPerPage: "10",
			expectedCount:   2,
		},
		{
			name:    "zero perPage defaults to maxUsersPerPage",
			perPage: 0,
			mockResponse: []User{
				{ID: 1, Login: "user1", Type: "User"},
			},
			mockStatusCode:  http.StatusOK,
			expectedPerPage: "100",
			expectedCount:   1,
		},
		{
			name:    "excessive perPage is capped to maxUsersPerPage",
			perPage: 500,
			mockResponse: []User{
				{ID: 1, Login: "user1", Type: "User"},
			},
			mockStatusCode:  http.StatusOK,
			expectedPerPage: "100",
			expectedCount:   1,
		},
		{
			name:           "API error response",
			perPage:        10,
			mockStatusCode: http.StatusForbidden,
			mockError: &APIError{
				Message:          "API rate limit exceeded",
				DocumentationURL: "https://docs.github.com/rest/overview/resources-in-the-rest-api#rate-limiting",
			},
			expectedPerPage: "10",
			expectedError:   "GitHub API error (403): API rate limit exceeded",
		},
		{
			name:            "non-JSON error response",
			perPage:         10,
			mockStatusCode:  http.StatusInternalServerError,
			expectedPerPage: "10",
			expectedError:   "GitHub API error (500): Internal Server Error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request headers
				assert.Equal(t, "application/vnd.github+json", r.Header.Get("Accept"))
				assert.Equal(t, "2022-11-28", r.Header.Get("X-GitHub-Api-Version"))
				assert.Equal(t, userAgent, r.Header.Get("User-Agent"))
				assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

				// Verify query parameters
				assert.Equal(t, "/users", r.URL.Path)
				assert.Equal(t, tt.expectedPerPage, r.URL.Query().Get("per_page"))
				if tt.since > 0 {
					assert.Equal(t, strconv.Itoa(tt.since), r.URL.Query().Get("since"))
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatusCode)

				if tt.mockStatusCode == http.StatusOK {
					json.NewEncoder(w).Encode(tt.mockResponse)
				} else if tt.mockError != nil {
					json.NewEncoder(w).Encode(tt.mockError)
				} else {
					w.Write([]byte("Internal Server Error"))
				}
			}))
			defer server.Close()

			// Create client with mock server URL
			cfg := &config.GitHubConfig{
				BaseURL:    server.URL,
				APIVersion: "2022-11-28",
				Token:      "test-token",
			}
			client := NewClient(cfg)

			// Execute test
			users, err := client.FetchUsers(context.Background(), tt.perPage, tt.since)

			// Assert results
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, users)
			} else {
				assert.NoError(t, err)
				assert.Len(t, users, tt.expectedCount)
				if tt.expectedCount > 0 {
					assert.Equal(t, tt.mockResponse, users)
				}
			}
		})
	}
}

// fakeUsersAPI serves GET /users like GitHub: each page starts after the "since"
// user ID, and only even IDs up to maxID exist (real GitHub user IDs have gaps too)
func fakeUsersAPI(maxID int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		since, _ := strconv.Atoi(r.URL.Query().Get("since"))
		perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))

		users := []User{}
		for id := since + 1; id <= maxID && len(users) < perPage; id++ {
			if id%2 == 0 {
				users = append(users, User{ID: id, Login: fmt.Sprintf("user%d", id), Type: "User"})
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(users)
	}))
}

func TestClient_FetchAllUsers(t *testing.T) {
	server := fakeUsersAPI(1000) // 500 users: IDs 2, 4, ..., 1000
	defer server.Close()

	client := NewClient(&config.GitHubConfig{BaseURL: server.URL})

	tests := []struct {
		name          string
		totalCount    int
		since         int
		expectedCount int
		expectedFirst int
	}{
		{name: "spans several pages", totalCount: 250, expectedCount: 250, expectedFirst: 2},
		{name: "starts after since", totalCount: 10, since: 100, expectedCount: 10, expectedFirst: 102},
		{name: "stops when users run out", totalCount: 600, expectedCount: 500, expectedFirst: 2},
		{name: "nothing after the last user", totalCount: 10, since: 1000, expectedCount: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			users, err := client.FetchAllUsers(context.Background(), tt.totalCount, tt.since)
			require.NoError(t, err)
			require.Len(t, users, tt.expectedCount)

			// Each page continues after the previous page's last ID: no duplicates, none skipped
			for i, user := range users {
				assert.Equal(t, tt.expectedFirst+2*i, user.ID)
			}
		})
	}
}

func TestClient_GetUserByUsername(t *testing.T) {
	tests := []struct {
		name           string
		username       string
		mockUser       *User
		mockStatusCode int
		mockError      *APIError
		expectedErr    error
		expectedError  string
	}{
		{
			name:     "successful user fetch",
			username: "testuser",
			mockUser: &User{
				ID:        123,
				Login:     "testuser",
				Type:      "User",
				AvatarURL: "https://github.com/images/error/testuser_happy.gif",
				URL:       "https://api.github.com/users/testuser",
			},
			mockStatusCode: http.StatusOK,
		},
		{
			name:           "user not found",
			username:       "nonexistentuser",
			mockStatusCode: http.StatusNotFound,
			expectedErr:    ErrNotFound,
		},
		{
			name:           "username is escaped, not sent as a query string",
			username:       "bad?name",
			mockStatusCode: http.StatusNotFound,
			expectedErr:    ErrNotFound,
		},
		{
			name:           "API error",
			username:       "testuser",
			mockStatusCode: http.StatusForbidden,
			mockError: &APIError{
				Message:          "API rate limit exceeded",
				DocumentationURL: "https://docs.github.com/rest/overview/resources-in-the-rest-api#rate-limiting",
			},
			expectedError: "GitHub API error (403): API rate limit exceeded",
		},
		{
			name:           "internal server error",
			username:       "testuser",
			mockStatusCode: http.StatusInternalServerError,
			expectedError:  "GitHub API error (500): Internal Server Error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request
				assert.Equal(t, "/users/"+tt.username, r.URL.Path)
				assert.Empty(t, r.URL.RawQuery)
				assert.Equal(t, "application/vnd.github+json", r.Header.Get("Accept"))
				assert.Equal(t, "2022-11-28", r.Header.Get("X-GitHub-Api-Version"))
				assert.Equal(t, userAgent, r.Header.Get("User-Agent"))
				assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatusCode)

				if tt.mockStatusCode == http.StatusOK && tt.mockUser != nil {
					json.NewEncoder(w).Encode(tt.mockUser)
				} else if tt.mockError != nil {
					json.NewEncoder(w).Encode(tt.mockError)
				} else if tt.mockStatusCode != http.StatusNotFound {
					w.Write([]byte("Internal Server Error"))
				}
			}))
			defer server.Close()

			cfg := &config.GitHubConfig{
				BaseURL:    server.URL,
				APIVersion: "2022-11-28",
				Token:      "test-token",
			}
			client := NewClient(cfg)

			user, err := client.GetUserByUsername(context.Background(), tt.username)

			switch {
			case tt.expectedErr != nil:
				assert.ErrorIs(t, err, tt.expectedErr)
				assert.Nil(t, user)
			case tt.expectedError != "":
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, user)
			default:
				assert.NoError(t, err)
				require.NotNil(t, user)
				assert.Equal(t, tt.mockUser.ID, user.ID)
				assert.Equal(t, tt.mockUser.Login, user.Login)
				assert.Equal(t, tt.mockUser.Type, user.Type)
				assert.Equal(t, tt.mockUser.AvatarURL, user.AvatarURL)
				assert.Equal(t, tt.mockUser.URL, user.URL)
			}
		})
	}
}

func TestClient_NetworkError(t *testing.T) {
	// Test network error handling by using a server that immediately closes
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Close connection immediately to simulate network error
		panic("connection closed")
	}))
	server.Close() // Close the server to simulate network failure

	cfg := &config.GitHubConfig{
		BaseURL:    server.URL,
		APIVersion: "2022-11-28",
		Token:      "test-token",
	}
	client := NewClient(cfg)

	users, err := client.FetchUsers(context.Background(), 10, 0)
	assert.Error(t, err)
	assert.Nil(t, users)
	assert.Contains(t, err.Error(), "failed to execute request")
}

func TestClient_InvalidJSON(t *testing.T) {
	// Test invalid JSON response handling
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	cfg := &config.GitHubConfig{
		BaseURL:    server.URL,
		APIVersion: "2022-11-28",
		Token:      "test-token",
	}
	client := NewClient(cfg)

	users, err := client.FetchUsers(context.Background(), 10, 0)
	assert.Error(t, err)
	assert.Nil(t, users)
	assert.Contains(t, err.Error(), "failed to parse JSON response")
}
