//go:build integrationtests

package controllers_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	miniov7 "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	redisv9 "github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/minio"
	"github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/zeddo123/mlsolid/solid/s3"
)

var (
	client      *redisv9.Client
	objectStore s3.Store
)

func TestMain(m *testing.M) {
	minioContainer, err := minio.Run(context.Background(),
		"minio/minio:RELEASE.2024-01-16T16-07-38Z",
		minio.WithUsername("access_key"),
		minio.WithPassword("secret_key"))
	defer func() {
		if err := testcontainers.TerminateContainer(minioContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()

	if err != nil {
		log.Printf("failed to start container: %s\n", err)
		panic(err)
	}

	minioConn, err := minioContainer.ConnectionString(context.Background())
	if err != nil {
		panic(err)
	}

	log.Printf("minio connection %s", minioConn)

	minioClient, err := miniov7.New(minioConn, &miniov7.Options{
		Creds: credentials.NewStaticV4("access_key", "secret_key", ""),
	})
	if err != nil {
		panic(err)
	}

	err = minioClient.MakeBucket(context.Background(), "mlsolid", miniov7.MakeBucketOptions{
		Region: "us-east-1",
	})
	if err != nil {
		panic(err)
	}

	log.Println("minio bucket created!")

	redisContainer, err := redis.Run(context.Background(),
		"redis:latest",
		redis.WithLogLevel(redis.LogLevelVerbose),
	)
	if err != nil {
		log.Printf("failed to start container: %s\n", err)
		panic(err)
	}

	conn, err := redisContainer.ConnectionString(context.Background())
	if err != nil {
		panic(err)
	}

	opts, err := redisv9.ParseURL(conn)
	if err != nil {
		panic(err)
	}

	client = redisv9.NewClient(opts)

	if err = client.Ping(context.Background()).Err(); err != nil {
		panic(err)
	}

	objectStore, err = s3.NewStore(s3.StoreOps{
		Bucket:          "mlsolid",
		Endpoint:        fmt.Sprintf("http://%s", minioConn),
		AccessKey:       "access_key",
		SecretAccessKey: "secret_key",
		Region:          "us-east-1",
		Prefix:          "mlsolid",
	})
	if err != nil {
		panic(err)
	}

	code := m.Run()

	if err := testcontainers.TerminateContainer(redisContainer); err != nil {
		log.Printf("failed to terminate container: %s\n", err)
	}

	os.Exit(code)
}
