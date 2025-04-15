package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/zeddo123/mlsolid/solid/types"
)

func (r *RedisStore) CreateModelRegistry(ctx context.Context, m types.ModelRegistry) error {
	fn := func(tx *redis.Tx) error {
		_, err := tx.Pipelined(ctx, func(p redis.Pipeliner) error {
			return r.createModelRegistry(ctx, p, m)
		})

		return err
	}

	return r.runTx(ctx, fn, TransactionMaxTries, r.makeModelRegistryInfoKey(m.Name))
}

func (r *RedisStore) createModelRegistry(ctx context.Context, p redis.Pipeliner, m types.ModelRegistry) error {
	err := r.ModelRegistryExists(ctx, m.Name)
	if !errors.Is(err, types.ErrNotFound) {
		return err
	}

	infoKey := r.makeModelRegistryInfoKey(m.Name)
	entriesKey := r.makeModelRegistryKey(m.Name)
	tagsIndexKey := r.makeModelRegistryTagsKey(m.Name)

	// Setting model info under key "info:registry:<name>"
	p.HSet(ctx, infoKey, map[string]string{
		"Name":      m.Name,
		"Timestamp": fmt.Sprint(time.Now().UnixMilli()),
	})

	// Setting model entries under key "registry:<name>"
	entries, err := m.MarshalEntries()
	if err != nil {
		return err
	}

	for _, entry := range entries {
		p.LPush(ctx, entriesKey, entry)
	}

	// Setting model tags under key "tag:registry:%s:%s" with
	// an index under "tag:registry:%s"
	for tag, entries := range m.Tags {
		p.SAdd(ctx, tagsIndexKey, tag)

		for _, entry := range entries {
			p.LPush(ctx, r.makeModelRegistryTagKey(m.Name, tag), entry)
		}
	}

	return nil
}

func (r *RedisStore) ModelRegistry(ctx context.Context, name string) (*types.ModelRegistry, error) { //nolint: cyclop
	if err := r.ModelRegistryExists(ctx, name); err != nil {
		return nil, err
	}

	infoKey := r.makeModelRegistryInfoKey(name)
	entriesKey := r.makeModelRegistryKey(name)
	tagsKey := r.makeModelRegistryTagsKey(name)

	p := r.Client.Pipeline()

	infores := p.HGetAll(ctx, infoKey)

	entriesres := p.LRange(ctx, entriesKey, 0, -1)

	tagsres := p.SMembers(ctx, tagsKey)

	_, err := p.Exec(ctx)
	if err != nil {
		return nil, types.NewInternalErr("could not fetch model registry")
	}

	info, err := infores.Result()
	if err != nil {
		return nil, types.NewInternalErr(err.Error())
	}

	entries, err := entriesres.Result()
	if err != nil {
		return nil, types.NewInternalErr(err.Error())
	}

	tags, err := tagsres.Result()
	if err != nil {
		return nil, types.NewInternalErr(err.Error())
	}

	p = r.Client.Pipeline()

	for _, tag := range tags {
		p.LRange(ctx, r.makeModelRegistryTagKey(name, tag), 0, -1)
	}

	cmds, err := p.Exec(ctx)
	if err != nil {
		return nil, types.NewInternalErr(err.Error())
	}

	registry := types.NewModelRegistry(info["Name"])

	for _, e := range entries {
		var entry types.ModelEntry

		err := json.Unmarshal([]byte(e), &entry)
		if err != nil {
			return nil, types.NewInternalErr(err.Error())
		}

		registry.Models = append(registry.Models, entry)
	}

	for i, cmd := range cmds {
		c, ok := cmd.(*redis.StringSliceCmd)
		if !ok {
			return nil, types.NewInternalErr("could not read tags")
		}

		idxs, err := c.Result()
		if err != nil {
			return nil, types.NewInternalErr("could not read tags")
		}

		intIdxs := make([]int, len(idxs))

		for i, idx := range idxs {
			id, err := strconv.Atoi(idx)
			if err != nil {
				return nil, types.NewInternalErr("could not read index")
			}

			intIdxs[i] = id
		}

		registry.Tags[tags[i]] = intIdxs
	}

	return registry, nil
}

func (r *RedisStore) LastModel(ctx context.Context, name string) (types.ModelEntry, error) {
	entries, err := r.Client.LRange(ctx, r.makeModelRegistryKey(name), 0, 0).Result()
	if err != nil {
		return types.ModelEntry{}, types.NewInternalErr("could not fetch model registry entries")
	}

	if len(entries) < 1 {
		return types.ModelEntry{}, types.NewNotFoundErr("could not find model registry entry")
	}

	var entry types.ModelEntry

	err = json.Unmarshal([]byte(entries[0]), &entry)
	if err != nil {
		return types.ModelEntry{}, types.NewInternalErr("could not process model entry")
	}

	return entry, nil
}

