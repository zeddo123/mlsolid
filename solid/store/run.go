package store

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/zeddo123/mlsolid/solid/types"
)

func (r *RedisStore) SetRun(ctx context.Context, run types.Run) error {
	key := r.makeRunKey(run.Name)

	fn := func(tx *redis.Tx) error {
		_, err := tx.Pipelined(ctx, func(p redis.Pipeliner) error {
			setRunHash(ctx, p, key, run)
			addRunToExperimentIndex(ctx, p, r.makeExpKey(run.ExperimentID), run.Name)
			r.setMetrics(ctx, p, run.Name, run.Metrics)

			return nil
		})
		return err
	}

	return r.runTx(ctx, fn, 10, key)
}

func (r *RedisStore) RunExists(ctx context.Context, runID string) (bool, error) {
	key := r.makeRunKey(runID)

	c, err := r.Client.Exists(ctx, key).Result()
	if err != nil {
		return false, types.NewInternalErr("could not check if run exists")
	}

	return c == 1, nil
}

func (r *RedisStore) Run(ctx context.Context, id string) (*types.Run, error) {
	key := r.makeRunKey(id)

	metricKeys, err := r.RunMetrics(ctx, id)
	if err != nil {
		return nil, types.NewInternalErr("could not fetch metrics")
	}

	p := r.Client.Pipeline()

	hashRes := readRunHash(ctx, p, key)
	metricsRes := r.metrics(ctx, p, metricKeys)

	_, err = p.Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: could not fetch run <%s>", types.ErrInternal, id)
	}

	return r.parseRun(ctx, hashRes, metricsRes)
}

func (r *RedisStore) parseRun(ctx context.Context, hashRes *redis.MapStringStringCmd, metricsRes []*redis.XMessageSliceCmd) (*types.Run, error) {
	metrics, err := r.parseMetrics(ctx, metricsRes)
	if err != nil {
		return nil, types.NewInternalErr("could not parse metrics")
	}

	mapping, err := hashRes.Result()
	if err != nil {
		return nil, types.NewInternalErr("could not fetch run")
	}

	timestamp, err := time.Parse(time.RFC3339, mapping["Timestamp"])
	if err != nil {
		return nil, types.NewInternalErr("malformed run data")
	}

	return &types.Run{
		Name:         mapping["Name"],
		Timestamp:    timestamp,
		ExperimentID: mapping["ExperimentID"],
		Metrics:      metrics,
	}, nil
}

func (r *RedisStore) Runs(ctx context.Context, ids []string) ([]*types.Run, error) {
	keys, err := r.pullRunMetricKeys(ctx, ids)
	if err != nil {
		return nil, err
	}

	p := r.Client.Pipeline()

	res := make(map[string]struct {
		Hash    *redis.MapStringStringCmd
		Metrics []*redis.XMessageSliceCmd
	})

	for _, id := range ids {
		res[id] = struct {
			Hash    *redis.MapStringStringCmd
			Metrics []*redis.XMessageSliceCmd
		}{
			Hash:    readRunHash(ctx, p, r.makeRunKey(id)),
			Metrics: r.metrics(ctx, p, keys[id]),
		}
	}

	_, err = p.Exec(ctx)
	if err != nil {
		return nil, types.NewInternalErr("could not fetch runs")
	}

	runs := make([]*types.Run, 0)
	for _, v := range res {
		run, err := r.parseRun(ctx, v.Hash, v.Metrics)
		if err == nil {
			runs = append(runs, run)
		}
	}

	return runs, nil
}

func (r *RedisStore) pullRunMetricKeys(ctx context.Context, ids []string) (map[string][]string, error) {
	p := r.Client.Pipeline()

	cmds := make([]*redis.ScanCmd, len(ids))
	for i, id := range ids {
		cmds[i] = r.runMetrics(ctx, p, id)
	}

	_, err := p.Exec(ctx)
	if err != nil {
		return nil, types.NewInternalErr("could not pull metrics")
	}

	keys := make(map[string][]string)

	for i, cmd := range cmds {
		ks, err := r.iterate(ctx, cmd.Iterator())
		if err != nil {
			continue
		}

		keys[ids[i]] = ks
	}

	return keys, nil
}

func setRunHash(ctx context.Context, p redis.Pipeliner, key string, run types.Run) *redis.IntCmd {
	return p.HSet(ctx, key, map[string]string{
		"Name":         run.Name,
		"Timestamp":    run.Timestamp.Format(time.RFC3339),
		"ExperimentID": run.ExperimentID,
	})
}

func readRunHash(ctx context.Context, p redis.Pipeliner, key string) *redis.MapStringStringCmd {
	return p.HGetAll(ctx, key)
}

func addRunToExperimentIndex(ctx context.Context, p redis.Pipeliner, expKey, id string) *redis.IntCmd {
	return p.SAdd(ctx, expKey, id)
}
