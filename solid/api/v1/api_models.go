package v1

import "time"

// Experiment struct returned by Experiment endpoint.
type Experiment struct {
	Details string    `json:"details"`
	Runs    []runInfo `json:"runs"`
	Metrics []string  `json:"metrics"`
}

type runInfo struct {
	RunID     string    `json:"runId"`
	CreatedAt time.Time `json:"createdAt,format:datetime"`
	Color     string    `json:"color"`
}

// Registry struct returned by registry endpoint.
type Registry struct {
	Details     string            `json:"details"`
	Name        string            `json:"name"`
	LastVer     int64             `json:"lastVer"`
	Tags        map[string][]int  `json:"tags"`
	CreatedAt   time.Time         `json:"createdAt,format:datetime"`
	EntriesInfo map[int]entryInfo `json:"entriesInfo"`
}

type entryInfo struct {
	CreatedAt time.Time `json:"createdAt,format:datetime"`
}
