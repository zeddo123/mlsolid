//go:build integrationtests

package controllers_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zeddo123/mlsolid/solid/controllers"
	"github.com/zeddo123/mlsolid/solid/s3"
	"github.com/zeddo123/mlsolid/solid/store"
	"github.com/zeddo123/mlsolid/solid/types"
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

func TestArtifact(t *testing.T) {
	t.Run("content_of_an_artifact_is_saved_correctly", func(t *testing.T) {
		t.Parallel()

		controller := controllers.Controller{Redis: store.RedisStore{Client: *client}, S3: objectStore}

		run := types.NewRun("run_artifact", "artifact_exp")
		artifactContent := []byte{1, 2, 3}
		artifact := types.CheckpointArtifact{Model: "model_path.pt", Checkpoint: bytes.NewReader(artifactContent)}

		// Act
		err := controller.CreateRun(context.Background(), run)
		require.NoError(t, err)

		err = controller.AddArtifacts(context.Background(), "run_artifact", []types.Artifact{artifact})
		require.NoError(t, err)

		savedArtifact, content, err := controller.Artifact(context.Background(), "run_artifact", "model_path.pt")

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, content)
		assert.NotNil(t, savedArtifact)

		defer content.Close()
		b, err := io.ReadAll(content)

		require.NoError(t, err)
		assert.Equal(t, artifactContent, b)
		assert.Equal(t, "model_path.pt", savedArtifact.Name)
		assert.Equal(t, types.ModelContentType, savedArtifact.ContentType)
		assert.NotZero(t, savedArtifact.S3Key)
	})

	t.Run("unknown_artifact_returns_an_error", func(t *testing.T) {
		t.Parallel()

		controller := controllers.Controller{Redis: store.RedisStore{Client: *client}, S3: objectStore}

		// Act
		savedArtifact, content, err := controller.Artifact(context.Background(), "random_run_id", "unknown_artifact_name")

		// Assert
		t.Log(err)
		assert.True(t, errors.Is(err, types.ErrNotFound))
		assert.Nil(t, savedArtifact)
		assert.Nil(t, content)
	})
}

func TestModelRegistryFlow(t *testing.T) {
	t.Run("create_and_pull_model_from_registry", func(t *testing.T) {
		const dockerImage string = "ghcr.io/zeddo123/bench-dummy:0.0.4"

		t.Parallel()
		// Arrange
		controller := controllers.Controller{Redis: store.RedisStore{Client: *client}, S3: objectStore}

		run := types.NewRun("run2", "exp2")
		checkpoint := make([]byte, 3024)
		artifact := types.CheckpointArtifact{Model: "model_path.pt", Checkpoint: bytes.NewReader(checkpoint)}

		// Act
		err := controller.CreateRun(t.Context(), run)
		require.NoError(t, err)

		err = controller.AddArtifacts(t.Context(), "run2", []types.Artifact{artifact})
		require.NoError(t, err)

		err = controller.CreateModelRegistry(t.Context(), "exp2-registry", types.RegistryBenchmarkOps{
			BenchmarkImage:          dockerImage,
			BenchmarkGpuPassthrough: true,
		})
		require.NoError(t, err)

		err = controller.AddArtifactToRegistry(t.Context(), "exp2-registry", "run2", "model_path.pt", "prod")
		require.NoError(t, err)

		// Assert
		_, err = controller.TaggedModel(t.Context(), "exp2-registry", "prod")
		require.NoError(t, err)

		_, err = controller.LastModelEntry(t.Context(), "exp2-registry")
		require.NoError(t, err)

		ids, err := controller.ModelRegistriesID(t.Context())
		t.Log(ids)
		require.NoError(t, err)
		assert.Contains(t, ids, "exp2-registry")

		registry, err := controller.ModelRegistry(t.Context(), "exp2-registry")
		require.NoError(t, err)
		t.Log(registry.LastModel())
		assert.Equal(t, "run2", registry.LastModel().Run)
		assert.Equal(t, "model_path.pt", registry.LastModel().Name)
		assert.Equal(t, 1, registry.LastModel().Version)
		assert.Equal(t, true, registry.BenchmarkGpuPassthrough)
		assert.Equal(t, dockerImage, registry.BenchmarkImage)
	})
}

