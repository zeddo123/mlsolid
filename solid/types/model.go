package types

import (
	"encoding/json"
	"fmt"
	"time"
)

// ModelEntry represents a model entry in a registry.
type ModelEntry struct {
	URL       string    `json:"url"`
	Tags      []string  `json:"tags"`
	Name      string    `json:"name"`
	Timestamp time.Time `json:"timestamp"`
}

// ModelRegistry holds data related to a registry.
type ModelRegistry struct {
	Name           string
	Models         []ModelEntry
	Tags           map[string][]int
	Timestamp      time.Time
	BenchmarkImage string
}

// NewModelRegistry creates a new registry.
func NewModelRegistry(name string) *ModelRegistry {
	return &ModelRegistry{ //nolint: exhaustruct
		Name:      name,
		Models:    make([]ModelEntry, 0),
		Tags:      make(map[string][]int),
		Timestamp: time.Now(),
	}
}

// NewModelRegistryWithTime creates a new registry with a specific Timestamp.
func NewModelRegistryWithTime(name string, t time.Time) *ModelRegistry {
	r := NewModelRegistry(name)
	r.Timestamp = t

	return r
}

// VersionTag builds the version tag of a model.
func (m *ModelRegistry) VersionTag(version int) string {
	return fmt.Sprintf("v%d", version)
}

// LatestVersion returns the latest version number in the registry.
func (m *ModelRegistry) LatestVersion() int {
	return len(m.Models)
}

// LastModel returns the last model push to the registry.
func (m *ModelRegistry) LastModel() ModelEntry {
	return m.Models[len(m.Models)-1]
}

// ModelByTag returns the last model entry tagged with tag.
func (m *ModelRegistry) ModelByTag(tag string) (ModelEntry, error) {
	versions, ok := m.Tags[tag]
	if !ok {
		return ModelEntry{}, NewNotFoundErr("unknown tag")
	}

	return m.Models[versions[len(versions)-1]-1], nil
}

// ModelsByTag returns the models tagged with tag.
func (m *ModelRegistry) ModelsByTag(tag string) ([]ModelEntry, error) {
	versions, ok := m.Tags[tag]
	if !ok {
		return nil, NewNotFoundErr("unknown tag")
	}

	models := make([]ModelEntry, len(versions))
	for i, version := range versions {
		models[i] = m.Models[version-1]
	}

	return models, nil
}

// ModelByVersion returns a model entry by its version.
func (m *ModelRegistry) ModelByVersion(version int) (ModelEntry, error) {
	if version < 1 || version > len(m.Models) {
		return ModelEntry{}, NewNotFoundErr("unknown version number")
	}

	return m.Models[version-1], nil
}

// Add registers a model checkpoint with a highest version number.
// Additional tags can be supplied such (`latest`, `prod`, `dev`).
func (m *ModelRegistry) Add(url string, tags ...string) {
	version := len(m.Models) + 1

	for _, tag := range tags {
		m.addTag(tag, version)
	}

	e := ModelEntry{
		URL:       url,
		Tags:      append(tags, m.VersionTag(version)),
		Timestamp: time.Now(),
	}

	m.pushEntry(e)
}

// MarshalEntries marshals all model entries to json.
func (m *ModelRegistry) MarshalEntries() ([][]byte, error) {
	res := make([][]byte, len(m.Models))

	for i, e := range m.Models {
		j, err := json.Marshal(e)
		if err != nil {
			return nil, NewInternalErr("could not process model entry")
		}

		res[i] = j
	}

	return res, nil
}

// AddTag adds a tag to model entry by its version number.
func (m *ModelRegistry) AddTag(tag string, version int) error {
	if version < 1 || version > len(m.Models) {
		return NewNotFoundErr("unknown version number")
	}

	m.addTag(tag, version)

	return nil
}

// SetBenchmarkImage sets the benchmark image of Model registry.
func (m *ModelRegistry) SetBenchmarkImage(image string) error {
	m.BenchmarkImage = image

	return nil
}

func (m *ModelRegistry) pushEntry(e ModelEntry) {
	m.Models = append(m.Models, e)
}

func (m *ModelRegistry) addTag(tag string, version int) {
	versions, ok := m.Tags[tag]
	if !ok {
		versions = make([]int, 0)
		m.Tags[tag] = versions
	}

	m.Tags[tag] = append(versions, version)
}
