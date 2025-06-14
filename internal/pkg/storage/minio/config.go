package minio

import (
	"time"
)

// Config holds the configuration for the MinIO client
type Config struct {
	// Endpoint is the MinIO server endpoint
	Endpoint string

	// AccessKeyID is the access key for MinIO
	AccessKeyID string

	// SecretAccessKey is the secret key for MinIO
	SecretAccessKey string

	// UseSSL indicates whether to use SSL/TLS
	UseSSL bool

	// Region is the MinIO region
	Region string

	// Bucket is the default bucket name
	Bucket string

	// DefaultExpiry is the default expiry time for presigned URLs
	DefaultExpiry time.Duration
}
