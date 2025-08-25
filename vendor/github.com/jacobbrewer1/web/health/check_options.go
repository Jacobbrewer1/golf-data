package health

import (
	"time"
)

// CheckOption is a function that modifies the configuration of a Check instance.
// It allows customization of various parameters such as timeout, error grace period,
// maximum failures, and status change listeners.
type CheckOption = func(*Check)

// WithCheckTimeout sets the timeout duration for the health check.
func WithCheckTimeout(timeout time.Duration) CheckOption {
	return func(c *Check) {
		c.timeout = timeout
	}
}

// WithNoCheckTimeout disables the timeout for the health check.
func WithNoCheckTimeout() CheckOption {
	return WithCheckTimeout(0)
}

// WithCheckErrorGracePeriod sets the maximum duration a check can remain in an error state before it is marked as down.
func WithCheckErrorGracePeriod(errorGracePeriod time.Duration) CheckOption {
	return func(c *Check) {
		c.errorGracePeriod = errorGracePeriod
	}
}

// WithCheckMaxFailures sets the maximum number of consecutive failures allowed before the check is marked as down.
func WithCheckMaxFailures(maxContiguousFails uint32) CheckOption {
	return func(c *Check) {
		c.maxContiguousFails = maxContiguousFails
	}
}

// WithCheckOnStatusChange sets a listener function to be called whenever the status
// of the health check changes.
func WithCheckOnStatusChange(statusListener StatusListenerFunc) CheckOption {
	return func(c *Check) {
		c.statusListener = statusListener
	}
}