// findRun returns the run matching registry+version from a slice pulled from the store.
func findRun(runs []*types.BenchRun, registry string, version int64) *types.BenchRun {
	for _, r := range runs {
		if r.Registry == registry && r.Version == version {
			return r
		}
	}

	return nil
}

func TestRecordRuns(t *testing.T) {
	t.Parallel()

	controller := controllers.Controller{Redis: store.RedisStore{Client: *client}, S3: objectStore}

	bench := types.Bench{ //nolint: exhaustruct
		Name:        "record-runs-bench",
		Registries:  []string{"dummy-registry", "shadow-registry"},
		Metrics:     []types.BenchMetric{{Name: "mae"}, {Name: "loss"}},
		DatasetName: "dummy-dataset",
		DatasetURL:  "https://example.com/dataset.zip",
		Timestamp:   time.Now(),
	}

	benchID, created, err := controller.CreateBenchmark(t.Context(), bench)
	require.NoError(t, err)
	assert.True(t, created)

	t.Run("multiple_runs_across_registries_and_versions_are_all_recorded", func(t *testing.T) {
		runs := []types.BenchRun{
			{
				Registry: "dummy-registry", Version: 1,
				Metrics:   map[string]float32{"mae": 92.0, "loss": 0.0023},
				Timestamp: time.Now(), Start: time.Now(), End: time.Now(),
			},
			{
				Registry: "dummy-registry", Version: 2,
				Metrics:   map[string]float32{"mae": 88.5, "loss": 0.0011},
				Timestamp: time.Now(), Start: time.Now(), End: time.Now(),
			},
			{
				Registry: "shadow-registry", Version: 1,
				Metrics:   map[string]float32{"mae": 100.2, "loss": 0.05},
				Timestamp: time.Now(), Start: time.Now(), End: time.Now(),
			},
		}

		err := controller.RecordRuns(t.Context(), benchID, runs)
		require.NoError(t, err)

		got, err := controller.BenchmarkRuns(t.Context(), benchID)
		require.NoError(t, err)

		for _, want := range runs {
			run := findRun(got, want.Registry, want.Version)
			require.NotNilf(t, run, "missing run %s:%d", want.Registry, want.Version)
			assert.InDeltaf(t, want.Metrics["mae"], run.Metrics["mae"], 1e-3, "%s:%d mae", want.Registry, want.Version)
			assert.InDeltaf(t, want.Metrics["loss"], run.Metrics["loss"], 1e-4, "%s:%d loss", want.Registry, want.Version)
		}
	})

	t.Run("metrics_not_declared_on_the_benchmark_are_still_recorded", func(t *testing.T) {
		// RecordRuns accepts whatever metrics a run reports, even ones not
		// (yet) part of the benchmark's declared Metrics list ("mae", "loss"
		// here). That way a run already carries a value once a matching
		// metric is later added to the benchmark.
		const registry, version = "undeclared-metric-registry", 1

		err := controller.RecordRuns(t.Context(), benchID, []types.BenchRun{{
			Registry:  registry,
			Version:   version,
			Metrics:   map[string]float32{"gpu_mem_gb": 14.2},
			Timestamp: time.Now(), Start: time.Now(), End: time.Now(),
		}})
		require.NoError(t, err)

		got, err := controller.BenchmarkRuns(t.Context(), benchID)
		require.NoError(t, err)

		run := findRun(got, registry, version)
		require.NotNil(t, run)
		assert.InDelta(t, 14.2, run.Metrics["gpu_mem_gb"], 1e-6)
	})

	t.Run("recording_runs_for_an_unknown_benchmark_returns_not_found", func(t *testing.T) {
		err := controller.RecordRuns(t.Context(), "unknown-benchmark-id", []types.BenchRun{{
			Registry:  "dummy-registry",
			Version:   1,
			Metrics:   map[string]float32{"mae": 1},
			Timestamp: time.Now(), Start: time.Now(), End: time.Now(),
		}})
		require.Error(t, err)
		assert.True(t, errors.Is(err, types.ErrNotFound))
	})

	t.Run("re-recording_same_registry_and_version_replaces_previous_metrics", func(t *testing.T) {
		// Special case: RecordRuns clears any previous run sharing the same
		// registry+version before writing, so a metric dropped between two
		// runs of the same version disappears instead of leaving a stale value.
		const registry, version = "merge-registry", 1

		err := controller.RecordRuns(t.Context(), benchID, []types.BenchRun{{
			Registry:  registry,
			Version:   version,
			Metrics:   map[string]float32{"mae": 10, "loss": 20, "extra": 30},
			Timestamp: time.Now(), Start: time.Now(), End: time.Now(),
		}})
		require.NoError(t, err)

		err = controller.RecordRuns(t.Context(), benchID, []types.BenchRun{{
			Registry:  registry,
			Version:   version,
			Metrics:   map[string]float32{"mae": 99},
			Timestamp: time.Now(), Start: time.Now(), End: time.Now(),
		}})
		require.NoError(t, err)

		got, err := controller.BenchmarkRuns(t.Context(), benchID)
		require.NoError(t, err)

		run := findRun(got, registry, version)
		require.NotNil(t, run)

		assert.InDelta(t, 99, run.Metrics["mae"], 1e-6)
		// fields from the first write no longer leak into the second.
		assert.NotContains(t, run.Metrics, "loss")
		assert.NotContains(t, run.Metrics, "extra")
	})

	t.Run("metric_names_are_sanitized", func(t *testing.T) {
		const registry, version = "sanitize-registry", 1

		err := controller.RecordRuns(t.Context(), benchID, []types.BenchRun{{
			Registry:  registry,
			Version:   version,
			Metrics:   map[string]float32{"  Mean Absolute Error ": 1.23},
			Timestamp: time.Now(), Start: time.Now(), End: time.Now(),
		}})
		require.NoError(t, err)

		got, err := controller.BenchmarkRuns(t.Context(), benchID)
		require.NoError(t, err)

		run := findRun(got, registry, version)
		require.NotNil(t, run)
		assert.InDelta(t, 1.23, run.Metrics["mean-absolute-error"], 1e-6)
	})

	t.Run("run_with_no_metrics_is_recorded_without_error", func(t *testing.T) {
		// Special case: an empty Metrics map used to turn into an HSET with
		// zero field/value pairs, which Redis rejects. RecordRuns now skips
		// the metrics HSET entirely when there are none, so the run is
		// still recorded (with an empty metric set) and no error is returned.
		const registry, version = "empty-metrics-registry", 1

		err := controller.RecordRuns(t.Context(), benchID, []types.BenchRun{{
			Registry:  registry,
			Version:   version,
			Metrics:   map[string]float32{},
			Timestamp: time.Now(), Start: time.Now(), End: time.Now(),
		}})
		require.NoError(t, err)

		got, err := controller.BenchmarkRuns(t.Context(), benchID)
		require.NoError(t, err)

		run := findRun(got, registry, version)
		require.NotNilf(t, run, "run metadata should be recorded even with no metrics")
		assert.Empty(t, run.Metrics)
	})
}

func TestExpInfo(t *testing.T) {
	t.Run("add_description_to_exp", func(t *testing.T) {
		t.Parallel()

		// Arrange
		expID := "exp_info_test"
		desc := "some_new_description"

		controller := controllers.Controller{Redis: store.RedisStore{Client: *client}, S3: objectStore}
		run := types.NewRun("exp_info_run", expID)

		err := controller.CreateRun(t.Context(), run)
		require.NoError(t, err)

		info, err := controller.ExpInfo(t.Context(), expID)
		require.NoError(t, err)
		require.Zero(t, info)

		// Act
		err = controller.SetExpInfo(t.Context(), expID, types.ExperimentInfo{
			Description: desc,
		})
		require.NoError(t, err)

		info, err = controller.ExpInfo(t.Context(), expID)
		require.NoError(t, err)
		assert.Equal(t, desc, info.Description)
	})
}
