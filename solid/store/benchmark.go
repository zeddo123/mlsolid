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

// activeBenchRunTTL bounds how long an "ActiveBenchRun" field can survive
// without being refreshed, so a run that never gets cleared (crashed
// process, RemActiveBenchRun failing to reach Redis) self-heals instead of
// showing as active forever. Set well above any expected benchmark run
// duration.
const activeBenchRunTTL = 5 * 24 * time.Hour

// CreateBenchmark creates a new benchmark.
func (r *RedisStore) CreateBenchmark(ctx context.Context, b types.Bench) (bool, error) {
	benchKey := r.makeBenchmarkKey(b.ID)

	score, err := r.Client.Incr(ctx, BenchmarksCounterKey).Result()
	if err != nil {
		return false, fmt.Errorf("%w: could not allocate benchmark index score: %w", types.ErrInternal, err)
	}

	p := r.Client.Pipeline()

	// add new benchmark to index
	p.ZAddNX(ctx, BenchmarksKey, redis.Z{Score: float64(score), Member: b.ID})
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

	_, err = p.Exec(ctx)
	if err != nil {
		return false, fmt.Errorf("%w: could not set benchmark: %w", types.ErrInternal, err)
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
		return fmt.Errorf("%w: could not set benchmark registries: %w", types.ErrInternal, err)
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
			log.Println("could not marshal metric to json:", err)

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

	data := r.Client.HGetAll(ctx, key)
	registries := r.Client.SMembers(ctx, r.makeBenchmarkRegistriesKey(benchID))
	metrics := r.Client.HGetAll(ctx, r.makeBenchmarkMetricsKey(benchID))

	return parseBenchmark(benchID, data, metrics, registries)
}

// BenchmarkMetrics pulls metrics linked to a benchmark.
func (r *RedisStore) BenchmarkMetrics(ctx context.Context, benchID string) ([]types.BenchMetric, error) {
	key := r.makeBenchmarkMetricsKey(benchID)

	return parseBenchmarkMetrics(r.Client.HGetAll(ctx, key))
}

// SelectBenchmarkMetrics searches for metrics linked to a benchmark.
func (r *RedisStore) SelectBenchmarkMetrics(ctx context.Context, benchID string,
	metrics []string,
) ([]types.BenchMetric, error) {
	key := r.makeBenchmarkMetricsKey(benchID)

	hash, err := r.Client.HMGet(ctx, key, metrics...).Result()
	if err != nil {
		return nil, fmt.Errorf("failed pulling metrics: %w", err)
	}

	info := types.BenchMetric{} //nolint: exhaustruct
	benchMetrics := make([]types.BenchMetric, 0, len(hash))

	for _, m := range hash {
		if m != nil {
			content, ok := m.(string)
			if !ok {
				continue
			}

			err := json.Unmarshal([]byte(content), &info)
			if err != nil {
				return nil, fmt.Errorf("could not unmarshall metric: %w", err)
			}

			benchMetrics = append(benchMetrics, info)
		}
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

		// Clear any previously recorded run so re-recording the same
		// registry+version replaces its metrics instead of merging with
		// stale fields left over from an earlier run.
		p.Del(ctx, runKey)

		p.HSet(ctx, runKey, map[string]any{
			"Registry":  run.Registry,
			"Version":   run.Version,
			"Timestamp": run.Timestamp,
			"Start":     run.Start,
			"End":       run.End,
		})

		metrics := make(map[string]any, len(run.Metrics))
		for k, metric := range run.SanitizedMetrics() {
			metrics[k] = metric
		}

		// Redis rejects a HSET with zero field/value pairs, so skip it
		// entirely for a run with no metrics rather than erroring out.
		if len(metrics) > 0 {
			p.HSet(ctx, runKey, metrics)
		}

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
			r.Logger.Error().
				Err(err).
				Str("benchID", benchID).
				Str("benchRun", runKeys[i]).
				Msg("failed fetching benchmark run")

			continue
		}

		run, err := r.parseBenchRun(m)
		if err != nil {
			r.Logger.Error().
				Err(err).
				Str("benchID", benchID).
				Str("benchRun", runKeys[i]).
				Msg("could not parse benchmark run")

			continue
		}

		runs[i] = run
	}

	return runs, nil
}

