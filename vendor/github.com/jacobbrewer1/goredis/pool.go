package goredis

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/gomodule/redigo/redis"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Latency is the duration of Redis queries.
var Latency = promauto.NewHistogramVec(
	prometheus.HistogramOpts{
		Name: "redis_latency",
		Help: "Duration of Redis queries",
	},
	[]string{"command"},
)

type Pool interface {
	// Do will send a command to the server and returns the received reply on a connection from the pool.
	Do(command string, args ...any) (reply any, err error)

	// DoCtx will send a command to the server with a context and returns the received reply on a connection from the pool.
	DoCtx(ctx context.Context, command string, args ...any) (reply any, err error)

	// Conn returns a redis connection from the pool.
	Conn() redis.Conn
}

// pool represents a redis connection pool.
type pool struct {
	*redis.Pool

	// l is the logger for the pool.
	l *slog.Logger

	// addr is the address of the redis server (host:port).
	addr string

	// network is the network type to use when connecting to the redis server.
	network string

	// dialOpts are the dial options to use when connecting to the redis server.
	dialOpts []redis.DialOption
}

// NewPool returns a new Pool.
func NewPool(connOpts ...PoolOption) (Pool, error) {
	poolConn := &pool{
		Pool: new(redis.Pool),
		l:    slog.Default(),
	}

	for _, opt := range connOpts {
		opt(poolConn)
	}

	switch {
	case poolConn.addr == "":
		return nil, errors.New("no address provided")
	case poolConn.network == "":
		return nil, errors.New("no network provided")
	}

	if poolConn.Dial == nil {
		poolConn.Dial = func() (redis.Conn, error) {
			return redis.Dial(poolConn.network, poolConn.addr, poolConn.dialOpts...)
		}
	}

	return poolConn, nil
}

// Do will send a command to the server and returns the received reply on a connection from the pool.
func (p *pool) Do(command string, args ...any) (reply any, err error) {
	return p.DoCtx(context.Background(), command, args...)
}

// DoCtx will send a command to the server with a context and returns the received reply on a connection from the pool.
func (p *pool) DoCtx(ctx context.Context, command string, args ...any) (reply any, err error) {
	// The context cannot be nil for the redis pool.
	if ctx == nil {
		ctx = context.Background()
	}

	t := prometheus.NewTimer(Latency.WithLabelValues(command))
	defer t.ObserveDuration()

	c, err := p.GetContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting connection from pool: %w", err)
	}

	defer func(c redis.Conn) {
		if err := c.Close(); err != nil {
			p.l.Error("error closing connection", slog.String(loggingKeyError, err.Error()))
		}
	}(c)

	return c.Do(command, args...)
}

// Conn returns a redis connection from the pool.
func (p *pool) Conn() redis.Conn {
	return p.Pool.Get()
}
