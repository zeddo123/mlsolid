package types //nolint: var-naming

type Experiment struct {
	Name string
	Runs []*Run
}
