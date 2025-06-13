package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all configuration for our application
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Logger   LoggerConfig
	Admin    AdminConfig
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port string
	Host string
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	DSN      string
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret string
}

// LoggerConfig holds logger configuration
type LoggerConfig struct {
	Level string
}

// AdminConfig holds admin configuration
type AdminConfig struct {
	Username string
	Password string
}

// Load loads configuration from .env file and environment variables
func Load() (*Config, error) {
	// Load .env file if it exists (ignore error if file doesn't exist)
	if err := godotenv.Load(); err != nil {
		// Only log the error, don't fail - useful for Docker environments
		fmt.Printf("Warning: Could not load .env file: %v\n", err)
	}

	config := &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "3306"),
			User:     getEnv("DB_USER", "root"),
			Password: getEnv("DB_PASSWORD", ""),
			Name:     getEnv("DB_NAME", "rh_pos"),
		},
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", "your-super-secret-jwt-key"),
		},
		Logger: LoggerConfig{
			Level: getEnv("LOG_LEVEL", "info"),
		},
		Admin: AdminConfig{
			Username: getEnv("ADMIN_USERNAME", ""),
			Password: getEnv("ADMIN_PASSWORD", ""),
		},
	}

	// Validate required fields
	if config.JWT.Secret == "your-super-secret-jwt-key" {
		return nil, fmt.Errorf("JWT_SECRET must be set to a secure value")
	}

	if config.Database.Name == "" {
		return nil, fmt.Errorf("DB_NAME is required")
	}

	if config.Admin.Username == "" || config.Admin.Password == "" {
		return nil, fmt.Errorf("ADMIN_USERNAME and ADMIN_PASSWORD are required")
	}

	// Construct DSN
	config.Database.DSN = fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.Database.User,
		config.Database.Password,
		config.Database.Host,
		config.Database.Port,
		config.Database.Name,
	)

	return config, nil
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
