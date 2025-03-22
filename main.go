package main

import (
	"fmt"
	"log"
	"net"

	"github.com/redis/go-redis/v9"
	"github.com/zeddo123/mlsolid/solid"
	"github.com/zeddo123/mlsolid/solid/controllers"
	mlsolidv1 "github.com/zeddo123/mlsolid/solid/gen/mlsolid/v1"
	"github.com/zeddo123/mlsolid/solid/grpcservice"
	"github.com/zeddo123/mlsolid/solid/store"
	"google.golang.org/grpc"
)

func main() {
	config, err := solid.LoadConfig(".")
	if err != nil {
		panic(err)
	}

	fmt.Println("mlsolid", config)

	redisClient := redis.NewClient(&redis.Options{
		Addr:     config.RedisAddr,
		Password: config.RedisPassword,
		DB:       config.RedisDB,
	})

	service := grpcservice.Service{
		Controller: controllers.Controller{
			Redis: store.RedisStore{Client: *redisClient},
		},
	}

	l, err := net.Listen("tcp", ":"+config.GrpcPort)
	if err != nil {
		log.Println("could not listen to port", config.GrpcPort)

		panic(err)
	}

	server := grpc.NewServer()

	mlsolidv1.RegisterMlsolidServer(server, &service)

	log.Println("gRPC server started at", config.GrpcPort)

	if err := server.Serve(l); err != nil {
		panic(err)
	}
}
