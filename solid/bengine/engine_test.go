package bengine_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zeddo123/mlsolid/solid/bengine"
	"github.com/zeddo123/mlsolid/solid/types"
	"github.com/zeddo123/pubgo"
)

const DatasetURL = "https://www.kaggle.com/api/v1/datasets/download/gpreda/chinese-mnist"

func TestEngineRun(t *testing.T) {
	t.Parallel()

	topic := "bench"
	bus := pubgo.NewBusWithContext(t.Context(), pubgo.DefaultOps())

	sub := bus.Subscribe(topic)

	ctx, cancel := context.WithCancel(context.Background())
	e := bengine.New(ctx, nil, sub, "", "")
	require.NotNil(t, e)

	err := bus.Publish(topic, &types.BenchEvent{ //nolint: exhaustruct
		DockerImage: "ghcr.io/zeddo123/bench-dummy:0.0.2",
		DatasetName: "chinese-mnist",
		DatasetURL:  DatasetURL,
	})
	require.NoError(t, err)

	time.Sleep(70 * time.Second)
	sub.Done()

	cancel()
}

func TestPullDataset(t *testing.T) {
	t.Parallel()

	path := t.TempDir()

	t.Log(path)

	err := bengine.PullDataset(t.Context(), DatasetURL, path)
	require.NoError(t, err)

	assert.DirExists(t, path)
}
