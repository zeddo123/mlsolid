package types_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zeddo123/mlsolid/solid/types"
)

func TestGenericMetric(t *testing.T) {
	t.Run("only_new_values_are_uncommitted", func(t *testing.T) {
		m := types.GenericMetric[string]{
			Key:    "paths",
			Values: []string{"path1", "path2"},
		}

		m.Add("path3")
		m.Add("path4")

		assert.Len(t, m.UnCommited(), 2)
		assert.Contains(t, m.UnCommited(), "path3")
		assert.Contains(t, m.UnCommited(), "path4")
		assert.Equal(t, "path2", m.LastVal())
	})

	t.Run("no_uncommitted_values_after_committing_them", func(t *testing.T) {
		m := types.GenericMetric[string]{
			Key:    "paths",
			Values: []string{"path1", "path2"},
		}

		m.Add("path3")
		m.Add("path4")
		m.Commit()

		assert.Empty(t, m.UnCommited())
		assert.Equal(t, "path4", m.LastVal())
	})

	t.Run("new_values_can_be_added_after_committing", func(t *testing.T) {
		m := types.GenericMetric[float64]{
			Key:    "paths",
			Values: []float64{0.23, 0.24},
		}

		m.Add(0.5)
		m.Add(0.6)
		m.Commit()
		m.Add(0.7)

		assert.Len(t, m.UnCommited(), 1)
		assert.Equal(t, m.LastVal(), 0.6)
	})

	t.Run("committing_multiple_times_has_same_effect", func(t *testing.T) {
		m := types.GenericMetric[float64]{
			Key:    "paths",
			Values: []float64{0.23, 0.24},
		}

		m.Add(0.5)
		m.Add(0.6)
		m.Commit()
		m.Commit()

		assert.Empty(t, m.UnCommited())
		assert.Equal(t, m.LastVal(), 0.6)
	})
}

func TestNewGenericMetric(t *testing.T) {
	t.Run("new_metric_name_is_normalized", func(t *testing.T) {
		m := types.NewGenericMetric[int]("M  S E", 10)

		assert.Equal(t, "m-s-e", m.Name())
	})
}

func TestGenericMetricTypeReflection(t *testing.T) {
	t.Run("type=metric/continuous", func(t *testing.T) {
		m := types.NewGenericMetric[float32]("mse", 2)

		m.Add(12.33)
		m.Add(10.23)
		m.Commit()

		assert.Equal(t, types.ContinuousMetric, m.Type())
	})
	t.Run("type=metric/multival", func(t *testing.T) {
		m := types.NewGenericMetric[string]("paths", 2)

		m.AddVal("path/to/image/1")
		m.AddVal("path/to/image/2")
		m.Commit()

		assert.Equal(t, types.MultiValueMetric, m.Type())
	})
	t.Run("type=metric/single-numeric", func(t *testing.T) {
		m := types.NewGenericMetric[float64]("best acc", 1)

		m.AddVal(99.0)
		m.Commit()

		assert.Equal(t, types.SingleNumericMetric, m.Type())
	})
	t.Run("type=metric/single", func(t *testing.T) {
		m := types.NewGenericMetric[string]("checkpoint", 1)

		m.AddVal("/path/to/checkpoint")
		m.Commit()

		assert.Equal(t, types.SingleMetric, m.Type())
	})
}
