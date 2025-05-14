package grpcservice

import (
	"bytes"
	"context"
	"io"

	mlsolidv1grpc "buf.build/gen/go/zeddo123/mlsolid/grpc/go/mlsolid/v1/mlsolidv1grpc"
	mlsolidv1 "buf.build/gen/go/zeddo123/mlsolid/protocolbuffers/go/mlsolid/v1"
	"github.com/zeddo123/mlsolid/solid/controllers"
	"github.com/zeddo123/mlsolid/solid/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Service struct {
	mlsolidv1grpc.UnimplementedMlsolidServiceServer
	Controller controllers.Controller
}

func (s *Service) Experiment(ctx context.Context,
	req *mlsolidv1.ExperimentRequest,
) (*mlsolidv1.ExperimentResponse, error) {
	id := req.GetExpId()

	runIDs, err := s.Controller.ExpRuns(ctx, id)
	if err != nil {
		return nil, ParseError(err)
	}

	return &mlsolidv1.ExperimentResponse{
		RunIds: runIDs,
	}, nil
}

func (s *Service) Experiments(ctx context.Context,
	_ *mlsolidv1.ExperimentsRequest,
) (*mlsolidv1.ExperimentsResponse, error) {
	ids, err := s.Controller.Exps(ctx)
	if err != nil {
		return nil, ParseError(err)
	}

	return &mlsolidv1.ExperimentsResponse{ExpIds: ids}, nil
}

func (s *Service) CreateRun(ctx context.Context,
	req *mlsolidv1.CreateRunRequest,
) (*mlsolidv1.CreateRunResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "req cannot be <nil>")
	}

	run := types.NewRun(req.RunId, req.ExperimentId)

	err := s.Controller.CreateRun(ctx, run)
	if err != nil {
		return nil, ParseError(err)
	}

	return &mlsolidv1.CreateRunResponse{RunId: run.Name}, nil
}

func (s *Service) Run(ctx context.Context, req *mlsolidv1.RunRequest) (*mlsolidv1.RunResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "req cannot be <nil>")
	}

	run, err := s.Controller.Run(ctx, req.RunId)
	if err != nil {
		return nil, ParseError(err)
	}

	return &mlsolidv1.RunResponse{
		RunId:        run.Name,
		ExperimentId: run.ExperimentID,
		Timestamp:    timestamppb.New(run.Timestamp),
		Metrics:      ParseMetrics(run.Metrics),
	}, nil
}

func (s *Service) Runs(ctx context.Context, req *mlsolidv1.RunsRequest) (*mlsolidv1.RunsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "req cannot be <nil>")
	}

	runs, err := s.Controller.Runs(ctx, req.GetRunIds())
	if err != nil {
		return nil, ParseError(err)
	}

	return NewRunsResponse(runs), nil
}

func (s *Service) AddMetrics(ctx context.Context,
	req *mlsolidv1.AddMetricsRequest,
) (*mlsolidv1.AddMetricsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "req cannot be <nil>")
	}

	err := s.Controller.AddMetrics(ctx, req.RunId, parseGrpcMetric(req.Metrics))
	if err != nil {
		return nil, ParseError(err)
	}

	return &mlsolidv1.AddMetricsResponse{Added: true}, nil
}

func (s *Service) Artifact(req *mlsolidv1.ArtifactRequest, stream mlsolidv1grpc.MlsolidService_ArtifactServer) error {
	if req == nil {
		return status.Error(codes.InvalidArgument, "req cannot be <nil>")
	}

	artifact, body, err := s.Controller.Artifact(stream.Context(), req.RunId, req.ArtifactName)
	if err != nil {
		return ParseError(err)
	}
	defer body.Close()

	bufferSize := 1024
	buffer := make([]byte, bufferSize)

	err = stream.Send(&mlsolidv1.ArtifactResponse{Request: &mlsolidv1.ArtifactResponse_Metadata{
		Metadata: &mlsolidv1.MetaData{
			Name:  artifact.Name,
			Type:  string(artifact.ContentType),
			RunId: req.RunId,
		},
	}})
	if err != nil {
		return status.Error(codes.Internal, "could not send metadata of artifact")
	}

	for {
		_, err := body.Read(buffer)
		if err == io.EOF {
			break
		}

		if err != nil {
			return status.Error(codes.Internal, "could not read artifact into buffer")
		}

		err = stream.Send(&mlsolidv1.ArtifactResponse{Request: &mlsolidv1.ArtifactResponse_Content{
			Content: &mlsolidv1.Content{
				Content: buffer,
			},
		}})
		if err != nil {
			return status.Error(codes.Internal, "could not send chunk to client")
		}
	}

	return nil
}

