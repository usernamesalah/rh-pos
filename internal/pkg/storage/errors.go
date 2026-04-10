package storage

import (
	"errors"
	"fmt"
)

// ErrNotSupported is returned when an operation is not supported by the storage backend.
var ErrNotSupported = errors.New("operation not supported by this storage backend")

// StorageError represents an error that occurred during storage operations.
type StorageError struct {
	Op  string
	Key string
	Err error
}

func (e *StorageError) Error() string {
	if e.Key != "" {
		return fmt.Sprintf("%s %s: %v", e.Op, e.Key, e.Err)
	}
	return fmt.Sprintf("%s: %v", e.Op, e.Err)
}

func (e *StorageError) Unwrap() error { return e.Err }

// NewStorageError creates a new StorageError.
func NewStorageError(op, key string, err error) error {
	return &StorageError{Op: op, Key: key, Err: err}
}
