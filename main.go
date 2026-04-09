package main

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
	"github.com/zeddo123/mlsolid/solid"
	"github.com/zeddo123/mlsolid/solid/api"
	"github.com/zeddo123/mlsolid/solid/controllers"
	"github.com/zeddo123/mlsolid/solid/grpcservice"
	"github.com/zeddo123/mlsolid/solid/s3"
	"github.com/zeddo123/mlsolid/solid/store"
)

func main() {
	config, err := solid.LoadConfig(".")
	if err != nil {
		panic(err)
	}

	log.Println("[MLSOLID] Configuration has been loaded")

	redisClient := redis.NewClient(&redis.Options{
		Addr:     config.RedisAddr,
		Password: config.RedisPassword,
		DB:       config.RedisDB,
	})

	err = redisClient.Ping(context.Background()).Err()
	if err != nil {
		panic(err)
	}

	objectStore, err := s3.NewStore(s3.StoreOps{
		Bucket:          config.S3Bucket,
		Endpoint:        config.S3Endpoint,
		AccessKey:       config.S3Key,
		SecretAccessKey: config.S3Secret,
		Region:          config.S3Region,
		Prefix:          config.S3Prefix,
	})
	if err != nil {
		panic(err)
	}

	controller := controllers.Controller{
		Redis: store.RedisStore{Client: *redisClient},
		S3:    objectStore,
	}

	go grpcservice.StartServer(config.GrpcPort, &controller)
	go api.StartServer(config.APIPort, &controller)

	select {}
}
