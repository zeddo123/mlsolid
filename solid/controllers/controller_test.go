//go:build integrationtests

package controllers_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zedd123/mlsolid/solid/controllers"
	"github.com/zedd123/mlsolid/solid/store"
	"github.com/zedd123/mlsolid/solid/types"
)

func TestRunFlow(t *testing.T) {
	controller := controllers.Controller{store.RedisStore{Client: *client}}

	run := types.NewRun("run1", "exp1")

	run.AddMetric("mse", &types.GenericMetric[float32]{Key: "mse", Values: []float32{0.23, 0.34}})
	run.AddMetric("loss", &types.GenericMetric[float32]{Key: "loss", Values: []float32{20, 10.5, 8.990}})

	err1 := controller.CreateRun(context.Background(), run)

	m1 := &types.GenericMetric[float32]{Key: "acc", Values: []float32{0.92}}
	m2 := &types.GenericMetric[string]{Key: "model_size", Values: []string{"huge"}}
	m3 := &types.GenericMetric[float32]{Key: "mse", Values: []float32{0.234}}

	err2 := controller.AddMetrics(context.Background(), run.Name, []types.Metric{m1, m2, m3})

	runs, err3 := controller.ExpRuns(context.Background(), "exp1")
	savedRun, err := controller.Run(context.Background(), "run1")

	require.NoError(t, err1)
	require.NoError(t, err2)
	require.NoError(t, err3)
	require.NoError(t, err)
	assert.Contains(t, runs, "run1")
	assert.Len(t, runs, 1)
	assert.Equal(t, savedRun.Name, run.Name)
	assert.Equal(t, savedRun.ExperimentID, run.ExperimentID)
	assert.Equal(t, savedRun.Metrics["mse"].LastVal(), 0.234)
	assert.Equal(t, savedRun.Metrics["acc"].LastVal(), 0.92)
	assert.Equal(t, savedRun.Metrics["model_size"].LastVal(), "huge")
}
