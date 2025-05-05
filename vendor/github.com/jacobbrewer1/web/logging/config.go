package logging

import (
	"log/slog"

	"github.com/caarlos0/env/v10"
)

// loggingConfig represents the configuration for logging.
//
// This struct defines the logging level, which can be set via the `LOG_LEVEL`
// environment variable. If the environment variable is not set, it defaults to "debug".
type loggingConfig struct {
	Level slog.Level `env:"LOG_LEVEL" envDefault:"debug"`
}

// newLoggingConfig creates and initializes a new loggingConfig instance.
//
// This function sets the default logging level to "debug" and attempts to parse
// environment variables to override the default values. Errors during parsing
// are ignored.
//
// Returns:
//   - *loggingConfig: A pointer to the initialized logging configuration.
func newLoggingConfig() *loggingConfig {
	logCfg := &loggingConfig{
		Level: slog.LevelDebug,
	}

	_ = env.Parse(logCfg) // We don't care about the error here

	return logCfg
}
