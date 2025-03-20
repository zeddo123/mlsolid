package types_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zedd123/mlsolid/solid/types"
)

func TestGenericMetric(t *testing.T) {
	t.Run("only_new_values_are_uncommited", func(t *testing.T) {
		m := types.GenericMetric[string]{
			Key:    "paths",
			Values: []string{"path1", "path2"},
		}

		m.Add("path3")
		m.Add("path4")

		assert.Len(t, m.UnCommited(), 2)
		assert.Contains(t, m.UnCommited(), "path3")
		assert.Contains(t, m.UnCommited(), "path4")
		assert.Equal(t, m.LastVal(), "path2")
	})

	t.Run("no_uncommitted_values_after_committing_them", func(t *testing.T) {
		m := types.GenericMetric[string]{
			Key:    "paths",
			Values: []string{"path1", "path2"},
		}

		m.Add("path3")
		m.Add("path4")
		m.Commit()

		assert.Len(t, m.UnCommited(), 0)
		assert.Equal(t, m.LastVal(), "path4")
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

		assert.Len(t, m.UnCommited(), 0)
		assert.Equal(t, m.LastVal(), 0.6)
	})
}

func TestNewGenericMetric(t *testing.T) {
	t.Run("new_metric_name_is_normalized", func(t *testing.T) {
		m := types.NewGenericMetric[int]("M  S E", 10)

		assert.Equal(t, m.Name(), "m-s-e")
	})
}
