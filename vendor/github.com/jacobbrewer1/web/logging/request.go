package logging

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/jacobbrewer1/uhttp"
)

// LoggerFromRequest returns a logger with the request ID from the provided HTTP request.
//
// This function extracts the context from the given HTTP request and uses it to
// add a request ID attribute to the provided logger. If the request is nil, the
// original logger is returned unmodified.
func LoggerFromRequest(l *slog.Logger, r *http.Request) *slog.Logger {
	if r != nil {
		l = LoggerFromContext(r.Context(), l)
	}
	return l
}

// LoggerFromContext returns a logger with the request ID from the provided context.
//
// This function retrieves the request ID from the given context and adds it as an
// attribute to the provided logger. If the context is nil, the original logger is
// returned unmodified.
func LoggerFromContext(ctx context.Context, l *slog.Logger) *slog.Logger {
	if ctx != nil {
		l = l.With(
			slog.String(KeyRequestID, uhttp.RequestIDFromContext(ctx)),
		)
	}
	return l
}
