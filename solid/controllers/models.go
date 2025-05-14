package controllers

import (
	"context"

	"github.com/zeddo123/mlsolid/solid/types"
)

func (c *Controller) ModelRegistry(ctx context.Context, name string) (*types.ModelRegistry, error) {
	registry, err := c.Redis.ModelRegistry(ctx, name)
	if err != nil {
		return nil, err
	}

	return registry, nil
}

func (c *Controller) CreateModelRegistry(ctx context.Context, name string) error {
	if name == "" {
		return types.NewBadRequest("model registry name cannot be empty")
	}

	return c.Redis.CreateModelRegistry(ctx, *types.NewModelRegistry(name))
}

func (c *Controller) LastModelEntry(ctx context.Context, registryName string) (types.ModelEntry, error) {
	return c.Redis.LastModel(ctx, registryName)
}

func (c *Controller) TaggedModel(ctx context.Context, registryName string, tag string) (types.ModelEntry, error) {
	return c.Redis.ModelByTag(ctx, registryName, tag)
}

func (c *Controller) AddModelEntry(ctx context.Context, registryName string, url string, tags ...string) error {
	registry, err := c.Redis.ModelRegistry(ctx, registryName)
	if err != nil {
		return err
	}

	registry.Add(url, tags...)

	return c.Redis.UpdateModelRegistry(ctx, *registry)
}

func (c *Controller) AddArtifactToRegistry(ctx context.Context, registryName string, runID string,
	artifactID string, tags ...string,
) error {
	artifact, err := c.Redis.Artifact(ctx, runID, artifactID)
	if err != nil {
		return err
	}

	return c.AddModelEntry(ctx, registryName, artifact.S3Key, tags...)
}

func (c *Controller) TagModel(ctx context.Context, registryName string, version int, tags ...string) error {
	registry, err := c.Redis.ModelRegistry(ctx, registryName)
	if err != nil {
		return err
	}

	for _, tag := range tags {
		err := registry.AddTag(tag, version)
		if err != nil {
			return err
		}
	}

	return c.Redis.UpdateModelRegistry(ctx, *registry)
}
