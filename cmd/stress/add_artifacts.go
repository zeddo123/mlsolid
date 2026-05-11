package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"

	"buf.build/gen/go/zeddo123/mlsolid/grpc/go/mlsolid/v1/mlsolidv1grpc"
	mlsolidv1 "buf.build/gen/go/zeddo123/mlsolid/protocolbuffers/go/mlsolid/v1"
	"github.com/anandvarma/namegen"
)

func addArtifacts(ctx context.Context, client mlsolidv1grpc.MlsolidServiceClient, workers int, url string) {
	var wg sync.WaitGroup

	id := namegen.New().Get()

	resp, err := client.CreateRun(ctx, &mlsolidv1.CreateRunRequest{
		RunId:        id,
		ExperimentId: "add_artifact_stress_test",
	})
	if err != nil {
		panic(err)
	}

	runID := resp.GetRunId()

	count := 0

	for {
		for wid := range workers {
			wg.Go(func() {
				addModelArtifact(ctx, client, runID, fmt.Sprintf("%d_%d", wid, count), url)
			})
		}

		wg.Wait()

		count++
	}
}

func addModelArtifact(ctx context.Context, client mlsolidv1grpc.MlsolidServiceClient, runID, modelName, url string) {
	log.Printf("[stress]: adding model artifact... runId=%s artifact=%s \n", runID, modelName)

	resp, err := http.Get(url) //nolint: noctx, gosec
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close() //nolint: errcheck

	log.Printf("[stress] Sending model artifact... \n")

	stream, err := client.AddArtifact(ctx)
	if err != nil {
		panic(err)
	}

	err = stream.Send(&mlsolidv1.AddArtifactRequest{Request: &mlsolidv1.AddArtifactRequest_Metadata{
		Metadata: &mlsolidv1.MetaData{
			Name:  modelName,
			Type:  "content-type/model",
			RunId: runID,
		},
	}})
	if err != nil {
		panic(err)
	}

	chunck := 1024
	buffer := make([]byte, chunck)

	log.Printf("[stress] Sending chunks artifact... \n")

	for {
		_, err := resp.Body.Read(buffer)
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			log.Printf("[ERROR] [stress] could not read next chuck... \n")
			panic(err)
		}

		err = stream.Send(&mlsolidv1.AddArtifactRequest{Request: &mlsolidv1.AddArtifactRequest_Content{
			Content: &mlsolidv1.Content{
				Content: buffer,
			},
		}})
		if err != nil {
			log.Printf("[ERROR] [stress] could not send next chuck... \n")

			break
		}
	}

	r, err := stream.CloseAndRecv()
	if err != nil {
		panic(err)
	}

	if r.GetStatus() == mlsolidv1.Status_STATUS_FAILED {
		log.Panicf("[ERRO] [stress] could not add artifact status=FAILED\n")
	}

	log.Printf("[stress]: added model artifact runId=%s model=%s \n", runID, modelName)
}
