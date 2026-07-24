package main

import (
	"context"
	"log"
	"math"
	"math/rand"
	"sort"

	"buf.build/gen/go/zeddo123/mlsolid/grpc/go/mlsolid/v1/mlsolidv1grpc"
	mlsolidv1 "buf.build/gen/go/zeddo123/mlsolid/protocolbuffers/go/mlsolid/v1"
)

const metricSeriesLen = 20

func addDescMetrics(client mlsolidv1grpc.MlsolidServiceClient, runID string, metricName string) {
	seq := randSlice(metricSeriesLen)
	sort.Slice(seq, func(i, j int) bool {
		return seq[i] <= seq[j]
	})

	commitDoubleMetric(client, runID, metricName, seq)
}

func addRandMetrics(client mlsolidv1grpc.MlsolidServiceClient, runID string, metricName string) {
	commitDoubleMetric(client, runID, metricName, randSlice(metricSeriesLen))
}

func addIncMetrics(client mlsolidv1grpc.MlsolidServiceClient, runID string, metricName string) {
	seq := randSlice(metricSeriesLen)
	sort.Slice(seq, func(i, j int) bool {
		return seq[i] >= seq[j]
	})

	commitDoubleMetric(client, runID, metricName, seq)
}

// addNoisyDecMetrics simulates a real training loss curve: a decaying trend
// with added gaussian noise, rather than a perfectly monotonic sequence.
func addNoisyDecMetrics(client mlsolidv1grpc.MlsolidServiceClient, runID string, metricName string) {
	seq := make([]float64, metricSeriesLen)

	for i := range seq {
		trend := 10.0 * math.Exp(-float64(i)/6.0) //nolint: mnd
		noise := rand.NormFloat64() * 0.2         //nolint: mnd, gosec
		seq[i] = math.Max(0, trend+noise)
	}

	commitDoubleMetric(client, runID, metricName, seq)
}

// addOscillatingMetrics simulates a metric that oscillates around a drifting
// midpoint, such as a cyclical learning rate schedule.
func addOscillatingMetrics(client mlsolidv1grpc.MlsolidServiceClient, runID string, metricName string) {
	seq := make([]float64, metricSeriesLen)

	for i := range seq {
		midpoint := 0.5 - float64(i)/float64(2*metricSeriesLen) //nolint: mnd
		seq[i] = midpoint + 0.3*math.Sin(float64(i)/2.0)        //nolint: mnd
	}

	commitDoubleMetric(client, runID, metricName, seq)
}

// addPlateauMetrics simulates a metric that improves quickly and then
// saturates, such as validation accuracy.
func addPlateauMetrics(client mlsolidv1grpc.MlsolidServiceClient, runID string, metricName string) {
	seq := make([]float64, metricSeriesLen)

	for i := range seq {
		seq[i] = 1 - math.Exp(-float64(i)/4.0) + rand.NormFloat64()*0.01 //nolint: mnd, gosec
	}

	commitDoubleMetric(client, runID, metricName, seq)
}

// addIntMetrics adds a strictly increasing integer-valued metric, such as a
// step or epoch counter.
func addIntMetrics(client mlsolidv1grpc.MlsolidServiceClient, runID string, metricName string) {
	vals := make([]*mlsolidv1.Val, metricSeriesLen)

	for i := range vals {
		vals[i] = &mlsolidv1.Val{Val: &mlsolidv1.Val_Int{Int: int64(i)}}
	}

	commitMetricVals(client, runID, metricName, vals)
}

// addStrMetrics adds a categorical string-valued metric, such as the current
// training phase.
func addStrMetrics(client mlsolidv1grpc.MlsolidServiceClient, runID string, metricName string) {
	phases := []string{"warmup", "train", "eval", "checkpoint"}
	vals := make([]*mlsolidv1.Val, metricSeriesLen)

	for i := range vals {
		phase := phases[rand.Intn(len(phases))] //nolint: gosec
		vals[i] = &mlsolidv1.Val{Val: &mlsolidv1.Val_Str{Str: phase}}
	}

	commitMetricVals(client, runID, metricName, vals)
}

func commitDoubleMetric(client mlsolidv1grpc.MlsolidServiceClient, runID, metricName string, seq []float64) {
	vals := make([]*mlsolidv1.Val, len(seq))

	for i := range vals {
		vals[i] = &mlsolidv1.Val{Val: &mlsolidv1.Val_Double{Double: seq[i]}}
	}

	commitMetricVals(client, runID, metricName, vals)
}

func commitMetricVals(client mlsolidv1grpc.MlsolidServiceClient, runID, metricName string, vals []*mlsolidv1.Val) {
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
