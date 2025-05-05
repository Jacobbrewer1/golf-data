package health

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/jacobbrewer1/web/logging"
)

// Status represents the health status of a service.
//
// This type is used to define various states of a service's health, such as
// whether it is operational, degraded, or down.
type Status int

const (
	// StatusDown indicates that the service is unhealthy.
	//
	// This status is used when the service is not operational and cannot
	// handle requests.
	StatusDown Status = iota

	// StatusDegraded indicates that the service is degraded.
	//
	// This status is used when the service is still operational but not
	// performing optimally. For example, a service may be considered degraded
	// if it is under high load but still responding to requests. This is
	// particularly useful in clustered environments where some nodes may be
	// degraded, but the overall service remains operational.
	StatusDegraded

	// StatusUp indicates that the service is healthy.
	//
	// This status is used when the service is fully operational and performing
	// as expected.
	StatusUp

	// StatusUnknown indicates that the service status is unknown.
	//
	// This status is used when the health of the service cannot be determined.
	StatusUnknown
)

// IsValid checks if the Status is valid.
//
// This method verifies whether the Status value is one of the predefined valid
// statuses: StatusUp, StatusDown, StatusDegraded, or StatusUnknown.
//
// Returns:
//   - bool: true if the Status is valid, false otherwise.
func (s Status) IsValid() bool {
	switch s {
	case StatusUp, StatusDown, StatusDegraded, StatusUnknown:
		return true
	}
	return false
}

// String returns the string representation of the Status.
//
// This method converts the Status value to its corresponding string
// representation. If the Status is invalid, it returns "invalid".
//
// Returns:
//   - string: The string representation of the Status.
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
//
// This method encodes the Status value as a JSON string. If the Status is
// invalid, it returns an error.
//
// Returns:
//   - []byte: The JSON-encoded string representation of the Status.
//   - error: An error if the Status is invalid or if encoding fails.
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

// StandardStatusListener creates a StatusListenerFunc that logs status changes.
//
// This function returns a StatusListenerFunc, which logs the health check's status changes
// using the provided logger. It is a standard implementation for monitoring and logging
// health check status transitions.
//
// Parameters:
//   - l (*slog.Logger): The logger used to log status changes.
//
// Returns:
//   - StatusListenerFunc: A function that logs the health check's status changes.
//
// Example usage:
//
//	health.NewCheck("example", func(ctx context.Context) error {
//		return nil
//	},
//		health.WithCheckOnStatusChange(health.StandardStatusListener(logging.LoggerWithComponent(l, "health-check"))),
//	)
func StandardStatusListener(l *slog.Logger) StatusListenerFunc {
	return func(ctx context.Context, name string, state *State) {
		l.Info("health check status changed",
			slog.String(logging.KeyName, name),
			slog.String(logging.KeyState, state.Status().String()),
		)
	}
}
