package main

import (
	"context"
	"log"
	"math/rand"
	"sort"

	"buf.build/gen/go/zeddo123/mlsolid/grpc/go/mlsolid/v1/mlsolidv1grpc"
	mlsolidv1 "buf.build/gen/go/zeddo123/mlsolid/protocolbuffers/go/mlsolid/v1"
)

func addDescMetrics(client mlsolidv1grpc.MlsolidServiceClient, runID string, metricName string) {
	count := 20
	seq := randSlice(count)
	sort.Slice(seq, func(i, j int) bool {
		return seq[i] <= seq[j]
	})

	vals := make([]*mlsolidv1.Val, len(seq))

	for i := range vals {
		vals[i] = &mlsolidv1.Val{Val: &mlsolidv1.Val_Double{Double: seq[i]}}
	}

	req := &mlsolidv1.AddMetricsRequest{
		RunId: runID,
		Metrics: []*mlsolidv1.Metric{
			{Name: metricName, Vals: vals},
		},
	}

	commitMetric(client, req)
}

func addRandMetrics(client mlsolidv1grpc.MlsolidServiceClient, runID string, metricName string) {
	count := 20
	seq := randSlice(count)

	vals := make([]*mlsolidv1.Val, len(seq))

	for i := range vals {
		vals[i] = &mlsolidv1.Val{Val: &mlsolidv1.Val_Double{Double: seq[i]}}
	}

	req := &mlsolidv1.AddMetricsRequest{
		RunId: runID,
		Metrics: []*mlsolidv1.Metric{
			{Name: metricName, Vals: vals},
		},
	}

	commitMetric(client, req)
}

func addIncMetrics(client mlsolidv1grpc.MlsolidServiceClient, runID string, metricName string) {
	count := 20
	seq := randSlice(count)
	sort.Slice(seq, func(i, j int) bool {
		return seq[i] >= seq[j]
	})

	vals := make([]*mlsolidv1.Val, len(seq))

	for i := range vals {
		vals[i] = &mlsolidv1.Val{Val: &mlsolidv1.Val_Double{Double: seq[i]}}
	}

	req := &mlsolidv1.AddMetricsRequest{
		RunId: runID,
		Metrics: []*mlsolidv1.Metric{
			{Name: metricName, Vals: vals},
		},
	}

	commitMetric(client, req)
}

func commitMetric(client mlsolidv1grpc.MlsolidServiceClient, m *mlsolidv1.AddMetricsRequest) {
	resp, err := client.AddMetrics(context.Background(), m)
	if err != nil {
		panic(err)
	}

	log.Printf("[populate]: metric added=%t runId=%s \n", resp.GetAdded(), m.GetRunId())
}

func randSlice(size int) []float64 {
	seq := make([]float64, size)

	for i := range seq {
		seq[i] = rand.ExpFloat64()
	}

	return seq
}
