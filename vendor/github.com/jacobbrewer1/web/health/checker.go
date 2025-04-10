package health

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
)

// Checker is a struct that handles the checking of multiple health checks.
//
// This is a group of checks that can be run in parallel.
type Checker struct {
	checks sync.Map

	httpStatusCodeUp   int
	httpStatusCodeDown int
}

// NewChecker creates a new Checker.
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

// Handler returns the handler for the check.
func (c *Checker) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		result := c.Check(r.Context())
		httpStatus := c.httpCodeFromStatus(result.Status)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(httpStatus)
		_ = json.NewEncoder(w).Encode(result)
	}
}

// Check returns the result of the check.
func (c *Checker) Check(ctx context.Context) *Result {
	if ctx == nil {
		ctx = context.Background()
	}

	result := NewResult()

	wg := new(sync.WaitGroup)
	c.checks.Range(func(key, value any) bool {
		check, ok := value.(*Check)
		if !ok {
			// This "should" never happen, but just in case
			return true
		}

		wg.Add(1)
		go func(check *Check) {
			defer wg.Done()

			checkResult := NewResult()

			checkStatus := StatusUp
			if err := check.Check(ctx); err != nil {
				checkStatus = StatusDown
				statusErr := new(StatusError)
				if errors.As(err, &statusErr) {
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

// AddCheck adds a check to the checker.
func (c *Checker) AddCheck(check *Check) error {
	if check == nil {
		return errors.New("check is nil")
	}

	if check.name == "" {
		return errors.New("check name is empty")
	}

	if _, ok := c.checks.Load(check.String()); ok {
		return fmt.Errorf("check already exists with the same key: %s", check.String())
	}

	c.checks.Store(check.String(), check)

	return nil
}
