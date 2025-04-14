package uhttp

import (
	"context"
	"log/slog"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/jacobbrewer1/goredis"
)

// redisRateLimiter is a rate limiter that uses Redis to store the rate limit.
type redisRateLimiter struct {
	keydb  goredis.Pool
	window time.Duration

	*rateLimiter
}

// NewRedisRateLimiter creates a new rate limiter.
func NewRedisRateLimiter(keydb goredis.Pool, rps float64, burst int, opts ...RateLimiterOption) RateLimiter {
	if burst == 0 || float64(burst) < rps {
		burst = int(rps)
	}

	rl := &redisRateLimiter{
		rateLimiter: &rateLimiter{
			rps:   rps,
			burst: burst,
		},
		keydb:  keydb,
		window: time.Second,
	}

	for _, opt := range opts {
		opt(rl.rateLimiter)
	}

	if rl.ctx == nil {
		rl.ctx = context.Background()
	}

	return rl
}

// Allow returns true if the request is allowed.
func (r *redisRateLimiter) Allow(key string) bool {
	key = "rate_limiter:" + key

	count, err := redis.Int64(r.keydb.DoCtx(r.ctx, "INCR", key))
	if err != nil {
		r.log(slog.LevelError, "failed to increment rate limiter", slog.String(loggingKeyKey, key), slog.String(loggingKeyError, err.Error()))
		return false
	}

	if count == 1 {
		_, err = r.keydb.DoCtx(r.ctx, "EXPIRE", key, int(r.window.Seconds()))
		if err != nil {
			r.log(slog.LevelError, "failed to set rate limiter expiration", slog.String(loggingKeyKey, key), slog.String(loggingKeyError, err.Error()))
			return false
		}
	}

	if count > int64(r.burst) {
		return false
	}

	return true
}
