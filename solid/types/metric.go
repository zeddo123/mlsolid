package types

import (
	"cmp"
	"reflect"
)

type MetricType string

const (
	ContinuousMetric    MetricType = "metric/continuous"
	MultiValueMetric    MetricType = "metric/multival"
	SingleNumericMetric MetricType = "metric/single-numeric"
	SingleMetric        MetricType = "metric/single"
	ComplexMetric       MetricType = "metric/complex"
)

type Metric interface {
	Name() string
	Vals() []any
	SetVals(vs []any)
	ValsToCommit() []any
	Commit()
	LastVal() any
	AddVal(v any)
	Type() MetricType
}

type GenericMetric[T cmp.Ordered] struct {
	Key        string
	Values     []T
	unCommited []T
}

func NewGenericMetric[T cmp.Ordered](key string, sizeAlloc int) *GenericMetric[T] {
	m := GenericMetric[T]{
		Key: normalizeName(key),
	}

	m.Values = make([]T, 0, sizeAlloc)
	m.unCommited = make([]T, 0, sizeAlloc)

	return &m
}

func (s GenericMetric[T]) Name() string {
	return s.Key
}

func (s GenericMetric[T]) Vals() []any {
	vals := make([]any, len(s.Values))

	for i, v := range s.Values {
		vals[i] = v
	}

	return vals
}

func (s *GenericMetric[T]) SetVals(vs []any) {
	for _, v := range vs {
		if g, ok := v.(T); ok {
			s.Values = append(s.Values, g)
		}
	}
}

func (s GenericMetric[T]) ValsToCommit() []any {
	vals := make([]any, len(s.unCommited))

	for i, v := range s.unCommited {
		vals[i] = v
	}

	return vals
}

func (s *GenericMetric[T]) Add(v T) {
	s.unCommited = append(s.unCommited, v)
}

func (s *GenericMetric[T]) AddVal(v any) {
	s.Add(v.(T))
}

func (s *GenericMetric[T]) UnCommited() []T {
	return s.unCommited
}

func (s *GenericMetric[T]) Commit() {
	s.Values = append(s.Values, s.unCommited...)
	s.unCommited = make([]T, 0)
}

func (s *GenericMetric[T]) LastVal() any {
	return s.Values[len(s.Values)-1]
}

// Type infers the MetricType using reflection. If the metric has multiple values
// it can either be a ContinuousMetric (numerical) or a MultiValueMetric. In the
// case where no value is present or a single value is present, it can be SingleMetric or SingleNumericMetric.
func (s *GenericMetric[T]) Type() MetricType {
	numeric := false
	str := false

	t := reflect.TypeFor[T]()

	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64,
		reflect.Complex64, reflect.Complex128:
		numeric = true
	case reflect.String:
		str = true
	}

	switch len(s.Values) {
	case 0, 1:
		if numeric {
			return SingleNumericMetric
		} else if str {
			return SingleMetric
		}

	default:
		if numeric {
			return ContinuousMetric
		} else if str {
			return MultiValueMetric
		}
	}

	return ComplexMetric
}

// metricTypePrededence given two types returns the metrics that has the most precedence.
// If two metric types are different underlying data types
// (SingleMetric vs SingleNumericMetric) a ComplexMetric is returned.
// Example:
// ContinuousMetric has a higher precedence than SingleNumericMetric.
func metricTypePrededence(m1, m2 MetricType) MetricType {
	if m1 == "" {
		return m2
	} else if m2 == "" || m1 == m2 {
		return m1
	} else if isNumericMetricType(m1) && isNumericMetricType(m2) {
		return ContinuousMetric
	} else if isNaNMetricType(m1) && isNaNMetricType(m2) {
		return MultiValueMetric
	}

	return ComplexMetric
}

func isNumericMetricType(m MetricType) bool {
	return m == SingleNumericMetric || m == ContinuousMetric
}

func isNaNMetricType(m MetricType) bool {
	return m == SingleMetric || m == MultiValueMetric
}
