package grpcservice

import (
	"context"
	"fmt"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/rs/zerolog"
	"github.com/zeddo123/mlsolid/solid/controllers"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type authenticated struct{}

func interceptorLogger(l zerolog.Logger) logging.Logger { //nolint: ireturn
	return logging.LoggerFunc(func(_ context.Context, level logging.Level, msg string, fields ...any) {
		l := l.With().Fields(fields).Logger()

		switch level {
		case logging.LevelDebug:
			l.Debug().Msg(msg)
		case logging.LevelInfo:
			l.Info().Msg(msg)
		case logging.LevelWarn:
			l.Warn().Msg(msg)
		case logging.LevelError:
			l.Error().Msg(msg)
		default:
			panic(fmt.Sprintf("unknown level %v", level))
		}
	})
}

func authInterceptor(c *controllers.Controller) auth.AuthFunc {
	return func(ctx context.Context) (context.Context, error) {
		token, err := auth.AuthFromMD(ctx, "bearer")
		if err != nil {
			return nil, err //nolint: wrapcheck
		}

		exists, err := c.IsValidAPIKey(ctx, token)
		if err != nil {
			return nil, status.Error(codes.Internal, "could not check if key is valid")
		}

		if !exists {
			return nil, status.Error(codes.PermissionDenied, "invalid api key")
		}

		return context.WithValue(ctx, authenticated{}, true), nil
	}
}
