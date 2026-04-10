package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all configuration for our application
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Logger   LoggerConfig
	Admin    AdminConfig
	Storage  StorageConfig
	MinIO    MinIOConfig
	Hash     HashConfig
}

// StorageConfig selects the storage backend and its settings.
type StorageConfig struct {
	// Type is "minio" or "local".
	Type string
	// LocalPath is the base directory for local storage (only used when Type == "local").
	LocalPath string
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port           string
	Host           string
	AllowedOrigins []string
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

// MinIOConfig holds MinIO configuration
type MinIOConfig struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	UseSSL          bool
	Region          string
	Bucket          string
	DefaultExpiry   time.Duration
}

// HashConfig holds hash ID configuration
type HashConfig struct {
	Salt string
}

// Load loads configuration from .env file and environment variables
func Load() (*Config, error) {
	// Load .env file if it exists (ignore error if file doesn't exist)
	if err := godotenv.Load(); err != nil {
		// Only log the error, don't fail - useful for Docker environments
		fmt.Printf("Warning: Could not load .env file: %v\n", err)
	}

	allowedOrigins := strings.Split(getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000"), ",")

	config := &Config{
		Server: ServerConfig{
			Port:           getEnv("SERVER_PORT", "8080"),
			Host:           getEnv("SERVER_HOST", "0.0.0.0"),
			AllowedOrigins: allowedOrigins,
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
		Storage: StorageConfig{
			Type:      getEnv("STORAGE_TYPE", "minio"),
			LocalPath: getEnv("LOCAL_STORAGE_PATH", "./storage"),
		},
		MinIO: MinIOConfig{
			Endpoint:        getEnv("MINIO_ENDPOINT", "minio:9000"),
			AccessKeyID:     getEnv("MINIO_ACCESS_KEY", ""),
			SecretAccessKey: getEnv("MINIO_SECRET_KEY", ""),
			UseSSL:          getEnv("MINIO_USE_SSL", "false") == "true",
			Region:          getEnv("MINIO_REGION", "us-east-1"),
			Bucket:          getEnv("MINIO_BUCKET", "rh-pos"),
			DefaultExpiry:   time.Hour * 1, // 1 hour default expiry
		},
		Hash: HashConfig{
			Salt: getEnv("HASHID_SALT", ""),
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

	// Validate MinIO configuration (only required when STORAGE_TYPE=minio)
	if config.Storage.Type == "minio" {
		if config.MinIO.AccessKeyID == "" || config.MinIO.SecretAccessKey == "" {
			return nil, fmt.Errorf("MINIO_ACCESS_KEY and MINIO_SECRET_KEY are required when STORAGE_TYPE=minio")
		}
	}

	// Validate HashID salt
	if config.Hash.Salt == "" {
		return nil, fmt.Errorf("HASHID_SALT is required (min 16 characters)")
	}
	if len(config.Hash.Salt) < 16 {
		return nil, fmt.Errorf("HASHID_SALT must be at least 16 characters")
	}

	// Convert local storage path to absolute path
	if config.Storage.Type == "local" && config.Storage.LocalPath != "" {
		absPath, err := filepath.Abs(config.Storage.LocalPath)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve local storage path: %w", err)
		}
		config.Storage.LocalPath = absPath
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
