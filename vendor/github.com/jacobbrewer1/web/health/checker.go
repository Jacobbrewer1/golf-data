package health

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"sync"
)

// Checker is an interface that defines a health checker.
type Checker interface {
	// Handler returns the handler for the check.
	Handler() http.HandlerFunc

	// Check returns the result of the check.
	Check(ctx context.Context) *Result
}

// checker is a struct that implements the Checker interface.
//
// This is a group of checks that can be run in parallel.
type checker struct {
	baseCtx context.Context
	cancel  context.CancelFunc
	mtx     *sync.Mutex
	checks  []*Check

	httpStatusCodeUp   int
	httpStatusCodeDown int
}

// NewChecker creates a new Checker.
func NewChecker(opts ...CheckerOption) Checker {
	c := &checker{
		mtx:                new(sync.Mutex),
		checks:             make([]*Check, 0),
		httpStatusCodeUp:   http.StatusOK,
		httpStatusCodeDown: http.StatusServiceUnavailable,
	}

	for _, opt := range opts {
		opt(c)
	}

	if c.baseCtx == nil {
		c.baseCtx, c.cancel = context.WithCancel(context.Background())
	}

	return c
}

func (c *checker) httpCodeFromStatus(status Status) int {
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
func (c *checker) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		result := c.Check(r.Context())
		httpStatus := c.httpCodeFromStatus(result.Status)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(httpStatus)
		_ = json.NewEncoder(w).Encode(result)
	}
}

// Check returns the result of the check.
func (c *checker) Check(ctx context.Context) *Result {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	result := newResult()

	wg := new(sync.WaitGroup)
	for _, check := range c.checks {
		wg.Add(1)
		go func(check *Check, result *Result) {
			defer wg.Done()
			checkResult := newResult()

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
			checkResult.SetTimestamp(Timestamp())

			result.SetStatus(checkResult.Status)
			result.addDetail(check.String(), *checkResult)
		}(check, result)
	}

	wg.Wait()
	return result
}
