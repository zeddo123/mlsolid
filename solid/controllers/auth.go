package controllers

import (
	"context"

	"github.com/zeddo123/mlsolid/solid/types"
)

func (c *Controller) CreateAPIKey(ctx context.Context, perm types.Permissions) (string, error) {
	return c.Redis.CreateAPIKey(ctx, perm)
}

func (c *Controller) GetPermissions(ctx context.Context, key string) (types.Permissions, error) {
	return c.Redis.APIKeyPermissions(ctx, key)
}
