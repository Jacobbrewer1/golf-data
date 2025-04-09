package health

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"

	"github.com/jacobbrewer1/web/logging"
)

type Status int

const (
	// StatusDown indicates that the service is unhealthy.
	StatusDown Status = iota

	// StatusDegraded indicates that the service is degraded.
	//
	// This is different to StatusDown in that the service is still operational,
	// but not performing at its best.
	// For example, a service may be degraded if it is running at 80% CPU usage,
	// but still responding to requests.
	// This is useful for services that are running in a cluster, where one or
	// more nodes may be degraded, but the service as a whole is still operational.
	StatusDegraded

	// StatusUp indicates that the service is healthy.
	StatusUp

	// StatusUnknown indicates that the service status is unknown.
	StatusUnknown
)

func (s Status) IsValid() bool {
	switch s {
	case StatusUp, StatusDown, StatusDegraded, StatusUnknown:
		return true
	}
	return false
}

func (s Status) String() string {
	switch s {
	case StatusUp:
		return "up"
	case StatusDown:
		return "down"
	case StatusDegraded:
		return "degraded"
	case StatusUnknown:
		return "unknown"
	default:
		return "invalid"
	}
}

func (s Status) MarshalJSON() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	if err := json.NewEncoder(buf).Encode(s.String()); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func StandardStatusListener(l *slog.Logger) StatusListenerFunc {
	return func(ctx context.Context, name string, state State) {
		l.Info("health check status changed",
			slog.String(logging.KeyName, name),
			slog.String(logging.KeyState, state.Status().String()),
		)
	}
}
