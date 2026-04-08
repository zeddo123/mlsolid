// Package main
package main

import (
	"context"
	"log"
	"math/rand"
	"os"

	"buf.build/gen/go/zeddo123/mlsolid/grpc/go/mlsolid/v1/mlsolidv1grpc"
	"github.com/anandvarma/namegen"
	"github.com/urfave/cli/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	var url string

	cmd := cli.Command{ //nolint: exhaustruct
		Name:  "populate",
		Usage: "populate a mlsolid service with test data",
		Flags: []cli.Flag{
			&cli.StringFlag{ //nolint: exhaustruct
				Name:        "url",
				Usage:       "mlsolid grpc url",
				Destination: &url,
				Required:    true,
			},
		},
		Action: func(_ context.Context, _ *cli.Command) error {
			log.Println("[populate] connecting to gRPC service...")

			conn, err := grpc.NewClient(url, grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				panic(err)
			}

			client := mlsolidv1grpc.NewMlsolidServiceClient(conn)

			createModelRegistry(client, "Yolo Prod")

			maxExps := 3
			maxRuns := 12

			for range max(1, rand.Intn(maxExps)) { //nolint: gosec
				exp := namegen.New().Get()

				for range rand.Intn(maxRuns) { //nolint: gosec
					createRun(client, exp) //nolint: contextcheck
				}
			}

			return nil
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		panic(err)
	}
}
