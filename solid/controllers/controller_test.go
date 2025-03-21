//go:build integrationtests

package controllers_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zedd123/mlsolid/solid/controllers"
	"github.com/zedd123/mlsolid/solid/s3"
	"github.com/zedd123/mlsolid/solid/store"
	"github.com/zedd123/mlsolid/solid/types"
)

func TestRunFlow(t *testing.T) {
	controller := controllers.Controller{Redis: store.RedisStore{Client: *client}, S3: s3.MockObjectStore{}}

	run := types.NewRun("run1", "exp1")

	mse := types.NewGenericMetric[float32]("mse", 10)
	mse.Add(0.23)
	mse.Add(0.123)

	loss := types.NewGenericMetric[float32]("loss", 10)
	loss.Add(23.342)
	loss.Add(13.99)
	loss.Add(1.99)
	loss.Add(0)

	run.AddMetric("mse", mse)
	run.AddMetric("loss", loss)

	err1 := controller.CreateRun(context.Background(), run)

	acc := types.NewGenericMetric[float64]("acc", 1)
	acc.Add(0.92)

	model_size := types.NewGenericMetric[string]("model_size", 1)
	model_size.Add("huge")

	moreMse := types.NewGenericMetric[float32]("mse", 10)
	moreMse.Add(0.234)

	err2 := controller.AddMetrics(context.Background(), run.Name, []types.Metric{acc, model_size, moreMse})

	runs, err3 := controller.ExpRuns(context.Background(), "exp1")
	savedRun, err := controller.Run(context.Background(), "run1")

	t.Log(savedRun)

	require.NoError(t, err1)
	require.NoError(t, err2)
	require.NoError(t, err3)
	require.NoError(t, err)
	require.NotNil(t, savedRun)
	assert.Contains(t, runs, "run1")
	assert.Len(t, runs, 1)
	assert.Equal(t, savedRun.Name, run.Name)
	assert.Equal(t, savedRun.ExperimentID, run.ExperimentID)
	assert.InDelta(t, 0.234, savedRun.Metrics["mse"].LastVal(), 0.001)
	assert.Equal(t, 0.92, savedRun.Metrics["acc"].LastVal())
	assert.Equal(t, "huge", savedRun.Metrics["model_size"].LastVal())
}
