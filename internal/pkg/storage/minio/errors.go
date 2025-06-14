package minio

import "fmt"

// StorageError represents an error that occurred during storage operations
type StorageError struct {
	Op  string // The operation that failed
	Key string // The object key that was being operated on
	Err error  // The underlying error
}

func (e *StorageError) Error() string {
	if e.Key != "" {
		return fmt.Sprintf("%s %s: %v", e.Op, e.Key, e.Err)
	}
	return fmt.Sprintf("%s: %v", e.Op, e.Err)
}

// Unwrap returns the underlying error
func (e *StorageError) Unwrap() error {
	return e.Err
}

// NewStorageError creates a new StorageError
func NewStorageError(op string, key string, err error) error {
	return &StorageError{
		Op:  op,
		Key: key,
		Err: err,
	}
}
