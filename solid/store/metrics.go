package store

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/zedd123/mlsolid/solid/types"
)

func (r *RedisStore) SetMetrics(ctx context.Context, runID string, ms map[string]types.Metric) error {
	_, err := r.Client.Pipelined(ctx, func(p redis.Pipeliner) error {
		r.setMetrics(ctx, p, runID, ms)

		return nil
	})

	return err
}

func (r *RedisStore) SetMetric(ctx context.Context, runID string, m types.Metric) error {
	_, err := r.Client.Pipelined(ctx, func(p redis.Pipeliner) error {
		r.setMetric(ctx, p, runID, m)

		return nil
	})

	return err
}

func (r *RedisStore) RunMetrics(ctx context.Context, runID string) ([]string, error) {
	pattern := fmt.Sprintf("metric:*:%s", runID)

	return r.scanKeys(ctx, pattern)
}

func (r *RedisStore) runMetrics(ctx context.Context, p redis.Pipeliner, runID string) *redis.ScanCmd {
	return p.Scan(ctx, 0, fmt.Sprintf("metric:*:%s", runID), 0)
}

func (r *RedisStore) Metrics(ctx context.Context, runID string) ([]types.Metric, error) {
	return nil, nil
}

func (r *RedisStore) setMetrics(ctx context.Context, p redis.Pipeliner, runID string, ms map[string]types.Metric) map[string][]*redis.StringCmd {
	res := make(map[string][]*redis.StringCmd, 0)

	for key, val := range ms {
		res[key] = r.setMetric(ctx, p, runID, val)
	}

	return res
}

func (r *RedisStore) setMetric(ctx context.Context, p redis.Pipeliner, runID string, m types.Metric) []*redis.StringCmd {
	key := r.makeMetricKey(m.Name(), runID)
	cmds := make([]*redis.StringCmd, 0)

	for _, val := range m.ValsToCommit() {
		cmds = append(cmds, p.XAdd(ctx, &redis.XAddArgs{
			Stream: key,
			Values: map[string]any{
				"Name": m.Name(),
				"Val":  val,
			},
		}))
	}

	return cmds
}

func (r *RedisStore) metrics(ctx context.Context, p redis.Pipeliner, keys []string) []*redis.XMessageSliceCmd {
	res := make([]*redis.XMessageSliceCmd, len(keys))

	for i, key := range keys {
		res[i] = p.XRange(ctx, key, "-", "+")
	}

	return res
}

func (r *RedisStore) parseMetric(ctx context.Context, res *redis.XMessageSliceCmd) (types.Metric, error) {
	msgs, err := res.Result()
	if err != nil {
		return nil, types.NewInternalErr("could not fetch metric")
	}

	vals := make([]any, len(msgs))

	for i, m := range msgs {
		vals[i] = m.Values["Val"]
	}

	var g types.Metric

	switch vals[0].(type) {
	case int:
		g = &types.GenericMetric[int64]{Key: msgs[0].Values["Name"].(string)}
	case float32, float64:
		g = &types.GenericMetric[float64]{Key: msgs[0].Values["Name"].(string)}
	case string:
		g = &types.GenericMetric[string]{Key: msgs[0].Values["Name"].(string)}
	}

	g.SetVals(vals)

	return g, nil
}

func (r *RedisStore) parseMetrics(ctx context.Context, res []*redis.XMessageSliceCmd) (map[string]types.Metric, error) {
	mapping := make(map[string]types.Metric, 0)

	for _, x := range res {
		m, err := r.parseMetric(ctx, x)
		if err != nil {
			mapping[m.Name()] = m
		}
	}

	return mapping, nil
}
