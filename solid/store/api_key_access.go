package store

import (
	"context"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/zeddo123/mlsolid/solid/types"
)

func (r *RedisStore) CreateAPIKey(ctx context.Context, perm types.Permissions) (string, error) {
	key := uuid.NewString()
	redisKey := r.makeAPIKeyKey(key)

	fn := func(tx *redis.Tx) error {
		_, err := tx.Pipelined(ctx, func(p redis.Pipeliner) error {
			p.HSet(ctx, redisKey, perm.Mapping())
			p.Expire(ctx, redisKey, perm.Expiry)

			return nil
		})

		return err
	}

	err := r.runTx(ctx, fn, TransactionMaxTries, redisKey)
	if err != nil {
		return "", err
	}

	return key, nil
}

func (r *RedisStore) APIKeyPermissions(ctx context.Context, key string) (types.Permissions, error) {
	redisKey := r.makeAPIKeyKey(key)

	m, err := r.Client.HGetAll(ctx, redisKey).Result()
	if err != nil {
		return types.Permissions{}, err
	}

	return types.NewPermissions(m)
}
