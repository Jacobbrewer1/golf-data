package health

import (
	"context"
)

type CheckerOption func(*checker)

// WithCheckerBaseContext sets the base context for the checker.
func WithCheckerBaseContext(baseCtx context.Context) CheckerOption {
	return func(c *checker) {
		ctx, cancel := context.WithCancel(baseCtx)
		c.baseCtx = ctx
		c.cancel = cancel
	}
}

// WithCheckerCheck adds a single check to the checker.
func WithCheckerCheck(check *Check) CheckerOption {
	return func(c *checker) {
		if check == nil {
			return
		}
		if c.checks == nil {
			c.checks = make([]*Check, 0)
		}
		c.checks = append(c.checks, check)
	}
}

// WithCheckerChecks adds multiple checks to the checker.
func WithCheckerChecks(checks ...*Check) CheckerOption {
	return func(c *checker) {
		for _, check := range checks {
			WithCheckerCheck(check)(c)
		}
	}
}

// WithCheckerHTTPCodeUp sets the HTTP status code when the system is up.
func WithCheckerHTTPCodeUp(code int) CheckerOption {
	return func(c *checker) {
		c.httpStatusCodeUp = code
	}
}

// WithCheckerHTTPCodeDown sets the HTTP status code when the system is down.
func WithCheckerHTTPCodeDown(code int) CheckerOption {
	return func(c *checker) {
		c.httpStatusCodeDown = code
	}
}
