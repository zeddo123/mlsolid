package grpcservice

import (
	"context"
	"fmt"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/rs/zerolog"
)

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
