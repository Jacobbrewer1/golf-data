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
func NewLogger(opts ...Option) *slog.Logger {
	return NewLoggerWithWriter(os.Stdout, opts...)
}

// NewLoggerWithWriter creates a new logger with the given writer and options.
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
