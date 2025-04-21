package health

import "time"

// State is the state of the health check.
type State struct {
	// lastCheckTime is the last time the check was performed.
	lastCheckTime time.Time

	// lastSuccess is the last time the check was successful.
	lastSuccess time.Time

	// lastFail is the last time the check failed.
	lastFail time.Time

	// firstFailInCycle is the first time the check failed in the current cycle.
	firstFailInCycle time.Time

	// contiguousFails is the number of contiguous failures.
	contiguousFails uint

	// checkErr is the last error returned by the check.
	checkErr error

	// status is the current status of the check.
	status Status
}

// CheckErr is the last error returned by the check.
//
// Note: This is cleared when the check is successful.
func (s *State) CheckErr() error {
	return s.checkErr
}

// ContiguousFails is the number of contiguous failures.
//
// Note: This is cleared when the check is successful.
func (s *State) ContiguousFails() uint {
	return s.contiguousFails
}

// LastFail is the last time the check failed.
//
// Note: This is persistent across successes and failures. If there is a success, this
// will be the last time the check failed.
func (s *State) LastFail() time.Time {
	return s.lastFail
}

// LastSuccess is the last time the check was successful.
//
// Note: This is persistent across successes and failures. If there is a failure, this
// will be the last time the check was successful.
func (s *State) LastSuccess() time.Time {
	return s.lastSuccess
}

// LastCheckTime is the last time the check was performed.
func (s *State) LastCheckTime() time.Time {
	return s.lastCheckTime
}

// Status is the current status of the check.
func (s *State) Status() Status {
	return s.status
}
