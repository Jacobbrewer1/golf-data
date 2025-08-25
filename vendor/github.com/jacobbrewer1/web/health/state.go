package health

import (
	"sync/atomic"
	"time"
)

// State represents the state of a health check.
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
func NewState() *State {
	return &State{
		status: StatusUnknown,
	}
}

// CheckErr retrieves the last error returned by the health check.
func (s *State) CheckErr() error {
	return s.checkErr
}

// ContiguousFails retrieves the number of consecutive failures.
func (s *State) ContiguousFails() uint32 {
	return s.contiguousFails.Load()
}

// LastFail retrieves the timestamp of the last failed health check.
func (s *State) LastFail() time.Time {
	return s.lastFail
}

// LastSuccess retrieves the timestamp of the last successful health check.
func (s *State) LastSuccess() time.Time {
	return s.lastSuccess
}

// LastCheckTime retrieves the timestamp of the last health check execution.
func (s *State) LastCheckTime() time.Time {
	return s.lastCheckTime
}

// Status retrieves the current health status of the check.
func (s *State) Status() Status {
	return s.status
}
