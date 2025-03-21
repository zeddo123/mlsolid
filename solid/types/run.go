package types

import (
	"fmt"
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
