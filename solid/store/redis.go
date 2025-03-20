package store

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/zedd123/mlsolid/solid/types"
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

// runTx runs a transaction function with an optimistic locks on the keys passed as argument.
func (r *RedisStore) runTx(ctx context.Context, fn func(tx *redis.Tx) error,
	maxRetries int, keys ...string, //nolint: unparam
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

	if err := iter.Err(); err != nil {
		return nil, types.NewInternalErr("could not retrieve keys")
	}

	return keys, nil
}

func (r *RedisStore) iterate(ctx context.Context, iter *redis.ScanIterator) ([]string, error) {
	vals := make([]string, 0)

	for iter.Next(ctx) {
		vals = append(vals, iter.Val())
	}

	if err := iter.Err(); err != nil {
		return nil, types.NewInternalErr("could not retrieve keys")
	}

	return vals, nil
}
