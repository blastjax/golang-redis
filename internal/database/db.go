package database

import (
	"fmt"
	"log"
	"time"

	"github.com/blastjax/golang-redis/internal/config"
	"github.com/blastjax/golang-redis/internal/github"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB represents a database connection using GORM
type DB struct {
	*gorm.DB
}

// New creates a new database connection using GORM
func New(cfg *config.DatabaseConfig) (*DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.GetDSN()), &gorm.Config{
		// Log slow queries and errors to stderr, so stdout stays clean for -format json
		Logger: logger.New(log.Default(), logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: true,
		}),
		// Report unique constraint violations as gorm.ErrDuplicatedKey
		TranslateError: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Get underlying sql.DB to configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxOpenConns(cfg.MaxConnections)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConnections)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Test the connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{db}, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}
	return sqlDB.Close()
}

// CreateSchema creates the necessary database schema using GORM auto-migration
func (db *DB) CreateSchema() error {
	// citext makes login comparisons case-insensitive, like GitHub logins
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS citext").Error; err != nil {
		return fmt.Errorf("failed to create citext extension: %w", err)
	}

	// Auto-migrate the User model
	if err := db.AutoMigrate(&github.User{}); err != nil {
		return fmt.Errorf("failed to auto-migrate schema: %w", err)
	}

	return nil
}
