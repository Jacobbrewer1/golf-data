package logging

import "log/slog"

// Option defines a function type that modifies a slog.Logger instance.
type Option = func(l *slog.Logger)

// WithDefaultLogger is a LoggingOption that sets the provided logger as the default logger.
func WithDefaultLogger() Option {
	return func(l *slog.Logger) {
		slog.SetDefault(l)
	}
}

// WithAppName is a LoggingOption that adds the application name to the logger.
func WithAppName(appName string) Option {
	return func(l *slog.Logger) {
		*l = *l.With(
			slog.String(KeyAppName, appName),
		)
	}
}

// WithComponent is a LoggingOption that adds a component name to the logger.
func WithComponent(component string) Option {
	return func(l *slog.Logger) {
		*l = *LoggerWithComponent(l, component)
	}
}
