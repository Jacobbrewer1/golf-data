package health

import (
	"context"
	"time"
)

type CheckOption = func(*Check)

type StatusListenerFunc = func(ctx context.Context, name string, state State)

// WithCheckTimeout sets the timeout for the check.
func WithCheckTimeout(timeout time.Duration) CheckOption {
	return func(c *Check) {
		c.timeout = timeout
	}
}

// WithCheckErrorGracePeriod sets the maximum time the check can be in an error state before it is marked as down.
func WithCheckErrorGracePeriod(maxTimeInError time.Duration) CheckOption {
	return func(c *Check) {
		c.maxTimeInError = maxTimeInError
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
