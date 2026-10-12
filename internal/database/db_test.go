package database

import (
	"os"
	"testing"

	"github.com/blastjax/golang-redis/internal/config"
	"github.com/blastjax/golang-redis/internal/github"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestNew_Unreachable(t *testing.T) {
	// Nothing listens on port 1, so the connection fails fast
	cfg := &config.DatabaseConfig{
		Host: "127.0.0.1",
		Port: 1,
		User: "testuser",
		Name: "testdb",
	}

	db, err := New(cfg)
	assert.Error(t, err)
	assert.Nil(t, db)
	assert.Contains(t, err.Error(), "failed to open database")
}

// The tests below run against a real PostgreSQL, using the same DB_* settings as the app
// (including the repo's .env). Each runs inside a transaction that is rolled back,
// so the database is left untouched. They are skipped unless TEST_POSTGRES is set:
//
//	TEST_POSTGRES=1 go test ./internal/database

// testID is far above real GitHub user IDs, so test rows never collide with synced users
const testID = 2_000_000_000

func testRepo(t *testing.T) *UserRepository {
	t.Helper()
	if os.Getenv("TEST_POSTGRES") == "" {
		t.Skip("set TEST_POSTGRES=1 to run the PostgreSQL integration tests")
	}

	// go test runs in the package directory, so load the repo's .env explicitly
	_ = godotenv.Load("../../.env")

	cfg, err := config.Load()
	require.NoError(t, err)
	db, err := New(&cfg.Database)
	require.NoError(t, err)
	require.NoError(t, db.CreateSchema())

	tx := db.Begin()
	t.Cleanup(func() {
		tx.Rollback()
		db.Close()
	})

	return NewUserRepository(&DB{tx})
}

func testUser(id int, login string) github.User {
	return github.User{ID: id, Login: login, Type: "User"}
}

func TestUserRepository_CreateUser(t *testing.T) {
	repo := testRepo(t)

	user := testUser(testID, "repo-test-user")
	require.NoError(t, repo.CreateUser(&user))

	// Logins are citext, so lookups ignore case like GitHub does
	got, err := repo.GetUserByUsername("Repo-Test-User")
	require.NoError(t, err)
	assert.Equal(t, testID, got.ID)
}

func TestUserRepository_CreateUserDuplicate(t *testing.T) {
	// The primary key and the case-insensitive unique login both reject duplicates
	duplicates := map[string]github.User{
		"same id":    testUser(testID, "repo-test-other"),
		"same login": testUser(testID+1, "REPO-TEST-USER"),
	}

	for name, dup := range duplicates {
		t.Run(name, func(t *testing.T) {
			// A fresh transaction per case: PostgreSQL aborts a transaction after an error
			repo := testRepo(t)

			user := testUser(testID, "repo-test-user")
			require.NoError(t, repo.CreateUser(&user))
			assert.ErrorIs(t, repo.CreateUser(&dup), gorm.ErrDuplicatedKey)
		})
	}
}

func TestUserRepository_NotFound(t *testing.T) {
	repo := testRepo(t)

	_, err := repo.GetUserByUsername("repo-test-missing")
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	assert.ErrorIs(t, repo.DeleteUser("repo-test-missing"), gorm.ErrRecordNotFound)
}

func TestUserRepository_UpdateAndDelete(t *testing.T) {
	repo := testRepo(t)

	user := testUser(testID, "repo-test-user")
	require.NoError(t, repo.CreateUser(&user))

	// Like the PUT handler: load the stored user, change a field, save it
	stored, err := repo.GetUserByUsername("repo-test-user")
	require.NoError(t, err)
	bio := "updated bio"
	stored.Bio = &bio
	require.NoError(t, repo.UpdateUser(stored))

	got, err := repo.GetUserByUsername("repo-test-user")
	require.NoError(t, err)
	require.NotNil(t, got.Bio)
	assert.Equal(t, "updated bio", *got.Bio)

	require.NoError(t, repo.DeleteUser("REPO-TEST-USER"))
	_, err = repo.GetUserByUsername("repo-test-user")
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestUserRepository_InsertUsersKeepsLocalEdits(t *testing.T) {
	repo := testRepo(t)

	name := "Local Name"
	user := testUser(testID, "repo-test-user")
	user.Name = &name
	require.NoError(t, repo.CreateUser(&user))

	// A re-sync only carries the fields GitHub's list endpoint returns
	synced := testUser(testID, "repo-test-user")
	synced.AvatarURL = "https://example.com/avatar.png"
	require.NoError(t, repo.InsertUsers([]github.User{synced}))

	got, err := repo.GetUserByUsername("repo-test-user")
	require.NoError(t, err)
	assert.Equal(t, "https://example.com/avatar.png", got.AvatarURL)
	require.NotNil(t, got.Name)
	assert.Equal(t, "Local Name", *got.Name)
}

func TestUserRepository_ListAndLastSyncedID(t *testing.T) {
	repo := testRepo(t)

	before, err := repo.GetTotalUsers()
	require.NoError(t, err)

	require.NoError(t, repo.InsertUsers([]github.User{
		testUser(testID, "repo-test-a"),
		testUser(testID+1, "repo-test-b"),
	}))

	total, err := repo.GetTotalUsers()
	require.NoError(t, err)
	assert.Equal(t, before+2, total)

	lastID, err := repo.GetLastSyncedUserID()
	require.NoError(t, err)
	assert.Equal(t, testID+1, lastID)

	// Users are ordered by ID, so the test users are the last page
	users, err := repo.GetUsers(2, total-2)
	require.NoError(t, err)
	require.Len(t, users, 2)
	assert.Equal(t, "repo-test-a", users[0].Login)
	assert.Equal(t, "repo-test-b", users[1].Login)
}
