package health

import "fmt"

// CheckerOption defines a function type that modifies a Checker instance.
// It returns an error if the modification fails.
type CheckerOption func(*Checker) error

// WithCheckerCheck adds a single check to the Checker.
//
// Parameters:
//   - check: A pointer to the Check struct to be added.
//
// Behavior:
//   - Calls the AddCheck method of the Checker to register the provided check.
//   - Wraps and returns any error encountered during the addition process.
//
// Returns:
//   - CheckerOption: A function that applies the check addition to a Checker.
func WithCheckerCheck(check *Check) CheckerOption {
	return func(c *Checker) error {
		if err := c.AddCheck(check); err != nil {
			return fmt.Errorf("failed to add check: %w", err)
		}

		return nil
	}
}

// WithCheckerChecks adds multiple checks to the Checker.
//
// Parameters:
//   - checks: A variadic list of pointers to Check structs to be added.
//
// Behavior:
//   - Iterates over the provided checks and applies WithCheckerCheck for each.
//   - Wraps and returns any error encountered during the addition process.
//
// Returns:
//   - CheckerOption: A function that applies the addition of multiple checks to a Checker.
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
//
// Parameters:
//   - code: An integer representing the HTTP status code to use when the system is up.
//
// Behavior:
//   - Updates the httpStatusCodeUp field of the Checker with the provided code.
//
// Returns:
//   - CheckerOption: A function that sets the HTTP status code for a healthy system.
func WithCheckerHTTPCodeUp(code int) CheckerOption {
	return func(c *Checker) error {
		c.httpStatusCodeUp = code
		return nil
	}
}

// WithCheckerHTTPCodeDown sets the HTTP status code for when the system is unhealthy.
//
// Parameters:
//   - code: An integer representing the HTTP status code to use when the system is down.
//
// Behavior:
//   - Updates the httpStatusCodeDown field of the Checker with the provided code.
//
// Returns:
//   - CheckerOption: A function that sets the HTTP status code for an unhealthy system.
func WithCheckerHTTPCodeDown(code int) CheckerOption {
	return func(c *Checker) error {
		c.httpStatusCodeDown = code
		return nil
	}
}
