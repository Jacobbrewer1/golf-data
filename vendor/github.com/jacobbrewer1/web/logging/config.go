package logging

import (
	"log/slog"

	"github.com/caarlos0/env/v10"
)

// loggingConfig represents the configuration for logging.
type loggingConfig struct {
	Level slog.Level `env:"LOG_LEVEL" envDefault:"debug"`
}

// newLoggingConfig creates and initializes a new loggingConfig instance.
func newLoggingConfig() *loggingConfig {
	logCfg := &loggingConfig{
		Level: slog.LevelInfo,
	}

	_ = env.Parse(logCfg) // We don't care about the error here

	return logCfg
}
