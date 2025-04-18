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
		// Cut the source file to a relative path.
		v := strings.Split(a.Value.String(), "/")
		idx := len(v) - 2
		if idx < 0 {
			idx = 0
		}
		a.Value = slog.StringValue(strings.Join(v[idx:], "/"))

		// Remove any curly braces from the source file. This is needed for the logstash parser.
		a.Value = slog.StringValue(strings.ReplaceAll(a.Value.String(), "{", ""))
		a.Value = slog.StringValue(strings.ReplaceAll(a.Value.String(), "}", ""))
	}

	return a
}
