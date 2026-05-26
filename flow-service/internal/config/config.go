// Package config loads environment-driven settings.
package config

import (
	"errors"
	"os"
	"strconv"
)

// Config holds the runtime configuration for the service.
type Config struct {
	Port             int
	TableName        string
	DynamoDBEndpoint string // empty in production (use real AWS); set to http://dynamodb:8000 locally
	AWSRegion        string
	CORSOrigin       string
	LogLevel         string
	JWTSecret        string // required in production; set via JWT_SECRET env var
}

// Load reads configuration from environment variables, falling back to defaults
// suitable for local development.
func Load() Config {
	return Config{
		Port:             intEnv("PORT", 8080),
		TableName:        strEnv("TABLE_NAME", "flow-table"),
		DynamoDBEndpoint: strEnv("DYNAMODB_ENDPOINT", ""),
		AWSRegion:        strEnv("AWS_REGION", "us-east-1"),
		CORSOrigin:       strEnv("CORS_ORIGIN", "http://localhost:4200"),
		LogLevel:         strEnv("LOG_LEVEL", "WARN"),
		JWTSecret:        strEnv("JWT_SECRET", ""),
	}
}

// Validate returns an error if the config is unsafe for production use.
func (c Config) Validate() error {
	if c.CORSOrigin == "*" {
		return errors.New("CORS_ORIGIN='*' is not permitted; set an explicit allowed origin")
	}
	if c.DynamoDBEndpoint == "" && c.JWTSecret == "" {
		return errors.New("JWT_SECRET is required in production (set JWT_SECRET env var)")
	}
	return nil
}

func strEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func intEnv(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
