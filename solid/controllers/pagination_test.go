//go:build integrationtests

package controllers_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zeddo123/mlsolid/solid/controllers"
	"github.com/zeddo123/mlsolid/solid/store"
	"github.com/zeddo123/mlsolid/solid/types"
)

const paginationPageSize = 5

func TestBenchmarksPage(t *testing.T) {
	t.Parallel()

	controller := controllers.Controller{Redis: store.RedisStore{Client: *client}}

	want := make(map[string]bool, 23)

	for range 23 {
		id, _, err := controller.CreateBenchmark(t.Context(), types.Bench{
			Name:        "pagination-bench-" + uuid.NewString(),
			Registries:  []string{"some-registry"},
			Metrics:     []types.BenchMetric{{Name: "metric", DescSort: false}},
			DatasetName: "dataset",
			DatasetURL:  "https://example.com/dataset",
			Timestamp:   time.Now(),
		})
		require.NoError(t, err)

		want[id] = false
	}

	var (
		cursor    uint64
		iteration int
	)

	for {
		iteration++
		require.Less(t, iteration, 10000, "pagination did not terminate")

		page, next, err := controller.BenchmarksPage(t.Context(), cursor, paginationPageSize)
		require.NoError(t, err)

		for id := range page {
			if _, ok := want[id]; ok {
				want[id] = true
			}
		}

		if next == 0 {
			break
		}

		cursor = next
	}

	for id, seen := range want {
		assert.True(t, seen, "benchmark %s was not returned while paginating", id)
	}
}

func TestModelRegistriesIDPage(t *testing.T) {
	t.Parallel()

	controller := controllers.Controller{Redis: store.RedisStore{Client: *client}}

	want := make(map[string]bool, 23)

	for range 23 {
		name := "pagination-registry-" + uuid.NewString()

		err := controller.CreateModelRegistry(t.Context(), name, types.RegistryBenchmarkOps{})
		require.NoError(t, err)

		want[name] = false
	}

	var (
		cursor    uint64
		iteration int
	)

	for {
		iteration++
		require.Less(t, iteration, 10000, "pagination did not terminate")

		page, next, err := controller.ModelRegistriesIDPage(t.Context(), cursor, paginationPageSize)
		require.NoError(t, err)

		for _, id := range page {
			if _, ok := want[id]; ok {
				want[id] = true
			}
		}

		if next == 0 {
			break
		}

		cursor = next
	}

	for id, seen := range want {
		assert.True(t, seen, "registry %s was not returned while paginating", id)
	}
}

func TestExpsPage(t *testing.T) {
	t.Parallel()

	controller := controllers.Controller{Redis: store.RedisStore{Client: *client}}

	want := make(map[string]bool, 23)

	for range 23 {
		expID := "pagination-exp-" + uuid.NewString()
		run := types.NewRun("pagination-run-"+uuid.NewString(), expID)

		err := controller.CreateRun(t.Context(), run)
		require.NoError(t, err)

		want[expID] = false
	}

	var (
		cursor    uint64
		iteration int
	)

	for {
		iteration++
		require.Less(t, iteration, 10000, "pagination did not terminate")

		page, next, err := controller.ExpsPage(t.Context(), cursor, paginationPageSize)
		require.NoError(t, err)

		for _, id := range page {
			if _, ok := want[id]; ok {
				want[id] = true
			}
		}

		if next == 0 {
			break
		}

		cursor = next
	}

	for id, seen := range want {
		assert.True(t, seen, "experiment %s was not returned while paginating", id)
	}
}
