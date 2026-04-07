package types

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"maps"
	"math/rand/v2"
	"slices"
	"sync"
	"time"
)

var (
	chacha8Source  *rand.ChaCha8 //nolint: gochecknoglobals
	initRandSource sync.Once     //nolint: gochecknoglobals
)

// Run struct holds all data (information, metrics, artifacts) related to a run.
type Run struct {
	Name         string
	Timestamp    time.Time
	ExperimentID string
	Color        string
	Metrics      map[string]Metric
	Artifacts    map[string]Artifact
}

// NewRun initializes a new run with a name and an experiment id.
// The supplied name is normalized to avoid whitespace.
func NewRun(name string, expID string) Run {
	return Run{
		Name:         normalizeName(name),
		ExperimentID: expID,
		Timestamp:    time.Now(),
		Color:        generateColor(),
		Metrics:      make(map[string]Metric),
		Artifacts:    make(map[string]Artifact),
	}
}

// AddMetric adds a new metric with a name.
// returns an error if the metric name is already in use `ErrAlreadyInUse`.
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

// AddAritifact adds a new artifact to the run. Retruns a non-nil
// error if the supplied name is already in use `ErrAlreadyInUse`.
func (r *Run) AddAritifact(name string, a Artifact) error {
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

func (r *Run) ArtifactsSlice() []Artifact {
	if r.Artifacts == nil {
		return []Artifact{}
	}

	artifacts := make([]Artifact, 0)

	for _, val := range r.Artifacts {
		artifacts = append(artifacts, val)
	}

	return artifacts
}

// UniqueMetrics returns all distinct/unique metric names ids from a slice of runs.
func UniqueMetrics(runs []*Run) []string {
	metrics := make(map[string]struct{})

	for _, run := range runs {
		for m := range run.Metrics {
			metrics[m] = struct{}{}
		}
	}

	return slices.Collect(maps.Keys(metrics))
}

// CollectMetric aggregates all values of a metric present in a slice of runs.
// Returns a key-value map of runIds (keys) and metric values.
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

// generateColor generetes a random color in hex format `#342d13`.
func generateColor() string {
	initRandSource.Do(func() {
		s := uint64(time.Now().UnixNano())

		uintSize := 8
		seed := make([]byte, uintSize)

		binary.LittleEndian.PutUint64(seed, s)

		var chaSeed [32]byte
		copy(chaSeed[:], seed)
		chacha8Source = rand.NewChaCha8(chaSeed)
	})

	size := 3
	bytes := make([]byte, size)

	_, err := chacha8Source.Read(bytes)
	if err != nil {
		panic(err)
	}

	return "#" + hex.EncodeToString(bytes)
}
