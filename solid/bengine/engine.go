// Package bengine handles running benchmarks.
package bengine

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"

	"github.com/docker/cli/opts"
	ctr "github.com/docker/go-sdk/container"
	"github.com/moby/moby/api/pkg/authconfig"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/registry"
	"github.com/moby/moby/client"
	"github.com/rs/zerolog"
	"github.com/zeddo123/mlsolid/solid/s3"
	"github.com/zeddo123/mlsolid/solid/store"
	"github.com/zeddo123/mlsolid/solid/types"
	"github.com/zeddo123/pubgo"
)

// Opts handler function for setting engine configuration.
type Opts func(cfg *Config)

// Engine is a benchmark runner with docker containers.
type Engine struct {
	store            *store.RedisStore
	sub              *pubgo.Subscription
	s3               s3.ObjectStore
	registryUsername string
	registryPassword string
	datasetsDest     string
	checkpointsDest  string
	l                zerolog.Logger
}

// Config struct for a bengine instance.
type Config struct {
	Store            *store.RedisStore
	Sub              *pubgo.Subscription
	S3               s3.ObjectStore
	RegistryUsername string
	RegistryPassword string
	DatasetsDest     string
	CheckpointsDest  string
	LoggingLevel     zerolog.Level
	HumanReadable    bool
}

func defaultOpts() Config {
	return Config{ //nolint: exhaustruct
		DatasetsDest:    "/mlsolid/datasets/",
		CheckpointsDest: "/mlsolid/checkpoints/",
		LoggingLevel:    zerolog.InfoLevel,
	}
}

// WithHumanReadableLogs disables structured logging.
func WithHumanReadableLogs() Opts {
	return func(cfg *Config) {
		cfg.HumanReadable = true
	}
}

// WithLoggingLevel sets logging level.
func WithLoggingLevel(lvl zerolog.Level) Opts {
	return func(cfg *Config) {
		cfg.LoggingLevel = lvl
	}
}

// WithDatasetsDest sets datasets path.
func WithDatasetsDest(dest string) Opts {
	return func(cfg *Config) {
		cfg.DatasetsDest = dest
	}
}

// WithCheckpointsDest sets checkpoints path.
func WithCheckpointsDest(dest string) Opts {
	return func(cfg *Config) {
		cfg.CheckpointsDest = dest
	}
}

// WithRegistryCreds sets the credentials of the docker registry to pull images from.
func WithRegistryCreds(username, password string) Opts {
	return func(cfg *Config) {
		cfg.RegistryUsername = username
		cfg.RegistryPassword = password
	}
}

// WithS3 enables pulling datasets from an S3 bucket.
func WithS3(store s3.ObjectStore) Opts {
	return func(cfg *Config) {
		cfg.S3 = store
	}
}

// WithRedisStore sets the redis store used to record benchmarking runs.
// If no redis store is provided, saving is skipped.
func WithRedisStore(store *store.RedisStore) Opts {
	return func(cfg *Config) {
		cfg.Store = store
	}
}

// New creates a new benchmark engine.
func New(sub *pubgo.Subscription, opts ...Opts) *Engine {
	cfg := defaultOpts()
	for _, fn := range opts {
		fn(&cfg)
	}

	var output io.Writer

	output = os.Stderr
	if cfg.HumanReadable {
		output = zerolog.ConsoleWriter{Out: os.Stderr} //nolint: exhaustruct
	}

	logger := zerolog.New(output).
		Level(cfg.LoggingLevel).With().
		Str("Layer", "bENGINE").
		Timestamp().Logger()

	return &Engine{
		store:            cfg.Store,
		sub:              sub,
		s3:               cfg.S3,
		registryUsername: cfg.RegistryUsername,
		registryPassword: cfg.RegistryPassword,
		checkpointsDest:  cfg.CheckpointsDest,
		datasetsDest:     cfg.DatasetsDest,
		l:                logger,
	}
}

// Start starts the engine instance.
func (e *Engine) Start(ctx context.Context) {
	cli, err := client.New(client.FromEnv)
	if err != nil {
		e.l.Error().Err(err).Msg("could not setup docker client")

		return
	}

	e.l.Info().Msg("listening to benchmarking events...")

	for {
		select {
		case msg, ok := <-e.sub.Msgs:
			if !ok {
				return
			}

			event, ok := msg.(*types.BenchEvent)
			if !ok {
				break
			}

			e.l.Info().
				Str("benchmark", event.BenchName).
				Str("registry", event.Registry).
				Str("version", event.Version).
				Str("image", event.DockerImage).
				Str("model", event.ModelURL).
				Str("dataset", event.DatasetName).
				Str("datasetURL", event.DatasetURL).
				Str("fromS3", strconv.FormatBool(event.FromS3)).
				Str("autoTag", strconv.FormatBool(event.AutoTag)).
				Str("tag", event.Tag).
				Msg("received benchmark event")

			err := e.ConsumeEvent(ctx, cli, event)
			if err != nil {
				e.l.Error().Err(err).Msg("could not run benchmark")
			}

		case <-ctx.Done():
			e.l.Info().Msg("shutting down bEngine")
		}
	}
}

