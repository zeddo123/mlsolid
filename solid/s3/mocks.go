package s3

import (
	"context"
	"io"

	"github.com/zedd123/mlsolid/solid/types"
)

type MockObjectStore struct{}

var _ ObjectStore = MockObjectStore{}

func (m MockObjectStore) UploadFile(ctx context.Context, key string, body io.Reader) (string, error) {
	return key, nil
}

func (m MockObjectStore) DownloadFile(ctx context.Context, key string) (io.ReadCloser, error) {
	return nil, nil
}

func (m MockObjectStore) UploadArtifacts(ctx context.Context, artifacts []types.Artifact) ([]types.SavedArtifact, error) {
	return []types.SavedArtifact{}, nil
}
