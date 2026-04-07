package types //nolint: var-naming

type Experiment struct {
	Name string
	Runs []*Run
}

// ExperimentInfo is struct that represent
// additional (user-set) information on an experiment.
type ExperimentInfo struct {
	Description string `redis:"Description"`
}

// NewExperimentInfo creates an ExperimentInfo struct from the provided hasmap.
func NewExperimentInfo(m map[string]string) ExperimentInfo {
	return ExperimentInfo{
		Description: m["Description"],
	}
}
