package config

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	Database DatabaseConfig
	GitHub   GitHubConfig
	Redis    RedisConfig
	API      APIConfig
	App      AppConfig
}

// DatabaseConfig holds database-related configuration
type DatabaseConfig struct {
	Host               string
	Port               int
	User               string
	Password           string
	Name               string
	MaxConnections     int
	MaxIdleConnections int
}

// GitHubConfig holds GitHub API-related configuration
type GitHubConfig struct {
	Token      string
	BaseURL    string
	APIVersion string
}

// RedisConfig holds Redis-related configuration
type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
	TTL      time.Duration
}

// APIConfig holds API server configuration
type APIConfig struct {
	Host string
	Port int
}

// AppConfig holds general application configuration
type AppConfig struct {
	Environment string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Try to load .env file (ignore error if file doesn't exist)
	_ = godotenv.Load()

	config := &Config{
		Database: DatabaseConfig{
			Host:               getEnv("DB_HOST", "localhost"),
			Port:               getEnvAsInt("DB_PORT", 5432),
			User:               getEnv("DB_USER", "postgres"),
			Password:           getEnv("DB_PASSWORD", ""),
			Name:               getEnv("DB_NAME", "github_users"),
			MaxConnections:     getEnvAsInt("DB_MAX_CONNECTIONS", 10),
			MaxIdleConnections: getEnvAsInt("DB_MAX_IDLE_CONNECTIONS", 5),
		},
		GitHub: GitHubConfig{
			Token:      getEnv("GITHUB_TOKEN", ""),
			BaseURL:    getEnv("GITHUB_API_BASE_URL", "https://api.github.com"),
			APIVersion: getEnv("GITHUB_API_VERSION", "2022-11-28"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnvAsInt("REDIS_PORT", 6379),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
			TTL:      time.Duration(getEnvAsInt("REDIS_TTL_SECONDS", 30)) * time.Second,
		},
		API: APIConfig{
			Host: getEnv("API_HOST", "localhost"),
			Port: getEnvAsInt("API_PORT", 8080),
		},
		App: AppConfig{
			Environment: getEnv("APP_ENV", "development"),
		},
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return config, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Database.Host == "" {
		return fmt.Errorf("DB_HOST is required")
	}
	if c.Database.User == "" {
		return fmt.Errorf("DB_USER is required")
	}
	if c.Database.Name == "" {
		return fmt.Errorf("DB_NAME is required")
	}
	return nil
}

// GetDSN returns the PostgreSQL connection URL. url.UserPassword escapes special
// characters in the credentials. With no sslmode in the URL, pgx tries TLS first and
// falls back to an unencrypted connection; set PGSSLMODE=require to insist on TLS.
func (c *DatabaseConfig) GetDSN() string {
	dsn := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(c.User, c.Password),
		Host:   net.JoinHostPort(c.Host, strconv.Itoa(c.Port)),
		Path:   "/" + c.Name,
	}
	return dsn.String()
}

// getEnv gets an environment variable with a fallback value
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

// getEnvAsInt gets an environment variable as integer with a fallback value
func getEnvAsInt(name string, fallback int) int {
	valueStr := getEnv(name, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return fallback
}
