package store

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/zeddo123/mlsolid/solid/types"
)

func (r *RedisStore) ArtifactExist(ctx context.Context, runID string, id string) error {
	c, err := r.Client.Exists(ctx, r.makeArtifactKey(id, runID)).Result()
	if err != nil {
		return types.NewInternalErr("could not check if artifact exists")
	}

	if c == 0 {
		return types.NewNotFoundErr("could not find artifact")
	}

	return nil
}

func (r *RedisStore) ArtifactsExist(ctx context.Context, runID string, ids []string) map[string]error {
	errs := make(map[string]error)

	for _, id := range ids {
		errs[id] = r.ArtifactExist(ctx, runID, id)
	}

	return errs
}

func (r *RedisStore) SetArtifact(ctx context.Context, runID string, a types.SavedArtifact) error {
	p := r.Client.Pipeline()

	r.setArtifact(ctx, p, runID, a)

	_, err := p.Exec(ctx)
	if err != nil {
		return types.NewInternalErr("could not save artifact url")
	}

	return nil
}

func (r *RedisStore) SetArtifacts(ctx context.Context, runID string, as []types.SavedArtifact) error {
	p := r.Client.Pipeline()

	for _, a := range as {
		r.setArtifact(ctx, p, runID, a)
	}

	_, err := p.Exec(ctx)
	if err != nil {
		return types.NewInternalErr("could not save artifacts to store")
	}

	return nil
}

func (r *RedisStore) setArtifact(ctx context.Context, p redis.Pipeliner,
	runID string, a types.SavedArtifact,
) *redis.IntCmd {
	return p.HSet(ctx, r.makeArtifactKey(a.Name, runID), map[string]string{
		"Name":  a.Name,
		"Type":  string(a.ContentType),
		"S3Key": a.S3Key,
	})
}

func (r *RedisStore) Artifacts(ctx context.Context, runID string) (map[string]types.SavedArtifact, error) {
	keys, err := r.scanKeys(ctx, fmt.Sprintf("artifact:*:%s", runID))
	if err != nil {
		return nil, err
	}

	var cmds []*redis.MapStringStringCmd

	_, err = r.Client.Pipelined(ctx, func(p redis.Pipeliner) error {
		cmds = r.artifacts(ctx, p, keys)
		return nil
	})
	if err != nil {
		return nil, types.NewInternalErr("could not fetch artifacts")
	}

	artifacts := make(map[string]types.SavedArtifact, 0)

	for _, cmd := range cmds {
		mapping, err := cmd.Result()
		if err != nil {
			continue
		}

		artifacts[mapping["Name"]] = types.SavedArtifact{
			Name:        mapping["Name"],
			ContentType: types.ContentType(mapping["Type"]),
			S3Key:       mapping["S3Key"],
		}
	}

	return artifacts, nil
}

func (r *RedisStore) Artifact(ctx context.Context, runID string, id string) (types.SavedArtifact, error) {
	p := r.Client.Pipeline()

	urlRes := r.artifact(ctx, p, runID, id)

	_, err := p.Exec(ctx)
	if err != nil {
		return types.SavedArtifact{}, types.NewInternalErr("could not pull artifact s3 key")
	}

	mapping, err := urlRes.Result()
	if err != nil {
		return types.SavedArtifact{}, types.NewInternalErr("could not pull artifact s3 key")
	}

	return types.SavedArtifact{
		Name:        mapping["Name"],
		ContentType: types.ContentType(mapping["Type"]),
		S3Key:       mapping["S3Key"],
	}, nil
}

func (r *RedisStore) artifact(ctx context.Context, p redis.Pipeliner, runID string, id string) *redis.MapStringStringCmd {
	return p.HGetAll(ctx, r.makeArtifactKey(id, runID))
}

func (r *RedisStore) artifacts(ctx context.Context, p redis.Pipeliner, keys []string) []*redis.MapStringStringCmd {
	cmds := make([]*redis.MapStringStringCmd, len(keys))

	for i, key := range keys {
		cmds[i] = p.HGetAll(ctx, key)
	}

	return cmds
}
