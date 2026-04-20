package types_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
