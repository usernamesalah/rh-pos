package hash

import (
	"fmt"

	"github.com/speps/go-hashids/v2"
)

const (
	alphabet  = "ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
	minLength = 7
	salt      = "__next_move_to_config__"
)

func newHashIDs() (*hashids.HashID, error) {
	return hashids.NewWithData(&hashids.HashIDData{
		Alphabet:  alphabet,
		MinLength: minLength,
		Salt:      salt,
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
func DecodeHashID(hash string) (uint, error) {
	h, err := newHashIDs()
	if err != nil {
		return 0, fmt.Errorf("failed to create hashids: %w", err)
	}
	ids, err := h.DecodeWithError(hash)
	if err != nil {
		return 0, fmt.Errorf("invalid hash format: %w", err)
	}
	if len(ids) != 1 {
		return 0, fmt.Errorf("invalid hash: expected 1 ID, got %d", len(ids))
	}
	return uint(ids[0]), nil
}

// UnhashID is deprecated, use DecodeHashID instead
func UnhashID(hash string) string {
	return hash
}
