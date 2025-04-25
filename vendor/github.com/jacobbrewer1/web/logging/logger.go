package logging

import (
	"io"
	"log/slog"
	"os"
	"strings"
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

// replaceAttrs is a slog.HandlerOptions.ReplaceAttr function that replaces some attributes.
func replaceAttrs(groups []string, a slog.Attr) slog.Attr {
	if a.Key == slog.SourceKey {
		valueStr := a.Value.String()

		// Cut the source file to a relative path.
		v := strings.Split(valueStr, "/")
		idx := len(v) - 2
		if idx < 0 {
			idx = 0
		}
		valueStr = strings.Join(v[idx:], "/")

		// Remove any curly braces from the source file. This is needed for the logstash parser.
		valueStr = strings.ReplaceAll(valueStr, "{", "")
		valueStr = strings.ReplaceAll(valueStr, "}", "")

		// Is the binary compiled with Bazel? We need to trim the path even more than that.
		// Example of source file at this point:
		// "gazelle~~go_deps~com_github_jacobbrewer1_web/app.go 439" need to be trimmed to web/app.go
		if strings.Contains(valueStr, bazelGazellePrefix) {
			v := strings.Split(valueStr, bazelGazellePrefix)
			if len(v) > 1 {
				// Extract everything after the gazelle~~go_deps~ prefix
				path := v[1]

				// Handle the repository name format (com_github_jacobbrewer1_web -> web)
				parts := strings.Split(path, "/")
				if len(parts) > 0 {
					repoNameParts := strings.Split(parts[0], "_")
					if len(repoNameParts) > 0 {
						// Take the last part of the repo name
						shortName := repoNameParts[len(repoNameParts)-1]
						// Reconstruct the path with the short name
						newPath := shortName
						if len(parts) > 1 {
							newPath += "/" + strings.Join(parts[1:], "/")
						}
						valueStr = newPath
					}
				}
			}
		}

		a.Value = slog.StringValue(valueStr)
	}

	return a
}
