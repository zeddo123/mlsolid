// Package bengine handles running benchmarks.
package bengine

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/docker/cli/opts"
	ctr "github.com/docker/go-sdk/container"
	"github.com/moby/moby/api/pkg/authconfig"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/registry"
	"github.com/moby/moby/client"
	"github.com/zeddo123/mlsolid/solid/s3"
	"github.com/zeddo123/mlsolid/solid/store"
	"github.com/zeddo123/mlsolid/solid/types"
	"github.com/zeddo123/pubgo"
)

// Engine is a benchmark runner with docker containers.
type Engine struct {
	store            *store.RedisStore
	sub              *pubgo.Subscription
	s3               s3.ObjectStore
	registryUsername string
	registryPassword string
}

// New creates a new benchmark engine.
func New(ctx context.Context, store *store.RedisStore,
	sub *pubgo.Subscription, registryUsername, registryPassword string,
) *Engine {
	e := &Engine{
		store:            store,
		sub:              sub,
		registryUsername: registryUsername,
		registryPassword: registryPassword,
	}

	go e.start(ctx)

	return e
}

func (e *Engine) start(ctx context.Context) {
	cli, err := client.New(client.FromEnv)
	if err != nil {
		log.Println("[BENGINE] Could not setup docker client")

		return
	}

	log.Println("[BENGINE] Listening to benchmark events...")

	for {
		select {
		case msg := <-e.sub.Msgs:
			event, ok := msg.(*types.BenchEvent)
			if !ok {
				break
			}

			log.Printf("[BENGINE] received event=%v+\n", event)

			err := e.consumeEvent(ctx, cli, event)
			if err != nil {
				log.Println("[BENGINE] could not run benchmark", err)
			}

		case <-ctx.Done():
			log.Println("shutting down engine")
		}
	}
}

func (e *Engine) consumeEvent(ctx context.Context, cli *client.Client, event *types.BenchEvent) error {
	err := e.pullImage(ctx, cli, event.DockerImage)
	if err != nil {
		log.Println("[BENGINE] could not pull docker image", err)

		return err
	}

	// Load dataset if not present
	datasetPath := fmt.Sprintf("/mlsolid/datasets/%s", event.DatasetName)
	if _, err := os.Stat(datasetPath); errors.Is(err, os.ErrNotExist) {
		// Downloading dataset
		if event.FromS3 {
			err := PullDatasetFromS3(event.DatasetURL, datasetPath)
			if err != nil {
				return err
			}
		} else {
			err := PullDataset(ctx, event.DatasetURL, datasetPath)
			if err != nil {
				return err
			}
		}
	}

	// Load model if not present
	checkpointPath := "/mlsolid/checkpoints/model.pt"

	result, err := RunContainer(ctx, event.DockerImage,
		event.DatasetName, datasetPath, checkpointPath)
	if err != nil {
		return err
	}

	log.Println("[BENGINE] Container finished running", result)

	return nil
}

func (e *Engine) pullImage(ctx context.Context, cli *client.Client, image string) error {
	opts := client.ImagePullOptions{} //nolint: exhaustruct

	if e.registryUsername != "" {
		authStr, err := authconfig.Encode(registry.AuthConfig{ //nolint: exhaustruct
			Username: e.registryUsername,
			Password: e.registryPassword,
		})
		if err != nil {
			log.Println("could not encode docker registry creds", err)
		} else {
			opts.RegistryAuth = authStr
		}
	}

	_, err := cli.ImagePull(ctx, image, opts)

	return err //nolint: wrapcheck
}

// PullDatasetFromS3 pulls a dataset from a S3 object.
func PullDatasetFromS3(url string, outputPath string) error {
	return nil
}

// PullDataset pulls a dataset from a public source with http.
func PullDataset(ctx context.Context, url string, outputPath string) error {
	fileName := path.Base(url)

	resp, err := http.Get(url) //nolint: gosec, noctx
	if err != nil {
		return fmt.Errorf("failed requesting file: %w", err)
	}

	defer resp.Body.Close() //nolint: errcheck

	fs, err := os.CreateTemp(os.TempDir(), "*."+fileName)
	if err != nil {
		return fmt.Errorf("could not create tmp file: %w", err)
	}

	defer fs.Close() //nolint: errcheck

	tmpPath := fs.Name()

	_, err = io.Copy(fs, resp.Body)
	if err != nil {
		return fmt.Errorf("could not write content to tmp file %s: %w", tmpPath, err)
	}

	defer os.Remove(tmpPath) //nolint: errcheck

	_, err = fs.Seek(0, io.SeekStart)
	if err != nil {
		return fmt.Errorf("could not rewind to start of tmpFile: %w", err)
	}

	err = ExtractArchiveFromReader(ctx, outputPath, fileName, fs)
	if err != nil {
		return fmt.Errorf("could not extract archive: %w", err)
	}

	return nil
}

// RunContainer runs a benchmark on a container with a specified image, dataset, and checkpoint.
func RunContainer(ctx context.Context, image, datasetName, datasetPath, checkpointPath string,
) (string, error) {
	outputPath := "/run/output.json"

	gpuOpts := opts.GpuOpts{}

	err := gpuOpts.Set("all")
	if err != nil {
		return "", errors.New("could not set GpuOpts")
	}

	log.Printf("[BENGINE] Starting container image=%q dataset=%q checkpoint=%q\n", image, datasetName, checkpointPath)

	c, err := ctr.Run(
		ctx,
		ctr.WithImage(image),
		ctr.WithCmd([]string{
			"-dn", datasetName,
			"-d", datasetPath,
			"-m", checkpointPath,
			"-o", outputPath,
		}...),
		ctr.WithHostConfigModifier(func(hostConfig *container.HostConfig) {
			hostConfig.Resources = container.Resources{ //nolint: exhaustruct
				DeviceRequests: gpuOpts.Value(),
			}
		}),
	)
	if err != nil {
		return "", fmt.Errorf("could not exec container %q: %w", image, err)
	}

	wait := c.Client().ContainerWait(ctx, c.ID(), client.ContainerWaitOptions{}) //nolint: exhaustruct
	select {
	case err := <-wait.Error:
		if err != nil {
			return "", fmt.Errorf("container %q exited with error: %w", c.ShortID(), err)
		}
	case <-wait.Result:
	}

	log.Println("[BENGINE] Copying benchmark results from container", c.ShortID())

	reader, err := c.CopyFromContainer(ctx, outputPath)
	if err != nil {
		return "", fmt.Errorf("could not read output file %q from container %q: %w", outputPath, c.ShortID(), err)
	}

	defer reader.Close() //nolint: errcheck

	results, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("could not read output file content: %w", err)
	}

	log.Println("[BENGINE] Removing container...", c.ShortID())

	err = c.Terminate(ctx)
	if err != nil {
		log.Printf("[BENGINE] could not terminate container %q: %v\n", c.ShortID(), err)
	}

	return string(results), nil
}
