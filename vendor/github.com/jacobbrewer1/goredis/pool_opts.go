package goredis

import (
	"log/slog"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/spf13/viper"
)

type PoolOption func(r *pool)

// WithLogger is a PoolOption that sets the logger for the pool.
func WithLogger(l *slog.Logger) PoolOption {
	return func(r *pool) {
		r.l = l
	}
}

// WithMaxIdle is a PoolOption that sets the maximum number of idle connections in the pool.
func WithMaxIdle(maxIdle int) PoolOption {
	return func(r *pool) {
		r.MaxIdle = maxIdle
	}
}

// WithMaxActive is a PoolOption that sets the maximum number of active connections in the pool.
func WithMaxActive(maxActive int) PoolOption {
	return func(r *pool) {
		r.MaxActive = maxActive
	}
}

// WithIdleTimeout is a PoolOption that sets the idle timeout for connections in the pool.
func WithIdleTimeout(idleTimeout int) PoolOption {
	return func(r *pool) {
		r.IdleTimeout = time.Duration(idleTimeout) * time.Second
	}
}

// WithAddress is a PoolOption that sets the address of the redis server (host:port).
func WithAddress(address string) PoolOption {
	return func(r *pool) {
		r.addr = address
	}
}

// WithNetwork is a PoolOption that sets the network type to use when connecting to the redis server.
func WithNetwork(network string) PoolOption {
	return func(r *pool) {
		r.network = network
	}
}

// WithDialOpts is a PoolOption that sets the dial options to use when connecting to the redis server.
func WithDialOpts(dialOpts ...redis.DialOption) PoolOption {
	return func(r *pool) {
		r.dialOpts = dialOpts
	}
}

// FromViper returns a slice of PoolOptions from a viper instance.s
func FromViper(v *viper.Viper) []PoolOption {
	return []PoolOption{
		WithMaxIdle(v.GetInt(viperMaxIdle)),
		WithMaxActive(v.GetInt(viperMaxActive)),
		WithIdleTimeout(v.GetInt(viperIdleTimeout)),
		WithAddress(v.GetString(viperAddress)),
		WithNetwork(NetworkTCP),
		WithDialOpts(
			redis.DialDatabase(v.GetInt(viperDatabase)),
			redis.DialUsername(v.GetString(viperUsername)),
			redis.DialPassword(v.GetString(viperPassword)),
		),
	}
}
