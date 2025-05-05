package health

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
)

// Checker is a struct that manages multiple health checks.
//
// This struct allows grouping and running health checks in parallel.
// It also provides functionality to determine HTTP status codes based on the
// overall health status of the checks.
type Checker struct {
	// checks is a thread-safe map that stores the health checks.
	checks sync.Map

	// httpStatusCodeUp is the HTTP status code returned when all checks are healthy.
	httpStatusCodeUp int

	// httpStatusCodeDown is the HTTP status code returned when one or more checks fail.
	httpStatusCodeDown int
}

// NewChecker creates a new instance of the Checker struct.
//
// This function initializes a Checker with default HTTP status codes for
// healthy and unhealthy states. It also applies any provided options to
// customize the Checker instance.
//
// Parameters:
//   - opts: A variadic list of CheckerOption functions used to configure the Checker.
//
// Returns:
//   - *Checker: A pointer to the newly created Checker instance.
//   - error: An error if any of the provided options fail to apply.
func NewChecker(opts ...CheckerOption) (*Checker, error) {
	c := &Checker{
		httpStatusCodeUp:   http.StatusOK,
		httpStatusCodeDown: http.StatusServiceUnavailable,
	}

	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, fmt.Errorf("failed to apply checker option: %w", err)
		}
	}

	return c, nil
}

// httpCodeFromStatus maps a given health check status to the corresponding HTTP status code.
//
// Parameters:
//   - status: The current health check status of type Status.
//
// Returns:
//   - int: The HTTP status code corresponding to the provided health check status.
//   - If the status is StatusUp, it returns the value of httpStatusCodeUp.
//   - If the status is StatusDown, StatusUnknown, or StatusDegraded, it returns the value of httpStatusCodeDown.
//   - For any other status, it defaults to http.StatusInternalServerError.
func (c *Checker) httpCodeFromStatus(status Status) int {
	switch status {
	case StatusUp:
		return c.httpStatusCodeUp
	case StatusDown, StatusUnknown, StatusDegraded:
		return c.httpStatusCodeDown
	default:
		return http.StatusInternalServerError
	}
}

// Handler returns an HTTP handler function for the Checker.
//
// This handler executes the health checks, determines the appropriate HTTP status code
// based on the overall health status, and writes the result as a JSON response.
//
// Behavior:
//   - Calls the Check method to execute all registered health checks.
//   - Maps the resulting health status to an HTTP status code using httpCodeFromStatus.
//   - Sets the "Content-Type" header to "application/json; charset=utf-8".
//   - Writes the HTTP status code and encodes the health check result as a JSON response.
//
// Returns:
//   - http.HandlerFunc: A function that handles HTTP requests and writes the health check result.
func (c *Checker) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		result := c.Check(r.Context())
		httpStatus := c.httpCodeFromStatus(result.Status)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(httpStatus)
		_ = json.NewEncoder(w).Encode(result)
	}
}

// Check executes all registered health checks and aggregates their results.
//
// This method iterates over all the health checks stored in the Checker, runs them
// concurrently, and collects their results into a single aggregated Result object.
//
// Parameters:
//   - ctx: The context for the health checks. If nil, a new background context is created.
//
// Behavior:
//   - If the context is nil, it defaults to context.Background().
//   - Creates a new Result object to store the aggregated results.
//   - Iterates over all registered checks using a thread-safe map.
//   - Runs each check in a separate goroutine, collecting its status and any errors.
//   - Aggregates the status and details of each check into the final Result object.
//
// Returns:
//   - *Result: An aggregated result containing the status and details of all checks.
func (c *Checker) Check(ctx context.Context) *Result {
	if ctx == nil {
		ctx = context.Background()
	}

	result := NewResult()
	wg := new(sync.WaitGroup)

	c.checks.Range(func(_, value any) bool {
		check, ok := value.(*Check)
		if !ok {
			return true
		}

		wg.Add(1)
		go func(check *Check) {
			defer wg.Done()

			checkResult := NewResult()
			checkStatus := StatusUp

			if err := check.Check(ctx); err != nil {
				checkStatus = StatusDown
				if statusErr := new(StatusError); errors.As(err, &statusErr) {
					checkStatus = statusErr.Status
				}
				checkResult.Error = err.Error()
			}

			checkResult.SetStatus(checkStatus)
			checkResult.SetTimestamp(timestamp())
			result.SetStatus(checkResult.Status)
			result.addDetail(check.String(), checkResult)
		}(check)

		return true
	})

	wg.Wait()
	return result
}

// AddCheck adds a health check to the Checker.
//
// This method registers a new health check in the Checker. Each check must have
// a unique name and cannot be nil.
//
// Parameters:
//   - check: A pointer to the Check struct representing the health check to be added.
//
// Behavior:
//   - Returns an error if the check is nil.
//   - Returns an error if the check's name is empty.
//   - Returns an error if a check with the same name already exists in the Checker.
//   - Stores the check in the thread-safe map of checks if all validations pass.
//
// Returns:
//   - error: An error if the check is nil, has an empty name, or already exists.
func (c *Checker) AddCheck(check *Check) error {
	if check == nil {
		return errors.New("check is nil")
	}

	if check.name == "" {
		return errors.New("check name is empty")
	}

	if _, exists := c.checks.Load(check.String()); exists {
		return fmt.Errorf("check already exists: %s", check.String())
	}

	c.checks.Store(check.String(), check)
	return nil
}
