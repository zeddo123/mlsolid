package types_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zeddo123/mlsolid/solid/types"
)

func TestModelRegistry(t *testing.T) {
	// Arrange
	r := types.NewModelRegistry("object-detection")

	// Act
	r.Add("<url-1>", "prod")
	r.Add("<url-2>")
	r.Add("<url-3>")
	r.Add("<url-4>", "prod")

	// Assert
	model, err := r.ModelByVersion(3)
	assert.NoError(t, err)
	assert.Equal(t, "<url-3>", model.URL)
	model, err = r.ModelByTag("prod")
	assert.NoError(t, err)
	assert.Equal(t, "<url-4>", model.URL)
}
