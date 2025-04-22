package s3

import (
	"context"
	"io"

	"github.com/zeddo123/mlsolid/solid/types"
)

type MockObjectStore struct{}

var _ ObjectStore = MockObjectStore{}

func (m MockObjectStore) UploadFile(_ context.Context, key string, _ io.Reader) (string, error) {
	return key, nil
}

func (m MockObjectStore) DownloadFile(_ context.Context, _ string) (io.ReadCloser, error) {
	return nil, nil
}

func (m MockObjectStore) UploadArtifacts(_ context.Context, _ []types.Artifact) ([]types.SavedArtifact, error) {
	return []types.SavedArtifact{}, nil
}
