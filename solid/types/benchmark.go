package types

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// Bench represents a benchmark.
type Bench struct {
	ID             string        `json:"id"             validate:"required"`
	Name           string        `json:"name"           validate:"required"`
	Paused         bool          `json:"paused"`
	EagerStart     bool          `json:"eagerStart"`
	AutoTag        bool          `json:"autoTag"`
	Tag            string        `json:"tag"            validate:"required"`
	DecisionMetric string        `json:"decisionMetric"`
	Registries     []string      `json:"registries"     validate:"required"`
	Metrics        []BenchMetric `json:"metrics"        validate:"required"`
	DatasetName    string        `json:"datasetName"    validate:"required"`
	DatasetURL     string        `json:"datasetUrl"     validate:"required,url"`
	FromS3         bool          `json:"fromS3"`
	Timestamp      time.Time     `json:"timestamp"      validate:"required"`
}

// BenchMetric represents a benchmark metric.
type BenchMetric struct {
	Name     string `json:"name"`
	DescSort bool   `json:"descSort"`
}

// UpdateBench struct used to update a benchmark.
type UpdateBench struct {
	Name           string
	AutoTag        *bool
	Tag            string
	DecisionMetric string
}

// BenchRun represents a benchmark run on a registry and version.
type BenchRun struct {
	Registry  string             `json:"registry"`
	Version   int64              `json:"version"`
	Metrics   map[string]float32 `json:"metrics"`
	Timestamp time.Time          `json:"timestamp"`
	Start     time.Time          `json:"start"`
	End       time.Time          `json:"end"`
}

// BenchEvent represents a benchmarking event.
type BenchEvent struct {
	BenchID     string
	BenchName   string
	Registry    string
	Version     int64
	DockerImage string
	ModelURL    string
	DatasetName string
	DatasetURL  string
	FromS3      bool
	AutoTag     bool
	Tag         string
}

// NewBenchMetric creates a new bench metric and sanitizes its name.
func NewBenchMetric(name string, descSort bool) BenchMetric {
	return BenchMetric{
		Name:     SanitizeName(name),
		DescSort: descSort,
	}
}

// GenerateID generates a new ID.
func (b *Bench) GenerateID() {
	b.ID = uuid.NewString()
}

// Validate validates Bench fields.
func (b *Bench) Validate() error {
	validate := validator.New(validator.WithRequiredStructEnabled())

	err := validate.Struct(b)
	if err != nil {
		return fmt.Errorf("bench validation error: %w", err)
	}

	return nil
}

// Sanitize validates and cleans benchmark fields.
func (b *Bench) Sanitize() {
	b.Name = SanitizeName(b.Name)
	b.DatasetName = strings.TrimSpace(b.DatasetName)
}

// BenchMetrics returns the names of metrics tracked.
func (b *Bench) BenchMetrics() []string {
	metrics := make([]string, len(b.Metrics))
	for i, m := range b.Metrics {
		metrics[i] = m.Name
	}

	return metrics
}

// SanitizedMetrics returns a metrics with sanitized metric names.
func (br *BenchRun) SanitizedMetrics() map[string]float32 {
	metrics := make(map[string]float32, len(br.Metrics))

	for k, v := range br.Metrics {
		metrics[SanitizeName(k)] = v
	}

	return metrics
}

// SanitizeName sanatizes a name by removing
// any whitespace and converting the string to lower case.
func SanitizeName(name string) string {
	return strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(name)), "-"))
}

// SanitizeNames sanitizes a list of names in place.
func SanitizeNames(names []string) []string {
	for i, name := range names {
		names[i] = SanitizeName(name)
	}

	return names
}

// BestRuns returns the best performing bechmark run for each metric provided.
func BestRuns(runs []*BenchRun, metrics ...BenchMetric) map[string]*BenchRun {
	out := make(map[string]*BenchRun, len(metrics))

	for _, run := range runs {
		for _, metric := range metrics {
			val, ok := run.Metrics[metric.Name]
			if !ok {
				continue
			}

			if out[metric.Name] == nil {
				out[metric.Name] = run

				continue
			}

			if metric.DescSort {
				if out[metric.Name].Metrics[metric.Name] > val {
					out[metric.Name] = run
				}
			} else {
				if out[metric.Name].Metrics[metric.Name] < val {
					out[metric.Name] = run
				}
			}
		}
	}

	return out
}
