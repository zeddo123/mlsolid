package store

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/zeddo123/mlsolid/solid/types"
)

// CreateBenchmark creates a new benchmark.
func (r *RedisStore) CreateBenchmark(ctx context.Context, b types.Bench) (bool, error) {
	benchKey := r.makeBenchmarkKey(b.ID)

	p := r.Client.Pipeline()

	// add new benchmark to index
	p.SAdd(ctx, BenchmarksKey, b.ID)
	p.HSet(ctx, benchKey, map[string]any{
		"Name":           b.Name,
		"Paused":         b.Paused,
		"EagerStart":     b.EagerStart,
		"AutoTag":        b.AutoTag,
		"Tag":            b.Tag,
		"DecisionMetric": b.DecisionMetric,
		"DatasetName":    b.DatasetName,
		"DatasetURL":     b.DatasetURL,
		"FromS3":         b.FromS3,
		"Timestamp":      b.Timestamp,
	})

	_, err := p.Exec(ctx)
	if err != nil {
		return false, fmt.Errorf("could not set benchmark: %w", err)
	}

	err = r.AddBenchmarkRegistries(ctx, b.ID, b.Registries)
	if err != nil {
		return false, err
	}

	err = r.AddBenchmarkMetrics(ctx, b.ID, b.Metrics)
	if err != nil {
		return false, err
	}

	return true, nil
}

// ToggleBenchmark toggles a benchmark state between paused and unpaused.
func (r *RedisStore) ToggleBenchmark(ctx context.Context, benchID string, paused bool) error {
	_, err := r.Client.HSet(ctx, r.makeBenchmarkKey(benchID), "Paused", paused).Result()
	if err != nil {
		return fmt.Errorf("could not set benchmark Paused status: %w", err)
	}

	return nil
}

// AddBenchmarkRegistries adds registries to a benchmark.
func (r *RedisStore) AddBenchmarkRegistries(ctx context.Context, benchID string, registries []string) error {
	benchRegistriesKey := r.makeBenchmarkRegistriesKey(benchID)

	p := r.Client.Pipeline()

	rs := make([]any, len(registries))
	for i, reg := range registries {
		rs[i] = reg

		// Setting the index (registry -> benchmarks)
		p.SAdd(ctx, r.makeRegistryBenchmarksKey(reg), benchID)
	}

	p.SAdd(ctx, benchRegistriesKey, rs...)

	_, err := p.Exec(ctx)
	if err != nil {
		return fmt.Errorf("could not set benchmark registries: %w", err)
	}

	return nil
}

// RemBenchmarkRegistries removes model registries from a benchmark.
func (r *RedisStore) RemBenchmarkRegistries(ctx context.Context, benchID string, registries []string) error {
	p := r.Client.Pipeline()

	rs := make([]any, len(registries))
	for i, reg := range registries {
		rs[i] = reg
		p.SRem(ctx, r.makeRegistryBenchmarksKey(reg), benchID)
	}

	p.SRem(ctx, r.makeBenchmarkRegistriesKey(benchID), rs...)

	_, err := p.Exec(ctx)
	if err != nil {
		return fmt.Errorf("could not remove benchmark registries: %w", err)
	}

	return nil
}

// AddBenchmarkMetrics adds metrics to a benchmark.
func (r *RedisStore) AddBenchmarkMetrics(ctx context.Context, benchID string, metrics []types.BenchMetric) error {
	key := r.makeBenchmarkMetricsKey(benchID)

	hash := make(map[string]any, len(metrics))
	for _, m := range metrics {
		info, err := json.Marshal(m)
		if err != nil {
			log.Println("could not marshal metric to json: %s", err)

			continue
		}

		hash[m.Name] = info
	}

	_, err := r.Client.HSet(ctx, key, hash).Result()
	if err != nil {
		return fmt.Errorf("could not set benchmark metrics hash: %w", err)
	}

	return nil
}

