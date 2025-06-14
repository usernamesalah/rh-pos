package minio

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/usernamesalah/rh-pos/internal/pkg/hash"
)

// Client implements the StorageClient interface for MinIO
type Client struct {
	client *minio.Client
	config *Config
}

// NewClient creates a new MinIO client with the given configuration
func NewClient(config *Config) (*Client, error) {
	if config == nil {
		return nil, fmt.Errorf("config is required")
	}

	fmt.Printf("Initializing MinIO client with endpoint: %s, bucket: %s\n", config.Endpoint, config.Bucket)

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

	fmt.Printf("Bucket %s exists: %v\n", config.Bucket, exists)

	// Create bucket if it doesn't exist
	if !exists {
		fmt.Printf("Creating bucket %s\n", config.Bucket)
		err = minioClient.MakeBucket(context.Background(), config.Bucket, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	client := &Client{
		client: minioClient,
		config: config,
	}

	return client, nil
}

// getTenantIDFromContext extracts and hashes the tenant ID from context
func (c *Client) getTenantIDFromContext(ctx context.Context) (string, error) {
	tenantID, ok := ctx.Value("tenant_id").(uint)
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
		return NewStorageError("upload", key, err)
	}

	_, err = c.client.PutObject(ctx, c.config.Bucket, objectKey, reader, -1,
		minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return NewStorageError("upload", key, err)
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
		return nil, NewStorageError("download", key, err)
	}

	object, err := c.client.GetObject(ctx, c.config.Bucket, objectKey, minio.GetObjectOptions{})
	if err != nil {
		return nil, NewStorageError("download", key, err)
	}
	return object, nil
}

// DownloadBytes downloads an object from MinIO and returns its contents as bytes
func (c *Client) DownloadBytes(ctx context.Context, key string) ([]byte, error) {
	reader, err := c.Download(ctx, key)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	return io.ReadAll(reader)
}

// Delete deletes an object from MinIO
func (c *Client) Delete(ctx context.Context, key string) error {
	objectKey, err := c.getTenantKey(ctx, key)
	if err != nil {
		return NewStorageError("delete", key, err)
	}

	err = c.client.RemoveObject(ctx, c.config.Bucket, objectKey, minio.RemoveObjectOptions{})
	if err != nil {
		return NewStorageError("delete", key, err)
	}
	return nil
}

// List lists objects in a prefix
func (c *Client) List(ctx context.Context, prefix string) ([]ObjectInfo, error) {
	tenantID, err := c.getTenantIDFromContext(ctx)
	if err != nil {
		return nil, NewStorageError("list", prefix, err)
	}

	tenantPrefix := path.Join(tenantID, prefix)
	objects := make([]ObjectInfo, 0)

	objectCh := c.client.ListObjects(ctx, c.config.Bucket, minio.ListObjectsOptions{
		Prefix:    tenantPrefix,
		Recursive: true,
	})

	for object := range objectCh {
		if object.Err != nil {
			return nil, NewStorageError("list", prefix, object.Err)
		}
		objects = append(objects, ObjectInfo{
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
		return "", NewStorageError("presign", key, err)
	}

	fmt.Printf("Generating presigned URL for key: %s, isUpload: %v\n", objectKey, isUpload)

	if isUpload {
		// Generate presigned URL for download
		presignedURL, err := c.client.PresignedPutObject(ctx, c.config.Bucket, objectKey, expiry)
		if err != nil {
			return "", NewStorageError("presign", key, err)
		}
		return presignedURL.String(), nil
	}

	// Generate presigned URL for download
	presignedURL, err := c.client.PresignedGetObject(ctx, c.config.Bucket, objectKey, expiry, nil)
	if err != nil {
		return "", NewStorageError("presign", key, err)
	}

	return presignedURL.String(), nil
}
