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
	ID             string `validate:"required"`
	Name           string `validate:"required"`
	Paused         bool
	EagerStart     bool
	AutoTag        bool
	Tag            string `validate:"required"`
	DecisionMetric string
	Registries     []string `validate:"required"`
	Metrics        []string `validate:"required"`
	DatasetName    string   `validate:"required"`
	DatasetURL     string   `validate:"required,url"`
	FromS3         bool
	Timestamp      time.Time `validate:"required"`
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
	Registry  string
	Version   int64
	Metrics   map[string]float32
	Timestamp time.Time
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

// Sanatize validates and cleans benchmark fields.
func (b *Bench) Sanatize() {
	b.Name = SanatizeName(b.Name)
	b.DatasetName = strings.TrimSpace(b.DatasetName)
}

// SanatizeName sanatizes a name by removing any whitespace
// and converting the string to lower case.
func SanatizeName(name string) string {
	return strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(name)), "-"))
}

// BestRuns returns the best performing bechmark run for each metric provided.
func BestRuns(runs []*BenchRun, metrics ...string) map[string]*BenchRun {
	out := make(map[string]*BenchRun, len(metrics))

	for _, run := range runs {
		for _, metric := range metrics {
			if out[metric] == nil {
				out[metric] = run

				continue
			}

			val, ok := run.Metrics[metric]
			if !ok {
				continue
			}

			if out[metric].Metrics[metric] < val {
				out[metric] = run
			}
		}
	}

	return out
}