// SetActiveBenchRun records run as the benchmark run currently in progress
// for benchID, so callers can tell which run is executing before it
// finishes and is written via RecordRuns. It is stored in the
// "ActiveBenchRun" field of the benchmark's hash as a JSON blob and
// overwrites whatever active run was previously set, so only one run is
// ever considered active per benchmark. Callers are responsible for
// clearing it with RemActiveBenchRun once the run completes.
func (r *RedisStore) SetActiveBenchRun(ctx context.Context, benchID string, run types.BenchRun) error {
	key := r.makeBenchmarkKey(benchID)

	content, err := json.Marshal(run)
	if err != nil {
		return fmt.Errorf("%w: could not marshal benchmark run due to %w", types.ErrInternal, err)
	}

	_, err = r.Client.HSet(ctx, key, "ActiveBenchRun", content).Result()
	if err != nil {
		return fmt.Errorf("%w: could not set active benchmark field of benchmark %q due to %w",
			types.ErrInternal, benchID, err)
	}

	_, err = r.Client.HExpire(ctx, key, activeBenchRunTTL, "ActiveBenchRun").Result()
	if err != nil {
		return fmt.Errorf("%w: could not set expiry on active benchmark field of benchmark %q due to %w",
			types.ErrInternal, benchID, err)
	}

	return nil
}

// RemActiveBenchRun clears the "ActiveBenchRun" field set by
// SetActiveBenchRun, e.g. once a run has finished and been recorded. It is
// a no-op, not an error, when no active run is set.
func (r *RedisStore) RemActiveBenchRun(ctx context.Context, benchID string) error {
	_, err := r.Client.HDel(ctx, r.makeBenchmarkKey(benchID), "ActiveBenchRun").Result()
	if err != nil {
		return fmt.Errorf("%w: could not delete active benchmark run field due to %w", types.ErrInternal, err)
	}

	return nil
}

// Benchmarks returns all known benchmarks.
func (r *RedisStore) Benchmarks(ctx context.Context) ([]string, error) {
	benchs, err := r.zIndexAll(ctx, BenchmarksKey)
	if err != nil {
		return nil, fmt.Errorf("%w: could not pull benchmarks: %w", types.ErrInternal, err)
	}

	return benchs, nil
}

// BenchmarksPage returns a single page of benchmark ids starting at cursor.
// A returned cursor of 0 means there are no more pages.
func (r *RedisStore) BenchmarksPage(ctx context.Context, cursor uint64, count int64) ([]string, uint64, error) {
	ids, next, err := r.zIndexPage(ctx, BenchmarksKey, cursor, count)
	if err != nil {
		return nil, 0, fmt.Errorf("%w: could not pull benchmarks page: %w", types.ErrInternal, err)
	}

	return ids, next, nil
}

// BackfillBenchmarksIndex migrates BenchmarksKey from its legacy Set representation
// to a Sorted Set, if it hasn't been already. It is idempotent and safe to run on
// every startup.
func (r *RedisStore) BackfillBenchmarksIndex(ctx context.Context) error {
	if err := r.migrateSetIndexToSortedSet(ctx, BenchmarksKey, BenchmarksCounterKey); err != nil {
		return fmt.Errorf("could not migrate benchmarks index: %w", err)
	}

	return nil
}

// BenchmarkNames returns a map of benchmark id to benchmark name for the given ids.
func (r *RedisStore) BenchmarkNames(ctx context.Context, ids []string) (map[string]string, error) {
	p := r.Client.Pipeline()

	cmds := make(map[string]*redis.MapStringStringCmd, len(ids))
	for _, id := range ids {
		cmds[id] = p.HGetAll(ctx, r.makeBenchmarkKey(id))
	}

	_, err := p.Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: could not pull benchmark names: %w", types.ErrInternal, err)
	}

	names := make(map[string]string, len(ids))
	for id, cmd := range cmds {
		names[id] = cmd.Val()["Name"]
	}

	return names, nil
}