func (r *RedisStore) ModelByVersion(ctx context.Context, name string, version int) (types.ModelEntry, error) {
	entries, err := r.Client.LRange(ctx, r.makeModelRegistryKey(name), int64(version)-1, int64(version)-1).Result()
	if err != nil {
		return types.ModelEntry{}, types.NewInternalErr("could not fetch model register entry")
	}

	if len(entries) < 1 {
		return types.ModelEntry{}, types.NewNotFoundErr("could not find model registry entry")
	}

	var entry types.ModelEntry

	err = json.Unmarshal([]byte(entries[0]), &entry)
	if err != nil {
		return types.ModelEntry{}, types.NewInternalErr("could not process model entry")
	}

	return entry, nil
}

func (r *RedisStore) ModelByTag(ctx context.Context, name string, tag string) (types.ModelEntry, error) {
	idxs, err := r.Client.LRange(ctx, r.makeModelRegistryTagKey(name, tag), 0, 0).Result()
	if err != nil {
		return types.ModelEntry{}, types.NewInternalErr("could not fetch model register entry")
	}

	if len(idxs) < 1 {
		return types.ModelEntry{}, types.NewNotFoundErr("could not find model registry entry")
	}

	idx, err := strconv.Atoi(idxs[0])
	if err != nil {
		return types.ModelEntry{}, types.NewInternalErr("could not read model registry idx")
	}

	return r.ModelByVersion(ctx, name, idx)
}

// ModelRegistryExists returns non-nil error if registry was not found.
func (r *RedisStore) ModelRegistryExists(ctx context.Context, name string) error {
	c, err := r.Client.Exists(ctx, r.makeModelRegistryInfoKey(name)).Result()
	if err != nil {
		return types.NewInternalErr(err.Error())
	}

	if c != 1 {
		return types.NewNotFoundErr("could not find model registry")
	}

	return nil
}

func (r *RedisStore) AddModel(ctx context.Context, name string, m types.ModelEntry, tags ...string) error {
	if err := r.ModelRegistryExists(ctx, name); err != nil {
		return err
	}

	key := r.makeModelRegistryKey(name)

	l, err := r.Client.LLen(ctx, key).Result()
	if err != nil {
		return types.NewInternalErr("could not query model register size")
	}

	b, err := json.Marshal(m)
	if err != nil {
		return types.NewInternalErr("could not process model entry")
	}

	p := r.Client.Pipeline()

	r.Client.LPush(ctx, key, b)

	for _, tag := range tags {
		r.Client.SAdd(ctx, r.makeModelRegistryTagsKey(name), tag)
		p.LPush(ctx, r.makeModelRegistryTagKey(name, tag), l)
	}

	_, err = p.Exec(ctx)
	if err != nil {
		return types.NewInternalErr("could not add model entry")
	}

	return nil
}

func (r *RedisStore) UpdateModelRegistry(ctx context.Context, m types.ModelRegistry) error {
	fn := func(tx *redis.Tx) error {
		_, err := tx.Pipelined(ctx, func(p redis.Pipeliner) error {
			return r.updateModelRegistry(ctx, p, m)
		})

		return err
	}

	return r.runTx(ctx, fn, TransactionMaxTries, r.makeModelRegistryInfoKey(m.Name),
		r.makeModelRegistryKey(m.Name), r.makeModelRegistryTagsKey(m.Name))
}

func (r *RedisStore) updateModelRegistry(ctx context.Context, p redis.Pipeliner, m types.ModelRegistry) error {
	if err := r.ModelRegistryExists(ctx, m.Name); err != nil {
		return err
	}

	entriesKey := r.makeModelRegistryKey(m.Name)
	tagsIndexKey := r.makeModelRegistryTagsKey(m.Name)

	// Setting model entries under key "registry:<name>"
	entries, err := m.MarshalEntries()
	if err != nil {
		return err
	}

	// TODO: delete any previous tags that were deleted
	p.Del(ctx, entriesKey, tagsIndexKey)

	for _, entry := range entries {
		p.LPush(ctx, entriesKey, entry)
	}

	// Setting model tags under key "tag:registry:%s:%s" with
	// an index under "tag:registry:%s"
	for tag, entries := range m.Tags {
		p.SAdd(ctx, tagsIndexKey, tag)

		tagKey := r.makeModelRegistryTagKey(m.Name, tag)

		p.Del(ctx, tagKey)

		for _, entry := range entries {
			p.LPush(ctx, tagKey, entry)
		}
	}

	return nil
}
