package database

import (
	"errors"
	"fmt"
	"time"

	"github.com/blastjax/golang-redis/internal/github"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// UserRepository handles user database operations
type UserRepository struct {
	db *DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *DB) *UserRepository {
	return &UserRepository{db: db}
}

// InsertUsers upserts users from GitHub's list endpoint in batches using GORM
func (r *UserRepository) InsertUsers(users []github.User) error {
	if len(users) == 0 {
		return nil
	}

	// Set SyncedAt for all users
	for i := range users {
		users[i].SyncedAt = time.Now()
	}

	// On conflict, refresh only the fields the list endpoint returns, so name, bio,
	// etc. edited through the API survive a re-sync. CreateInBatches runs multiple
	// batches in one transaction.
	result := r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"login", "avatar_url", "url", "type", "synced_at"}),
	}).CreateInBatches(users, 100) // Insert in batches of 100

	if result.Error != nil {
		return fmt.Errorf("failed to batch insert users: %w", result.Error)
	}

	return nil
}

// GetUsers retrieves users with pagination using GORM
func (r *UserRepository) GetUsers(limit, offset int) ([]github.User, error) {
	var users []github.User
	result := r.db.Order("id").Limit(limit).Offset(offset).Find(&users)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to query users: %w", result.Error)
	}

	return users, nil
}

// GetTotalUsers returns the total number of users in the database using GORM
func (r *UserRepository) GetTotalUsers() (int, error) {
	var count int64
	result := r.db.Model(&github.User{}).Count(&count)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to count users: %w", result.Error)
	}
	return int(count), nil
}

// GetLastSyncedUserID returns the highest user ID that has been synced using GORM
func (r *UserRepository) GetLastSyncedUserID() (int, error) {
	var user github.User
	result := r.db.Order("id desc").First(&user)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return 0, nil // No users found
		}
		return 0, fmt.Errorf("failed to get last synced user ID: %w", result.Error)
	}

	return user.ID, nil
}
