package controllers

import (
	"context"
	"errors"
	"fmt"

	"github.com/zeddo123/mlsolid/solid/types"
)

// CreateBenchmark creates a new benchmark.
func (c *Controller) CreateBenchmark(ctx context.Context, b types.Bench) (string, bool, error) {
	b.GenerateID()
	b.Sanitize()

	if err := b.Validate(); err != nil {
		return "", false, fmt.Errorf("%w: invalid benchmark: %w", types.ErrBadRequest, err)
	}

	created, err := c.Redis.CreateBenchmark(ctx, b)
	if err != nil {
		return "", false, fmt.Errorf("saving benchmark failed: %w", err)
	}

	return b.ID, created, nil
}

// Benchmark returns a benchmark with id `benchID` if present.
func (c *Controller) Benchmark(ctx context.Context, benchID string) (*types.Bench, error) {
	exists, err := c.Redis.BenchmarkExists(ctx, benchID)
	if err != nil {
		return nil, fmt.Errorf("%w: could not pull benchmark: %w", types.ErrInternal, err)
	}

	if !exists {
		return nil, fmt.Errorf("%w: benchmark does not exist", types.ErrNotFound)
	}

	bench, err := c.Redis.Benchmark(ctx, benchID)
	if err != nil {
		return nil, fmt.Errorf("%w: could not pull benchmark: %w", types.ErrInternal, err)
	}

	return bench, nil
}

// BenchmarkMetrics pulls metrics tracked by a benchmark.
func (c *Controller) BenchmarkMetrics(ctx context.Context, benchID string) ([]types.BenchMetric, error) {
	exists, err := c.Redis.BenchmarkExists(ctx, benchID)
	if err != nil {
		return nil, fmt.Errorf("could not pull benchmark metrics: %w", err)
	}

	if !exists {
		return nil, errors.New("could not pull benchmark metrics: benchmark does not exist")
	}

	metrics, err := c.Redis.BenchmarkMetrics(ctx, benchID)
	if err != nil {
		return nil, fmt.Errorf("could not pull benchmark metrics: %w", err)
	}

	return metrics, nil
}

// BenchmarkRegistries pulls registries linked the benchmark.
func (c *Controller) BenchmarkRegistries(ctx context.Context, benchID string) ([]string, error) {
	exists, err := c.Redis.BenchmarkExists(ctx, benchID)
	if err != nil {
		return nil, fmt.Errorf("checking if benchmark is present failed: %w", err)
	}

	if !exists {
		return nil, errors.New("could not pull benchmark registries: benchmark does not exist")
	}

	metrics, err := c.Redis.BenchmarkRegistries(ctx, benchID)
	if err != nil {
		return nil, fmt.Errorf("could not pull benchmark registries: %w", err)
	}

	return metrics, nil
}

// BenchmarkRuns pulls all recorded benchmark runs.
func (c *Controller) BenchmarkRuns(ctx context.Context, benchID string) ([]*types.BenchRun, error) {
	exists, err := c.Redis.BenchmarkExists(ctx, benchID)
	if err != nil {
		return nil, fmt.Errorf("%w: checking if benchmark is present failed: %w", types.ErrInternal, err)
	}

	if !exists {
		return nil, fmt.Errorf("%w: benchmark does not exist", types.ErrNotFound)
	}

	runs, err := c.Redis.BenchmarkRuns(ctx, benchID)
	if err != nil {
		return nil, fmt.Errorf("%w: could not pull benchmark runs: %w", types.ErrInternal, err)
	}

	return runs, nil
}

// RecordRuns records new benchmark runs. All metrics reported by a run are
// stored as-is, including ones not (yet) declared on the benchmark, so a run
// already carries a value once a matching metric is added to the benchmark.
func (c *Controller) RecordRuns(ctx context.Context, benchID string, runs []types.BenchRun) error {
	exists, err := c.Redis.BenchmarkExists(ctx, benchID)
	if err != nil {
		return fmt.Errorf("%w: checking if benchmark is present failed: %w", types.ErrInternal, err)
	}

	if !exists {
		return fmt.Errorf("%w: benchmark does not exist", types.ErrNotFound)
	}

	if err := c.Redis.RecordRuns(ctx, benchID, runs); err != nil {
		return fmt.Errorf("%w: could not record benchmark runs: %w", types.ErrInternal, err)
	}

	return nil
}

