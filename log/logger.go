package log

import (
	"os"

	"github.com/rs/zerolog"
)

// DefaultLogger is a functional logger that can be used ad-hoc.
var DefaultLogger zerolog.Logger

func init() {
	DefaultLogger = zerolog.New(os.Stdout).With().Timestamp().Logger().Level(zerolog.InfoLevel)
}
