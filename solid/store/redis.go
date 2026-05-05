package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/zeddo123/mlsolid/solid/types"
)

const (
	// APIKeyPattern pattern that represents a api key.
	APIKeyPattern = "api-key:%s"

	// ExpInfoKeyPattern is a pattern to a key that holds
	// information on experiments.
	ExpInfoKeyPattern = "info:exp:%s"

	// ExpKeyPattern is a pattern to an experiment key
	// that holds index of all runs linked to that exp.
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

	// ModelRegistryInfoKeyPattern pattern of model registry's data info value.
	ModelRegistryInfoKeyPattern = "info:registry:%s"

	// ModelRegistryKeyPattern pattern of a model registry key
	// Example
	// registry:yolov12 maps to a list of model entries
	ModelRegistryKeyPattern = "registry:%s"

	// ModelRegistryMatchPattern match pattern for all model registries.
	ModelRegistryMatchPattern = "registry:*"

	// ModelRegistryTagsKeyPattern pattern a model registry's tags
	// Example
	// tag:registry:yolov12 -> [prod, latest, ...]
	ModelRegistryTagsKeyPattern = "tag:registry:%s"

	// ModelRegistryTagKeyPattern pattern a model registry tag
	// Example
	// tag:registry:yolov12:prod
	ModelRegistryTagKeyPattern = "tag:registry:%s:%s"

	// ModelRegistryBenchmarksIndexPattern (Set holding all linked benchmarks)
	// form: index:registry:<registry-name>:benchs
	ModelRegistryBenchmarksIndexPattern = "index:registry:%s:benchs"

	// BenchmarksKey.
	BenchmarksKey = "index:benchs"

	// BenchmarkKeyPattern.
	BenchmarkKeyPattern = "bench:%s"

	// BenchmarkMetricsKeyPattern.
	BenchmarkMetricsKeyPattern = "bench:%s:metrics"

	// BenchmarkRegistriesKeyPattern.
	BenchmarkRegistriesKeyPattern = "bench:%s:registries"

	// BenchmarkRunsKeyPattern index used to pull all benchmark runs
	// It follows this form: index:bench:<bench-id>:runs.
	BenchmarkRunsKeyPattern = "index:bench:%s:runs"

	// BenchmarkRunKeyPattern key pattern used to save a benchmark run
	// It follows this order: bench:<bench-id>:run:<registry-name>:<version>.
	BenchmarkRunKeyPattern = "bench:%s:run:%s:%d"

	transactionMaxTries = 10
)

type RedisStore struct {
	Client redis.Client
}

func (r *RedisStore) makeAPIKey(key string) string {
	return fmt.Sprintf(APIKeyPattern, key)
}

func (r *RedisStore) makeExpKey(id string) string {
	return fmt.Sprintf(ExpKeyPattern, id)
}

func (r *RedisStore) makeExperimentInfoKey(expID string) string {
	return fmt.Sprintf(ExpInfoKeyPattern, expID)
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

func (r *RedisStore) makeBenchmarkKey(id string) string {
	return fmt.Sprintf(BenchmarkKeyPattern, id)
}

func (r *RedisStore) makeBenchmarkRunsKey(id string) string {
	return fmt.Sprintf(BenchmarkRunsKeyPattern, id)
}

func (r *RedisStore) makeBenchmarkRunKey(benchID, registryName string, version int64) string {
	return fmt.Sprintf(BenchmarkRunKeyPattern, benchID, registryName, version)
}

func (r *RedisStore) makeRegistryBenchmarksKey(registry string) string {
	return fmt.Sprintf(ModelRegistryBenchmarksIndexPattern, registry)
}

func (r *RedisStore) makeBenchmarkMetricsKey(benchID string) string {
	return fmt.Sprintf(BenchmarkMetricsKeyPattern, benchID)
}

func (r *RedisStore) makeBenchmarkRegistriesKey(benchID string) string {
	return fmt.Sprintf(BenchmarkRegistriesKeyPattern, benchID)
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

		if errors.Is(err, redis.TxFailedErr) {
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

	if err := iter.Err(); !errors.Is(err, redis.Nil) && err != nil {
		return nil, types.NewInternalErr("could not retrieve keys")
	}

	return keys, nil
}
