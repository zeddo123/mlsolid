package controllers

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/zedd123/mlsolid/solid/types"
)

func (c *Controller) CreateRun(ctx context.Context, run types.Run) error {
	ok, err := c.Redis.RunExists(ctx, run.Name)
	if err != nil {
		return err
	}

	if ok {
		return types.NewBadRequest(fmt.Sprintf("run id <%s> already in use", run.Name))
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
	ids := types.ArtifactIds(as)
	artifactsMap := types.ArtifactIdMap(as)

	errs := c.Redis.ArtifactsExist(ctx, runID, ids)

	to_upload := make([]types.Artifact, 0, len(as))

	for id, err := range errs {
		if errors.Is(err, types.ErrNotFound) {
			to_upload = append(to_upload, artifactsMap[id])
		} else if err != nil {
			return err
		}
	}

	artifacts, uploadErr := c.S3.UploadArtifacts(ctx, to_upload)

	err := c.Redis.SetArtifacts(ctx, runID, artifacts)
	if err != nil {
		return err
	}

	return uploadErr
}
