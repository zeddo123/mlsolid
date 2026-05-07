package controllers

import (
	"context"
	"log"

	"github.com/zeddo123/mlsolid/solid/s3"
	"github.com/zeddo123/mlsolid/solid/store"
	"github.com/zeddo123/mlsolid/solid/types"
	"github.com/zeddo123/pubgo"
)

// Controller defines methods to interact with mlsolid.
type Controller struct {
	Redis store.RedisStore
	S3    s3.ObjectStore
	Bus   *pubgo.Bus
}

func (c *Controller) pushBengineEvent(ctx context.Context, registryName string, version int) {
	registry, err := c.Redis.ModelRegistry(ctx, registryName)
	if err != nil {
		log.Println("could not pull registry from db", err)

		return
	}

	modelEntry, err := registry.ModelByVersion(version)
	if err != nil {
		log.Println("could not find model entry", err)

		return
	}

	benchmarks, err := c.Redis.RegistryBenchmarks(ctx, registryName)
	if err != nil {
		log.Println("could not pull registry benchmarks", err)

		return
	}

	benchs, err := c.Redis.BenchmarksWithId(ctx, benchmarks)
	if err != nil {
		log.Println("could not pull benchmarks", err)

		return
	}

	for _, bench := range benchs {
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
			log.Println("could not publish benchmark event", bench.ID)

			continue
		}
	}
}
