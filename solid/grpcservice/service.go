package grpcservice

import (
	"context"

	"github.com/zeddo123/mlsolid/solid/controllers"
	mlsolidv1 "github.com/zeddo123/mlsolid/solid/gen/mlsolid/v1"
	"github.com/zeddo123/mlsolid/solid/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Service struct {
	mlsolidv1.UnimplementedMlsolidServer
	Controller controllers.Controller
}

func (s *Service) Experiments(ctx context.Context, req *mlsolidv1.ExperimentsRequest) (*mlsolidv1.ExperimentsResponse, error) {
	ids, err := s.Controller.Exps(ctx)
	if err != nil {
		return nil, ParseError(err)
	}

	return &mlsolidv1.ExperimentsResponse{ExpIds: ids}, nil
}

func (s *Service) CreateRun(ctx context.Context, req *mlsolidv1.CreateRunRequest) (*mlsolidv1.CreateRunResponse, error) {
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

func (s *Service) AddMetrics(ctx context.Context, req *mlsolidv1.AddMetricsRequest) (*mlsolidv1.AddMetricsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "req cannot be <nil>")
	}

	err := s.Controller.AddMetrics(ctx, req.RunId, parseGrpcMetric(req.Metrics))
	if err != nil {
		return nil, ParseError(err)
	}

	return &mlsolidv1.AddMetricsResponse{}, nil
}

func (s *Service) AddArtifacts(ctx context.Context, req *mlsolidv1.AddArtifactsRequest) (*mlsolidv1.AddArtifactsResponse, error) {
	return nil, nil
}
