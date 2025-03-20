package controllers

import (
	"github.com/zedd123/mlsolid/solid/s3"
	"github.com/zedd123/mlsolid/solid/store"
)

type Controller struct {
	Redis store.RedisStore
	S3    s3.ObjectStore
}
