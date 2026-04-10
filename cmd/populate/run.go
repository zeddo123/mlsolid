package main

import (
	"context"
	"errors"
	"io"
	"log"
	"math/rand"
	"net/http"

	"buf.build/gen/go/zeddo123/mlsolid/grpc/go/mlsolid/v1/mlsolidv1grpc"
	mlsolidv1 "buf.build/gen/go/zeddo123/mlsolid/protocolbuffers/go/mlsolid/v1"
	"github.com/anandvarma/namegen"
)

func createRun(client mlsolidv1grpc.MlsolidServiceClient, expID string) {
	id := namegen.New().Get()

	resp, err := client.CreateRun(context.Background(), &mlsolidv1.CreateRunRequest{
		RunId:        id,
		ExperimentId: expID,
	})
	if err != nil {
		panic(err)
	}

	log.Printf("[populate]: run created runId=%s expId=%s \n", resp.GetRunId(), expID)

	addDescMetrics(client, resp.GetRunId(), "loss")
	addRandMetrics(client, resp.GetRunId(), "mae")
	addIncMetrics(client, resp.GetRunId(), "acc")

	addTxtAtrifact(client, resp.GetRunId(), "log.txt")
	addTxtAtrifact(client, resp.GetRunId(), "data.txt")

	addModelArtifact(client, resp.GetRunId(), "yolo.pt")

	tags := make([]string, 0, 1)
	if rand.Intn(2) == 1 {
		tags = append(tags, "latest")
	}

	_, err = client.AddModelEntry(context.Background(), &mlsolidv1.AddModelEntryRequest{
		Name:       "Yolo Prod",
		RunId:      resp.GetRunId(),
		ArtifactId: "yolo.pt",
		Tags:       tags,
	})
	if err != nil {
		panic(err)
	}

	log.Printf("[populate]: added model to registry runId=%s expId=%s tags=%v \n", resp.GetRunId(), expID, tags)
}

func addModelArtifact(client mlsolidv1grpc.MlsolidServiceClient, runID, modelName string) {
	log.Printf("[populate]: adding model artifact... runId=%s artifact=%s \n", runID, modelName)

	resp, err := http.Get("https://huggingface.co/Ultralytics/YOLO11/resolve/d3043e98a1ad0e2956728c13cf1e041e0fa4220f/yolo11s.pt")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	log.Printf("[populate] Sending model artifact... \n")

	stream, err := client.AddArtifact(context.Background())
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

	log.Printf("[populate] Sending chunks artifact... \n")
	for {
		_, err := resp.Body.Read(buffer)
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			log.Printf("[ERROR] [populate] could not read next chuck... \n")
			panic(err)
		}

		err = stream.Send(&mlsolidv1.AddArtifactRequest{Request: &mlsolidv1.AddArtifactRequest_Content{
			Content: &mlsolidv1.Content{
				Content: buffer,
			},
		}})
		if err != nil {
			log.Printf("[ERROR] [populate] could not send next chuck... \n")
			break
		}
	}

	r, err := stream.CloseAndRecv()
	if err != nil {
		panic(err)
	}

	if r.GetStatus() == mlsolidv1.Status_STATUS_FAILED {
		log.Panicf("[ERRO] [populate] could not add artifact status=FAILED\n")
	}

	log.Printf("[populate]: added model artifact runId=%s model=%s \n", runID, modelName)
}

func addTxtAtrifact(client mlsolidv1grpc.MlsolidServiceClient, runID, artifactName string) {
	log.Printf("[populate]: adding artifact... runId=%s artifact=%s \n", runID, artifactName)

	stream, err := client.AddArtifact(context.Background())
	if err != nil {
		panic(err)
	}

	data := []byte("logs...\n")

	err = stream.Send(&mlsolidv1.AddArtifactRequest{Request: &mlsolidv1.AddArtifactRequest_Metadata{
		Metadata: &mlsolidv1.MetaData{
			Name:  artifactName,
			Type:  "content-type/text",
			RunId: runID,
		},
	}})
	if err != nil {
		panic(err)
	}

	err = stream.Send(&mlsolidv1.AddArtifactRequest{Request: &mlsolidv1.AddArtifactRequest_Content{
		Content: &mlsolidv1.Content{
			Content: data,
		},
	}})
	if err != nil {
		panic(err)
	}

	_, err = stream.CloseAndRecv()
	if err != nil {
		panic(err)
	}

	log.Printf("[populate]: added artifact runId=%s artifact=%s \n", runID, artifactName)
}
