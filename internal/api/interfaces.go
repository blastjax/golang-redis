package api

import (
	"context"

	"github.com/blastjax/golang-redis/internal/github"
)

// UserRepository defines the database operations the handlers use
type UserRepository interface {
	GetUsers(limit, offset int) ([]github.User, error)
	GetTotalUsers() (int, error)
	GetUserByUsername(username string) (*github.User, error)
	CreateUser(user *github.User) error
	UpdateUser(user *github.User) error
	DeleteUser(username string) error
}

// RedisClient defines the cache operations the handlers use
type RedisClient interface {
	GetUser(ctx context.Context, username string) (*github.User, error)
	SetUser(ctx context.Context, username string, user *github.User) error
	DeleteUser(ctx context.Context, username string) error
}

// GitHubClient defines the GitHub API operations the handlers use
type GitHubClient interface {
	GetUserByUsername(ctx context.Context, username string) (*github.User, error)
}
