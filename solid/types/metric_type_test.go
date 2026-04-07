package types //nolint: testpackage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetricTypePrecedence(t *testing.T) {
	tt := []struct {
		m1 MetricType
		m2 MetricType
		p  MetricType
	}{
		{
			m1: SingleMetric,
			m2: SingleMetric,
			p:  SingleMetric,
		},
		{
			m1: SingleNumericMetric,
			m2: SingleNumericMetric,
			p:  SingleNumericMetric,
		},
		{
			m1: SingleNumericMetric,
			m2: ContinuousMetric,
			p:  ContinuousMetric,
		},
		{
			m1: ContinuousMetric,
			m2: SingleNumericMetric,
			p:  ContinuousMetric,
		},
		{
			m1: ComplexMetric,
			m2: SingleMetric,
			p:  ComplexMetric,
		},
		{
			m1: SingleMetric,
			m2: SingleNumericMetric,
			p:  ComplexMetric,
		},
		{
			m1: SingleMetric,
			m2: MultiValueMetric,
			p:  MultiValueMetric,
		},
		{
			m1: "",
			m2: "",
			p:  "",
		},
		{
			m1: "",
			m2: SingleMetric,
			p:  SingleMetric,
		},
		{
			m1: ContinuousMetric,
			m2: "",
			p:  ContinuousMetric,
		},
	}

	for _, tc := range tt {
		t.Run("", func(t *testing.T) {
			assert.Equal(t, tc.p, metricTypePrededence(tc.m1, tc.m2))
		})
	}
}