// RemBenchmarkMetrics removes metrics from a benchmark.
func (r *RedisStore) RemBenchmarkMetrics(ctx context.Context, benchID string, metrics []string) error {
	_, err := r.Client.HDel(ctx, r.makeBenchmarkMetricsKey(benchID), metrics...).Result()
	if err != nil {
		return fmt.Errorf("could not remove benchmark metrics: %w", err)
	}

	return nil
}

// UpdateBenchmark updates the settings fields of a benchmark.
func (r *RedisStore) UpdateBenchmark(ctx context.Context, benchID string, update types.UpdateBench) error {
	keyVals := make(map[string]any, 5) //nolint: mnd

	if update.AutoTag != nil {
		keyVals["AutoTag"] = update.AutoTag
	}

	if update.DecisionMetric != "" {
		keyVals["DecisionMetric"] = update.DecisionMetric
	}

	if update.Tag != "" {
		keyVals["Tag"] = update.Tag
	}

	if update.Name != "" {
		keyVals["Name"] = update.Name
	}

	_, err := r.Client.HSet(ctx, r.makeBenchmarkKey(benchID), keyVals).Result()
	if err != nil {
		return fmt.Errorf("could not update benchmark settings: %w", err)
	}

	return nil
}

// Benchmark pulls a benchmark by its name from the redis store.
func (r *RedisStore) Benchmark(ctx context.Context, benchID string) (*types.Bench, error) {
	key := r.makeBenchmarkKey(benchID)

	c, err := r.Client.Exists(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("could not connect to db: %w", err)
	}

	if c != 1 {
		return nil, fmt.Errorf("could not find benchmark %s", benchID)
	}

	mapping, err := r.Client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("could not pull benchmark: %w", err)
	}

	paused, err := strconv.ParseBool(mapping["Paused"])
	if err != nil {
		return nil, fmt.Errorf("could not parse benchmark: %w", err)
	}

	eager, err := strconv.ParseBool(mapping["EagerStart"])
	if err != nil {
		return nil, fmt.Errorf("could not parse benchmark: %w", err)
	}

	froms3, err := strconv.ParseBool(mapping["FromS3"])
	if err != nil {
		return nil, fmt.Errorf("could not parse benchmark: %w", err)
	}

	autotag, err := strconv.ParseBool(mapping["AutoTag"])
	if err != nil {
		return nil, fmt.Errorf("could not parse benchmark: %w", err)
	}

	timestamp, err := time.Parse(time.RFC3339, mapping["FromS3"])
	if err != nil {
		return nil, fmt.Errorf("could not parse benchmark timestamp: %w", err)
	}

	registries, err := r.BenchmarkRegistries(ctx, benchID)
	if err != nil {
		return nil, err
	}

	metrics, err := r.BenchmarkMetrics(ctx, benchID)
	if err != nil {
		return nil, err
	}

	return &types.Bench{
		ID:             benchID,
		Name:           mapping["Name"],
		Paused:         paused,
		EagerStart:     eager,
		AutoTag:        autotag,
		Tag:            mapping["Tag"],
		DecisionMetric: mapping["DecisionMetric"],
		DatasetName:    mapping["DatasetName"],
		DatasetURL:     mapping["DatasetURL"],
		FromS3:         froms3,
		Timestamp:      timestamp,
		Registries:     registries,
		Metrics:        metrics,
	}, nil
}

// BenchmarkMetrics pulls metrics linked to a benchmark.
func (r *RedisStore) BenchmarkMetrics(ctx context.Context, benchID string) ([]types.BenchMetric, error) {
	key := r.makeBenchmarkMetricsKey(benchID)

	metrics, err := r.Client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("could not pull benchmark metrics: %w", err)
	}

	benchMetrics := make([]types.BenchMetric, 0, len(metrics))

	metric := types.BenchMetric{} //nolint: exhaustruct

	for _, v := range metrics {
		err := json.Unmarshal([]byte(v), &metric)
		if err != nil {
			return nil, fmt.Errorf("could not unmarshall metric: %w", err)
		}

		benchMetrics = append(benchMetrics, metric)
	}

	return benchMetrics, nil
}

