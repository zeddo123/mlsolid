// Package grpcservice implements the grpc server methods
package grpcservice

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"time"

	mlsolidv1grpc "buf.build/gen/go/zeddo123/mlsolid/grpc/go/mlsolid/v1/mlsolidv1grpc"
	mlsolidv1 "buf.build/gen/go/zeddo123/mlsolid/protocolbuffers/go/mlsolid/v1"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/rs/zerolog"
	"github.com/zeddo123/mlsolid/solid/controllers"
	"github.com/zeddo123/mlsolid/solid/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Service struct {
	mlsolidv1grpc.UnimplementedMlsolidServiceServer

	Controller *controllers.Controller
}

// StartServer starts a grpc server instance.
func StartServer(port string, ctrl *controllers.Controller) {
	l, err := net.Listen("tcp", ":"+port) //nolint: noctx
	if err != nil {
		log.Println("could not listen to port", port)

		panic(err)
	}

	service := Service{ //nolint: exhaustruct
		Controller: ctrl,
	}

	logger := zerolog.New(os.Stderr)
	opts := []logging.Option{
		logging.WithLogOnEvents(logging.StartCall, logging.FinishCall),
	}

	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			logging.UnaryServerInterceptor(interceptorLogger(logger), opts...),
		),
		grpc.ChainStreamInterceptor(
			logging.StreamServerInterceptor(interceptorLogger(logger), opts...),
		),
	)

	mlsolidv1grpc.RegisterMlsolidServiceServer(server, &service)

	log.Println("gRPC server started at", port)

	if err := server.Serve(l); err != nil {
		panic(err)
	}
}

