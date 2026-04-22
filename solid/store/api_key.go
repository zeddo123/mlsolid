package store

import (
	"context"
	"fmt"
	"time"

	"github.com/zeddo123/mlsolid/solid/types"
)

const apiKeyLength = 32

const apiKeyDuration = 2048 * time.Hour

// IsValidAPIKey returns true if the API key is valid.
func (r *RedisStore) IsValidAPIKey(ctx context.Context, key string) (bool, error) {
	c, err := r.Client.Exists(ctx, r.makeAPIKey(key)).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check if api is present: %w", err)
	}

	return c == 1, nil
}

// GenerateKey generates a new key and saves it to the store.
func (r *RedisStore) GenerateKey(ctx context.Context) (string, error) {
	key, err := types.NewAPIKey(apiKeyLength)
	if err != nil {
		return "", fmt.Errorf("failed to generate: %w", err)
	}

	_, err = r.Client.Set(ctx, r.makeAPIKey(key), "1", apiKeyDuration).Result()
	if err != nil {
		return "", fmt.Errorf("could not set api key: %w", err)
	}

	return key, nil
}
