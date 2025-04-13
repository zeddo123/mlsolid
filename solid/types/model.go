package types

import "fmt"

type ModelEntry struct {
	URL  string
	Tags []string
}

type ModelRegistry struct {
	Name   string
	models []ModelEntry
	tags   map[string][]int
}

func NewModelRegistry(name string) *ModelRegistry {
	return &ModelRegistry{
		Name:   name,
		models: make([]ModelEntry, 0),
		tags:   make(map[string][]int),
	}
}

func (m ModelRegistry) VersionTag(version int) string {
	return fmt.Sprintf("v%d", version)
}

func (m ModelRegistry) LatestVersion() int {
	return len(m.models)
}

func (m ModelRegistry) ModelByTag(tag string) (ModelEntry, error) {
	versions, ok := m.tags[tag]
	if !ok {
		return ModelEntry{}, NewNotFoundErr("unknown tag")
	}

	return m.models[versions[len(versions)-1]-1], nil
}

func (m ModelRegistry) ModelsByTag(tag string) ([]ModelEntry, error) {
	versions, ok := m.tags[tag]
	if !ok {
		return nil, NewNotFoundErr("unknown tag")
	}

	models := make([]ModelEntry, len(versions))
	for i, version := range versions {
		models[i] = m.models[version-1]
	}

	return models, nil
}

func (m ModelRegistry) ModelByVersion(version int) (ModelEntry, error) {
	if version < 1 || version > len(m.models) {
		return ModelEntry{}, NewNotFoundErr("unknown version number")
	}

	return m.models[version-1], nil
}

// Add registers a model checkpoint with a highest version number.
// Additional tags can be supplied such (`latest`, `prod`, `dev`)
func (m *ModelRegistry) Add(url string, tags ...string) {
	version := len(m.models) + 1

	for _, tag := range tags {
		m.addTag(tag, version)
	}

	e := ModelEntry{
		URL:  url,
		Tags: append(tags, m.VersionTag(version)),
	}

	m.pushEntry(e)
}

func (m *ModelRegistry) pushEntry(e ModelEntry) {
	m.models = append(m.models, e)
}

func (m *ModelRegistry) addTag(tag string, version int) {
	versions, ok := m.tags[tag]
	if !ok {
		versions = make([]int, 0)
		m.tags[tag] = versions
	}

	m.tags[tag] = append(versions, version)
}
