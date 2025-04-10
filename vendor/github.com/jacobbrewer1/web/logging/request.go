package logging

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/jacobbrewer1/uhttp"
)

// LoggerFromRequest returns a logger with the request ID from the request.
func LoggerFromRequest(l *slog.Logger, r *http.Request) *slog.Logger {
	if r != nil {
		l = LoggerFromContext(r.Context(), l)
	}
	return l
}

// LoggerFromContext returns a logger with the request ID from the context.
func LoggerFromContext(ctx context.Context, l *slog.Logger) *slog.Logger {
	if ctx != nil {
		l = l.With(
			slog.String(KeyRequestID, uhttp.RequestIDFromContext(ctx)),
		)
	}
	return l
}
