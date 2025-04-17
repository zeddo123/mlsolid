package controllers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"

	"github.com/zeddo123/mlsolid/solid/types"
)

func (c *Controller) CreateRun(ctx context.Context, run types.Run) error {
	if c.S3 == nil {
		return types.NewInternalErr("object store is not configured")
	}

	ok, err := c.Redis.RunExists(ctx, run.Name)
	if err != nil {
		return err
	}

	if ok {
		return types.NewAlreadyInUseErr(fmt.Sprintf("run id <%s> already in use", run.Name))
	}

	err = c.Redis.SetRun(ctx, run)
	if err != nil {
		return err
	}

	artifacts, uploaderr := c.S3.UploadArtifacts(ctx, run.ArtifactsSlice())
	if uploaderr != nil {
		log.Println("not all artifacts were uploaded", uploaderr)
	}

	err = c.Redis.SetArtifacts(ctx, run.Name, artifacts)
	if err != nil {
		return err
	}

	return uploaderr
}

func (c *Controller) Run(ctx context.Context, runID string) (*types.Run, error) {
	id := types.NormalizeID(runID)

	ok, err := c.Redis.RunExists(ctx, id)
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, types.NewNotFoundErr(fmt.Sprintf("run <%s> not found", id))
	}

	return c.Redis.Run(ctx, id)
}

func (c *Controller) Runs(ctx context.Context, ids []string) ([]*types.Run, error) {
	normalized := make([]string, len(ids))

	for i, id := range ids {
		normalized[i] = types.NormalizeID(id)
	}

	runs, err := c.Redis.Runs(ctx, normalized)
	if err != nil {
		return nil, err
	}

	return runs, nil
}

func (c *Controller) AddMetrics(ctx context.Context, runID string, m []types.Metric) error {
	ok, err := c.Redis.RunExists(ctx, types.NormalizeID(runID))
	if err != nil {
		return err
	}

	if !ok {
		return types.NewNotFoundErr("could not find run")
	}

	mapping := make(map[string]types.Metric)

	for _, m := range m {
		mapping[m.Name()] = m
	}

	err = c.Redis.SetMetrics(ctx, runID, mapping)
	if err != nil {
		return types.NewInternalErr("could not set metrics")
	}

	return nil
}

func (c *Controller) AddArtifacts(ctx context.Context, runID string, as []types.Artifact) error {
	ids := types.ArtifactIDs(as)
	artifactsMap := types.ArtifactIDMap(as)

	errs := c.Redis.ArtifactsExist(ctx, runID, ids)

	toUpload := make([]types.Artifact, 0, len(as))

	for id, err := range errs {
		if errors.Is(err, types.ErrNotFound) {
			toUpload = append(toUpload, artifactsMap[id])
		} else if err != nil {
			return err
		}
	}

	artifacts, uploadErr := c.S3.UploadArtifacts(ctx, toUpload)

	err := c.Redis.SetArtifacts(ctx, runID, artifacts)
	if err != nil {
		return err
	}

	return uploadErr
}

// Artifact fetches an artifact and returns a Reader to its content.
func (c *Controller) Artifact(ctx context.Context, runID string,
	artifact string,
) (*types.SavedArtifact, io.ReadCloser, error) {
	a, err := c.Redis.Artifact(ctx, runID, artifact)
	if err != nil {
		return nil, nil, err
	}

	body, err := c.S3.DownloadFile(ctx, a.S3Key)
	if err != nil {
		return nil, nil, err
	}

	return &a, body, nil
}
