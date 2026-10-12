package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDatabaseConfig_GetDSN(t *testing.T) {
	tests := []struct {
		name     string
		password string
		expected string
	}{
		{name: "with password", password: "testpass", expected: "postgres://testuser:testpass@localhost:5432/testdb"},
		{name: "empty password", password: "", expected: "postgres://testuser:@localhost:5432/testdb"},
		{name: "special characters are escaped", password: "p@ss/w:rd", expected: "postgres://testuser:p%40ss%2Fw%3Ard@localhost:5432/testdb"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &DatabaseConfig{
				Host:     "localhost",
				Port:     5432,
				User:     "testuser",
				Password: tt.password,
				Name:     "testdb",
			}

			assert.Equal(t, tt.expected, cfg.GetDSN())
		})
	}
}
