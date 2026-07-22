// Package store implements the persistence layer with redis.
package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/zeddo123/mlsolid/solid/types"
)

const (
	// APIKeyPattern pattern that represents a api key.
	APIKeyPattern = "api-key:%s" //nolint: gosec

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

	// ModelRegistriesKey is a Sorted Set index of all known registry names, scored by
	// insertion order (see RegistriesCounterKey). Populated on registry creation
	// regardless of whether it has any model entries yet, since a registry's
	// "registry:<name>" list key (ModelRegistryKeyPattern) only comes into existence
	// once the first model entry is pushed to it.
	ModelRegistriesKey = "index:registries"

	// RegistriesCounterKey is an auto-incrementing counter used to assign strictly
	// increasing scores to new entries in ModelRegistriesKey.
	RegistriesCounterKey = "counter:registries"

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

	// BenchmarksKey is a Sorted Set index of all known benchmark ids, scored by
	// insertion order (see BenchmarksCounterKey).
	BenchmarksKey = "index:benchs"

	// BenchmarksCounterKey is an auto-incrementing counter used to assign strictly
	// increasing scores to new entries in BenchmarksKey.
	BenchmarksCounterKey = "counter:benchs"

	// ExpsIndexKey is a Sorted Set index of all known experiment ids, scored by
	// insertion order (see ExpsCounterKey). Populated the first time a run is
	// recorded under a given experiment id.
	ExpsIndexKey = "index:exps"

	// ExpsCounterKey is an auto-incrementing counter used to assign strictly
	// increasing scores to new entries in ExpsIndexKey.
	ExpsCounterKey = "counter:exps"

	// BenchmarkKeyPattern represents the key for a benchmark.
	BenchmarkKeyPattern = "bench:%s"

	// BenchmarkMetricsKeyPattern represents all metrics linked to a benchmark.
	BenchmarkMetricsKeyPattern = "bench:%s:metrics"

	// BenchmarkRegistriesKeyPattern represents all registries linked to a benchmark.
	BenchmarkRegistriesKeyPattern = "bench:%s:registries"

	// BenchmarkRunsKeyPattern index used to pull all benchmark runs
	// It follows this form: index:bench:<bench-id>:runs.
	BenchmarkRunsKeyPattern = "index:bench:%s:runs"

	// BenchmarkRunKeyPattern key pattern used to save a benchmark run
	// It follows this order: bench:<bench-id>:run:<registry-name>:<version>.
	BenchmarkRunKeyPattern = "bench:%s:run:%s:%d"

	transactionMaxTries = 10
)

// RedisStore implements the store for mlsolid.
type RedisStore struct {
	Client redis.Client
	Logger zerolog.Logger
}

func (r *RedisStore) makeAPIKey(key string) string {
	return fmt.Sprintf(APIKeyPattern, key)
}

func (r *RedisStore) makeAPIKeyMatchingPattern() string {
	return fmt.Sprintf(APIKeyPattern, "*")
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

// zIndexAdd adds member to the Sorted Set index at indexKey with a fresh, strictly
// increasing score drawn from counterKey. It is a no-op if member is already present
// (ZADD NX leaves its existing score untouched), so it is safe to call unconditionally
// regardless of whether the entity is already indexed.
func (r *RedisStore) zIndexAdd(ctx context.Context, indexKey, counterKey, member string) error {
	score, err := r.Client.Incr(ctx, counterKey).Result()
	if err != nil {
		return fmt.Errorf("could not allocate index score: %w", err)
	}

	if err := r.Client.ZAddNX(ctx, indexKey, redis.Z{Score: float64(score), Member: member}).Err(); err != nil {
		return fmt.Errorf("could not add to index: %w", err)
	}

	return nil
}

// zIndexAddAll adds every member to the Sorted Set index at indexKey via zIndexAdd.
// Safe to call with members that are already indexed.
func (r *RedisStore) zIndexAddAll(ctx context.Context, indexKey, counterKey string, members []string) error {
	for _, member := range members {
		if err := r.zIndexAdd(ctx, indexKey, counterKey, member); err != nil {
			return err
		}
	}

	return nil
}

// zIndexAll returns every member of the Sorted Set index at indexKey, ordered by score.
func (r *RedisStore) zIndexAll(ctx context.Context, indexKey string) ([]string, error) {
	members, err := r.Client.ZRange(ctx, indexKey, 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("could not read index: %w", err)
	}

	return members, nil
}

// zIndexPage returns a single page of members from the Sorted Set index at indexKey,
// ordered by score, starting strictly after cursor. A returned cursor of 0 means there
// are no more pages. Unlike SCAN's COUNT (a hint Redis may ignore for small
// collections), this always returns at most limit members.
func (r *RedisStore) zIndexPage(ctx context.Context, indexKey string,
	cursor uint64, limit int64,
) ([]string, uint64, error) {
	// Uses the legacy ZRANGEBYSCORE form (via ZRangeByScoreWithScores) rather than
	// ZRANGE ... BYSCORE, since the latter's LIMIT support requires Redis 6.2+.
	res, err := r.Client.ZRangeByScoreWithScores(ctx, indexKey, &redis.ZRangeBy{
		Min:    fmt.Sprintf("(%d", cursor),
		Max:    "+inf",
		Offset: 0,
		Count:  limit + 1,
	}).Result()
	if err != nil {
		return nil, 0, fmt.Errorf("could not scan index: %w", err)
	}

	hasMore := int64(len(res)) > limit
	if hasMore {
		res = res[:limit]
	}

	members := make([]string, len(res))

	var next uint64

	for i, z := range res {
		members[i], _ = z.Member.(string)
	}

	if hasMore {
		next = uint64(res[len(res)-1].Score)
	}

	return members, next, nil
}

// migrateSetIndexToSortedSet converts a legacy Set-based index at key into a Sorted
// Set scored by counterKey, preserving its members. It is a no-op if key does not
// exist or is not a Set (e.g. it was already migrated). Intended to run once at
// startup, before the server accepts traffic, since it is not atomic with respect to
// concurrent writers.
func (r *RedisStore) migrateSetIndexToSortedSet(ctx context.Context, key, counterKey string) error {
	typ, err := r.Client.Type(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("could not check index type: %w", err)
	}

	if typ != "set" {
		return nil
	}

	members, err := r.Client.SMembers(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("could not read legacy set index: %w", err)
	}

	if err := r.Client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("could not clear legacy set index: %w", err)
	}

	return r.zIndexAddAll(ctx, key, counterKey, members)
}
