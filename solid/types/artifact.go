package types

type ContentType string

const (
	TextContentType  ContentType = "content-type/text"
	ModelContentType ContentType = "content-type/model"
)

type Artifact interface {
	Name() string
	Content() []byte
	ContentType() ContentType
}

type PlainTextArtifact struct {
	FileName    string
	FileContent string
}

type CheckpointArtifact struct {
	Model      string
	Checkpoint []byte
}

type SavedArtifact struct {
	Name        string
	ContentType ContentType
	S3Key       string
}

func (p PlainTextArtifact) Name() string {
	return p.FileName
}

func (p PlainTextArtifact) Content() []byte {
	return []byte(p.FileContent)
}

func (p PlainTextArtifact) ContentType() ContentType {
	return TextContentType
}

func (c CheckpointArtifact) Name() string {
	return c.Model
}

func (c CheckpointArtifact) Content() []byte {
	return c.Checkpoint
}

func (c CheckpointArtifact) ContentType() ContentType {
	return ModelContentType
}

func ArtifactIds(artifacts []Artifact) []string {
	ids := make([]string, len(artifacts))

	for i, a := range artifacts {
		ids[i] = a.Name()
	}

	return ids
}

func ArtifactIdMap(artifacts []Artifact) map[string]Artifact {
	mapping := make(map[string]Artifact, len(artifacts))

	for _, a := range artifacts {
		mapping[a.Name()] = a
	}

	return mapping
}
