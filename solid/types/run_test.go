package types_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zeddo123/mlsolid/solid/types"
)

func TestRun(t *testing.T) {
	r := types.NewRun("Linear regression 1", "linreg")

	m := &types.GenericMetric[float64]{
		Key:    "mse",
		Values: []float64{50, 30.34, 20.34, 10.34, 7.03, 1.01},
	}

	err := r.AddMetric("mse", m)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "linear-regression-1", r.Name)
	assert.Equal(t, r.Metrics["mse"], m)
	assert.Equal(t, "linreg", r.ExperimentID)
}
