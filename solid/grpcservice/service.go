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
		Status: mlsolidv1.Status_SUCCESS,
		Size:   uint64(buf.Len()), //nolint: gosec
	})
}
