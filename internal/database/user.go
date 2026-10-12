package database

import (
	"fmt"

	"github.com/blastjax/golang-redis/internal/github"
	"gorm.io/gorm"
)

// GetUserByUsername retrieves a user by username from the database using GORM.
// A missing user returns an error wrapping gorm.ErrRecordNotFound.
func (r *UserRepository) GetUserByUsername(username string) (*github.User, error) {
	var user github.User
	if err := r.db.Where("login = ?", username).First(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to get user %s: %w", username, err)
	}

	return &user, nil
}

// UpdateUser saves all fields of an existing user using GORM, which also sets UpdatedAt
func (r *UserRepository) UpdateUser(user *github.User) error {
	if err := r.db.Save(user).Error; err != nil {
		return fmt.Errorf("failed to update user %s: %w", user.Login, err)
	}

	return nil
}

// DeleteUser deletes a user from the database by username using GORM.
// A missing user returns gorm.ErrRecordNotFound.
func (r *UserRepository) DeleteUser(username string) error {
	result := r.db.Where("login = ?", username).Delete(&github.User{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete user %s: %w", username, result.Error)
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

// CreateUser inserts a new user into the database using GORM, which sets CreatedAt
// and UpdatedAt. A duplicate ID or login returns an error wrapping gorm.ErrDuplicatedKey.
func (r *UserRepository) CreateUser(user *github.User) error {
	if err := r.db.Create(user).Error; err != nil {
		return fmt.Errorf("failed to create user %s: %w", user.Login, err)
	}

	return nil
}
