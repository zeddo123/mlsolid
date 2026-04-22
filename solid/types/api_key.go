package types

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// NewAPIKey generates a new API key.
func NewAPIKey(length int) (string, error) {
	bytes := make([]byte, length)

	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random key: %w", err)
	}

	return hex.EncodeToString(bytes), nil
}
