package store

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/zeddo123/mlsolid/solid/types"
)

// CreateBenchmark creates a new benchmark.
func (r *RedisStore) CreateBenchmark(ctx context.Context, b types.Bench) (bool, error) {
	benchKey := r.makeBenchmarkKey(b.Name)

	p := r.Client.Pipeline()

	// add new benchmark to index
	p.SAdd(ctx, BenchmarksKey, b.Name)
	p.HSet(ctx, benchKey, map[string]any{
		"Name":        b.Name,
		"Paused":      b.Paused,
		"EagerStart":  b.EagerStart,
		"AutoTag":     b.AutoTag,
		"Tag":         b.Tag,
		"DatasetName": b.DatasetName,
		"DatasetURL":  b.DatasetURL,
		"FromS3":      b.FromS3,
		"Timestamp":   b.Timestamp,
	})

	_, err := p.Exec(ctx)
	if err != nil {
		return false, fmt.Errorf("could not set benchmark: %w", err)
	}

	err = r.AddBenchmarkRegistries(ctx, b.Name, b.Registries)
	if err != nil {
		return false, err
	}

	err = r.AddBenchmarkMetrics(ctx, b.Name, b.Metrics)
	if err != nil {
		return false, err
	}

	return true, nil
}

// ToggleBenchmark toggles a benchmark state between paused and unpaused.
func (r *RedisStore) ToggleBenchmark(ctx context.Context, benchName string, paused bool) error {
	_, err := r.Client.HSet(ctx, r.makeBenchmarkKey(benchName), "Paused", paused).Result()
	if err != nil {
		return fmt.Errorf("could not set benchmark Paused status: %w", err)
	}

	return nil
}

// AddBenchmarkRegistries adds registries to a benchmark.
func (r *RedisStore) AddBenchmarkRegistries(ctx context.Context, benchName string, registries []string) error {
	benchRegistriesKey := r.makeBenchmarkRegistriesKey(benchName)

	p := r.Client.Pipeline()

	rs := make([]any, len(registries))
	for i, reg := range registries {
		rs[i] = reg

		// Setting the index (registry -> benchmarks)
		p.SAdd(ctx, r.makeRegistryBenchmarksKey(reg), benchName)
	}

	p.SAdd(ctx, benchRegistriesKey, rs...)

	_, err := p.Exec(ctx)
	if err != nil {
		return fmt.Errorf("could not set benchmark registries: %w", err)
	}

	return nil
}

// RemBenchmarkRegistries removes model registries from a benchmark.
func (r *RedisStore) RemBenchmarkRegistries(ctx context.Context, benchName string, registries []string) error {
	p := r.Client.Pipeline()

	rs := make([]any, len(registries))
	for i, reg := range registries {
		rs[i] = reg
		p.SRem(ctx, r.makeRegistryBenchmarksKey(reg), benchName)
	}

	p.SRem(ctx, r.makeBenchmarkRegistriesKey(benchName), rs...)

	_, err := p.Exec(ctx)
	if err != nil {
		return fmt.Errorf("could not remove benchmark registries: %w", err)
	}

	return nil
}

// AddBenchmarkMetrics adds metrics to a benchmark.
func (r *RedisStore) AddBenchmarkMetrics(ctx context.Context, benchName string, metrics []string) error {
	key := r.makeBenchmarkMetricsKey(benchName)

	ms := make([]any, len(metrics))
	for i, m := range metrics {
		ms[i] = m
	}

	_, err := r.Client.SAdd(ctx, key, ms...).Result()
	if err != nil {
		return fmt.Errorf("could not set benchmark metrics: %w", err)
	}

	return nil
}

// RemBenchmarkMetrics removes metrics from a benchmark.
func (r *RedisStore) RemBenchmarkMetrics(ctx context.Context, benchName string, metrics []string) error {
	ms := make([]any, len(metrics))
	for i, m := range metrics {
		ms[i] = m
	}

	_, err := r.Client.SRem(ctx, r.makeBenchmarkMetricsKey(benchName), ms...).Result()
	if err != nil {
		return fmt.Errorf("could not remove benchmark metrics: %w", err)
	}

	return nil
}

// Benchmark pulls a benchmark by its name from the redis store.
func (r *RedisStore) Benchmark(ctx context.Context, benchName string) (*types.Bench, error) {
	key := r.makeBenchmarkKey(benchName)

	c, err := r.Client.Exists(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("could not connect to db: %w", err)
	}

	if c != 1 {
		return nil, fmt.Errorf("could not find benchmark %s", benchName)
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

	registries, err := r.BenchmarkRegistries(ctx, benchName)
	if err != nil {
		return nil, err
	}

	metrics, err := r.BenchmarkMetrics(ctx, benchName)
	if err != nil {
		return nil, err
	}

	return &types.Bench{
		Name:        mapping["Name"],
		Paused:      paused,
		EagerStart:  eager,
		AutoTag:     autotag,
		Tag:         mapping["Tag"],
		DatasetName: mapping["DatasetName"],
		DatasetURL:  mapping["DatasetURL"],
		FromS3:      froms3,
		Timestamp:   timestamp,
		Registries:  registries,
		Metrics:     metrics,
	}, nil
}

// BenchmarkMetrics pulls metrics linked to a benchmark.
func (r *RedisStore) BenchmarkMetrics(ctx context.Context, benchName string) ([]string, error) {
	key := r.makeBenchmarkMetricsKey(benchName)

	registries, err := r.Client.SMembers(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("could not pull benchmark metrics: %w", err)
	}

	return registries, nil
}

// BenchmarkRegistries pulls model registries linked to a benchmark.
func (r *RedisStore) BenchmarkRegistries(ctx context.Context, benchName string) ([]string, error) {
	key := r.makeBenchmarkRegistriesKey(benchName)

	registries, err := r.Client.SMembers(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("could not pull benchmark metrics: %w", err)
	}

	return registries, nil
}

// RecordRuns records new benchmark runs to the store.
func (r *RedisStore) RecordRuns(ctx context.Context, benchName string, runs []types.BenchRun) error {
	indexKey := r.makeBenchmarkRunsKey(benchName)
	p := r.Client.Pipeline()

	for _, run := range runs {
		runKey := r.makeBenchmarkRunKey(benchName, run.Registry, run.Version)

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
func (r *RedisStore) BenchmarkRuns(ctx context.Context, benchName string) ([]*types.BenchRun, error) {
	runKeys, err := r.Client.SMembers(ctx, r.makeBenchmarkRunsKey(benchName)).Result()
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

	metrics := make(map[string]float64, len(m))

	for k, v := range m {
		metric, err := strconv.ParseFloat(v, 64)
		if err != nil {
			// TODO: log error
			continue
		}

		metrics[k] = metric
	}

	return &types.BenchRun{
		Timestamp: timestamp,
		Registry:  reg,
		Version:   v,
		Metrics:   metrics,
	}, nil
}
