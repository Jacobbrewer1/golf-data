package health

import (
	"context"
	"time"
)

// CheckOption is a function that modifies the check configuration.
type CheckOption = func(*Check)

// StatusListenerFunc is a function that is called when the status of the check changes.
type StatusListenerFunc = func(ctx context.Context, name string, state State)

// WithCheckTimeout sets the timeout for the check.
func WithCheckTimeout(timeout time.Duration) CheckOption {
	return func(c *Check) {
		c.timeout = timeout
	}
}

// WithNoCheckTimeout has the same effect as WithCheckTimeout(0).
// It disables the timeout for the check and allows the check to run indefinitely. This should be used with caution,
// as it can lead to hangs if the check never returns.
func WithNoCheckTimeout() CheckOption {
	return WithCheckTimeout(0)
}

// WithCheckErrorGracePeriod sets the maximum time the check can be in an error state before it is marked as down.
func WithCheckErrorGracePeriod(errorGracePeriod time.Duration) CheckOption {
	return func(c *Check) {
		c.errorGracePeriod = errorGracePeriod
	}
}

// WithCheckMaxFailures sets the maximum number of contiguous fails.
func WithCheckMaxFailures(maxContiguousFails uint) CheckOption {
	return func(c *Check) {
		c.maxContiguousFails = maxContiguousFails
	}
}

// WithCheckOnStatusChange sets the status listener for the check.
func WithCheckOnStatusChange(statusListener StatusListenerFunc) CheckOption {
	return func(c *Check) {
		c.statusListener = statusListener
	}
}
