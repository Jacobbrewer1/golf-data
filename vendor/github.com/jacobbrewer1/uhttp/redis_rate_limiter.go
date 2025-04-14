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
	ctx := context.Background()
	redisKey := "rate_limit:" + key

	// Try to set key with value 1 and 1-second TTL if not exists
	setReply, err := redis.String(r.keydb.DoCtx(ctx, "SET", redisKey, 1, "EX", int(r.window.Seconds()), "NX"))
	if err != nil {
		r.log(slog.LevelError, "failed to set rate limit key", slog.String(loggingKeyKey, redisKey), slog.String(loggingKeyError, err.Error()))
		return false
	} else if setReply == "OK" {
		// Key was created — allow request
		return true
	}

	// Key exists, increment count
	reply, err := r.keydb.DoCtx(ctx, "INCR", redisKey)
	if err != nil {
		return false
	}

	count, ok := reply.(int64)
	if !ok {
		return false
	}

	return count <= int64(r.burst)
}