func (s *Service) Experiment(ctx context.Context,
	req *mlsolidv1.ExperimentRequest,
) (*mlsolidv1.ExperimentResponse, error) {
	id := req.GetExpId()

	runIDs, err := s.Controller.ExpRuns(ctx, id)
	if err != nil {
		return nil, ParseError(err)
	}

	info, err := s.Controller.ExpInfo(ctx, id)
	if err != nil {
		return nil, ParseError(err)
	}

	return &mlsolidv1.ExperimentResponse{
		RunIds: runIDs,
		Desc:   info.Description,
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

	run := types.NewRun(req.GetRunId(), req.GetExperimentId())

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

	run, err := s.Controller.Run(ctx, req.GetRunId())
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

	err := s.Controller.AddMetrics(ctx, req.GetRunId(), parseGrpcMetric(req.GetMetrics()))
	if err != nil {
		return nil, ParseError(err)
	}

	return &mlsolidv1.AddMetricsResponse{Added: true}, nil
}

func (s *Service) Artifact(req *mlsolidv1.ArtifactRequest, stream mlsolidv1grpc.MlsolidService_ArtifactServer) error {
	if req == nil {
		return status.Error(codes.InvalidArgument, "req cannot be <nil>")
	}

	artifact, body, err := s.Controller.Artifact(stream.Context(), req.GetRunId(), req.GetArtifactName())
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
			RunId: req.GetRunId(),
		},
	}})
	if err != nil {
		return status.Error(codes.Internal, "could not send metadata of artifact")
	}

	eof := false

	for {
		_, err := body.Read(buffer)
		if errors.Is(err, io.EOF) {
			eof = true
		} else if err != nil {
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

		if eof {
			break
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
		if errors.Is(err, io.EOF) {
			break
		}

		switch request.GetRequest().(type) {
		case *mlsolidv1.AddArtifactRequest_Metadata:
			metadata, ok := request.GetRequest().(*mlsolidv1.AddArtifactRequest_Metadata)
			if !ok {
				return status.Errorf(codes.InvalidArgument, "could not read metadata")
			}

			runID = metadata.Metadata.GetRunId()
			artifactName = metadata.Metadata.GetName()

			contentType = metadata.Metadata.GetType()
			if !types.IsValidContentType(contentType) {
				return status.Error(codes.InvalidArgument, "unknown content type for artifact")
			}

		case *mlsolidv1.AddArtifactRequest_Content:
			content, ok := request.GetRequest().(*mlsolidv1.AddArtifactRequest_Content)
			if !ok {
				break
			}

			_, err = buf.Write(content.Content.GetContent())
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

	err := s.Controller.CreateModelRegistry(ctx, req.GetName())
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

	registry, err := s.Controller.ModelRegistry(ctx, req.GetName())
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

	err := s.Controller.AddArtifactToRegistry(ctx,
		req.GetName(), req.GetRunId(), req.GetArtifactId(), req.GetTags()...)
	if err != nil {
		log.Println(err)

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

	entry, err := s.Controller.TaggedModel(ctx, req.GetName(), req.GetTag())
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

func (s *Service) StreamTaggedModel(req *mlsolidv1.StreamTaggedModelRequest,
	stream mlsolidv1grpc.MlsolidService_StreamTaggedModelServer,
) error {
	if req == nil {
		return status.Error(codes.InvalidArgument, "req cannot be <nil>")
	}

	entry, err := s.Controller.TaggedModel(stream.Context(), req.GetName(), req.GetTag())
	if err != nil {
		return ParseError(err)
	}

	bufferSize := 1024
	buffer := make([]byte, bufferSize)

	fileName := strings.ReplaceAll(entry.URL, "/", "_")

	err = stream.Send(&mlsolidv1.StreamTaggedModelResponse{
		Response: &mlsolidv1.StreamTaggedModelResponse_Metadata{
			Metadata: &mlsolidv1.MetaData{
				Name: fmt.Sprintf("%s_%s_%s", req.GetName(), req.GetTag(), fileName),
				Type: string(types.ModelContentType),
			},
		},
	})
	if err != nil {
		return status.Error(codes.Internal, "could not send metadata of model entry")
	}

	body, err := s.Controller.S3.DownloadFile(stream.Context(), entry.URL)
	if err != nil {
		return ParseError(err)
	}
	defer body.Close()

	for {
		_, err := body.Read(buffer)
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return status.Error(codes.Internal, "could not read artifact into buffer")
		}

		err = stream.Send(&mlsolidv1.StreamTaggedModelResponse{
			Response: &mlsolidv1.StreamTaggedModelResponse_Content{
				Content: &mlsolidv1.Content{
					Content: buffer,
				},
			},
		})
		if err != nil {
			return status.Error(codes.Internal, "could not send chunk to client")
		}
	}

	return nil
}

func (s *Service) TagModel(ctx context.Context, req *mlsolidv1.TagModelRequest) (*mlsolidv1.TagModelResponse, error) {
	err := s.Controller.TagModel(ctx, req.GetName(), int(req.GetVersion()), req.GetTags()...)
	if err != nil {
		return nil, ParseError(err)
	}

	return &mlsolidv1.TagModelResponse{Added: true}, nil
}

// SetBenchmarkContainer sets the benchmark container of a registry.
func (s *Service) SetBenchmarkContainer(ctx context.Context, req *mlsolidv1.SetBenchmarkContainerRequest) (*mlsolidv1.SetBenchmarkContainerResponse, error) {
	err := s.Controller.UpdateRegistryDockerImage(ctx, req.GetRegistryName(), req.GetContainerUrl())
	if err != nil {
		return nil, ParseError(err)
	}

	return &mlsolidv1.SetBenchmarkContainerResponse{Set: true}, nil
}

// Benchmarks returns a list of benchmark ids.
func (s *Service) Benchmarks(ctx context.Context, req *mlsolidv1.BenchmarksRequest) (*mlsolidv1.BenchmarksResponse, error) {
	benchs, err := s.Controller.Benchmarks(ctx)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &mlsolidv1.BenchmarksResponse{
		Benchmarks: benchs,
	}, nil
}

// Benchmark pull a benchmark by its id.
func (s *Service) Benchmark(ctx context.Context,
	req *mlsolidv1.BenchmarkRequest,
) (*mlsolidv1.BenchmarkResponse, error) {
	bench, err := s.Controller.Benchmark(ctx, req.GetBenchmarkId())
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &mlsolidv1.BenchmarkResponse{
		BenchmarkId:     bench.ID,
		Name:            bench.Name,
		EagerStart:      bench.EagerStart,
		AutoTag:         bench.AutoTag,
		Tag:             bench.Tag,
		DecisionMetric:  bench.DecisionMetric,
		ModelRegistries: bench.Registries,
		Metrics:         bench.BenchMetrics(),
		DatasetName:     bench.DatasetName,
		DatasetUrl:      bench.DatasetURL,
		FromS3:          bench.FromS3,
	}, nil
}

// CreateBenchmark grpc method to create a new benchmark.
func (s *Service) CreateBenchmark(ctx context.Context, req *mlsolidv1.CreateBenchmarkRequest) (*mlsolidv1.CreateBenchmarkResponse, error) {
	metrics := make([]types.BenchMetric, len(req.GetMetrics()))

	for i, m := range req.GetMetrics() {
		metrics[i] = types.BenchMetric{Name: m}
	}

	id, created, err := s.Controller.CreateBenchmark(ctx, types.Bench{
		Timestamp:      time.Now(),
		Paused:         false,
		Name:           req.GetName(),
		EagerStart:     req.GetEagerStart(),
		AutoTag:        req.GetAutoTag(),
		Tag:            req.GetTag(),
		DecisionMetric: req.GetDecisionMetric(),
		Registries:     req.GetModelRegistries(),
		Metrics:        metrics,
		DatasetName:    req.GetDatasetName(),
		DatasetURL:     req.GetDatasetUrl(),
		FromS3:         req.GetFromS3(),
	})
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &mlsolidv1.CreateBenchmarkResponse{
		BenchmarkId: id,
		Created:     created,
	}, nil
}

// ToggleBenchmark toogles a benchmark's pause state.
func (s *Service) ToggleBenchmark(ctx context.Context, req *mlsolidv1.ToggleBenchmarkRequest) (*mlsolidv1.ToggleBenchmarkResponse, error) {
	err := s.Controller.ToggleBenchmark(ctx, req.GetBenchmarkId(), req.GetPaused())
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &mlsolidv1.ToggleBenchmarkResponse{Paused: req.GetPaused()}, nil
}

// UpdateBenchmark updates an existant benchmark.
func (s *Service) UpdateBenchmark(ctx context.Context, req *mlsolidv1.UpdateBenchmarkRequest) (*mlsolidv1.UpdateBenchmarkResponse, error) {
	err := s.Controller.UpdateBenchmark(ctx, req.GetBenchmarkId(), types.UpdateBench{
		Name:           req.GetName(),
		AutoTag:        req.AutoTag,
		Tag:            req.GetTag(),
		DecisionMetric: req.GetDecisionMetric(),
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if len(req.GetAddMetrics()) > 0 {
		metrics := make([]types.BenchMetric, len(req.GetAddMetrics()))
		for i, m := range req.GetAddMetrics() {
			metrics[i] = types.BenchMetric{
				Name: m,
			}
		}
		err := s.Controller.AddBenchmarkMetrics(ctx, req.GetBenchmarkId(), metrics)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	if len(req.GetRemoveMetrics()) > 0 {
		err := s.Controller.RemBenchmarkMetrics(ctx, req.GetBenchmarkId(), req.GetRemoveMetrics())
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	if len(req.GetAddRegistires()) > 0 {
		err := s.Controller.AddBenchmarkRegistries(ctx, req.GetBenchmarkId(), req.GetAddRegistires())
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	if len(req.GetRemoveRegistries()) > 0 {
		err := s.Controller.RemBenchmarkRegistries(ctx, req.GetBenchmarkId(), req.GetRemoveRegistries())
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	benchmark, err := s.Controller.Benchmark(ctx, req.GetBenchmarkId())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &mlsolidv1.UpdateBenchmarkResponse{
		Name:            benchmark.Name,
		AutoTag:         benchmark.AutoTag,
		Tag:             benchmark.Tag,
		ModelRegistries: benchmark.Registries,
		Metrics:         benchmark.BenchMetrics(),
		DecisionMetric:  benchmark.DecisionMetric,
	}, nil
}

// DeleteBenchmark rpc method.
func (s *Service) DeleteBenchmark(_ context.Context, _ *mlsolidv1.DeleteBenchmarkRequest) (*mlsolidv1.DeleteBenchmarkResponse, error) {
	return nil, status.Error(codes.Unimplemented, "DeleteBenchmark is not implemented")
}

// BenchmarkRuns rpc method.
func (s *Service) BenchmarkRuns(ctx context.Context,
	req *mlsolidv1.BenchmarkRunsRequest,
) (*mlsolidv1.BenchmarkRunsResponse, error) {
	runs, err := s.Controller.BenchmarkRuns(ctx, req.GetBenchmarkId())
	if err != nil {
		return nil, status.Error(codes.NotFound, "could not pull benchmark runs")
	}

	rs := make([]*mlsolidv1.Metrics, len(runs))

	for i, run := range runs {
		if run == nil {
			continue
		}

		rs[i] = &mlsolidv1.Metrics{
			Registry:  run.Registry,
			Version:   run.Version,
			Timestamp: timestamppb.New(run.Timestamp),
			Metrics:   run.Metrics,
		}
	}

	return &mlsolidv1.BenchmarkRunsResponse{
		Runs: rs,
	}, nil
}

// BestModel rpc method.
func (s *Service) BestModel(ctx context.Context, req *mlsolidv1.BestModelRequest) (*mlsolidv1.BestModelResponse, error) {
	runs, err := s.Controller.BenchmarkRuns(ctx, req.GetBenchmarkId())
	if err != nil {
		return nil, ParseError(err)
	}

	_ = types.BestRuns(runs, req.GetMetric())

	resp := &mlsolidv1.BestModelResponse{}

	return resp, status.Error(codes.Unimplemented, "BestModel is not implemented yet")
}
