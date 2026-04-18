package storage

import (
	"context"
	"io"
	"time"
)

// ObjectInfo represents metadata about an object in storage.
type ObjectInfo struct {
	Key          string
	Size         int64
	LastModified time.Time
	ETag         string
}

// StorageClient defines the interface for storage operations.
// Both the S3-compatible and local filesystem implementations satisfy this interface.
type StorageClient interface {
	Upload(ctx context.Context, key string, reader io.Reader, contentType string) error
	UploadBytes(ctx context.Context, key string, data []byte, contentType string) error
	Download(ctx context.Context, key string) (io.ReadCloser, error)
	DownloadBytes(ctx context.Context, key string) ([]byte, error)
	Delete(ctx context.Context, key string) error
	List(ctx context.Context, prefix string) ([]ObjectInfo, error)
	// GeneratePresignedURL generates a pre-signed URL for direct client upload/download.
	// Returns ErrNotSupported for backends that do not support pre-signed URLs (e.g. local).
	GeneratePresignedURL(ctx context.Context, key string, expiry time.Duration, isUpload bool) (string, error)
}
