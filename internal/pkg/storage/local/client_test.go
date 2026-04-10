package local_test

import (
	"bytes"
	"context"
	"io"
	"os"
	"testing"
	"time"

	"github.com/usernamesalah/rh-pos/internal/pkg/ctxkey"
	"github.com/usernamesalah/rh-pos/internal/pkg/hash"
	"github.com/usernamesalah/rh-pos/internal/pkg/storage"
	"github.com/usernamesalah/rh-pos/internal/pkg/storage/local"
)

func TestMain(m *testing.M) {
	hash.Init("test-salt-for-local-storage-tests")
	os.Exit(m.Run())
}

// ctxWithTenant returns a context with the given tenant ID set.
func ctxWithTenant(tenantID uint) context.Context {
	return ctxkey.WithTenantID(context.Background(), tenantID)
}

func TestUploadAndDownload(t *testing.T) {
	dir := t.TempDir()
	client, err := local.NewClient(dir, nil)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	ctx := ctxWithTenant(1)
	data := []byte("hello world")
	key := "products/test.jpg"

	if err := client.UploadBytes(ctx, key, data, "image/jpeg"); err != nil {
		t.Fatalf("UploadBytes: %v", err)
	}

	got, err := client.DownloadBytes(ctx, key)
	if err != nil {
		t.Fatalf("DownloadBytes: %v", err)
	}
	if !bytes.Equal(got, data) {
		t.Errorf("got %q, want %q", got, data)
	}
}

func TestDownload_ReturnsReadCloser(t *testing.T) {
	dir := t.TempDir()
	client, _ := local.NewClient(dir, nil)
	ctx := ctxWithTenant(1)
	data := []byte("stream test")

	_ = client.UploadBytes(ctx, "products/stream.jpg", data, "image/jpeg")

	rc, err := client.Download(ctx, "products/stream.jpg")
	if err != nil {
		t.Fatalf("Download: %v", err)
	}
	defer rc.Close()

	got, _ := io.ReadAll(rc)
	if !bytes.Equal(got, data) {
		t.Errorf("got %q, want %q", got, data)
	}
}

func TestDelete(t *testing.T) {
	dir := t.TempDir()
	client, _ := local.NewClient(dir, nil)
	ctx := ctxWithTenant(1)
	key := "products/delete_me.jpg"

	_ = client.UploadBytes(ctx, key, []byte("x"), "image/jpeg")
	if err := client.Delete(ctx, key); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err := client.DownloadBytes(ctx, key)
	if err == nil {
		t.Error("expected error after delete, got nil")
	}
}

func TestList(t *testing.T) {
	dir := t.TempDir()
	client, _ := local.NewClient(dir, nil)
	ctx := ctxWithTenant(1)

	_ = client.UploadBytes(ctx, "products/a.jpg", []byte("a"), "image/jpeg")
	_ = client.UploadBytes(ctx, "products/b.jpg", []byte("b"), "image/jpeg")

	objects, err := client.List(ctx, "products/")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(objects) != 2 {
		t.Errorf("expected 2 objects, got %d", len(objects))
	}
}

func TestList_IsolatedByTenant(t *testing.T) {
	dir := t.TempDir()
	client, _ := local.NewClient(dir, nil)

	ctx1 := ctxWithTenant(1)
	ctx2 := ctxWithTenant(2)

	_ = client.UploadBytes(ctx1, "products/a.jpg", []byte("a"), "image/jpeg")
	_ = client.UploadBytes(ctx2, "products/b.jpg", []byte("b"), "image/jpeg")

	objects, err := client.List(ctx1, "products/")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(objects) != 1 {
		t.Errorf("tenant isolation broken: expected 1 object for tenant 1, got %d", len(objects))
	}
}

func TestGeneratePresignedURL_NotSupported(t *testing.T) {
	dir := t.TempDir()
	client, _ := local.NewClient(dir, nil)
	ctx := ctxWithTenant(1)

	_, err := client.GeneratePresignedURL(ctx, "products/x.jpg", time.Hour, true)
	if err == nil {
		t.Error("expected ErrNotSupported, got nil")
	}
}

func TestUpload_MissingTenantID(t *testing.T) {
	dir := t.TempDir()
	client, _ := local.NewClient(dir, nil)
	ctx := context.Background() // no tenant

	err := client.UploadBytes(ctx, "products/x.jpg", []byte("x"), "image/jpeg")
	if err == nil {
		t.Error("expected error for missing tenant ID, got nil")
	}
}

func TestDownloadBytes_ExceedsMaxSize(t *testing.T) {
	// Write a file larger than 10MB directly (bypass client to set up the state).
	dir := t.TempDir()
	client, _ := local.NewClient(dir, nil)
	ctx := ctxWithTenant(1)

	big := make([]byte, 11<<20) // 11 MB
	// Write directly so we bypass the upload path (which doesn't enforce a limit).
	_ = os.MkdirAll(dir+"/tenant_path", 0755)

	// Use upload path — we want DownloadBytes to enforce the limit.
	if err := client.Upload(ctx, "products/big.jpg", bytes.NewReader(big), "image/jpeg"); err != nil {
		t.Fatalf("Upload: %v", err)
	}

	_, err := client.DownloadBytes(ctx, "products/big.jpg")
	if err == nil {
		t.Error("expected error for oversized object, got nil")
	}
}

// Ensure *Client satisfies the storage.StorageClient interface at compile time.
var _ storage.StorageClient = (*local.Client)(nil)
