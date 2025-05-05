package health

import (
	"context"
	"errors"
	"time"
)

// CheckFunc defines the signature for a health check function.
//
// Parameters:
//   - ctx: A context.Context object that provides a deadline, cancellation signal, or other request-scoped values.
//
// Returns:
//   - An error if the check fails, or nil if the check succeeds.
//
// Notes:
//   - This function is expected to perform the actual health check logic and return an appropriate error
//     to indicate failure or success.
type CheckFunc = func(ctx context.Context) error

// StatusListenerFunc defines the signature for a function that listens for status changes in a health check.
//
// Parameters:
//   - ctx: A context.Context object that provides a deadline, cancellation signal, or other request-scoped values.
//   - name: A string representing the name of the health check whose status has changed.
//   - state: A pointer to the State object representing the current state of the health check.
//
// Notes:
//   - This function is invoked whenever the status of a health check changes.
//   - Implementations of this function can be used to log status changes or trigger other actions.
type StatusListenerFunc = func(ctx context.Context, name string, state *State)

// Check represents a health check with configurable parameters and state tracking.
// It handles running health checks, managing error states, enforcing timeouts,
// and providing notifications when health status changes.
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
	maxContiguousFails uint32

	// statusListener is the function that will be called when the status changes.
	statusListener StatusListenerFunc

	// state is the state of the check.
	state *State
}

// NewCheck creates a new health check with the specified name, check function, and optional configuration.
//
// Parameters:
//   - name: A unique string identifier for the health check.
//   - checkerFunc: A CheckFunc that defines the logic for the health check.
//   - options: A variadic list of CheckOption functions to customize the check's behavior.
//
// Returns:
//   - A pointer to a newly created Check instance.
//
// Notes:
//   - The `name` parameter must be unique to avoid conflicts between health checks.
//   - Custom statuses can be returned by using a StatusError in the `checkerFunc`.
//     For example, a "degraded" status can be returned for partial failures.
//   - Default timeout for the check is set to 5 seconds unless overridden by options.
func NewCheck(name string, checkerFunc CheckFunc, options ...CheckOption) *Check {
	c := &Check{
		name:    name,
		check:   checkerFunc,
		timeout: 5 * time.Second,
		state:   NewState(),
	}

	for _, option := range options {
		option(c)
	}

	return c
}

// String returns the name of the health check.
//
// This method provides a string representation of the health check by
// returning its unique name.
//
// Returns:
//   - A string representing the name of the health check.
func (c *Check) String() string {
	return c.name
}

// Check performs the health check and updates the state of the check.
//
// This method executes the health check function, manages the state transitions,
// and updates the status of the check based on the results.
//
// Parameters:
//   - ctx: A context.Context object that provides a deadline, cancellation signal, or other request-scoped values.
//     If nil, an error is returned.
//
// Behavior:
//   - Records the current timestamp as the last check time.
//   - Executes the health check function within the specified timeout or cancellation context.
//   - Updates the state of the check based on the result of the health check function.
//   - Handles error states, including tracking consecutive failures and applying error grace periods.
//   - Invokes the status listener function if the status changes.
//
// Returns:
//   - An error if the health check fails, or nil if it succeeds.
func (c *Check) Check(ctx context.Context) error {
	if ctx == nil {
		return errors.New("context cannot be nil")
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
			c.statusListener(checkCtx, c.name, c.state)
		} else {
			c.state.status = newStatus
		}
	}()

	if err := c.check(checkCtx); err != nil {
		c.state.contiguousFails.Add(1)
		c.state.checkErr = err
		c.state.lastFail = now

		if c.state.firstFailInCycle.IsZero() {
			c.state.firstFailInCycle = now
		}

		// Determine status based on grace period and contiguous fails
		newStatus = StatusDown
		if c.errorGracePeriod > 0 && now.Sub(c.state.firstFailInCycle) <= c.errorGracePeriod {
			newStatus = StatusUp // Still within grace period
		} else if c.maxContiguousFails > 0 && c.state.contiguousFails.Load() < c.maxContiguousFails {
			newStatus = StatusUp // Still within fail threshold
		}

		if statusErr := new(StatusError); errors.As(err, &statusErr) {
			newStatus = StatusUnknown
			if statusErr.Status.IsValid() {
				newStatus = statusErr.Status
			}
		}

		return err
	}

	newStatus = StatusUp
	c.state.lastSuccess = now
	c.state.contiguousFails.Store(0)
	c.state.checkErr = nil
	c.state.firstFailInCycle = time.Time{}

	return nil
}
