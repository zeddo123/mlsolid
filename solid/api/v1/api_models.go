package v1

import (
	"time"

	"github.com/zeddo123/mlsolid/solid/types"
)

// ErrorResponse struct returned when api call is not successful.
type ErrorResponse struct {
	Error string `json:"error"`
}

// ExperimentResponse struct returned by Experiment endpoint.
type ExperimentResponse struct {
	Details string    `json:"details"`
	Runs    []runInfo `json:"runs"`
	Metrics []string  `json:"metrics"`
}

// ExperimentsResponse struct returned by Experiments endpoint.
type ExperimentsResponse struct {
	Details string              `json:"details"`
	Exps    map[string][]string `json:"exps"`
	Cursor  string              `json:"cursor"`
}

type runInfo struct {
	RunID     string    `json:"runId"`
	CreatedAt time.Time `json:"createdAt,format:datetime"`
	Color     string    `json:"color"`
}

// RegistryResponse struct returned by registry endpoint.
type RegistryResponse struct {
	Details                 string            `json:"details"`
	Name                    string            `json:"name"`
	LastVer                 int64             `json:"lastVer"`
	Tags                    map[string][]int  `json:"tags"`
	CreatedAt               time.Time         `json:"createdAt,format:datetime"`
	EntriesInfo             map[int]entryInfo `json:"entriesInfo"`
	BenchmarkImage          string            `json:"benchmarkImage"`
	BenchmarkGpuPassthrough bool              `json:"benchmarkGpuPassthrough"`
}

type entryInfo struct {
	CreatedAt time.Time `json:"createdAt,format:datetime"`
	Tags      []string  `json:"tags"`
	Name      string    `json:"name"`
	Run       string    `json:"run"`
}

// RegistriesResponse response to registries request.
type RegistriesResponse struct {
	Details    string   `json:"details"`
	Registries []string `json:"registries"`
	Cursor     string   `json:"cursor"`
}

// CreateRegistryRequest payload to create a new registry model.
type CreateRegistryRequest struct {
	Name                    string `json:"name"`
	BenchmarkGpuPassthrough bool   `json:"benchmarkGpuPassthrough"`
	BenchmarkImage          string `json:"benchmarkImage"`
}

// CreateRegistryResponse response to a registry creation request.
type CreateRegistryResponse struct {
	Details string `json:"details"`
	ID      string `json:"id"`
}

// BenchmarksResponse response returned for Benchmarks endpoint.
type BenchmarksResponse struct {
	Details    string            `json:"details"`
	Benchmarks map[string]string `json:"benchmarks"`
	Cursor     string            `json:"cursor"`
}

// BenchmarkResponse response returned for Benchmark endpoint.
type BenchmarkResponse struct {
	Details   string      `json:"details"`
	Benchmark types.Bench `json:"benchmark"`
}

// CreateBenchmarkRequest represents a request to create a new benchmark.
type CreateBenchmarkRequest struct {
	Name           string
	EagerStart     bool
	AutoTag        bool
	Tag            string
	DecisionMetric string
	Registries     []string
	Metrics        []types.BenchMetric
	DatasetName    string
	DatasetURL     string
	DatasetFromS3  bool
}

// CreateBenchmarkResponse response to a CreateBenchmark request.
type CreateBenchmarkResponse struct {
	ID      string `json:"id"`
	Created bool   `json:"created"`
	Details string `json:"details"`
}

// BechmarkBestRequest represents a request to pull best models for each metric.
type BechmarkBestRequest struct {
	Metrics []string
}

// BenchmarkToggleResponse response to benchmark toggle request.
type BenchmarkToggleResponse struct {
	Details string `json:"details"`
}

// BenchmarkUpdateResponse response to Benchmark update request.
type BenchmarkUpdateResponse struct {
	Details string `json:"details"`
}

// BenchmarkRunsResponse response to benchmark runs request.
type BenchmarkRunsResponse struct {
	Details string            `json:"details"`
	Runs    []*types.BenchRun `json:"runs"`
}

// BenchmarkBestResponse response to benchmark best request.
type BenchmarkBestResponse struct {
	Details string                     `json:"details"`
	Runs    map[string]*types.BenchRun `json:"runs"`
}

// ArtifactsResponse response to artifacts request.
type ArtifactsResponse struct {
	Details   string              `json:"details"`
	Artifacts map[string][]string `json:"artifacts"`
}

// MetricsResponse response to metrics request.
type MetricsResponse struct {
	Details string   `json:"details"`
	Metrics []string `json:"metrics"`
}

// MetricResponse response to metric request.
type MetricResponse struct {
	Details string         `json:"details"`
	Metric  map[string]any `json:"metric"`
	Kind    string         `json:"kind"`
}

// KeyLabelsResponse response to keys request.
type KeyLabelsResponse struct {
	Details string               `json:"details"`
	Labels  map[string]time.Time `json:"labels"`
}

// KeyResponse response to generating api key.
type KeyResponse struct {
	Details string `json:"details"`
	Key     string `json:"key"`
}

func (request CreateBenchmarkRequest) bench() types.Bench {
	return types.Bench{ //nolint: exhaustruct
		Name:           request.Name,
		EagerStart:     request.EagerStart,
		AutoTag:        request.AutoTag,
		Tag:            request.Tag,
		DecisionMetric: request.DecisionMetric,
		Registries:     request.Registries,
		Metrics:        request.Metrics,
		DatasetName:    request.DatasetName,
		DatasetURL:     request.DatasetURL,
		FromS3:         request.DatasetFromS3,
		Timestamp:      time.Now(),
	}
}
