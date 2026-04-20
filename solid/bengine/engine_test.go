//go:build integrationtests

package bengine_test

import (
	"sync"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zeddo123/mlsolid/solid/bengine"
	"github.com/zeddo123/mlsolid/solid/types"
	"github.com/zeddo123/pubgo"
)

const DatasetURL = "https://www.kaggle.com/api/v1/datasets/download/gpreda/chinese-mnist"

func TestEngineRun(t *testing.T) {
	t.Parallel()

	var wg sync.WaitGroup

	topic := "bench"
	bus := pubgo.NewBusWithContext(t.Context(), pubgo.DefaultOps())

	sub := bus.Subscribe(topic)

	e := bengine.New(sub,
		bengine.WithRootDest("./.mlsolid/"),
		bengine.WithHumanReadableLogs(),
		bengine.WithLoggingLevel(zerolog.DebugLevel))

	wg.Go(func() {
		e.Start(t.Context())
	})

	require.NotNil(t, e)

	err := bus.Publish(topic, &types.BenchEvent{ //nolint: exhaustruct
		DockerImage: "ghcr.io/zeddo123/bench-dummy:0.0.4",
		DatasetName: "chinese-mnist",
		DatasetURL:  DatasetURL,
	})
	require.NoError(t, err)

	sub.Done()
	wg.Wait()
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
