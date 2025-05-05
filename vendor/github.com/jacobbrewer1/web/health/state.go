package health

import (
	"sync/atomic"
	"time"
)

// State represents the state of a health check.
//
// This struct tracks various timestamps, error information, and the current status
// of a health check. It is designed to provide detailed context about the health
// check's performance and outcomes.
type State struct {
	// lastCheckTime is the timestamp of the most recent health check execution.
	lastCheckTime time.Time

	// lastSuccess is the timestamp of the most recent successful health check.
	lastSuccess time.Time

	// lastFail is the timestamp of the most recent failed health check.
	lastFail time.Time

	// firstFailInCycle is the timestamp of the first failure in the current failure cycle.
	firstFailInCycle time.Time

	// contiguousFails is the count of consecutive failed health checks.
	contiguousFails atomic.Uint32

	// checkErr is the last error encountered during a health check.
	checkErr error

	// status represents the current health status of the check.
	status Status
}

// NewState creates and initializes a new State instance.
//
// This function returns a pointer to a State struct with the following default value:
// - status: Initialized to StatusUnknown, representing an unknown health status.
func NewState() *State {
	return &State{
		status: StatusUnknown,
	}
}

// CheckErr retrieves the last error returned by the health check.
//
// This method provides access to the most recent error encountered during a health check.
// Note: The error is cleared when the check is successful.
//
// Returns:
//   - error: The last error encountered during the health check.
func (s *State) CheckErr() error {
	return s.checkErr
}

// ContiguousFails retrieves the number of consecutive failures.
//
// This method provides the count of consecutive failed health checks.
// Note: The count is reset when the check is successful.
//
// Returns:
//   - uint32: The number of contiguous failures.
func (s *State) ContiguousFails() uint32 {
	return s.contiguousFails.Load()
}

// LastFail retrieves the timestamp of the last failed health check.
//
// This method provides the time of the most recent failure. The value persists
// across successes and failures. If there is a success, this will still reflect
// the last failure.
//
// Returns:
//   - time.Time: The timestamp of the last failure.
func (s *State) LastFail() time.Time {
	return s.lastFail
}

// LastSuccess retrieves the timestamp of the last successful health check.
//
// This method provides the time of the most recent success. The value persists
// across successes and failures. If there is a failure, this will still reflect
// the last success.
//
// Returns:
//   - time.Time: The timestamp of the last success.
func (s *State) LastSuccess() time.Time {
	return s.lastSuccess
}

// LastCheckTime retrieves the timestamp of the last health check execution.
//
// This method provides the time when the health check was last performed.
//
// Returns:
//   - time.Time: The timestamp of the last health check.
func (s *State) LastCheckTime() time.Time {
	return s.lastCheckTime
}

// Status retrieves the current health status of the check.
//
// This method provides the current status of the health check.
//
// Returns:
//   - Status: The current health status.
func (s *State) Status() Status {
	return s.status
}
