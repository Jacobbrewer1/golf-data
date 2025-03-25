package goredis

const (
	NetworkTCP = "tcp"

	loggingKeyError = "err"

	viperPrefix      = "redis"
	viperMaxIdle     = viperPrefix + ".max_idle"
	viperMaxActive   = viperPrefix + ".max_active"
	viperIdleTimeout = viperPrefix + ".idle_timeout_secs"
	viperAddress     = viperPrefix + ".address"
	viperDatabase    = viperPrefix + ".db"
	viperUsername    = viperPrefix + ".username"
	viperPassword    = viperPrefix + ".password"
)
