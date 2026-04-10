package minio

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"path"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/usernamesalah/rh-pos/internal/pkg/ctxkey"
	"github.com/usernamesalah/rh-pos/internal/pkg/hash"
	"github.com/usernamesalah/rh-pos/internal/pkg/storage"
)

const maxDownloadSize = 10 << 20 // 10 MB

// Client implements the StorageClient interface for MinIO
type Client struct {
	client *minio.Client
	config *Config
	logger *slog.Logger
}

// NewClient creates a new MinIO client with the given configuration
func NewClient(config *Config, logger *slog.Logger) (*Client, error) {
	if config == nil {
		return nil, fmt.Errorf("config is required")
	}

	logger.Info("initializing MinIO client", "endpoint", config.Endpoint, "bucket", config.Bucket)

	// Create MinIO client
	minioClient, err := minio.New(config.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.AccessKeyID, config.SecretAccessKey, ""),
		Secure: config.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	// Check if bucket exists
	exists, err := minioClient.BucketExists(context.Background(), config.Bucket)
	if err != nil {
		return nil, fmt.Errorf("failed to check if bucket exists: %w", err)
	}

	logger.Info("bucket check", "bucket", config.Bucket, "exists", exists)

	// Create bucket if it doesn't exist
	if !exists {
		logger.Info("creating bucket", "bucket", config.Bucket)
		err = minioClient.MakeBucket(context.Background(), config.Bucket, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	return &Client{
		client: minioClient,
		config: config,
		logger: logger,
	}, nil
}

// getTenantIDFromContext extracts and hashes the tenant ID from context
func (c *Client) getTenantIDFromContext(ctx context.Context) (string, error) {
	tenantID, ok := ctxkey.TenantIDFromContext(ctx)
	if !ok {
		return "", fmt.Errorf("tenant ID not found in context")
	}

	hashedID := hash.HashID(tenantID)
	return hashedID, nil
}

// getTenantKey returns the full key with tenant prefix
func (c *Client) getTenantKey(ctx context.Context, key string) (string, error) {
	tenantID, err := c.getTenantIDFromContext(ctx)
	if err != nil {
		return "", err
	}
	return path.Join(tenantID, key), nil
}

// Upload uploads an object to MinIO
func (c *Client) Upload(ctx context.Context, key string, reader io.Reader, contentType string) error {
	objectKey, err := c.getTenantKey(ctx, key)
	if err != nil {
		return storage.NewStorageError("upload", key, err)
	}

	_, err = c.client.PutObject(ctx, c.config.Bucket, objectKey, reader, -1,
		minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return storage.NewStorageError("upload", key, err)
	}
	return nil
}

// UploadBytes uploads a byte array to MinIO
func (c *Client) UploadBytes(ctx context.Context, key string, data []byte, contentType string) error {
	reader := bytes.NewReader(data)
	return c.Upload(ctx, key, reader, contentType)
}

// Download downloads an object from MinIO
func (c *Client) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	objectKey, err := c.getTenantKey(ctx, key)
	if err != nil {
		return nil, storage.NewStorageError("download", key, err)
	}

	object, err := c.client.GetObject(ctx, c.config.Bucket, objectKey, minio.GetObjectOptions{})
	if err != nil {
		return nil, storage.NewStorageError("download", key, err)
	}
	return object, nil
}

// DownloadBytes downloads an object from MinIO and returns its contents as bytes.
// Returns an error if the object exceeds maxDownloadSize (10 MB).
func (c *Client) DownloadBytes(ctx context.Context, key string) ([]byte, error) {
	reader, err := c.Download(ctx, key)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	limited := io.LimitReader(reader, maxDownloadSize+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, fmt.Errorf("failed to read object: %w", err)
	}
	if int64(len(data)) > maxDownloadSize {
		return nil, fmt.Errorf("object exceeds maximum allowed size of %d bytes", maxDownloadSize)
	}

	return data, nil
}

// Delete deletes an object from MinIO
func (c *Client) Delete(ctx context.Context, key string) error {
	objectKey, err := c.getTenantKey(ctx, key)
	if err != nil {
		return storage.NewStorageError("delete", key, err)
	}

	err = c.client.RemoveObject(ctx, c.config.Bucket, objectKey, minio.RemoveObjectOptions{})
	if err != nil {
		return storage.NewStorageError("delete", key, err)
	}
	return nil
}

// List lists objects in a prefix
func (c *Client) List(ctx context.Context, prefix string) ([]storage.ObjectInfo, error) {
	tenantID, err := c.getTenantIDFromContext(ctx)
	if err != nil {
		return nil, storage.NewStorageError("list", prefix, err)
	}

	tenantPrefix := path.Join(tenantID, prefix)
	objects := make([]storage.ObjectInfo, 0)

	objectCh := c.client.ListObjects(ctx, c.config.Bucket, minio.ListObjectsOptions{
		Prefix:    tenantPrefix,
		Recursive: true,
	})

	for object := range objectCh {
		if object.Err != nil {
			return nil, storage.NewStorageError("list", prefix, object.Err)
		}
		objects = append(objects, storage.ObjectInfo{
			Key:          object.Key,
			Size:         object.Size,
			LastModified: object.LastModified,
			ETag:         object.ETag,
		})
	}

	return objects, nil
}

// GeneratePresignedURL generates a presigned URL for upload or download
func (c *Client) GeneratePresignedURL(ctx context.Context, key string, expiry time.Duration, isUpload bool) (string, error) {
	if expiry == 0 {
		expiry = c.config.DefaultExpiry
	}

	objectKey, err := c.getTenantKey(ctx, key)
	if err != nil {
		return "", storage.NewStorageError("presign", key, err)
	}

	c.logger.DebugContext(ctx, "generating presigned URL", "key", objectKey, "is_upload", isUpload)

	if isUpload {
		presignedURL, err := c.client.PresignedPutObject(ctx, c.config.Bucket, objectKey, expiry)
		if err != nil {
			return "", storage.NewStorageError("presign", key, err)
		}
		return presignedURL.String(), nil
	}

	presignedURL, err := c.client.PresignedGetObject(ctx, c.config.Bucket, objectKey, expiry, nil)
	if err != nil {
		return "", storage.NewStorageError("presign", key, err)
	}

	return presignedURL.String(), nil
}
