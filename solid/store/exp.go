package store

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/zeddo123/mlsolid/solid/types"
)

func (r *RedisStore) Exp(ctx context.Context, expID string) (*types.Experiment, error) {
	if err := r.ExpExists(ctx, expID); err != nil {
		return nil, err
	}

	ids, err := r.ExpRunIDs(ctx, expID)
	if err != nil {
		return nil, err
	}

	runs, err := r.Runs(ctx, ids)
	if err != nil {
		return nil, err
	}

	return &types.Experiment{
		Name: expID,
		Runs: runs,
	}, nil
}

func (r *RedisStore) ExpRunIDs(ctx context.Context, expID string) ([]string, error) {
	p := r.Client.Pipeline()

	res := r.expRunIDs(ctx, p, expID)

	_, err := p.Exec(ctx)
	if err != nil {
		return nil, types.NewInternalErr("could not fetch experiment runs")
	}

	ids, err := res.Result()
	if err != nil {
		return nil, types.NewInternalErr("could not fetch experiment runs")
	}

	return ids, nil
}

func (r *RedisStore) expRunIDs(ctx context.Context, p redis.Pipeliner, expID string) *redis.StringSliceCmd {
	return p.SMembers(ctx, r.makeExpKey(expID))
}

func (r *RedisStore) ExpExists(ctx context.Context, expID string) error {
	c, err := r.Client.Exists(ctx, r.makeExpKey(expID)).Result()
	if err != nil {
		return types.NewInternalErr("could not check if Experiment exists")
	}

	if c == 0 {
		return types.NewNotFoundErr("could not find experiment")
	}

	return nil
}

func (r *RedisStore) Exps(ctx context.Context) ([]string, error) {
	keys, err := r.scanKeys(ctx, "exp:*")
	if err != nil {
		return nil, types.NewInternalErr("could not fetch experiments")
	}

	ids := make([]string, len(keys))

	for i, key := range keys {
		var id string

		_, err := fmt.Sscanf(key, "exp:%s", &id)
		if err == nil {
			ids[i] = id
		}
	}

	return ids, nil
}
