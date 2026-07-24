package main

import (
	"context"
	"log"

	"buf.build/gen/go/zeddo123/mlsolid/grpc/go/mlsolid/v1/mlsolidv1grpc"
	mlsolidv1 "buf.build/gen/go/zeddo123/mlsolid/protocolbuffers/go/mlsolid/v1"
	"github.com/anandvarma/namegen"
)

// DummyImage's entrypoint sleeps for 20s and always writes a fixed set of
// metrics to its output file, see bench-container-example/entrypoint.py.
const DummyImage = "ghcr.io/zeddo123/bench-dummy:0.0.4"

const (
	benchmarkDatasetName = "chinese-mnist"
	benchmarkDatasetURL  = "https://www.kaggle.com/api/v1/datasets/download/gpreda/chinese-mnist"
)

// createBenchmark creates a benchmark watching registryName and points the
// registry's docker image at DummyImage. Once wired up this way, every
// subsequent model entry added to registryName (e.g. via createRun's
// AddModelEntry call) makes the server publish a real bEngine event that
// pulls and runs the dummy benchmark container.
func createBenchmark(client mlsolidv1grpc.MlsolidServiceClient, registryName string) string {
	image := DummyImage

	_, err := client.SetRegistryBenchmarkOps(context.Background(), &mlsolidv1.SetRegistryBenchmarkOpsRequest{
		Name:           registryName,
		BenchmarkImage: &image,
	})
	if err != nil {
		panic(err)
	}

	log.Printf("[populate]: set benchmark image registry=%s image=%s \n", registryName, image)

	resp, err := client.CreateBenchmark(context.Background(), &mlsolidv1.CreateBenchmarkRequest{ //nolint: exhaustruct
		Name:            namegen.New().Get(),
		ModelRegistries: []string{registryName},
		Metrics: []*mlsolidv1.BenchmarkMetric{
			{Name: "mae"},  //nolint: exhaustruct
			{Name: "loss"}, //nolint: exhaustruct
		},
		DecisionMetric: "loss",
		DatasetName:    benchmarkDatasetName,
		DatasetUrl:     benchmarkDatasetURL,
		AutoTag:        true,
	})
	if err != nil {
		panic(err)
	}

	log.Printf("[populate]: benchmark created benchId=%s registry=%s \n", resp.GetBenchmarkId(), registryName)

	return resp.GetBenchmarkId()
}
