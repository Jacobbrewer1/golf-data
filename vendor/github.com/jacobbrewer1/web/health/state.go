package health

import "time"

type State struct {
	// lastCheckTime is the last time the check was performed.
	lastCheckTime time.Time

	// lastSuccess is the last time the check was successful.
	lastSuccess time.Time

	// lastFail is the last time the check failed.
	lastFail time.Time

	// contiguousFails is the number of contiguous failures.
	contiguousFails uint

	// checkErr is the last error returned by the check.
	checkErr error

	// status is the current status of the check.
	status Status
}

func (s *State) CheckErr() error {
	return s.checkErr
}

func (s *State) ContiguousFails() uint {
	return s.contiguousFails
}

func (s *State) LastFail() time.Time {
	return s.lastFail
}

func (s *State) LastSuccess() time.Time {
	return s.lastSuccess
}

func (s *State) LastCheckTime() time.Time {
	return s.lastCheckTime
}

func (s *State) Status() Status {
	return s.status
}
