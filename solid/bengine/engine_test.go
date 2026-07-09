//go:build integrationtests

package bengine_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zeddo123/mlsolid/solid/bengine"
	"github.com/zeddo123/mlsolid/solid/controllers"
	"github.com/zeddo123/mlsolid/solid/store"
	"github.com/zeddo123/mlsolid/solid/types"
	"github.com/zeddo123/pubgo"
)

const (
	DatasetURL = "https://www.kaggle.com/api/v1/datasets/download/gpreda/chinese-mnist"

	// DummyImage's entrypoint always writes a fixed set of metrics to its output file,
	// see bench-container-example/entrypoint.py.
	DummyImage = "ghcr.io/zeddo123/bench-dummy:0.0.4"

	expectedMAE  = 92.0
	expectedLoss = 0.0023
)

func TestEngineRun(t *testing.T) {
	t.Parallel()

	var wg sync.WaitGroup

	topic := "bench"
	bus := pubgo.NewBusWithContext(t.Context(), pubgo.DefaultOps())

	sub := bus.Subscribe(topic)

	controller := &controllers.Controller{Redis: store.RedisStore{Client: *client}} //nolint: exhaustruct

	benchID, _, err := controller.CreateBenchmark(t.Context(), types.Bench{ //nolint: exhaustruct
		Name:        "engine-run-bench",
		Registries:  []string{"dummy-registry"},
		Metrics:     []types.BenchMetric{{Name: "mae"}, {Name: "loss"}},
		DatasetName: "chinese-mnist",
		DatasetURL:  DatasetURL,
		Timestamp:   time.Now(),
	})
	require.NoError(t, err)

	e := bengine.New(sub,
		bengine.WithRootDest("./.mlsolid/"),
		bengine.WithHumanReadableLogs(),
		bengine.WithLoggingLevel(zerolog.DebugLevel),
		bengine.WithRunRecorder(controller))

	wg.Go(func() {
		e.Start(t.Context())
	})

	require.NotNil(t, e)

	event := &types.BenchEvent{ //nolint: exhaustruct
		BenchID:     benchID,
		Registry:    "dummy-registry",
		Version:     1,
		DockerImage: DummyImage,
		DatasetName: "chinese-mnist",
		DatasetURL:  DatasetURL,
	}

	err = bus.Publish(topic, event)
	require.NoError(t, err)

	sub.Done()
	wg.Wait()

	runs, err := controller.BenchmarkRuns(t.Context(), event.BenchID)
	require.NoError(t, err)
	require.Len(t, runs, 1)

	metrics := runs[0].Metrics
	assert.InDelta(t, expectedMAE, metrics["mae"], 1e-3)
	assert.InDelta(t, expectedLoss, metrics["loss"], 1e-4)
	assert.Equal(t, event.Registry, runs[0].Registry)
	assert.Equal(t, event.Version, runs[0].Version)
}

// TestRunContainerExtractsFixedMetrics drives RunContainer directly (bypassing
// pubsub/redis) to check that metrics reported by the dummy benchmark image
// are extracted correctly from its output file.
func TestRunContainerExtractsFixedMetrics(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	datasetPath := filepath.Join(root, "datasets", "dummy-dataset")
	require.NoError(t, os.MkdirAll(datasetPath, 0o755))

	engine := bengine.New(nil,
		bengine.WithRootDest(root),
		bengine.WithHumanReadableLogs(),
		bengine.WithLoggingLevel(zerolog.DebugLevel))

	result, err := engine.RunContainer(t.Context(), DummyImage, "dummy-dataset",
		datasetPath, filepath.Join(root, "checkpoints", "model.pth"))
	require.NoError(t, err)

	var metrics map[string]float32

	require.NoError(t, json.Unmarshal([]byte(result), &metrics))

	assert.InDelta(t, expectedMAE, metrics["mae"], 1e-3)
	assert.InDelta(t, expectedLoss, metrics["loss"], 1e-4)
}

func TestPullDataset(t *testing.T) {
	t.Parallel()

	path := t.TempDir()

	t.Log(path)

	engine := bengine.New(nil)

	err := engine.PullDataset(t.Context(), DatasetURL, path, false)
	require.NoError(t, err)

	assert.DirExists(t, path)
}
