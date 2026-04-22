package controllers

import "context"

// IsValidAPIKey checks if an API Key is valid.
func (c *Controller) IsValidAPIKey(ctx context.Context, key string) (bool, error) {
	return c.Redis.IsValidAPIKey(ctx, key) //nolint: wrapcheck
}

// GenerateKey generates a new API Key.
func (c *Controller) GenerateKey(ctx context.Context) (string, error) {
	return c.Redis.GenerateKey(ctx) //nolint: wrapcheck
}
