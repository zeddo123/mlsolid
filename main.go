// Package main is the entrypoint for running the service.
package main

import (
	"context"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/zeddo123/mlsolid/solid"
	"github.com/zeddo123/mlsolid/solid/api"
	"github.com/zeddo123/mlsolid/solid/bengine"
	"github.com/zeddo123/mlsolid/solid/controllers"
	"github.com/zeddo123/mlsolid/solid/grpcservice"
	"github.com/zeddo123/mlsolid/solid/logger"
	"github.com/zeddo123/mlsolid/solid/oauth"
	"github.com/zeddo123/mlsolid/solid/s3"
	"github.com/zeddo123/mlsolid/solid/store"
	"github.com/zeddo123/pubgo"
)

const (
	BengineBufferSize = 100
	BusTopicsCap      = 10
	BusSubsCap        = 2
)

var log zerolog.Logger //nolint: gochecknoglobals

func main() {
	config, err := solid.LoadConfig(".")
	if err != nil {
		panic(err)
	}

	log = logger.New(!config.Prod)

	log.Info().
		Str("APIPort", config.APIPort).
		Str("gRPCPort", config.GrpcPort).
		Bool("prod", config.Prod).
		Bool("api_access", config.APIKeyAccess).
		Bool("bEngine", config.EnableBEngine).
		Msg("configuration loaded successfully")

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

	store := store.RedisStore{
		Client: *redisClient,
		Logger: logger.NewSub(log, "store"),
	}
	controller := controllers.Controller{
		Redis:              store,
		S3:                 objectStore,
		Bus:                bus,
		Logger:             logger.NewSub(log, "controller"),
		PublishBenchEvents: config.EnableBEngine,
	}

	log.Info().Msg("starting servers")

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

	go grpcservice.StartServer(config.GrpcPort, &controller, config.APIKeyAccess, logger.NewSub(log, "gRPC"))
	go api.StartServer(config.APIPort, &controller, oauth.Config{
		Prod:                 config.Prod,
		RootURL:              config.RootURL,
		GoogleClientID:       config.GoogleClientID,
		GoogleClientSecret:   config.GoogleSecretID,
		GoogleAllowedDomains: config.GoogleAllowedDomains,
		FrontendURL:          config.FrontendURL,
	})

	select {}
}
