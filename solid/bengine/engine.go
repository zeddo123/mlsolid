// Package bengine handles running benchmarks.
package bengine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/docker/cli/opts"
	ctr "github.com/docker/go-sdk/container"
	"github.com/moby/moby/api/pkg/authconfig"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/mount"
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
	rootDest         string
	hostSourceVolume string
	l                zerolog.Logger
}

// Config struct for a bengine instance.
type Config struct {
	Store            *store.RedisStore
	Sub              *pubgo.Subscription
	S3               s3.ObjectStore
	RegistryUsername string
	RegistryPassword string
	RootDest         string
	LoggingLevel     zerolog.Level
	HumanReadable    bool
	hostSourceVolume string
}

func defaultOpts() Config {
	return Config{ //nolint: exhaustruct
		RootDest:     "/mlsolid/",
		LoggingLevel: zerolog.InfoLevel,
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

// WithRootDest sets root destination where datasets & checkpoints are saved.
func WithRootDest(dest string) Opts {
	return func(cfg *Config) {
		path, err := filepath.Abs(dest)
		if err != nil {
			panic(err)
		}

		cfg.RootDest = path
	}
}

// WithHostSourceVolume sets host source volume if service is running inside a container.
func WithHostSourceVolume(source string) Opts {
	return func(cfg *Config) {
		cfg.hostSourceVolume = source
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
		rootDest:         cfg.RootDest,
		hostSourceVolume: cfg.hostSourceVolume,
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
				Int64("version", event.Version).
				Str("image", event.DockerImage).
				Str("model", event.ModelURL).
				Str("dataset", event.DatasetName).
				Str("datasetURL", event.DatasetURL).
				Bool("fromS3", event.FromS3).
				Bool("autoTag", event.AutoTag).
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

	datasetPath := filepath.Join(e.rootDest, "datasets", event.DatasetName)

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
	checkpointPath := filepath.Join(e.rootDest, "checkpoints", "model.pth")

	result, err := e.RunContainer(ctx, event.DockerImage,
		event.DatasetName, datasetPath, checkpointPath)
	if err != nil {
		return err
	}

	e.l.Info().Str("result", result).Msg("container exited successfully")

	if e.store == nil {
		e.l.Info().Str("result", result).Msg("Redis store not configured skipping")

		return nil
	}

	return e.RecordRun(ctx, event, result)
}

// PullDatasetFromS3 pulls a dataset from a S3 object.
func (e *Engine) PullDatasetFromS3(ctx context.Context, url string, outputPath string) error {
	if e.s3 == nil {
		return fmt.Errorf("could not pull obj: s3 store not configured")
	}

	fileName := path.Base(url)

	content, err := e.s3.DownloadURL(ctx, url)
	if err != nil {
		return fmt.Errorf("could not download obj: %w", err)
	}

	defer content.Close() //nolint: errcheck

	err = e.extractArchive(ctx, fileName, content, outputPath)
	if err != nil {
		return fmt.Errorf("could not extract archive: %w", err)
	}

	return nil
}

// PullDataset pulls a dataset from a public source with http.
func (e *Engine) PullDataset(ctx context.Context, url string, outputPath string, fromS3 bool) error {
	if fromS3 {
		return e.PullDatasetFromS3(ctx, url, outputPath)
	}

	fileName := path.Base(url)

	e.l.Info().Str("url", url).Msg("Downloading dataset")

	resp, err := http.Get(url) //nolint: gosec, noctx
	if err != nil {
		return fmt.Errorf("failed requesting file: %w", err)
	}

	defer resp.Body.Close() //nolint: errcheck

	err = e.extractArchive(ctx, fileName, resp.Body, outputPath)
	if err != nil {
		return fmt.Errorf("could not extract archive: %w", err)
	}

	return nil
}

func (e *Engine) extractArchive(ctx context.Context, fileName string, content io.ReadCloser, outputPath string) error {
	e.l.Debug().Str("filename", fileName).Msg("creating temp file for dataset")

	fs, err := os.CreateTemp(os.TempDir(), "*."+fileName)
	if err != nil {
		return fmt.Errorf("could not create tmp file: %w", err)
	}

	defer fs.Close() //nolint: errcheck

	tmpPath := fs.Name()

	e.l.Debug().Str("tmpPath", tmpPath).Msg("Saving dataset to temp file")

	_, err = io.Copy(fs, content)
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

	source := e.rootDest
	target := e.rootDest

	if e.hostSourceVolume != "" {
		source = e.hostSourceVolume
	}

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
			hostConfig.Mounts = []mount.Mount{
				{Type: mount.TypeBind, Source: source, Target: target, ReadOnly: true}, //nolint: exhaustruct
			}
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

	logs, err := c.Logs(ctx)
	if err != nil {
		e.l.Error().Err(err).
			Str("container", c.ShortID()).
			Msg("could not pull logs from container")
	}
	defer logs.Close() //nolint: errcheck

	logsContent, err := io.ReadAll(logs)
	if err != nil {
		e.l.Error().Err(err).
			Str("container", c.ShortID()).
			Msg("could not read container logs")
	}

	e.l.Debug().
		Str("container", c.ShortID()).
		Str("logs", string(logsContent)).
		Msg("docker container logs")

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

// RecordRun records a run into the store.
func (e *Engine) RecordRun(ctx context.Context, event *types.BenchEvent, result string) error {
	metrics := make(map[string]float64)

	e.l.Debug().Str("result", result).Msg("unmarshalling result")

	err := json.Unmarshal([]byte(result), &metrics)
	if err != nil {
		return fmt.Errorf("could not unmarshal results: %w", err)
	}

	e.l.Info().
		Str("result", result).
		Str("benchName", event.BenchName).
		Str("Registry", event.Registry).
		Int("Version", int(event.Version)).
		Msg("recording bench run into the store")

	err = e.store.RecordRuns(ctx, event.BenchName, []types.BenchRun{{
		Registry:  event.Registry,
		Version:   event.Version,
		Metrics:   metrics,
		Timestamp: time.Now(),
	}})
	if err != nil {
		return fmt.Errorf("could not record run into the store: %w", err)
	}

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
