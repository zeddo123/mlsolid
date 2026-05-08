package logger

import (
	"io"
	"os"

	"github.com/rs/zerolog"
)

// New creates a new logger using zerolog package.
func New(humanReadable bool) zerolog.Logger {
	var out io.Writer

	out = os.Stderr

	if humanReadable {
		out = zerolog.ConsoleWriter{Out: os.Stderr} //nolint: exhaustruct
	}

	return zerolog.New(out).With().Timestamp().Str("app", "[MLSolid]").Logger()
}

// NewSub creates a sub logger from an existing one and prepends the component name.
func NewSub(logger zerolog.Logger, component string) zerolog.Logger {
	return logger.With().Str("component", component).Logger()
}
