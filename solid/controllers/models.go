package controllers

import (
	"context"
	"fmt"

	"github.com/zeddo123/mlsolid/solid/types"
)

// ModelRegistry pulls a model registry.
func (c *Controller) ModelRegistry(ctx context.Context, name string) (*types.ModelRegistry, error) {
	registry, err := c.Redis.ModelRegistry(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed pulling model registry %q: %w", name, err)
	}

	return registry, nil
}

// CreateModelRegistry creates a new model registry with the specified name.
func (c *Controller) CreateModelRegistry(ctx context.Context,
	name string, benchmarkingOps types.RegistryBenchmarkOps,
) error {
	if name == "" {
		return types.NewBadRequest("model registry name cannot be empty") //nolint: wrapcheck
	}

	err := c.Redis.CreateModelRegistry(ctx, *types.NewModelRegistryWithBenchmarkOps(name, benchmarkingOps))
	if err != nil {
		return fmt.Errorf("failed creating model registry %q : %w", name, err)
	}

	return nil
}

// LastModelEntry returns the last model entry in the model registry.
func (c *Controller) LastModelEntry(ctx context.Context, registryName string) (types.ModelEntry, error) {
	entry, err := c.Redis.LastModel(ctx, registryName)
	if err != nil {
		return types.ModelEntry{}, fmt.Errorf("failed getting last model of %q: %w", registryName, err)
	}

	return entry, nil
}

// TaggedModel returns the latest model entry with the specified tag.
func (c *Controller) TaggedModel(ctx context.Context, registryName string, tag string) (types.ModelEntry, error) {
	entry, err := c.Redis.ModelByTag(ctx, registryName, tag)
	if err != nil {
		return types.ModelEntry{}, fmt.Errorf("failed getting latest enty with %q model from %q: %w", tag, registryName, err)
	}

	return entry, nil
}

// AddModelEntry adds a new model entry to the registry.
func (c *Controller) AddModelEntry(ctx context.Context, registryName string, url string, tags ...string) error {
	registry, err := c.Redis.ModelRegistry(ctx, registryName)
	if err != nil {
		return fmt.Errorf("failed retrieving model registry %q: %w", registryName, err)
	}

	registry.Add(url, tags...)

	err = c.Redis.UpdateModelRegistry(ctx, registry)
	if err != nil {
		return fmt.Errorf("update model registry failed: %w", err)
	}

	if c.PublishBenchEvents {
		go c.pushBengineEvent(ctx, registryName, registry.LatestVersion())
	}

	return nil
}

// AddArtifactToRegistry adds a model artifact as a new model entry to a registry.
func (c *Controller) AddArtifactToRegistry(ctx context.Context, registryName string, runID string,
	artifactID string, tags ...string,
) error {
	registry, err := c.Redis.ModelRegistry(ctx, registryName)
	if err != nil {
		return fmt.Errorf("failed retrieving model registry %q: %w", registryName, err)
	}

	artifact, err := c.Redis.Artifact(ctx, runID, artifactID)
	if err != nil {
		return fmt.Errorf("failed pulling artifact: %w", err)
	}

	registry.AddArtifact(runID, artifactID, artifact.S3Key, tags...)

	err = c.Redis.UpdateModelRegistry(ctx, registry)
	if err != nil {
		return fmt.Errorf("update registry failed: %w", err)
	}

	if c.PublishBenchEvents {
		go c.pushBengineEvent(ctx, registryName, registry.LatestVersion())
	}

	return nil
}

// TagModel tags a model entry of a registry with the specified tag.
func (c *Controller) TagModel(ctx context.Context, registryName string, version int, tags ...string) error {
	registry, err := c.Redis.ModelRegistry(ctx, registryName)
	if err != nil {
		return fmt.Errorf("failed pulling model registry: %w", err)
	}

	for _, tag := range tags {
		err := registry.AddTag(tag, version)
		if err != nil {
			return fmt.Errorf("failed adding tag %q to version %q: %w", tag, version, err)
		}
	}

	err = c.Redis.UpdateModelRegistry(ctx, registry)
	if err != nil {
		return fmt.Errorf("failed updating model registry: %w", err)
	}

	return nil
}

// ModelRegistriesID retrieves all known model registry ids.
func (c *Controller) ModelRegistriesID(ctx context.Context) ([]string, error) {
	ids, err := c.Redis.ModelRegistriesID(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not pull registry ids: %w", err)
	}

	return ids, nil
}

// ModelRegistriesIDPage retrieves a single page of model registry ids starting at cursor.
// A returned cursor of 0 means there are no more pages.
func (c *Controller) ModelRegistriesIDPage(ctx context.Context, cursor uint64, limit int64) ([]string, uint64, error) {
	ids, next, err := c.Redis.ModelRegistriesIDPage(ctx, cursor, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("could not pull registry ids page: %w", err)
	}

	return ids, next, nil
}

// UpdateRegistryDockerImage updates the docker image of a registry.
func (c *Controller) UpdateRegistryDockerImage(ctx context.Context, registry, dockerImage string) error {
	err := c.Redis.UpdateRegistryDockerImage(ctx, registry, dockerImage)
	if err != nil {
		return fmt.Errorf("update failed: %w", err)
	}

	return nil
}

func (c *Controller) UpdateRegistryBenchmarkGpuPassthrough(ctx context.Context, registry string, pass bool) error {
	err := c.Redis.UpdateRegistryBenchmarkGpuPassthrough(ctx, registry, pass)
	if err != nil {
		return fmt.Errorf("update failed: %w", err)
	}

	return nil
}