// ConsumeEvent handles a benchmarking event.
func (e *Engine) ConsumeEvent(ctx context.Context, cli *client.Client, event *types.BenchEvent) error {
	err := e.pullImage(ctx, cli, event.DockerImage)
	if err != nil {
		e.l.Error().Err(err).Msg("could not pull docker image")

		return err
	}

	datasetPath := filepath.Join(e.datasetsDest, event.DatasetName)

	e.l.Info().Str("datasetPath", datasetPath).
		Msg("checking if dataset is already present")

	// Load dataset if not present
	if _, err := os.Stat(datasetPath); errors.Is(err, os.ErrNotExist) {
		// Downloading dataset
		err := e.PullDataset(ctx, event.DatasetURL, datasetPath, event.FromS3)
		if err != nil {
			return err
		}
	}

	// Load model if not present
	checkpointPath := filepath.Join(e.checkpointsDest, "model.pth")

	result, err := e.RunContainer(ctx, event.DockerImage,
		event.DatasetName, datasetPath, checkpointPath)
	if err != nil {
		return err
	}

	e.l.Info().Str("result", result).Msg("container exited successfully")

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
			e.l.Error().Err(err).Msg("could not encode docker registry creds")
		} else {
			opts.RegistryAuth = authStr
		}
	}

	_, err := cli.ImagePull(ctx, image, opts)
	if err != nil {
		return fmt.Errorf("could not pull image: %w", err)
	}

	return nil
}

// PullDatasetFromS3 pulls a dataset from a S3 object.
func (e *Engine) PullDatasetFromS3(url string, outputPath string) error {
	return nil
}

// PullDataset pulls a dataset from a public source with http.
func (e *Engine) PullDataset(ctx context.Context, url string, outputPath string, fromS3 bool) error {
	if fromS3 {
		return e.PullDatasetFromS3(url, outputPath)
	}

	fileName := path.Base(url)

	e.l.Info().Str("url", url).Msg("Downloading dataset")

	resp, err := http.Get(url) //nolint: gosec, noctx
	if err != nil {
		return fmt.Errorf("failed requesting file: %w", err)
	}

	defer resp.Body.Close() //nolint: errcheck

	e.l.Debug().Str("filename", fileName).Msg("creating temp file for dataset")

	fs, err := os.CreateTemp(os.TempDir(), "*."+fileName)
	if err != nil {
		return fmt.Errorf("could not create tmp file: %w", err)
	}

	defer fs.Close() //nolint: errcheck

	tmpPath := fs.Name()

	e.l.Debug().Str("tmpPath", tmpPath).Msg("Saving dataset to temp file")

	_, err = io.Copy(fs, resp.Body)
	if err != nil {
		return fmt.Errorf("could not write content to tmp file %s: %w", tmpPath, err)
	}

	defer os.Remove(tmpPath) //nolint: errcheck

	e.l.Debug().Msg("seeking to start of file descriptor")

	_, err = fs.Seek(0, io.SeekStart)
	if err != nil {
		return fmt.Errorf("could not rewind to start of tmpFile: %w", err)
	}

	e.l.Info().Str("outputPath", outputPath).Msg("extracting archive")

	err = ExtractArchiveFromReader(ctx, outputPath, fileName, fs)
	if err != nil {
		return fmt.Errorf("could not extract archive: %w", err)
	}

	return nil
}

// RunContainer runs a benchmark on a container with a specified image, dataset, and checkpoint.
func (e *Engine) RunContainer(ctx context.Context, image, datasetName, datasetPath, checkpointPath string,
) (string, error) {
	outputPath := "/run/output.json"

	gpuOpts := opts.GpuOpts{}

	err := gpuOpts.Set("all")
	if err != nil {
		return "", errors.New("could not set GpuOpts")
	}

	e.l.Info().
		Str("image", image).
		Str("dataset", datasetPath).
		Str("checkpoint", checkpointPath).
		Msg("starting container")

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

	e.l.Debug().
		Str("container", c.ShortID()).
		Msg("waiting for container to exit")

	wait := c.Client().ContainerWait(ctx, c.ID(), client.ContainerWaitOptions{}) //nolint: exhaustruct
	select {
	case err := <-wait.Error:
		if err != nil {
			return "", fmt.Errorf("container %q exited with error: %w", c.ShortID(), err)
		}
	case <-wait.Result:
	}

	e.l.Debug().
		Str("container", c.ShortID()).
		Msg("copying benchmark results from container")

	reader, err := c.CopyFromContainer(ctx, outputPath)
	if err != nil {
		return "", fmt.Errorf("could not read output file %q from container %q: %w", outputPath, c.ShortID(), err)
	}

	defer reader.Close() //nolint: errcheck

	results, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("could not read output file content: %w", err)
	}

	e.l.Debug().
		Str("container", c.ShortID()).
		Msg("removing container from docker")

	err = c.Terminate(ctx)
	if err != nil {
		e.l.Error().
			Err(err).
			Str("container", c.ShortID()).
			Msg("could not terminate container")
	}

	return string(results), nil
}
