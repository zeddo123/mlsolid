package main

import (
	"context"
	"log"

	"buf.build/gen/go/zeddo123/mlsolid/grpc/go/mlsolid/v1/mlsolidv1grpc"
	mlsolidv1 "buf.build/gen/go/zeddo123/mlsolid/protocolbuffers/go/mlsolid/v1"
)

func createModelRegistry(client mlsolidv1grpc.MlsolidServiceClient, registryName string) {
	_, err := client.CreateModelRegistry(context.Background(), &mlsolidv1.CreateModelRegistryRequest{
		Name: registryName,
	})
	if err != nil {
		panic(err)
	}

	log.Printf("[populate] Model Registry %s created!\n", registryName)
}
