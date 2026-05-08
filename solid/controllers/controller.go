package controllers

import (
	"context"

	"github.com/rs/zerolog"
	"github.com/zeddo123/mlsolid/solid/s3"
	"github.com/zeddo123/mlsolid/solid/store"
	"github.com/zeddo123/mlsolid/solid/types"
	"github.com/zeddo123/pubgo"
)

// Controller defines methods to interact with mlsolid.
type Controller struct {
	Redis              store.RedisStore
	S3                 s3.ObjectStore
	Bus                *pubgo.Bus
	Logger             zerolog.Logger
	PublishBenchEvents bool
}

func (c *Controller) pushBengineEvent(ctx context.Context, registryName string, version int) {
	registry, err := c.Redis.ModelRegistry(ctx, registryName)
	if err != nil {
		c.Logger.Error().Err(err).Msg("could not pull registry from db")

		return
	}

	modelEntry, err := registry.ModelByVersion(version)
	if err != nil {
		c.Logger.Error().Err(err).Msg("could not find model entry")

		return
	}

	benchmarks, err := c.Redis.RegistryBenchmarks(ctx, registryName)
	if err != nil {
		c.Logger.Error().Err(err).Msg("could not pull registry benchmarks")

		return
	}

	benchs, err := c.Redis.BenchmarksWithIDs(ctx, benchmarks)
	if err != nil {
		c.Logger.Error().Err(err).Msg("could not pull benchmarks")

		return
	}

	for _, bench := range benchs {
		c.Logger.Info().
			Str("registry", registryName).
			Int("version", version).
			Str("benchID", bench.ID).
			Msg("publishing benchmark event")

		err = c.Bus.Publish("bengine", types.BenchEvent{
			BenchID:     bench.ID,
			BenchName:   bench.Name,
			Registry:    registryName,
			Version:     int64(version),
			DockerImage: registry.BenchmarkImage,
			ModelURL:    modelEntry.URL,
			DatasetName: bench.DatasetName,
			DatasetURL:  bench.DatasetURL,
			FromS3:      bench.FromS3,
			AutoTag:     bench.AutoTag,
			Tag:         bench.Tag,
		})
		if err != nil {
			c.Logger.Error().
				Err(err).
				Str("benchID", bench.ID).
				Msg("could not publish benchmark event")

			continue
		}
	}
}
