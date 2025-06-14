package minio

import (
	"context"
	"io"
	"time"
)

// ObjectInfo represents metadata about an object in storage
type ObjectInfo struct {
	Key          string
	Size         int64
	LastModified time.Time
	ETag         string
}

// StorageClient defines the interface for storage operations
type StorageClient interface {
	// Upload uploads an object to storage
	Upload(ctx context.Context, key string, reader io.Reader, contentType string) error

	// UploadBytes uploads a byte array to storage
	UploadBytes(ctx context.Context, key string, data []byte, contentType string) error

	// Download downloads an object from storage
	Download(ctx context.Context, key string) (io.ReadCloser, error)

	// DownloadBytes downloads a byte array from storage
	DownloadBytes(ctx context.Context, key string) ([]byte, error)

	// Delete deletes an object from storage
	Delete(ctx context.Context, key string) error

	// List lists objects in a prefix
	List(ctx context.Context, prefix string) ([]ObjectInfo, error)

	// GeneratePresignedURL generates a presigned URL for upload or download
	GeneratePresignedURL(ctx context.Context, key string, expiry time.Duration, isUpload bool) (string, error)
}
