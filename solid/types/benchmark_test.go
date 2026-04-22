package types_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zeddo123/mlsolid/solid/types"
)

func TestSanatizeName(t *testing.T) {
	t.Parallel()

	tt := []struct {
		Name string
		Out  string
	}{
		{
			Name: "benchmark Number 1",
			Out:  "benchmark-number-1",
		},
		{
			Name: "BENCH#2",
			Out:  "bench#2",
		},
		{
			Name: "BENCH       #2",
			Out:  "bench-#2",
		},
		{
			Name: "   TEST   BENCH       #2",
			Out:  "test-bench-#2",
		},
	}

	for _, tc := range tt {
		t.Run("", func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.Out, types.SanatizeName(tc.Name))
		})
	}
}

func TestBestRuns(t *testing.T) {
	t.Parallel()

	tt := []struct {
		runs    []*types.BenchRun
		results map[string]*types.BenchRun
	}{
		{
			runs: []*types.BenchRun{
				{
					Registry: "registry#1",
					Version:  2,
					Metrics:  nil,
				},
				{
					Registry: "registry#1",
					Version:  4,
					Metrics: map[string]float32{
						"loss": 32.3,
						"acc":  0.43,
						"mae":  0.87,
					},
				},
				{
					Registry: "registry#1",
					Version:  8,
					Metrics: map[string]float32{
						"loss": 50.3,
						"acc":  0.23,
						"mae":  0.37,
					},
				},
				{
					Registry: "registry#1",
					Version:  8,
					Metrics: map[string]float32{
						"loss": 50.3,
						"acc":  0.23,
						"mae":  0.99,
					},
				},
			},
			results: map[string]*types.BenchRun{
				"loss": {
					Registry: "registry#1",
					Version:  8,
					Metrics: map[string]float32{
						"loss": 50.3,
						"acc":  0.23,
						"mae":  0.37,
					},
				},
				"acc": {
					Registry: "registry#1",
					Version:  4,
					Metrics: map[string]float32{
						"loss": 32.3,
						"acc":  0.43,
						"mae":  0.87,
					},
				},
				"mae": {
					Registry: "registry#1",
					Version:  8,
					Metrics: map[string]float32{
						"loss": 50.3,
						"acc":  0.23,
						"mae":  0.99,
					},
				},
			},
		},
	}

	for _, tc := range tt {
		t.Run("", func(t *testing.T) {
			t.Parallel()

			result := types.BestRuns(tc.runs, "loss", "acc", "mae")
			t.Log(result)
			require.Len(t, result, 3)

			assert.Equal(t, result["loss"], tc.results["loss"])
			assert.Equal(t, result["acc"], tc.results["acc"])
			assert.Equal(t, result["mae"], tc.results["mae"])
		})
	}
}
