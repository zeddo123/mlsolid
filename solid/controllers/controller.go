package controllers

import (
	"context"
	"fmt"

	"github.com/zeddo123/mlsolid/solid"
	"github.com/zeddo123/mlsolid/solid/s3"
	"github.com/zeddo123/mlsolid/solid/store"
	"github.com/zeddo123/mlsolid/solid/types"
)

type Controller struct {
	Redis  store.RedisStore
	S3     s3.ObjectStore
	Config solid.Config
	APIKey string
}

func (c *Controller) Permissions() (types.Permissions, error) {
	return c.GetPermissions(context.Background(), c.APIKey)
}

func (c *Controller) HasPermission(perm types.PermissionType) error {
	if !c.Config.APIKeyAccess {
		return nil
	}

	perms, err := c.Permissions()
	if err != nil {
		return err
	}

	if !perms.HasPermission(perm) {
		return fmt.Errorf("unauthorized")
	}

	return nil
}
