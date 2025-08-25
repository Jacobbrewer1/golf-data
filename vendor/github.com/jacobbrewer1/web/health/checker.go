package health

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
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

	// firstFailInCycle is the timestamp of the first failure in the current cycle.
	firstFailInCycle atomic.Value

	// errorGracePeriod is the time.Duration during which errors are tolerated.
	errorGracePeriod time.Duration
}

// NewChecker creates a new instance of the Checker struct.
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

			now := time.Now().UTC()

			checkResult := NewResult()
			checkStatus := StatusUp

			if err := check.Check(ctx); err != nil {
				checkStatus = StatusDown

				firstFail, ok := c.firstFailInCycle.Load().(time.Time)
				if !ok {
					firstFail = time.Time{}
				}
				if firstFail.IsZero() {
					c.firstFailInCycle.Store(now)
					firstFail = now
				}

				if c.errorGracePeriod > 0 && now.Sub(firstFail) <= c.errorGracePeriod {
					checkStatus = StatusUp // Still within grace period
				}

				if statusErr := new(StatusError); errors.As(err, &statusErr) {
					checkStatus = statusErr.Status
				}
				checkResult.Error = err.Error()
			} else {
				c.firstFailInCycle.Store(time.Time{})
			}

			checkResult.SetStatus(checkStatus)
			checkResult.SetTimestamp(now)
			result.SetStatus(checkResult.Status)
			result.addDetail(check.String(), checkResult)
		}(check)

		return true
	})

	wg.Wait()
	return result
}

// AddCheck adds a health check to the Checker.
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
