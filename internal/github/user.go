package github

import (
	"context"
	"net/url"
)

// GetUserByUsername fetches a specific user by username from GitHub API.
// It returns ErrNotFound if GitHub has no such user.
func (c *Client) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	// Escape the username so it can't change the request path or add a query string
	var user User
	if err := c.get(ctx, "/users/"+url.PathEscape(username), &user); err != nil {
		return nil, err
	}

	return &user, nil
}
