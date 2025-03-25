package logging

import "log/slog"

type Option func(l *slog.Logger)

// WithDefaultLogger is a LoggingOption that sets up the default logger.
func WithDefaultLogger() Option {
	return func(l *slog.Logger) {
		slog.SetDefault(l)
	}
}

// WithAppName is a LoggingOption that sets up the application name for the logger.
func WithAppName(appName string) Option {
	return func(l *slog.Logger) {
		*l = *l.With(
			slog.String(KeyAppName, appName),
		)
	}
}

// WithComponent is a LoggingOption that sets up the component for the logger.
func WithComponent(component string) Option {
	return func(l *slog.Logger) {
		*l = *LoggerWithComponent(l, component)
	}
}
