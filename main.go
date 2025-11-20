package main

import (
	"context"
	"fmt"
	"log"
	"net"

	mlsolidv1grpc "buf.build/gen/go/zeddo123/mlsolid/grpc/go/mlsolid/v1/mlsolidv1grpc"
	"github.com/redis/go-redis/v9"
	"github.com/zeddo123/mlsolid/solid"
	"github.com/zeddo123/mlsolid/solid/controllers"
	"github.com/zeddo123/mlsolid/solid/grpcservice"
	"github.com/zeddo123/mlsolid/solid/s3"
	"github.com/zeddo123/mlsolid/solid/store"
	"google.golang.org/grpc"
)

func main() {
	config, err := solid.LoadConfig(".")
	if err != nil {
		panic(err)
	}

	fmt.Println("loaded config", config)

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

	service := grpcservice.Service{
		Controller: controllers.Controller{
			Redis: store.RedisStore{Client: *redisClient},
			S3:    objectStore,
		},
	}

	l, err := net.Listen("tcp", ":"+config.GrpcPort)
	if err != nil {
		log.Println("could not listen to port", config.GrpcPort)

		panic(err)
	}

	server := grpc.NewServer()

	mlsolidv1grpc.RegisterMlsolidServiceServer(server, &service)

	log.Println("gRPC server started at", config.GrpcPort)

	if err := server.Serve(l); err != nil {
		panic(err)
	}
}
