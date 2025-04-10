package health

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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

// IsValid checks if the Status is valid.
func (s Status) IsValid() bool {
	switch s {
	case StatusUp, StatusDown, StatusDegraded, StatusUnknown:
		return true
	}
	return false
}

// String returns the string representation of the Status.
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

// MarshalJSON marshals the Status to JSON as a string.
func (s Status) MarshalJSON() ([]byte, error) {
	if !s.IsValid() {
		return nil, fmt.Errorf("%s is not a valid status", s)
	}

	buf := bytes.NewBuffer(nil)
	if err := json.NewEncoder(buf).Encode(s.String()); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// StandardStatusListener is a standard implementation of the StatusListenerFunc that logs the
// status change to the provided logger.
//
// Note: This is not implemented into the health check itself and you will need to
// implement this yourself if you want to use it. This can be done by using the
// WithCheckOnStatusChange option when creating a new check.
//
// Example:
//
//	health.NewCheck("example", func(ctx context.Context) error {
//		return nil
//	},
//		health.WithCheckOnStatusChange(health.StandardStatusListener(logging.LoggerWithComponent(l, "health-check"))),
//	)
func StandardStatusListener(l *slog.Logger) StatusListenerFunc {
	return func(ctx context.Context, name string, state State) {
		l.Info("health check status changed",
			slog.String(logging.KeyName, name),
			slog.String(logging.KeyState, state.Status().String()),
		)
	}
}
