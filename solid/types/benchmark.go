package types

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
)

// Bench represents a benchmark.
type Bench struct {
	Name        string `validate:"required"`
	Paused      bool
	EagerStart  bool
	AutoTag     bool
	Tag         string   `validate:"required"`
	Registries  []string `validate:"required"`
	Metrics     []string `validate:"required"`
	DatasetName string   `validate:"required"`
	DatasetURL  string   `validate:"required,url"`
	FromS3      bool
	Timestamp   time.Time `validate:"required"`
}

// BenchRun represents a benchmark run on a registry and version.
type BenchRun struct {
	Registry  string
	Version   int64
	Metrics   map[string]float64
	Timestamp time.Time
}

// BenchEvent represents a benchmarking event.
type BenchEvent struct {
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
