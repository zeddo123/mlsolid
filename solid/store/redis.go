package store

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/zeddo123/mlsolid/solid/types"
)

const (
	ExpKeyPattern = "exp:%s"
	// RunKeyPattern pattern of a Run's key
	// Example:
	// run:linear-regression
	RunKeyPattern = "run:%s"
	// MetricKeyPattern pattern of a metric metric:<metric_name>:<run_id>
	// Example
	// metric:mse:linear-regression
	MetricKeyPattern = "metric:%s:%s"
	// ArtifactKeyPattern pattern of a artifact's key
	// Example
	// artifact:logs:linear-regression
	ArtifactKeyPattern = "artifact:%s:%s"

	ModelRegistryInfoKeyPattern = "info:registry:%s"
	// ModelRegistryKeyPattern pattern a model registry key
	// Example
	// registry:yolov12
	ModelRegistryKeyPattern = "registry:%s"
	// ModelRegistryTagsKeyPattern pattern a model registry's tags
	// Example
	// tag:registry:yolov12 -> [prod, latest, ...]
	ModelRegistryTagsKeyPattern = "tag:registry:%s"
	// ModelRegistryTagKeyPattern pattern a model registry tag
	// Example
	// tag:registry:yolov12:prod
	ModelRegistryTagKeyPattern = "tag:registry:%s:%s"

	TransactionMaxTries = 10
)

type RedisStore struct {
	Client redis.Client
}

func (r *RedisStore) makeExpKey(id string) string {
	return fmt.Sprintf(ExpKeyPattern, id)
}

func (r *RedisStore) makeRunKey(name string) string {
	return fmt.Sprintf(RunKeyPattern, name)
}

func (r *RedisStore) makeMetricKey(name string, runID string) string {
	return fmt.Sprintf(MetricKeyPattern, name, runID)
}

func (r *RedisStore) makeArtifactKey(name string, runID string) string {
	return fmt.Sprintf(ArtifactKeyPattern, name, runID)
}

func (r *RedisStore) makeModelRegistryKey(name string) string {
	return fmt.Sprintf(ModelRegistryKeyPattern, name)
}

func (r *RedisStore) makeModelRegistryInfoKey(name string) string {
	return fmt.Sprintf(ModelRegistryInfoKeyPattern, name)
}

func (r *RedisStore) makeModelRegistryTagsKey(name string) string {
	return fmt.Sprintf(ModelRegistryTagsKeyPattern, name)
}

func (r *RedisStore) makeModelRegistryTagKey(name, tag string) string {
	return fmt.Sprintf(ModelRegistryTagKeyPattern, name, tag)
}

// runTx runs a transaction function with an optimistic locks on the keys passed as argument.
func (r *RedisStore) runTx(ctx context.Context, fn func(tx *redis.Tx) error,
	maxRetries int, keys ...string,
) error {
	for range maxRetries {
		err := r.Client.Watch(ctx, fn, keys...)
		if err == nil {
			return nil
		}

		if err == redis.TxFailedErr {
			continue
		}

		return fmt.Errorf("%w: %w", types.ErrInvalidInput, err)
	}

	return types.NewInternalErr("maxRetries exceeded")
}

func (r *RedisStore) scanKeys(ctx context.Context, pattern string) ([]string, error) {
	keys := make([]string, 0)

	iter := r.Client.Scan(ctx, 0, pattern, 0).Iterator()

	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	if err := iter.Err(); err != redis.Nil && err != nil {
		return nil, types.NewInternalErr("could not retrieve keys")
	}

	return keys, nil
}
