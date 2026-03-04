package types

import (
	"fmt"
	"maps"
	"slices"
	"time"
)

type Run struct {
	Name         string
	Timestamp    time.Time
	ExperimentID string
	Metrics      map[string]Metric
	Artifacts    map[string]Artifact
}

func NewRun(name string, expID string) Run {
	return Run{
		Name:         normalizeName(name),
		ExperimentID: expID,
		Timestamp:    time.Now(),
		Metrics:      make(map[string]Metric),
		Artifacts:    make(map[string]Artifact),
	}
}

func (r *Run) AddMetric(name string, m Metric) error {
	if r.Metrics == nil {
		return ErrInvalidInput
	}

	metric, ok := r.Metrics[name]
	if ok && metric != nil {
		return fmt.Errorf("%w: metric name is already used", ErrAlreadyInUse)
	}

	r.Metrics[name] = m

	return nil
}

func (r Run) AddAritifact(name string, a Artifact) error {
	if r.Artifacts == nil {
		return ErrInvalidInput
	}

	art, ok := r.Artifacts[name]
	if ok && art != nil {
		return fmt.Errorf("%w: already name is already used", ErrAlreadyInUse)
	}

	r.Artifacts[name] = a

	return nil
}

func (r Run) ArtifactsSlice() []Artifact {
	if r.Artifacts == nil {
		return []Artifact{}
	}

	artifacts := make([]Artifact, 0)

	for _, val := range r.Artifacts {
		artifacts = append(artifacts, val)
	}

	return artifacts
}

func UniqueMetrics(runs []*Run) []string {
	metrics := make(map[string]struct{})

	for _, run := range runs {
		for m := range run.Metrics {
			metrics[m] = struct{}{}
		}
	}

	return slices.Collect(maps.Keys(metrics))
}

func CollectMetric(runs []*Run, metric string) (map[string]any, MetricType) {
	metrics := make(map[string]any, len(runs))

	var kind MetricType

	for _, run := range runs {
		if m, ok := run.Metrics[metric]; ok {
			metrics[run.Name] = m.Vals()
			curr := m.Type()

			kind = metricTypePrededence(kind, curr)
		}
	}

	return metrics, kind
}
