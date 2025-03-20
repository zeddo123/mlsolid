//go:build integrationtests

package controllers_test

import (
	"context"
	"log"
	"os"
	"testing"

	redisv9 "github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/redis"
)

var client *redisv9.Client

func TestMain(m *testing.M) {
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

	code := m.Run()

	if err := testcontainers.TerminateContainer(redisContainer); err != nil {
		log.Printf("failed to terminate container: %s\n", err)
	}

	os.Exit(code)
}