// SetActiveBenchRun marks run as the one currently executing for benchID, so
// that Benchmark's ActiveBenchRun field reflects it before the run finishes
// and is persisted via RecordRuns. Setting a new active run replaces any
// previous one. Callers should follow up with RemActiveBenchRun once the
// run completes, whether it succeeded or failed.
func (c *Controller) SetActiveBenchRun(ctx context.Context, benchID string, run types.BenchRun) error {
	exists, err := c.Redis.BenchmarkExists(ctx, benchID)
	if err != nil {
		return fmt.Errorf("%w: checking if benchmark is present failed: %w", types.ErrInternal, err)
	}

	if !exists {
		return fmt.Errorf("%w: benchmark does not exist", types.ErrNotFound)
	}

	if err := c.Redis.SetActiveBenchRun(ctx, benchID, run); err != nil {
		return fmt.Errorf("%w: could not set active benchmark run: %w", types.ErrInternal, err)
	}

	return nil
}

// RemActiveBenchRun clears the run marked as currently executing for
// benchID by SetActiveBenchRun. It succeeds without error when no run is
// active, so it is safe to call unconditionally once a run finishes.
func (c *Controller) RemActiveBenchRun(ctx context.Context, benchID string) error {
	exists, err := c.Redis.BenchmarkExists(ctx, benchID)
	if err != nil {
		return fmt.Errorf("%w: checking if benchmark is present failed: %w", types.ErrInternal, err)
	}

	if !exists {
		return fmt.Errorf("%w: benchmark does not exist", types.ErrNotFound)
	}

	if err := c.Redis.RemActiveBenchRun(ctx, benchID); err != nil {
		return fmt.Errorf("%w: could not remove active benchmark run: %w", types.ErrInternal, err)
	}

	return nil
}

// Benchmarks pulls all known benchmark ids.
func (c *Controller) Benchmarks(ctx context.Context) ([]string, error) {
	benchs, err := c.Redis.Benchmarks(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed pulling benchmarks: %w", err)
	}

	return benchs, nil
}

// BenchmarkNames pulls a map of benchmark id to benchmark name for all known benchmarks.
func (c *Controller) BenchmarkNames(ctx context.Context) (map[string]string, error) {
	ids, err := c.Redis.Benchmarks(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed pulling benchmarks: %w", err)
	}

	names, err := c.Redis.BenchmarkNames(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("failed pulling benchmark names: %w", err)
	}

	return names, nil
}

