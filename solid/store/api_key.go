package store

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/zeddo123/mlsolid/solid/types"
)

const apiKeyLength = 32

const apiKeyDuration = 2048 * time.Hour

// KeyLabels returns all api key labels still active.
func (r *RedisStore) KeyLabels(ctx context.Context) (map[string]time.Time, error) {
	keys, err := r.scanKeys(ctx, r.makeAPIKeyMatchingPattern())
	if err != nil {
		return nil, fmt.Errorf("could not get all keys: %w", err)
	}

	p := r.Client.Pipeline()
	cmds := make([]*redis.StringCmd, len(keys))
	ttls := make([]*redis.DurationCmd, len(keys))

	for i, k := range keys {
		cmds[i] = p.Get(ctx, k)
		ttls[i] = p.TTL(ctx, k)
	}

	_, err = p.Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve keys: %w", err)
	}

	results := make(map[string]time.Time, len(keys))

	for i, cmd := range cmds {
		label, err := cmd.Result()
		if err != nil {
			r.Logger.Error().Err(err).Msg("could not read label of api key")

			continue
		}

		ttl, err := ttls[i].Result()
		if err != nil {
			r.Logger.Error().
				Err(err).
				Str("label", label).
				Msg("could not read timestamp of api key")

			results[label] = time.Time{}
		}

		results[label] = time.Now().Add(ttl)
	}

	return results, nil
}

// IsValidAPIKey returns true if the API key is valid.
func (r *RedisStore) IsValidAPIKey(ctx context.Context, key string) (bool, error) {
	c, err := r.Client.Exists(ctx, r.makeAPIKey(hashKey(key))).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check if api is present: %w", err)
	}

	return c == 1, nil
}

// GenerateKey generates a new key and saves it to the store.
func (r *RedisStore) GenerateKey(ctx context.Context, label string) (string, error) {
	key, err := types.NewAPIKey(apiKeyLength)
	if err != nil {
		return "", fmt.Errorf("failed to generate: %w", err)
	}

	_, err = r.Client.Set(ctx, r.makeAPIKey(hashKey(key)), label, apiKeyDuration).Result()
	if err != nil {
		return "", fmt.Errorf("could not set api key: %w", err)
	}

	return key, nil
}

func hashKey(key string) string {
	h := sha256.New()

	h.Write([]byte(key))

	return string(h.Sum(nil))
}
