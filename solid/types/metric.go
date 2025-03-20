package types

import "golang.org/x/exp/constraints"

type Metric interface {
	Name() string
	Vals() []any
	SetVals(vs []any)
	ValsToCommit() []any
	Commit()
	LastVal() any
}

type GenericMetric[T constraints.Ordered] struct {
	Key        string
	Values     []T
	unCommited []T
}

func NewGenericMetric[T constraints.Ordered](key string, sizeAlloc int) *GenericMetric[T] {
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
