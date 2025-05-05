package logging

import "log/slog"

// Option defines a function type that modifies a slog.Logger instance.
//
// This type is used to apply various configurations or settings to a logger.
type Option = func(l *slog.Logger)

// WithDefaultLogger is a LoggingOption that sets the provided logger as the default logger.
//
// This function returns an Option that, when applied, sets the given slog.Logger
// instance as the default logger for the application.
//
// Returns:
//   - Option: A function that sets the default logger.
func WithDefaultLogger() Option {
	return func(l *slog.Logger) {
		slog.SetDefault(l)
	}
}

// WithAppName is a LoggingOption that adds the application name to the logger.
//
// This function returns an Option that, when applied, adds an attribute to the logger
// with the key "KeyAppName" and the provided application name.
//
// Parameters:
//   - appName (string): The name of the application to associate with the logger.
//
// Returns:
//   - Option: A function that adds the application name to the logger.
func WithAppName(appName string) Option {
	return func(l *slog.Logger) {
		*l = *l.With(
			slog.String(KeyAppName, appName),
		)
	}
}

// WithComponent is a LoggingOption that adds a component name to the logger.
//
// This function returns an Option that, when applied, adds a component attribute
// to the logger using the LoggerWithComponent function.
//
// Parameters:
//   - component (string): The name of the component to associate with the logger.
//
// Returns:
//   - Option: A function that adds the component name to the logger.
func WithComponent(component string) Option {
	return func(l *slog.Logger) {
		*l = *LoggerWithComponent(l, component)
	}
}
