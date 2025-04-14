package uhttp

import (
	"context"
	"log/slog"
	"sync"

	"golang.org/x/time/rate"
)

type RateLimiter interface {
	// Allow returns true if the request is allowed.
	Allow(key string) bool
}

type rateLimiter struct {
	ctx context.Context
	l   *slog.Logger

	// limiter is the limiter.
	limiters sync.Map

	// rps is the requests per second.
	rps float64

	// burst is the burst.
	burst int
}

// NewRateLimiter creates a new rate limiter.
func NewRateLimiter(rps float64, burst int, opts ...RateLimiterOption) RateLimiter {
	if burst == 0 || float64(burst) < rps {
		burst = int(rps)
	}

	rl := &rateLimiter{
		ctx:   context.Background(),
		rps:   rps,
		burst: burst,
	}

	for _, opt := range opts {
		opt(rl)
	}

	return rl
}

// Allow returns true if the request is allowed.
func (r *rateLimiter) Allow(key string) bool {
	// Rate limits the request.
	gotLimiter, _ := r.limiters.LoadOrStore(key, rate.NewLimiter(rate.Limit(r.rps), r.burst))

	limiter, ok := gotLimiter.(*rate.Limiter)
	if !ok {
		r.log(slog.LevelError, "failed to cast rate limiter", slog.String(loggingKeyKey, key))
		return false
	}

	return limiter.Allow()
}

func (r *rateLimiter) log(level slog.Level, msg string, args ...any) {
	if r.l == nil {
		return
	}

	r.l.Log(context.Background(), level, msg, args...) // nolint:sloglint // Handler around the messages passed in to prevent panics on nil logger
}