func (s *Service) AddArtifact(stream mlsolidv1grpc.MlsolidService_AddArtifactServer) error { //nolint: cyclop
	buf := bytes.Buffer{}

	var contentType string

	var artifactName string

	var runID string

	for {
		request, err := stream.Recv()
		if err == io.EOF {
			break
		}

		switch request.Request.(type) {
		case *mlsolidv1.AddArtifactRequest_Metadata:
			metadata, ok := request.Request.(*mlsolidv1.AddArtifactRequest_Metadata)
			if !ok {
				return status.Errorf(codes.InvalidArgument, "could not read metadata")
			}

			runID = metadata.Metadata.RunId
			artifactName = metadata.Metadata.Name

			contentType = metadata.Metadata.Type
			if !types.IsValidContentType(contentType) {
				return status.Error(codes.InvalidArgument, "unknown content type for artifact")
			}

		case *mlsolidv1.AddArtifactRequest_Content:
			content, ok := request.Request.(*mlsolidv1.AddArtifactRequest_Content)
			if !ok {
				break
			}

			_, err = buf.Write(content.Content.Content)
			if err != nil {
				return status.Errorf(codes.Internal, "could not write data chunk %v", err)
			}
		}
	}

	artifact, err := types.NewArtifact(artifactName, contentType, buf.Bytes())
	if err != nil {
		return ParseError(err)
	}

	err = s.Controller.AddArtifacts(stream.Context(), runID, []types.Artifact{artifact})
	if err != nil {
		return ParseError(err)
	}

	return stream.SendAndClose(&mlsolidv1.AddArtifactResponse{
		Name:   artifactName,
		Status: mlsolidv1.Status_STATUS_SUCCESS,
		Size:   uint64(buf.Len()), //nolint: gosec
	})
}

func (s *Service) CreateModelRegistry(ctx context.Context,
	req *mlsolidv1.CreateModelRegistryRequest,
) (*mlsolidv1.CreateModelRegistryResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "req cannot be <nil>")
	}

	err := s.Controller.CreateModelRegistry(ctx, req.Name)
	if err != nil {
		return nil, ParseError(err)
	}

	return &mlsolidv1.CreateModelRegistryResponse{Created: true}, nil
}

func (s *Service) ModelRegistry(ctx context.Context,
	req *mlsolidv1.ModelRegistryRequest,
) (*mlsolidv1.ModelRegistryResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "req cannot be <nil>")
	}

	registry, err := s.Controller.ModelRegistry(ctx, req.Name)
	if err != nil {
		return nil, ParseError(err)
	}

	return parseModelRegistry(registry), nil
}

func (s *Service) AddModelEntry(ctx context.Context,
	req *mlsolidv1.AddModelEntryRequest,
) (*mlsolidv1.AddModelEntryResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "req cannot be <nil>")
	}

	err := s.Controller.AddArtifactToRegistry(ctx, req.Name, req.RunId, req.ArtifactId, req.Tags...)
	if err != nil {
		return nil, ParseError(err)
	}

	return &mlsolidv1.AddModelEntryResponse{Added: true}, nil
}

func (s *Service) TaggedModel(ctx context.Context, req *mlsolidv1.TaggedModelRequest) (*mlsolidv1.TaggedModelResponse,
	error,
) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "req cannot be <nil>")
	}

	entry, err := s.Controller.TaggedModel(ctx, req.Name, req.Tag)
	if err != nil {
		return nil, ParseError(err)
	}

	return &mlsolidv1.TaggedModelResponse{
		Entry: &mlsolidv1.ModelEntry{
			Url:  entry.URL,
			Tags: entry.Tags,
		},
	}, nil
}

func (s *Service) TagModel(ctx context.Context, req *mlsolidv1.TagModelRequest) (*mlsolidv1.TagModelResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "req cannot be <nil>")
	}

	err := s.Controller.TagModel(ctx, req.Name, int(req.Version), req.Tags...)
	if err != nil {
		return nil, ParseError(err)
	}

	return &mlsolidv1.TagModelResponse{Added: true}, nil
}
