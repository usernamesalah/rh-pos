package storage

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/usernamesalah/rh-pos/internal/pkg/hash"
)

// GenerateImageKey generates a unique image key for a product
// Format: products/{hash_id}_{timestamp}.{ext}
func GenerateImageKey(productID uint, ext string) string {
	// Generate hash ID from product ID
	hashID := hash.HashID(productID)

	// Generate timestamp
	timestamp := time.Now().Unix()

	// Clean extension (remove leading dot if present)
	if len(ext) > 0 && ext[0] == '.' {
		ext = ext[1:]
	}

	// Generate key
	key := fmt.Sprintf("products/%s_%d.%s", hashID, timestamp, ext)

	return key
}

// GetImageExtension extracts the file extension from a filename
func GetImageExtension(filename string) string {
	return filepath.Ext(filename)
}
