package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// Config contains all application settings
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Kafka    KafkaConfig
}

// ServerConfig defines settings for HTTP server
type ServerConfig struct {
	Host         string
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// DatabaseConfig defines settings for database
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// JWTConfig defines settings for JWT tokens
type JWTConfig struct {
	Secret    string
	ExpiresIn time.Duration
}

// KafkaConfig defines settings for Kafka
type KafkaConfig struct {
	Brokers []string
	Topic   string
}

// Load sets up configuration with env variables
func Load() (*Config, error) {
	config := &Config{
		Server: ServerConfig{
			Host:         getEnv("SERVER_HOST", "localhost"),
			Port:         getEnv("SERVER_PORT", "8080"),
			ReadTimeout:  parseDuration(getEnv("SERVER_READ_TIMEOUT", "10s")),
			WriteTimeout: parseDuration(getEnv("SERVER_WRITE_TIMEOUT", "10s")),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			DBName:   getEnv("DB_NAME", "talkify"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
		},
		JWT: JWTConfig{
			Secret:    getEnv("JWT_SECRET", ""),
			ExpiresIn: parseDuration(getEnv("JWT_EXPIRES_IN", "24h")),
		},
		Kafka: KafkaConfig{
			Brokers: parseStringSlice(getEnv("KAFKA_BROKERS", "localhost:9092")),
			Topic:   getEnv("KAFKA_TOPIC", "talkify-messages"),
		},
	}

	if err := config.validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return config, nil
}

// validate checks required configuration parameters
func (c *Config) validate() error {
	if c.Database.Password == "" {
		return fmt.Errorf("DB_PASSWORD is required")
	}

	if c.JWT.Secret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}

	if len(c.JWT.Secret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters")
	}

	return nil
}

// getEnv retrieves value of env variable or returns default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return defaultValue
}

// parseDuration parses string in time.Duration with errors handling
func parseDuration(s string) time.Duration {
	duration, err := time.ParseDuration(s)
	if err != nil {
		return 10 * time.Second
	}
	return duration
}

// parseStringSlice parses string with splitter into slice of strings
func parseStringSlice(s string) []string {
	if s == "" {
		return []string{}
	}

	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}

// GetDSN returns PostgreSQL connection string
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode,
	)
}

// GetServerAddress returns server address in host:port format
func (c *ServerConfig) GetServerAddress() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}
