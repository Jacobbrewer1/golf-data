package health

import (
	"time"
)

// CheckOption is a function that modifies the configuration of a Check instance.
// It allows customization of various parameters such as timeout, error grace period,
// maximum failures, and status change listeners.
type CheckOption = func(*Check)

// WithCheckTimeout sets the timeout duration for the health check.
//
// Parameters:
//   - timeout: The maximum duration the check is allowed to run before timing out.
//
// Returns:
//   - A CheckOption function that sets the timeout for the check.
func WithCheckTimeout(timeout time.Duration) CheckOption {
	return func(c *Check) {
		c.timeout = timeout
	}
}

// WithNoCheckTimeout disables the timeout for the health check.
//
// Behavior:
//   - This function has the same effect as calling WithCheckTimeout(0).
//   - It allows the check to run indefinitely, which can lead to hangs if the check never returns.
//     Use this option with caution.
//
// Returns:
//   - A CheckOption function that disables the timeout for the check.
func WithNoCheckTimeout() CheckOption {
	return WithCheckTimeout(0)
}

// WithCheckErrorGracePeriod sets the maximum duration a check can remain in an error state
// before it is marked as down.
//
// Parameters:
//   - errorGracePeriod: The duration for which the check can remain in an error state.
//
// Returns:
//   - A CheckOption function that sets the error grace period for the check.
func WithCheckErrorGracePeriod(errorGracePeriod time.Duration) CheckOption {
	return func(c *Check) {
		c.errorGracePeriod = errorGracePeriod
	}
}

// WithCheckMaxFailures sets the maximum number of consecutive failures allowed
// before the check is marked as down.
//
// Parameters:
//   - maxContiguousFails: The maximum number of consecutive failures.
//
// Returns:
//   - A CheckOption function that sets the maximum number of contiguous failures.
func WithCheckMaxFailures(maxContiguousFails uint32) CheckOption {
	return func(c *Check) {
		c.maxContiguousFails = maxContiguousFails
	}
}

// WithCheckOnStatusChange sets a listener function to be called whenever the status
// of the health check changes.
//
// Parameters:
//   - statusListener: A StatusListenerFunc that will be invoked on status changes.
//
// Returns:
//   - A CheckOption function that sets the status listener for the check.
func WithCheckOnStatusChange(statusListener StatusListenerFunc) CheckOption {
	return func(c *Check) {
		c.statusListener = statusListener
	}
}
