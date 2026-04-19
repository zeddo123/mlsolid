package types

import "time"

// Bench represents a benchmark.
type Bench struct {
	Name        string
	Paused      bool
	EagerStart  bool
	AutoTag     bool
	Tag         string
	Registries  []string
	Metrics     []string
	DatasetName string
	DatasetURL  string
	FromS3      bool
	Timestamp   time.Time
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

// Sanatize validates and cleans benchmark fields.
func (b *Bench) Sanatize() {
}
