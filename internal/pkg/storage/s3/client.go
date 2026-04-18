package s3

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"path"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/usernamesalah/rh-pos/internal/pkg/ctxkey"
	"github.com/usernamesalah/rh-pos/internal/pkg/hash"
	"github.com/usernamesalah/rh-pos/internal/pkg/storage"
)

const maxDownloadSize = 10 << 20 // 10 MB

type Client struct {
	client *s3.Client
	bucket string
	logger *slog.Logger
}

func NewClient(cfg *Config, logger *slog.Logger) (*Client, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is required")
	}

	logger.Info("initializing S3 client", "endpoint", cfg.Endpoint, "bucket", cfg.Bucket)

	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AccessKeyID,
			cfg.SecretAccessKey,
			"",
		)),
		awsconfig.WithRegion(cfg.Region),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	s3Client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(cfg.Endpoint)
		o.UsePathStyle = false
	})

	return &Client{
		client: s3Client,
		bucket: cfg.Bucket,
		logger: logger,
	}, nil
}

func (c *Client) getTenantIDFromContext(ctx context.Context) (string, error) {
	tenantID, ok := ctxkey.TenantIDFromContext(ctx)
	if !ok {
		return "", fmt.Errorf("tenant ID not found in context")
	}

	hashedID := hash.HashID(tenantID)
	return hashedID, nil
}

func (c *Client) getTenantKey(ctx context.Context, key string) (string, error) {
	tenantID, err := c.getTenantIDFromContext(ctx)
	if err != nil {
		return "", err
	}
	return path.Join(tenantID, key), nil
}

func (c *Client) Upload(ctx context.Context, key string, reader io.Reader, contentType string) error {
	objectKey, err := c.getTenantKey(ctx, key)
	if err != nil {
		return storage.NewStorageError("upload", key, err)
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		return storage.NewStorageError("upload", key, err)
	}

	_, err = c.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(c.bucket),
		Key:         aws.String(objectKey),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return storage.NewStorageError("upload", key, err)
	}
	return nil
}

func (c *Client) UploadBytes(ctx context.Context, key string, data []byte, contentType string) error {
	reader := bytes.NewReader(data)
	return c.Upload(ctx, key, reader, contentType)
}

func (c *Client) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	objectKey, err := c.getTenantKey(ctx, key)
	if err != nil {
		return nil, storage.NewStorageError("download", key, err)
	}

	result, err := c.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		return nil, storage.NewStorageError("download", key, err)
	}

	return result.Body, nil
}

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

func (c *Client) Delete(ctx context.Context, key string) error {
	objectKey, err := c.getTenantKey(ctx, key)
	if err != nil {
		return storage.NewStorageError("delete", key, err)
	}

	_, err = c.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		return storage.NewStorageError("delete", key, err)
	}
	return nil
}

func (c *Client) List(ctx context.Context, prefix string) ([]storage.ObjectInfo, error) {
	tenantID, err := c.getTenantIDFromContext(ctx)
	if err != nil {
		return nil, storage.NewStorageError("list", prefix, err)
	}

	tenantPrefix := path.Join(tenantID, prefix)

	result, err := c.client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(c.bucket),
		Prefix: aws.String(tenantPrefix),
	})
	if err != nil {
		return nil, storage.NewStorageError("list", prefix, err)
	}

	objects := make([]storage.ObjectInfo, 0, len(result.Contents))
	for _, obj := range result.Contents {
		objects = append(objects, storage.ObjectInfo{
			Key:          *obj.Key,
			Size:         *obj.Size,
			LastModified: *obj.LastModified,
			ETag:          *obj.ETag,
		})
	}

	return objects, nil
}

func (c *Client) GeneratePresignedURL(ctx context.Context, key string, expiry time.Duration, isUpload bool) (string, error) {
	if expiry == 0 {
		expiry = time.Hour
	}

	objectKey, err := c.getTenantKey(ctx, key)
	if err != nil {
		return "", storage.NewStorageError("presign", key, err)
	}

	c.logger.DebugContext(ctx, "generating presigned URL", "key", objectKey, "is_upload", isUpload)

	presignClient := s3.NewPresignClient(c.client)

	if isUpload {
		presigned, err := presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
			Bucket: aws.String(c.bucket),
			Key:    aws.String(objectKey),
		}, s3.WithPresignExpires(expiry))
		if err != nil {
			return "", storage.NewStorageError("presign", key, err)
		}
		return presigned.URL, nil
	}

	presigned, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(objectKey),
	}, s3.WithPresignExpires(expiry))
	if err != nil {
		return "", storage.NewStorageError("presign", key, err)
	}

	return presigned.URL, nil
}