// BenchmarksPage pulls a single page of benchmark id to benchmark name, starting at cursor.
// A returned cursor of 0 means there are no more pages.
func (c *Controller) BenchmarksPage(ctx context.Context,
	cursor uint64, limit int64,
) (map[string]string, uint64, error) {
	ids, next, err := c.Redis.BenchmarksPage(ctx, cursor, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("failed pulling benchmarks page: %w", err)
	}

	names, err := c.Redis.BenchmarkNames(ctx, ids)
	if err != nil {
		return nil, 0, fmt.Errorf("failed pulling benchmark names: %w", err)
	}

	return names, next, nil
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
func (c *Controller) ToggleBenchmark(ctx context.Context, benchID string, paused bool) error {
	exists, err := c.Redis.BenchmarkExists(ctx, benchID)
	if err != nil {
		return fmt.Errorf("%w: could not toggle benchmark: %w", types.ErrInternal, err)
	}

	if !exists {
		return fmt.Errorf("%w: benchmark does not exist", types.ErrNotFound)
	}

	err = c.Redis.ToggleBenchmark(ctx, benchID, paused)
	if err != nil {
		return fmt.Errorf("%w: could not toggle benchmark paused: %w", types.ErrInternal, err)
	}

	return nil
}

// UpdateBenchmark updates an existing benchmark.
func (c *Controller) UpdateBenchmark(ctx context.Context, benchID string, update types.UpdateBench) error {
	exists, err := c.Redis.BenchmarkExists(ctx, benchID)
	if err != nil {
		return fmt.Errorf("%w: checking if benchmark exists failed: %w", types.ErrInternal, err)
	}

	if !exists {
		return fmt.Errorf("%w: could not find benchmark %q", types.ErrNotFound, benchID)
	}

	err = c.Redis.UpdateBenchmark(ctx, benchID, update)
	if err != nil {
		return fmt.Errorf("%w: could not update benchmark: %w", types.ErrInternal, err)
	}

	return nil
}

// AddBenchmarkRegistries adds new registries to a benchmark.
func (c *Controller) AddBenchmarkRegistries(ctx context.Context, benchID string, registries []string) error {
	exists, err := c.Redis.BenchmarkExists(ctx, benchID)
	if err != nil {
		return fmt.Errorf("could not add registries: %w", err)
	}

	if !exists {
		return errors.New("could not add registries: benchmark does not exist")
	}

	err = c.Redis.AddBenchmarkRegistries(ctx, benchID, registries)
	if err != nil {
		return fmt.Errorf("could not add registries: %w", err)
	}

	return nil
}

// RemBenchmarkRegistries removes registries from a benchmark.
func (c *Controller) RemBenchmarkRegistries(ctx context.Context, benchID string, registries []string) error {
	exists, err := c.Redis.BenchmarkExists(ctx, benchID)
	if err != nil {
		return fmt.Errorf("could not remove registries: %w", err)
	}

	if !exists {
		return errors.New("could not remove registries: benchmark does not exist")
	}

	err = c.Redis.RemBenchmarkRegistries(ctx, benchID, registries)
	if err != nil {
		return fmt.Errorf("could not remove registries: %w", err)
	}

	return nil
}

// AddBenchmarkMetrics adds metrics to a benchmark.
func (c *Controller) AddBenchmarkMetrics(ctx context.Context, benchID string, metrics []types.BenchMetric) error {
	exists, err := c.Redis.BenchmarkExists(ctx, benchID)
	if err != nil {
		return fmt.Errorf("could not add metrics: %w", err)
	}

	if !exists {
		return errors.New("could not add metrics: benchmark does not exist")
	}

	err = c.Redis.AddBenchmarkMetrics(ctx, benchID, metrics)
	if err != nil {
		return fmt.Errorf("could not add metrics: %w", err)
	}

	return nil
}

// RemBenchmarkMetrics removes metrics from a benchmark.
func (c *Controller) RemBenchmarkMetrics(ctx context.Context, benchID string, metrics []string) error {
	exists, err := c.Redis.BenchmarkExists(ctx, benchID)
	if err != nil {
		return fmt.Errorf("could not remove metrics: %w", err)
	}

	if !exists {
		return errors.New("could not remove metrics: benchmark does not exist")
	}

	err = c.Redis.RemBenchmarkMetrics(ctx, benchID, metrics)
	if err != nil {
		return fmt.Errorf("could not remove registries: %w", err)
	}

	return nil
}

// BestRuns returns the best run for each metric selected.
func (c *Controller) BestRuns(ctx context.Context, benchID string,
	metrics ...string,
) (map[string]*types.BenchRun, error) {
	metrics = types.SanitizeNames(metrics)

	metricsInfo, err := c.Redis.SelectBenchmarkMetrics(ctx, benchID, metrics)
	if err != nil {
		return nil, fmt.Errorf("failed getting metrics: %w", err)
	}

	runs, err := c.Redis.BenchmarkRuns(ctx, benchID)
	if err != nil {
		return nil, fmt.Errorf("failed getting benchmark runs: %w", err)
	}

	return types.BestRuns(runs, metricsInfo...), nil
}
