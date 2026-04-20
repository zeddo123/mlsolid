package main

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
	"github.com/zeddo123/mlsolid/solid"
	"github.com/zeddo123/mlsolid/solid/api"
	"github.com/zeddo123/mlsolid/solid/bengine"
	"github.com/zeddo123/mlsolid/solid/controllers"
	"github.com/zeddo123/mlsolid/solid/grpcservice"
	"github.com/zeddo123/mlsolid/solid/s3"
	"github.com/zeddo123/mlsolid/solid/store"
	"github.com/zeddo123/pubgo"
)

const (
	BengineBufferSize = 100
	BusTopicsCap      = 10
	BusSubsCap        = 2
)

func main() {
	config, err := solid.LoadConfig(".")
	if err != nil {
		panic(err)
	}

	log.Println("[MLSOLID] Configuration has been loaded")

	bus := pubgo.NewBus(pubgo.BusOps{
		InitialTopicsCap: BusTopicsCap,
		InitialSubsCap:   BusSubsCap,
		PublishStrat:     pubgo.NonBlockingPublish(),
	})

	redisClient := redis.NewClient(&redis.Options{ //nolint: exhaustruct
		Addr:     config.RedisAddr,
		Password: config.RedisPassword,
		DB:       config.RedisDB,
	})

	err = redisClient.Ping(context.Background()).Err()
	if err != nil {
		panic(err)
	}

	objectStore, err := s3.NewStore(s3.StoreOps{ //nolint: exhaustruct
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

	store := store.RedisStore{Client: *redisClient}
	controller := controllers.Controller{
		Redis: store,
		S3:    objectStore,
	}

	if config.EnableBEngine {
		sub := bus.Subscribe("bengine", pubgo.WithBufferSize(BengineBufferSize))

		engine := bengine.New(sub,
			bengine.WithRedisStore(&store),
			bengine.WithS3(objectStore),
			bengine.WithHostSourceVolume(config.HostSourceVolume),
			bengine.WithRegistryCreds(config.DockerRegistryUsername, config.DockerRegistryPassword),
		)

		go engine.Start(context.Background())
	}

	go grpcservice.StartServer(config.GrpcPort, &controller)
	go api.StartServer(config.APIPort, &controller)

	select {}
}
