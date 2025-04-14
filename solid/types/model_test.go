package types_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zeddo123/mlsolid/solid/types"
)

func TestModelRegistry(t *testing.T) {
	t.Run("last_version_number_is_correct", func(t *testing.T) {
		t.Parallel()

		// Arrange
		r := types.NewModelRegistry("object-detection")

		// Act
		r.Add("<url-1>", "prod")
		r.Add("<url-2>")
		r.Add("<url-3>")
		r.Add("<url-4>", "prod")

		assert.Equal(t, 4, r.LatestVersion())
		model, err := r.ModelByVersion(4)
		assert.NoError(t, err)
		assert.Equal(t, "<url-4>", model.URL)
		assert.Equal(t, "<url-4>", r.LastModel().URL)
	})

	t.Run("last_tagged_model_is_correct", func(t *testing.T) {
		t.Parallel()

		// Arrange
		r := types.NewModelRegistry("object-detection")

		// Act
		r.Add("<url-1>", "prod")
		r.Add("<url-2>")
		r.Add("<url-3>", "prod")
		r.Add("<url-4>")

		// Assert
		model, err := r.ModelByTag("prod")
		assert.NoError(t, err)
		assert.Equal(t, "<url-3>", model.URL)
		assert.Contains(t, model.Tags, "v3")
		assert.Contains(t, model.Tags, "prod")
	})
}
