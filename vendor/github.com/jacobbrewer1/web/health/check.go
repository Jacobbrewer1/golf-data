package health

import (
	"context"
	"errors"
	"time"
)

// CheckFunc is a function that performs the check.
type CheckFunc = func(ctx context.Context) error

// Check is a struct that represents a health check.
type Check struct {
	// name is the name of the check.
	name string

	// check is the function that performs the check.
	check CheckFunc

	// timeout is the timeout for the check.
	timeout time.Duration

	// errorGracePeriod is the maximum time the check can be in an error state.
	errorGracePeriod time.Duration

	// maxContiguousFails is the maximum number of contiguous fails.
	maxContiguousFails uint

	// statusListener is the function that will be called when the status changes.
	statusListener func(ctx context.Context, name string, state State)

	// state is the state of the check.
	state *State
}

// NewCheck creates a new Check. Every check should have a unique name.
//
// You are able to return custom statuses by returning a StatusError from the check function. This way you can perform
// checks that return a status other than up or down. For example, you can return a status of "degraded" if the check
// is partially failing. This is useful for checks that are not binary in nature.
func NewCheck(name string, checkerFunc CheckFunc, options ...CheckOption) *Check {
	c := &Check{
		name:    name,
		check:   checkerFunc,
		timeout: 5 * time.Second,
		state:   &State{status: StatusUnknown},
	}

	for _, option := range options {
		option(c)
	}

	return c
}

// String returns the name of the check.
func (c *Check) String() string {
	return c.name
}

// Check performs the check and updates the state of the check.
func (c *Check) Check(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	now := timestamp()
	c.state.lastCheckTime = now

	var (
		checkCtx context.Context
		cancel   context.CancelFunc
	)
	if c.timeout > 0 {
		checkCtx, cancel = context.WithTimeout(ctx, c.timeout)
	} else {
		checkCtx, cancel = context.WithCancel(ctx)
	}
	defer cancel()

	newStatus := StatusUnknown
	defer func() {
		if c.statusListener != nil && c.state.status != newStatus {
			c.state.status = newStatus // Set the new status before calling the listener
			c.statusListener(checkCtx, c.name, *c.state)
		} else {
			c.state.status = newStatus
		}
	}()

	if err := c.check(checkCtx); err != nil {
		c.state.contiguousFails++
		c.state.checkErr = err
		c.state.lastFail = now

		if c.state.firstFailInCycle.IsZero() {
			c.state.firstFailInCycle = now
		}

		// Determine status based on grace period and contiguous fails
		newStatus = StatusDown
		if c.errorGracePeriod > 0 && now.Sub(c.state.firstFailInCycle) <= c.errorGracePeriod {
			newStatus = StatusUp // Still within grace period
		} else if c.maxContiguousFails > 0 && c.state.contiguousFails < c.maxContiguousFails {
			newStatus = StatusUp // Still within fail threshold
		}

		// Handle custom status errors
		statusErr := new(StatusError)
		if errors.As(err, &statusErr) {
			if !statusErr.Status.IsValid() {
				statusErr.Status = StatusUnknown
			}
			newStatus = statusErr.Status
		}

		return err
	}

	newStatus = StatusUp
	c.state.lastSuccess = now
	c.state.contiguousFails = 0
	c.state.checkErr = nil
	c.state.firstFailInCycle = time.Time{}

	return nil
}
