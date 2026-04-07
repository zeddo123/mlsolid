package v1

// Experiment struct returned by Experiment endpoint.
type Experiment struct {
	Details string    `json:"details"`
	Runs    []runInfo `json:"runs"`
	Metrics []string  `json:"metrics"`
}

type runInfo struct {
	RunID     string `json:"runId"`
	Timestamp string `json:"timestamp"`
	Color     string `json:"color"`
}
