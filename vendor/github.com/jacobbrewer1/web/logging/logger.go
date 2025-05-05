package logging

import (
	"io"
	"log/slog"
	"os"
	"strings"
)

const (
	// bazelGazellePrefix is a constant used to identify the prefix added by Bazel's Gazelle tool.
	//
	// This prefix is used in paths to indicate dependencies managed by Gazelle.
	// It is typically used to trim or process paths in logging or other operations.
	bazelGazellePrefix = "gazelle~~go_deps~"
)

// NewLogger creates a new logger with the default configuration.
//
// This function initializes a logger that writes to the standard output (os.Stdout)
// and applies any provided options to customize the logger's behavior.
//
// Parameters:
//   - opts (...Option): A variadic list of options to configure the logger.
//
// Returns:
//   - *slog.Logger: A pointer to the newly created logger instance.
func NewLogger(opts ...Option) *slog.Logger {
	return NewLoggerWithWriter(os.Stdout, opts...)
}

// NewLoggerWithWriter creates a new logger with the given writer and options.
//
// This function initializes a logger that writes to the specified writer and
// applies any provided options to customize the logger's behavior. It uses a
// JSON handler with source information and a configurable logging level.
//
// Parameters:
//   - writer (io.Writer): The writer to which log messages will be written.
//   - opts (...Option): A variadic list of options to configure the logger.
//
// Returns:
//   - *slog.Logger: A pointer to the newly created logger instance.
func NewLoggerWithWriter(writer io.Writer, opts ...Option) *slog.Logger {
	logCfg := newLoggingConfig()

	l := slog.New(NewDedupeHandler(slog.NewJSONHandler(writer, &slog.HandlerOptions{
		AddSource:   true,
		Level:       logCfg.Level,
		ReplaceAttr: replaceAttrs,
	})))

	for _, opt := range opts {
		opt(l)
	}

	return l
}

// LoggerWithComponent returns a new logger with the given component name.
//
// This function creates a new logger instance by adding a component attribute
// to the provided logger. The component attribute is useful for categorizing
// log messages by their source or purpose.
//
// Parameters:
//   - l (*slog.Logger): The base logger to which the component attribute will be added.
//   - component (string): The name of the component to associate with the logger.
//
// Returns:
//   - *slog.Logger: A new logger instance with the component attribute added.
func LoggerWithComponent(l *slog.Logger, component string) *slog.Logger {
	return l.With(
		slog.String(KeyComponent, component),
	)
}

// replaceAttrs simplifies and adjusts attributes for logging.
//
// This function processes a given slog.Attr and modifies its value based on specific rules:
//   - If the attribute key matches slog.SourceKey, the value is treated as a file path and adjusted.
//   - Converts absolute file paths to relative paths by keeping only the last two segments.
//   - Removes curly braces from the value for compatibility with logstash.
//   - Handles Bazel-compiled paths by extracting and simplifying the repository name and path.
//
// Parameters:
//   - groups ([]string): A slice of group names (not used in this function but part of the signature).
//   - a (slog.Attr): The attribute to be processed and potentially modified.
//
// Returns:
//   - slog.Attr: The modified attribute with adjusted values.
func replaceAttrs(groups []string, a slog.Attr) slog.Attr {
	if a.Key == slog.SourceKey {
		valueStr := a.Value.String()

		// Convert source file path to a relative path.
		parts := strings.Split(valueStr, "/")
		if len(parts) > 2 {
			valueStr = strings.Join(parts[len(parts)-2:], "/")
		}

		// Remove curly braces for compatibility with logstash.
		valueStr = strings.ReplaceAll(valueStr, "{", "")
		valueStr = strings.ReplaceAll(valueStr, "}", "")

		// Handle Bazel-compiled paths.
		if strings.Contains(valueStr, bazelGazellePrefix) {
			trimmedPath := strings.SplitN(valueStr, bazelGazellePrefix, 2)[1]
			pathParts := strings.Split(trimmedPath, "/")
			if len(pathParts) > 0 {
				repoParts := strings.Split(pathParts[0], "_")
				shortName := repoParts[len(repoParts)-1]
				valueStr = shortName
				if len(pathParts) > 1 {
					valueStr += "/" + strings.Join(pathParts[1:], "/")
				}
			}
		}

		a.Value = slog.StringValue(valueStr)
	}
	return a
}
