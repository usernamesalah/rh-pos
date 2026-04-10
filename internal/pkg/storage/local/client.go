package local

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/usernamesalah/rh-pos/internal/pkg/ctxkey"
	"github.com/usernamesalah/rh-pos/internal/pkg/hash"
	"github.com/usernamesalah/rh-pos/internal/pkg/storage"
)

const maxDownloadSize = 10 << 20 // 10 MB

// Client implements storage.StorageClient using the local filesystem.
// Files are stored under: <baseDir>/<hashed_tenant_id>/<key>
type Client struct {
	baseDir string
	logger  *slog.Logger
}

// NewClient creates a new local storage client.
// baseDir is the root directory where files will be stored.
func NewClient(baseDir string, logger *slog.Logger) (*Client, error) {
	if baseDir == "" {
		return nil, fmt.Errorf("baseDir is required")
	}
	if logger == nil {
		logger = slog.Default()
	}
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}
	return &Client{baseDir: baseDir, logger: logger}, nil
}

// resolvePath returns the absolute filesystem path for a given context + key.
// The path is: <baseDir>/<hashed_tenant_id>/<key>
func (c *Client) resolvePath(ctx context.Context, key string) (string, error) {
	tenantID, ok := ctxkey.TenantIDFromContext(ctx)
	if !ok {
		return "", fmt.Errorf("tenant ID not found in context")
	}
	hashedID := hash.HashID(tenantID)
	return filepath.Join(c.baseDir, hashedID, filepath.FromSlash(key)), nil
}

// Upload writes reader contents to the local filesystem.
func (c *Client) Upload(ctx context.Context, key string, reader io.Reader, contentType string) error {
	fpath, err := c.resolvePath(ctx, key)
	if err != nil {
		return storage.NewStorageError("upload", key, err)
	}
	if err := os.MkdirAll(filepath.Dir(fpath), 0755); err != nil {
		return storage.NewStorageError("upload", key, err)
	}
	f, err := os.Create(fpath)
	if err != nil {
		return storage.NewStorageError("upload", key, err)
	}
	defer f.Close()
	if _, err := io.Copy(f, reader); err != nil {
		return storage.NewStorageError("upload", key, err)
	}
	return nil
}

// UploadBytes writes a byte slice to the local filesystem.
func (c *Client) UploadBytes(ctx context.Context, key string, data []byte, contentType string) error {
	return c.Upload(ctx, key, bytes.NewReader(data), contentType)
}

// Download opens the file for reading. Caller must close the returned ReadCloser.
func (c *Client) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	fpath, err := c.resolvePath(ctx, key)
	if err != nil {
		return nil, storage.NewStorageError("download", key, err)
	}
	f, err := os.Open(fpath)
	if err != nil {
		return nil, storage.NewStorageError("download", key, err)
	}
	return f, nil
}

// DownloadBytes reads the entire file into memory. Returns an error if the file exceeds 10 MB.
func (c *Client) DownloadBytes(ctx context.Context, key string) ([]byte, error) {
	rc, err := c.Download(ctx, key)
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	limited := io.LimitReader(rc, maxDownloadSize+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	if int64(len(data)) > maxDownloadSize {
		return nil, fmt.Errorf("object exceeds maximum allowed size of %d bytes", maxDownloadSize)
	}
	return data, nil
}

// Delete removes the file from the local filesystem.
func (c *Client) Delete(ctx context.Context, key string) error {
	fpath, err := c.resolvePath(ctx, key)
	if err != nil {
		return storage.NewStorageError("delete", key, err)
	}
	if err := os.Remove(fpath); err != nil {
		return storage.NewStorageError("delete", key, err)
	}
	return nil
}

// List returns metadata for all files under the given prefix, scoped to the tenant.
func (c *Client) List(ctx context.Context, prefix string) ([]storage.ObjectInfo, error) {
	tenantID, ok := ctxkey.TenantIDFromContext(ctx)
	if !ok {
		return nil, storage.NewStorageError("list", prefix, fmt.Errorf("tenant ID not found in context"))
	}
	hashedID := hash.HashID(tenantID)
	searchDir := filepath.Join(c.baseDir, hashedID, filepath.FromSlash(prefix))

	var objects []storage.ObjectInfo
	err := filepath.Walk(searchDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil // empty prefix is fine
			}
			return err
		}
		if info.IsDir() {
			return nil
		}
		// Return the key relative to <baseDir>/<hashedID>
		rel, _ := filepath.Rel(filepath.Join(c.baseDir, hashedID), path)
		objects = append(objects, storage.ObjectInfo{
			Key:          filepath.ToSlash(rel),
			Size:         info.Size(),
			LastModified: info.ModTime(),
		})
		return nil
	})
	if err != nil {
		return nil, storage.NewStorageError("list", prefix, err)
	}
	return objects, nil
}

// GeneratePresignedURL is not supported for local storage.
// Use the regular upload/download HTTP endpoints instead.
func (c *Client) GeneratePresignedURL(_ context.Context, key string, _ time.Duration, _ bool) (string, error) {
	return "", storage.NewStorageError("presign", key, storage.ErrNotSupported)
}
