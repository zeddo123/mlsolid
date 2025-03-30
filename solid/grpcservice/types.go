package grpcservice

import (
	"errors"

	mlsolidv1 "buf.build/gen/go/zeddo123/mlsolid/protocolbuffers/go/mlsolid/v1"
	"github.com/zeddo123/mlsolid/solid/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func ParseError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, types.ErrInternal) {
		return status.Error(codes.Internal, err.Error())
	} else if errors.Is(err, types.ErrBadRequest) {
		return status.Error(codes.InvalidArgument, err.Error())
	} else if errors.Is(err, types.ErrInvalidInput) {
		return status.Error(codes.FailedPrecondition, err.Error())
	} else if errors.Is(err, types.ErrNotFound) {
		return status.Error(codes.NotFound, err.Error())
	} else if errors.Is(err, types.ErrAlreadyInUse) {
		return status.Error(codes.AlreadyExists, err.Error())
	}

	return status.Error(codes.Internal, err.Error())
}

func NewRunsResponse(runs []*types.Run) *mlsolidv1.RunsResponse {
	rs := make([]*mlsolidv1.Run, len(runs))

	for i, run := range runs {
		rs[i] = parseRun(run)
	}

	return &mlsolidv1.RunsResponse{Runs: rs}
}

func parseRun(run *types.Run) *mlsolidv1.Run {
	return &mlsolidv1.Run{
		RunId:        run.Name,
		ExperimentId: run.ExperimentID,
		Timestamp:    timestamppb.New(run.Timestamp),
		Metrics:      ParseMetrics(run.Metrics),
	}
}

func ParseMetrics(m map[string]types.Metric) map[string]*mlsolidv1.Metric {
	res := make(map[string]*mlsolidv1.Metric)

	for k, v := range m {
		res[k] = ParseMetric(v)
	}

	return res
}

func ParseMetric(m types.Metric) *mlsolidv1.Metric {
	var fn func(v any) *mlsolidv1.Val

	switch m.(type) {
	case *types.GenericMetric[float64]:
		fn = func(v any) *mlsolidv1.Val {
			return &mlsolidv1.Val{Val: &mlsolidv1.Val_Double{Double: v.(float64)}}
		}
	case *types.GenericMetric[string]:
		fn = func(v any) *mlsolidv1.Val {
			return &mlsolidv1.Val{Val: &mlsolidv1.Val_Str{Str: v.(string)}}
		}
	case *types.GenericMetric[int64]:
		fn = func(v any) *mlsolidv1.Val {
			return &mlsolidv1.Val{Val: &mlsolidv1.Val_Int{Int: v.(int64)}}
		}
	}

	vals := make([]*mlsolidv1.Val, len(m.Vals()))

	for i, val := range m.Vals() {
		vals[i] = fn(val)
	}

	return &mlsolidv1.Metric{
		Name: m.Name(),
		Vals: vals,
	}
}

func parseGrpcMetric(ms []*mlsolidv1.Metric) []types.Metric {
	metrics := make([]types.Metric, len(ms))

	for i, m := range ms {
		if m == nil || len(m.Vals) == 0 {
			continue
		}

		var metric types.Metric

		switch m.Vals[0].GetVal().(type) {
		case *mlsolidv1.Val_Str:
			metric = types.NewGenericMetric[string](m.Name, len(m.Vals))
		case *mlsolidv1.Val_Double:
			metric = types.NewGenericMetric[float64](m.Name, len(m.Vals))
		case *mlsolidv1.Val_Int:
			metric = types.NewGenericMetric[int64](m.Name, len(m.Vals))
		}

		for _, val := range m.Vals {
			metric.AddVal(val.GetVal())
		}

		metrics[i] = metric
	}

	return metrics
}