// BenchmarksWithIDs returns all benchmarks with specified ids without checking if they exist in the store.
func (r *RedisStore) BenchmarksWithIDs(ctx context.Context, ids []string) ([]*types.Bench, error) {
	type tempResult struct {
		data       *redis.MapStringStringCmd
		metrics    *redis.MapStringStringCmd
		registries *redis.StringSliceCmd
	}

	results := make([]*types.Bench, len(ids))
	partialResults := make(map[string]tempResult, len(ids))

	p := r.Client.Pipeline()

	for _, id := range ids {
		partialResults[id] = tempResult{
			data:       p.HGetAll(ctx, r.makeBenchmarkKey(id)),
			registries: p.SMembers(ctx, r.makeBenchmarkRegistriesKey(id)),
			metrics:    p.HGetAll(ctx, r.makeBenchmarkMetricsKey(id)),
		}
	}

	_, err := p.Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed pulling benchmarks: %w", err)
	}

	for idx, id := range ids {
		prslt := partialResults[id]

		bench, err := parseBenchmark(id, prslt.data, prslt.metrics, prslt.registries)
		if err != nil {
			log.Println("could not parse benchmark:", err)

			continue
		}

		results[idx] = bench
	}

	return results, nil
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

func parseBenchmark(benchID string, data, metrics *redis.MapStringStringCmd,
	registries *redis.StringSliceCmd,
) (*types.Bench, error) {
	mapping, err := data.Result()
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

	timestamp, err := time.Parse(time.RFC3339, mapping["Timestamp"])
	if err != nil {
		return nil, fmt.Errorf("could not parse benchmark timestamp: %w", err)
	}

	regs, err := registries.Result()
	if err != nil {
		return nil, fmt.Errorf("could not get registries: %w", err)
	}

	mets, err := parseBenchmarkMetrics(metrics)
	if err != nil {
		return nil, fmt.Errorf("could not parse benchmark metrics: %w", err)
	}

	var benchRun types.BenchRun

	// "ActiveBenchRun" is absent for the common case of a benchmark with no
	// run in progress, so only attempt to parse it when present.
	if run, ok := mapping["ActiveBenchRun"]; ok && run != "" {
		err = json.Unmarshal([]byte(mapping["ActiveBenchRun"]), &benchRun)
		if err != nil {
			log.Printf("could not parse active benchmark run Run=%s err=%s\n", mapping["ActiveBenchRun"], err)
		}
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
		Registries:     regs,
		Metrics:        mets,
		ActiveBenchRun: benchRun,
	}, nil
}

func parseBenchmarkMetrics(metrics *redis.MapStringStringCmd) ([]types.BenchMetric, error) {
	mapping, err := metrics.Result()
	if err != nil {
		return nil, fmt.Errorf("could not pull benchmark metrics: %w", err)
	}

	benchMetrics := make([]types.BenchMetric, 0, len(mapping))

	metric := types.BenchMetric{} //nolint: exhaustruct

	for _, v := range mapping {
		err := json.Unmarshal([]byte(v), &metric)
		if err != nil {
			return nil, fmt.Errorf("could not unmarshall metric: %w", err)
		}

		benchMetrics = append(benchMetrics, metric)
	}

	return benchMetrics, nil
}

func (r *RedisStore) parseBenchRun(m map[string]string) (*types.BenchRun, error) {
	reg := m["Registry"]

	v, err := strconv.ParseInt(m["Version"], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("could not parse run Version: %w", err)
	}

	timestamp, err := time.Parse(time.RFC3339, m["Timestamp"])
	if err != nil {
		return nil, fmt.Errorf("could not parse run Timestamp: %w", err)
	}

	start, err := time.Parse(time.RFC3339, m["Start"])
	if err != nil {
		start = time.Time{}
	}

	end, err := time.Parse(time.RFC3339, m["End"])
	if err != nil {
		end = time.Time{}
	}

	delete(m, "Registry")
	delete(m, "Version")
	delete(m, "Timestamp")
	delete(m, "Start")
	delete(m, "End")

	metrics := make(map[string]float32, len(m))

	for k, v := range m {
		metric, err := strconv.ParseFloat(v, 32)
		if err != nil {
			r.Logger.Error().
				Err(err).
				Str("metric", k).
				Str("value", v).
				Msg("parsing benchmark run: could not parse version number to float")

			continue
		}

		metrics[k] = float32(metric)
	}

	return &types.BenchRun{
		Timestamp: timestamp,
		Start:     start,
		End:       end,
		Registry:  reg,
		Version:   v,
		Metrics:   metrics,
	}, nil
}
