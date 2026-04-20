package controllers

import (
	"context"
	"errors"
	"fmt"

	"github.com/zeddo123/mlsolid/solid/types"
)

// CreateBenchmark creates a new benchmark.
func (c *Controller) CreateBenchmark(ctx context.Context, b types.Bench) (bool, error) {
	if err := b.Validate(); err != nil {
		return false, fmt.Errorf("could not create benchmark: %w", err)
	}

	b.Sanatize()

	created, err := c.Redis.CreateBenchmark(ctx, b)
	if err != nil {
		return false, fmt.Errorf("saving benchmark failed: %w", err)
	}

	return created, nil
}

// Benchmark returns a benchmark with id `benchName` if present.
func (c *Controller) Benchmark(ctx context.Context, benchName string) (*types.Bench, error) {
	exists, err := c.Redis.BenchmarkExists(ctx, benchName)
	if err != nil {
		return nil, fmt.Errorf("could not pull benchmark: %w", err)
	}

	if !exists {
		return nil, errors.New("could not pull benchmark: benchmark does not exist")
	}

	bench, err := c.Redis.Benchmark(ctx, benchName)
	if err != nil {
		return nil, fmt.Errorf("could not pull benchmark: %w", err)
	}

	return bench, nil
}

// BenchmarkMetrics pulls metrics tracked by a benchmark.
func (c *Controller) BenchmarkMetrics(ctx context.Context, benchName string) ([]string, error) {
	exists, err := c.Redis.BenchmarkExists(ctx, benchName)
	if err != nil {
		return nil, fmt.Errorf("could not pull benchmark metrics: %w", err)
	}

	if !exists {
		return nil, errors.New("could not pull benchmark metrics: benchmark does not exist")
	}

	metrics, err := c.Redis.BenchmarkMetrics(ctx, benchName)
	if err != nil {
		return nil, fmt.Errorf("could not pull benchmark metrics: %w", err)
	}

	return metrics, nil
}

// BenchmarkRegistries pulls registries linked the benchmark.
func (c *Controller) BenchmarkRegistries(ctx context.Context, benchName string) ([]string, error) {
	exists, err := c.Redis.BenchmarkExists(ctx, benchName)
	if err != nil {
		return nil, fmt.Errorf("checking if benchmark is present failed: %w", err)
	}

	if !exists {
		return nil, errors.New("could not pull benchmark registries: benchmark does not exist")
	}

	metrics, err := c.Redis.BenchmarkRegistries(ctx, benchName)
	if err != nil {
		return nil, fmt.Errorf("could not pull benchmark registries: %w", err)
	}

	return metrics, nil
}

// BenchmarkRuns pulls all recorded benchmark runs.
func (c *Controller) BenchmarkRuns(ctx context.Context, benchName string) ([]*types.BenchRun, error) {
	exists, err := c.Redis.BenchmarkExists(ctx, benchName)
	if err != nil {
		return nil, fmt.Errorf("checking if benchmark is present failed: %w", err)
	}

	if !exists {
		return nil, errors.New("benchmark does not exist")
	}

	runs, err := c.Redis.BenchmarkRuns(ctx, benchName)
	if err != nil {
		return nil, fmt.Errorf("could not pull benchmark runs: %w", err)
	}

	return runs, nil
}

// Benchmarks pulls all known benchmark names.
func (c *Controller) Benchmarks(ctx context.Context) ([]string, error) {
	benchs, err := c.Redis.Benchmarks(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed pulling benchmarks: %w", err)
	}

	return benchs, nil
}

// RegistryBenchmarks pulls all benchmarks linked to a registry.
func (c *Controller) RegistryBenchmarks(ctx context.Context, registry string) ([]string, error) {
	benchs, err := c.Redis.RegistryBenchmarks(ctx, registry)
	if err != nil {
		return nil, fmt.Errorf("failed pulling benchmarks: %w", err)
	}

	return benchs, nil
}

// ToggleBenchmark toggle a benchmark's paused state.
func (c *Controller) ToggleBenchmark(ctx context.Context, benchName string, paused bool) error {
	exists, err := c.Redis.BenchmarkExists(ctx, benchName)
	if err != nil {
		return fmt.Errorf("could not toggle benchmark: %w", err)
	}

	if !exists {
		return errors.New("could not toggle benchmark: benchmark does not exist")
	}

	err = c.Redis.ToggleBenchmark(ctx, benchName, paused)
	if err != nil {
		return fmt.Errorf("could not toggle benchmark paused: %w", err)
	}

	return nil
}

// AddBenchmarkRegistries adds new registries to a benchmark.
func (c *Controller) AddBenchmarkRegistries(ctx context.Context, benchName string, registries []string) error {
	exists, err := c.Redis.BenchmarkExists(ctx, benchName)
	if err != nil {
		return fmt.Errorf("could not add registries: %w", err)
	}

	if !exists {
		return errors.New("could not add registries: benchmark does not exist")
	}

	err = c.Redis.AddBenchmarkRegistries(ctx, benchName, registries)
	if err != nil {
		return fmt.Errorf("could not add registries: %w", err)
	}

	return nil
}

// RemBenchmarkRegistries removes registries from a benchmark.
func (c *Controller) RemBenchmarkRegistries(ctx context.Context, benchName string, registries []string) error {
	exists, err := c.Redis.BenchmarkExists(ctx, benchName)
	if err != nil {
		return fmt.Errorf("could not remove registries: %w", err)
	}

	if !exists {
		return errors.New("could not remove registries: benchmark does not exist")
	}

	err = c.Redis.RemBenchmarkRegistries(ctx, benchName, registries)
	if err != nil {
		return fmt.Errorf("could not remove registries: %w", err)
	}

	return nil
}

// AddBenchmarkMetrics adds metrics to a benchmark.
func (c *Controller) AddBenchmarkMetrics(ctx context.Context, benchName string, metrics []string) error {
	exists, err := c.Redis.BenchmarkExists(ctx, benchName)
	if err != nil {
		return fmt.Errorf("could not add metrics: %w", err)
	}

	if !exists {
		return errors.New("could not add metrics: benchmark does not exist")
	}

	err = c.Redis.AddBenchmarkMetrics(ctx, benchName, metrics)
	if err != nil {
		return fmt.Errorf("could not add metrics: %w", err)
	}

	return nil
}

// RemBenchmarkMetrics removes metrics from a benchmark.
func (c *Controller) RemBenchmarkMetrics(ctx context.Context, benchName string, metrics []string) error {
	exists, err := c.Redis.BenchmarkExists(ctx, benchName)
	if err != nil {
		return fmt.Errorf("could not remove metrics: %w", err)
	}

	if !exists {
		return errors.New("could not remove metrics: benchmark does not exist")
	}

	err = c.Redis.RemBenchmarkMetrics(ctx, benchName, metrics)
	if err != nil {
		return fmt.Errorf("could not remove registries: %w", err)
	}

	return nil
}
