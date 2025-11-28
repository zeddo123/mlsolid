package controllers

import (
	"context"

	"github.com/zeddo123/mlsolid/solid/types"
)

func (c *Controller) Exp(ctx context.Context, expID string) (*types.Experiment, error) {
	if err := c.HasPermission(types.PushExperimentsPermission); err != nil {
		return nil, err
	}

	return c.Redis.Exp(ctx, types.NormalizeID(expID))
}

func (c *Controller) ExpRuns(ctx context.Context, expID string) ([]string, error) {
	if err := c.HasPermission(types.PushExperimentsPermission); err != nil {
		return nil, err
	}

	return c.Redis.ExpRunIDs(ctx, expID)
}

func (c *Controller) Exps(ctx context.Context) ([]string, error) {
	if err := c.HasPermission(types.PushExperimentsPermission); err != nil {
		return nil, err
	}

	return c.Redis.Exps(ctx)
}