// BenchmarkRegistries pulls model registries linked to a benchmark.
func (r *RedisStore) BenchmarkRegistries(ctx context.Context, benchID string) ([]string, error) {
	key := r.makeBenchmarkRegistriesKey(benchID)

	registries, err := r.Client.SMembers(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("could not pull benchmark metrics: %w", err)
	}

	return registries, nil
}

// RecordRuns records new benchmark runs to the store.
func (r *RedisStore) RecordRuns(ctx context.Context, benchID string, runs []types.BenchRun) error {
	indexKey := r.makeBenchmarkRunsKey(benchID)
	p := r.Client.Pipeline()

	for _, run := range runs {
		runKey := r.makeBenchmarkRunKey(benchID, run.Registry, run.Version)

		p.HSet(ctx, runKey, run.Metrics)
		p.HSet(ctx, runKey, map[string]any{
			"Registry":  run.Registry,
			"Version":   run.Version,
			"Timestamp": run.Timestamp,
		})
		// set index
		p.SAdd(ctx, indexKey, runKey)
	}

	_, err := p.Exec(ctx)
	if err != nil {
		return fmt.Errorf("could not record runs: %w", err)
	}

	return nil
}

// BenchmarkRuns returns all benchmark runs recorded.
func (r *RedisStore) BenchmarkRuns(ctx context.Context, benchID string) ([]*types.BenchRun, error) {
	runKeys, err := r.Client.SMembers(ctx, r.makeBenchmarkRunsKey(benchID)).Result()
	if err != nil {
		return nil, fmt.Errorf("could not get benchmark runs: %w", err)
	}

	cmds := make([]*redis.MapStringStringCmd, len(runKeys))

	p := r.Client.Pipeline()

	for i, key := range runKeys {
		cmds[i] = p.HGetAll(ctx, key)
	}

	_, err = p.Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not pull benchmark runs: %w", err)
	}

	runs := make([]*types.BenchRun, len(runKeys))

	for i, cmd := range cmds {
		m, err := cmd.Result()
		if err != nil {
			// TODO: log error
			continue
		}

		run, err := parseBenchRun(m)
		if err != nil {
			// TODO: log error
			continue
		}

		runs[i] = run
	}

	return runs, nil
}

// Benchmarks returns all known benchmarks.
func (r *RedisStore) Benchmarks(ctx context.Context) ([]string, error) {
	benchs, err := r.Client.SMembers(ctx, BenchmarksKey).Result()
	if err != nil {
		return nil, fmt.Errorf("could not pull benchmarks: %w", err)
	}

	return benchs, nil
}

// RegistryBenchmarks returns all benchmarks linked with a registry.
func (r *RedisStore) RegistryBenchmarks(ctx context.Context, registry string) ([]string, error) {
	benchs, err := r.Client.SMembers(ctx, r.makeRegistryBenchmarksKey(registry)).Result()
	if err != nil {
		return nil, fmt.Errorf("could not pull registry benchmarks: %w", err)
	}

	return benchs, nil
}

// BenchmarkExists checks if a benchmark with BenchName exists.
func (r *RedisStore) BenchmarkExists(ctx context.Context, benchID string) (bool, error) {
	key := r.makeBenchmarkKey(benchID)

	c, err := r.Client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("could not check store: %w", err)
	}

	return c == 1, nil
}

func parseBenchRun(m map[string]string) (*types.BenchRun, error) {
	reg := m["Registry"]

	v, err := strconv.ParseInt(m["Version"], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("could not parse run Version: %w", err)
	}

	timestamp, err := time.Parse(time.RFC3339, m["Timestamp"])
	if err != nil {
		return nil, fmt.Errorf("could not parse run Timestamp: %w", err)
	}

	delete(m, "Registry")
	delete(m, "Version")
	delete(m, "Timestamp")

	metrics := make(map[string]float32, len(m))

	for k, v := range m {
		metric, err := strconv.ParseFloat(v, 32)
		if err != nil {
			// TODO: log error
			continue
		}

		metrics[k] = float32(metric)
	}

	return &types.BenchRun{
		Timestamp: timestamp,
		Registry:  reg,
		Version:   v,
		Metrics:   metrics,
	}, nil
}
