package s3

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/zeddo123/mlsolid/solid/types"
)

const IDByteSize = 6

type ObjectStore interface {
	UploadFile(ctx context.Context, key string, body io.Reader) (string, error)
	DownloadFile(ctx context.Context, key string) (io.ReadCloser, error)
	UploadArtifacts(ctx context.Context, artifacts []types.Artifact) ([]types.SavedArtifact, error)
}

type Store struct {
	Bucket          string
	Endpoint        string
	AccessKey       string
	SecretAccessKey string
	Signature       string
	Region          string
	Prefix          string
	client          *s3.Client
}

var _ ObjectStore = Store{}

type StoreOps struct {
	Bucket          string
	Endpoint        string
	AccessKey       string
	SecretAccessKey string
	Region          string
	Signature       string
	Prefix          string
}

func NewStore(ops StoreOps) (Store, error) {
	service := Store{
		Bucket:          ops.Bucket,
		Endpoint:        ops.Bucket,
		AccessKey:       ops.AccessKey,
		SecretAccessKey: ops.SecretAccessKey,
		Signature:       ops.Signature,
		Region:          ops.Region,
		Prefix:          ops.Prefix,
	}

	creds := credentials.NewStaticCredentialsProvider(service.AccessKey, service.SecretAccessKey, service.Signature)

	config, err := config.LoadDefaultConfig(
		context.Background(),
		config.WithRegion(service.Region),
		config.WithCredentialsProvider(creds))
	if err != nil {
		return service, err
	}

	service.client = s3.NewFromConfig(config, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(service.Endpoint)
	})

	return service, nil
}

func (s Store) UploadArtifacts(ctx context.Context, as []types.Artifact) ([]types.SavedArtifact, error) {
	artifacts := make([]types.SavedArtifact, 0, len(as))

	var errs error

	for _, a := range as {
		key := s.GenerateKey(a.Name())

		_, err := s.UploadFile(ctx, key, bytes.NewReader(a.Content()))
		if err != nil {
			errs = fmt.Errorf("%w | could not upload artifact <%s> : %w", errs, a.Name(), err)

			continue
		}

		artifacts = append(artifacts, types.SavedArtifact{
			Name:        a.Name(),
			ContentType: a.ContentType(),
			S3Key:       key,
		})
	}

	return artifacts, errs
}

func (s Store) UploadFile(ctx context.Context, key string, body io.Reader) (string, error) {
	if s.client == nil {
		return "", types.ErrNotInitialized
	}

	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: &s.Bucket,
		Key:    &key,
		Body:   body,
	})
	if err != nil {
		return "", err
	}

	return key, nil
}

// DownloadFile
func (s Store) DownloadFile(ctx context.Context, key string) (io.ReadCloser, error) {
	if s.client == nil {
		return nil, types.ErrNotInitialized
	}

	obj, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &s.Bucket,
		Key:    &key,
	})
	if err != nil {
		return nil, err
	}

	return obj.Body, nil
}

func (s *Store) GenerateKey(name string) string {
	r, err := generateID(IDByteSize)
	if err != nil {
		panic(err)
	}

	return fmt.Sprintf("%s/%s-%s", s.Prefix, name, r)
}

func generateID(b int) (string, error) {
	bytes := make([]byte, b)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return hex.EncodeToString(bytes), nil
}
