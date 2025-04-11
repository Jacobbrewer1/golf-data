package uhttp

import "log/slog"

type RateLimiterOption = func(*rateLimiter)

// WithLogger sets the logger for the rate limiter.
func WithLogger(l *slog.Logger) RateLimiterOption {
	return func(r *rateLimiter) {
		r.l = l
	}
}
