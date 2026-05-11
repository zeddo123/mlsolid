// Package main
package main

import (
	"context"
	"log"
	"os"

	"buf.build/gen/go/zeddo123/mlsolid/grpc/go/mlsolid/v1/mlsolidv1grpc"
	"github.com/urfave/cli/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const DefaultWorkers = 5

func main() {
	var url string

	var workers int

	var client mlsolidv1grpc.MlsolidServiceClient

	cmd := cli.Command{ //nolint: exhaustruct
		Name:  "stress",
		Usage: "stress tests an mlsolid service",
		Flags: []cli.Flag{
			&cli.StringFlag{ //nolint: exhaustruct
				Name:        "url",
				Usage:       "mlsolid grpc url",
				Required:    true,
				Destination: &url,
			},
			&cli.IntFlag{ //nolint: exhaustruct
				Name:        "workers",
				Usage:       "number of workers",
				Value:       DefaultWorkers,
				Destination: &workers,
			},
		},
		Before: func(ctx context.Context, _ *cli.Command) (context.Context, error) {
			conn, err := grpc.NewClient(url, grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				panic(err)
			}

			client = mlsolidv1grpc.NewMlsolidServiceClient(conn)

			return ctx, nil
		},
		Commands: []*cli.Command{
			{
				Name:  "artifact",
				Usage: "stress test adding artifacts",
				Flags: []cli.Flag{
					&cli.StringFlag{ //nolint: exhaustruct
						Name:  "artifact",
						Usage: "artifact url source",
						Value: "https://huggingface.co/Ultralytics/YOLO11/resolve/d3043e98a1ad0e2956728c13cf1e041e0fa4220f/yolo11s.pt",
					},
				},
				Action: func(ctx context.Context, c *cli.Command) error {
					addArtifacts(ctx, client, workers, c.String("artifact"))

					return nil
				},
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
