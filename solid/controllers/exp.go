package controllers

import (
	"context"
	"fmt"

	"github.com/zeddo123/mlsolid/solid/types"
)

func (c *Controller) Exp(ctx context.Context, expID string) (*types.Experiment, error) {
	return c.Redis.Exp(ctx, types.NormalizeID(expID))
}

func (c *Controller) ExpRuns(ctx context.Context, expID string) ([]string, error) {
	return c.Redis.ExpRunIDs(ctx, expID)
}

func (c *Controller) RunsFromExp(ctx context.Context, expID string) ([]*types.Run, error) {
	runs, err := c.ExpRuns(ctx, expID)
	if err != nil {
		return nil, err
	}

	return c.Runs(ctx, runs)
}

func (c *Controller) Exps(ctx context.Context) ([]string, error) {
	return c.Redis.Exps(ctx)
}

// ExpInfo returns info data linked to an experiment (description, etc).
func (c *Controller) ExpInfo(ctx context.Context, expID string) (types.ExperimentInfo, error) {
	info, err := c.Redis.ExpInfo(ctx, expID)
	if err != nil {
		return info, fmt.Errorf("could not pull ExpInfo: %w", err)
	}

	return info, nil
}

// SetExpInfo updates an experiment's info data.
func (c *Controller) SetExpInfo(ctx context.Context, expID string,
	info types.ExperimentInfo,
) error {
	err := c.Redis.SetExpInfo(ctx, expID, info)
	if err != nil {
		return fmt.Errorf("could not set ExpInfo: %w", err)
	}

	return nil
}
