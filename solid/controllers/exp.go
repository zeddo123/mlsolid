package controllers

import (
	"context"

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
