package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/blastjax/golang-redis/internal/config"
	"github.com/blastjax/golang-redis/internal/github"
	"github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRedisClient(t *testing.T) (*RedisClient, redismock.ClientMock) {
	client, mock := redismock.NewClientMock()

	redisClient := &RedisClient{
		client: client,
		ttl:    30 * time.Second,
	}

	return redisClient, mock
}

func createTestUser() *github.User {
	name := "Test User"
	company := "Test Company"
	return &github.User{
		ID:        123,
		Login:     "testuser",
		Type:      "User",
		Name:      &name,
		Company:   &company,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func TestNewRedisClient(t *testing.T) {
	// Nothing listens on port 1, so the ping fails fast
	cfg := &config.RedisConfig{
		Host: "127.0.0.1",
		Port: 1,
		TTL:  30 * time.Second,
	}

	client, err := NewRedisClient(cfg)
	assert.Error(t, err)
	assert.Nil(t, client)
	assert.Contains(t, err.Error(), "failed to connect to Redis")
}

func TestRedisClient_GetUser_CacheHit(t *testing.T) {
	redisClient, mock := setupTestRedisClient(t)
	defer redisClient.Close()

	testUser := createTestUser()
	userData, _ := json.Marshal(testUser)

	ctx := context.Background()
	key := "user:testuser"

	// Mock successful GET
	mock.ExpectGet(key).SetVal(string(userData))

	user, err := redisClient.GetUser(ctx, "testuser")
	assert.NoError(t, err)
	require.NotNil(t, user)
	assert.Equal(t, testUser.ID, user.ID)
	assert.Equal(t, testUser.Login, user.Login)
	assert.Equal(t, testUser.Type, user.Type)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRedisClient_GetUser_CacheMiss(t *testing.T) {
	redisClient, mock := setupTestRedisClient(t)
	defer redisClient.Close()

	ctx := context.Background()
	key := "user:testuser"

	// Mock cache miss (redis.Nil error)
	mock.ExpectGet(key).RedisNil()

	user, err := redisClient.GetUser(ctx, "testuser")
	assert.NoError(t, err)
	assert.Nil(t, user) // Should return nil for cache miss

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRedisClient_GetUser_CachedNotFound(t *testing.T) {
	redisClient, mock := setupTestRedisClient(t)
	defer redisClient.Close()

	// SetUser(nil) stores JSON null to mark a user that doesn't exist
	mock.ExpectGet("user:ghost").SetVal("null")

	user, err := redisClient.GetUser(context.Background(), "ghost")
	assert.ErrorIs(t, err, ErrNotFound)
	assert.Nil(t, user)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRedisClient_GetUser_RedisError(t *testing.T) {
	redisClient, mock := setupTestRedisClient(t)
	defer redisClient.Close()

	ctx := context.Background()
	key := "user:testuser"

	// Mock Redis error
	mock.ExpectGet(key).SetErr(fmt.Errorf("connection error"))

	user, err := redisClient.GetUser(ctx, "testuser")
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "failed to get user from Redis")

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRedisClient_GetUser_InvalidJSON(t *testing.T) {
	redisClient, mock := setupTestRedisClient(t)
	defer redisClient.Close()

	ctx := context.Background()
	key := "user:testuser"

	// Mock GET with invalid JSON
	mock.ExpectGet(key).SetVal("invalid json")

	user, err := redisClient.GetUser(ctx, "testuser")
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "failed to unmarshal user from Redis")

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRedisClient_SetUser(t *testing.T) {
	redisClient, mock := setupTestRedisClient(t)
	defer redisClient.Close()

	testUser := createTestUser()
	userData, _ := json.Marshal(testUser)

	ctx := context.Background()
	key := "user:testuser"

	// Mock successful SET
	mock.ExpectSet(key, userData, 30*time.Second).SetVal("OK")

	err := redisClient.SetUser(ctx, "testuser", testUser)
	assert.NoError(t, err)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRedisClient_SetUser_NilUser(t *testing.T) {
	redisClient, mock := setupTestRedisClient(t)
	defer redisClient.Close()

	// A nil user is stored as JSON null, the "does not exist" marker
	mock.ExpectSet("user:ghost", []byte("null"), 30*time.Second).SetVal("OK")

	err := redisClient.SetUser(context.Background(), "ghost", nil)
	assert.NoError(t, err)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRedisClient_SetUser_RedisError(t *testing.T) {
	redisClient, mock := setupTestRedisClient(t)
	defer redisClient.Close()

	testUser := createTestUser()
	userData, _ := json.Marshal(testUser)

	ctx := context.Background()
	key := "user:testuser"

	// Mock SET with error
	mock.ExpectSet(key, userData, 30*time.Second).SetErr(fmt.Errorf("connection error"))

	err := redisClient.SetUser(ctx, "testuser", testUser)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to set user in Redis")

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRedisClient_DeleteUser(t *testing.T) {
	redisClient, mock := setupTestRedisClient(t)
	defer redisClient.Close()

	ctx := context.Background()
	key := "user:testuser"

	// Mock successful DEL
	mock.ExpectDel(key).SetVal(1)

	err := redisClient.DeleteUser(ctx, "testuser")
	assert.NoError(t, err)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRedisClient_DeleteUser_RedisError(t *testing.T) {
	redisClient, mock := setupTestRedisClient(t)
	defer redisClient.Close()

	ctx := context.Background()
	key := "user:testuser"

	// Mock DEL with error
	mock.ExpectDel(key).SetErr(fmt.Errorf("connection error"))

	err := redisClient.DeleteUser(ctx, "testuser")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete user from Redis")

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRedisClient_Close(t *testing.T) {
	redisClient, _ := setupTestRedisClient(t)

	// Note: redismock doesn't support ExpectClose, so we just test that Close doesn't panic
	err := redisClient.Close()
	assert.NoError(t, err)
}

// Verify every operation uses the same key, lowercased because GitHub logins are case-insensitive
func TestRedisClient_KeyFormat(t *testing.T) {
	redisClient, mock := setupTestRedisClient(t)
	defer redisClient.Close()

	ctx := context.Background()
	username := "Test-User-123"
	expectedKey := "user:test-user-123"

	mock.ExpectGet(expectedKey).RedisNil()
	_, err := redisClient.GetUser(ctx, username)
	assert.NoError(t, err)

	testUser := createTestUser()
	userData, _ := json.Marshal(testUser)
	mock.ExpectSet(expectedKey, userData, 30*time.Second).SetVal("OK")
	err = redisClient.SetUser(ctx, username, testUser)
	assert.NoError(t, err)

	mock.ExpectDel(expectedKey).SetVal(1)
	err = redisClient.DeleteUser(ctx, username)
	assert.NoError(t, err)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}
