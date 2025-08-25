package health

import (
	"fmt"
	"time"
)

// CheckerOption defines a function type that modifies a Checker instance.
type CheckerOption func(*Checker) error

// WithCheckerCheck adds a single check to the Checker.
func WithCheckerCheck(check *Check) CheckerOption {
	return func(c *Checker) error {
		if err := c.AddCheck(check); err != nil {
			return fmt.Errorf("failed to add check: %w", err)
		}

		return nil
	}
}

// WithCheckerChecks adds multiple checks to the Checker.
func WithCheckerChecks(checks ...*Check) CheckerOption {
	return func(c *Checker) error {
		for _, check := range checks {
			if err := WithCheckerCheck(check)(c); err != nil {
				return fmt.Errorf("failed to add check %s: %w", check.String(), err)
			}
		}

		return nil
	}
}

// WithCheckerHTTPCodeUp sets the HTTP status code for when the system is healthy.
func WithCheckerHTTPCodeUp(code int) CheckerOption {
	return func(c *Checker) error {
		c.httpStatusCodeUp = code
		return nil
	}
}

// WithCheckerHTTPCodeDown sets the HTTP status code for when the system is unhealthy.
func WithCheckerHTTPCodeDown(code int) CheckerOption {
	return func(c *Checker) error {
		c.httpStatusCodeDown = code
		return nil
	}
}

// WithCheckerErrorGracePeriod sets the grace period for errors. This will ignore errors from any checks.
func WithCheckerErrorGracePeriod(gracePeriod time.Duration) CheckerOption {
	return func(c *Checker) error {
		c.errorGracePeriod = gracePeriod
		return nil
	}
}
