package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/blastjax/golang-redis/internal/config"
	"github.com/blastjax/golang-redis/internal/github"

	"github.com/redis/go-redis/v9"
)

// ErrNotFound is returned by GetUser when the cache holds a "user does not exist" entry
var ErrNotFound = errors.New("cache: user does not exist")

// RedisClient wraps the Redis client with application-specific methods
type RedisClient struct {
	client *redis.Client
	ttl    time.Duration
}

// NewRedisClient creates a new Redis client
func NewRedisClient(cfg *config.RedisConfig) (*RedisClient, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisClient{
		client: rdb,
		ttl:    cfg.TTL,
	}, nil
}

// Close closes the Redis connection
func (r *RedisClient) Close() error {
	return r.client.Close()
}

// userKey returns the cache key for a username. GitHub logins are case-insensitive,
// so every spelling of a login shares one entry.
func userKey(username string) string {
	return "user:" + strings.ToLower(username)
}

// GetUser retrieves a user from Redis cache by username. It returns nil, nil on a
// cache miss and ErrNotFound if the user is cached as not existing.
func (r *RedisClient) GetUser(ctx context.Context, username string) (*github.User, error) {
	val, err := r.client.Get(ctx, userKey(username)).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Cache miss
		}
		return nil, fmt.Errorf("failed to get user from Redis: %w", err)
	}

	// SetUser stores a nil user as JSON null
	if val == "null" {
		return nil, ErrNotFound
	}

	var user github.User
	if err := json.Unmarshal([]byte(val), &user); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user from Redis: %w", err)
	}

	return &user, nil
}

// SetUser stores a user in Redis cache with TTL. A nil user records that the user
// does not exist, so repeated lookups don't reach GitHub.
func (r *RedisClient) SetUser(ctx context.Context, username string, user *github.User) error {
	data, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("failed to marshal user for Redis: %w", err)
	}

	if err := r.client.Set(ctx, userKey(username), data, r.ttl).Err(); err != nil {
		return fmt.Errorf("failed to set user in Redis: %w", err)
	}

	return nil
}

// DeleteUser removes a user from Redis cache
func (r *RedisClient) DeleteUser(ctx context.Context, username string) error {
	if err := r.client.Del(ctx, userKey(username)).Err(); err != nil {
		return fmt.Errorf("failed to delete user from Redis: %w", err)
	}

	return nil
}
