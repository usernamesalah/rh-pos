package hash

import (
	"fmt"

	"github.com/speps/go-hashids/v2"
)

const (
	alphabet  = "ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
	minLength = 7
)

var hashSalt string

// Init sets the salt used for all hash operations.
// Must be called once during application startup before any hash/decode calls.
func Init(salt string) {
	hashSalt = salt
}

func newHashIDs() (*hashids.HashID, error) {
	if hashSalt == "" {
		return nil, fmt.Errorf("hash salt not initialized: call hash.Init() at startup")
	}
	return hashids.NewWithData(&hashids.HashIDData{
		Alphabet:  alphabet,
		MinLength: minLength,
		Salt:      hashSalt,
	})
}

// HashID encodes an ID into a short, reversible hash
func HashID(id uint) string {
	h, err := newHashIDs()
	if err != nil {
		return ""
	}
	idHash, err := h.Encode([]int{int(id)})
	if err != nil {
		return ""
	}
	return idHash
}

// DecodeHashID decodes a hashed ID back to its original uint value
func DecodeHashID(hashed string) (uint, error) {
	h, err := newHashIDs()
	if err != nil {
		return 0, fmt.Errorf("failed to create hashids: %w", err)
	}
	ids, err := h.DecodeWithError(hashed)
	if err != nil {
		return 0, fmt.Errorf("invalid hash format: %w", err)
	}
	if len(ids) != 1 {
		return 0, fmt.Errorf("invalid hash: expected 1 ID, got %d", len(ids))
	}
	return uint(ids[0]), nil
}
