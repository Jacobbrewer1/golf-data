package logging

import (
	"log/slog"

	"github.com/caarlos0/env/v10"
)

type loggingConfig struct {
	Level slog.Level `env:"LOG_LEVEL" envDefault:"debug"`
}

func newLoggingConfig() *loggingConfig {
	logCfg := &loggingConfig{
		Level: slog.LevelDebug,
	}

	_ = env.Parse(logCfg) // We don't care about the error here

	return logCfg
}
