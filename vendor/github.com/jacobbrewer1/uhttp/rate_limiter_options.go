package uhttp

import (
	"context"
	"log/slog"
)

type RateLimiterOption = func(*rateLimiter)

// WithContext sets the context for the rate limiter.
func WithContext(ctx context.Context) RateLimiterOption {
	return func(r *rateLimiter) {
		r.ctx = ctx
	}
}

// WithLogger sets the logger for the rate limiter.
func WithLogger(l *slog.Logger) RateLimiterOption {
	return func(r *rateLimiter) {
		r.l = l
	}
}
