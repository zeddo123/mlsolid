package types_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zeddo123/mlsolid/solid/types"
)

func TestRun(t *testing.T) {
	t.Parallel()

	r := types.NewRun("Linear regression 1", "linreg")

	m := &types.GenericMetric[float64]{
		Key:    "mse",
		Values: []float64{50, 30.34, 20.34, 10.34, 7.03, 1.01},
	}

	err := r.AddMetric("mse", m)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "linear-regression-1", r.Name)
	assert.Equal(t, r.Metrics["mse"], m)
	assert.Equal(t, "linreg", r.ExperimentID)
}

func TestRunColorGeneration(t *testing.T) {
	t.Parallel()

	r := types.NewRun("run_name", "exp23")

	t.Log(r.Color)
	assert.NotEmpty(t, r.Color)
	assert.Regexp(t, "^#[a-fA-F0-9]{6}", r.Color)
}
