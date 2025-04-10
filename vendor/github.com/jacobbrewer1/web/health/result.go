package health

import (
	"sync"
	"time"
)

// Result is the result of a health check.
type Result struct {
	mtx *sync.RWMutex

	// Status is the status of the check.
	Status Status `json:"status"`

	// Timestamp is the time the check was performed.
	Timestamp *time.Time `json:"timestamp,omitempty"`

	// Details is the details of the check.
	Details map[string]*Result `json:"details,omitempty"`

	// Error is the error returned by the check.
	Error string `json:"error,omitempty"`
}

// NewResult creates a new Result with initialized and default fields.
func NewResult() *Result {
	return &Result{
		mtx:     new(sync.RWMutex),
		Status:  StatusUnknown,
		Details: make(map[string]*Result),
	}
}

// SetTimestamp sets the timestamp and is thread-safe.
func (r *Result) SetTimestamp(t time.Time) {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	if r.Timestamp == nil {
		r.Timestamp = new(time.Time)
	}

	*r.Timestamp = t
}

// SetStatus sets the status and is thread-safe.
// It only sets the status if it is worse than the current status.
func (r *Result) SetStatus(status Status) {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	// Only set the status if it is worse than the current status.
	if status < r.Status {
		r.Status = status
	}
}

// addDetail adds a detail to the result and is thread-safe.
func (r *Result) addDetail(name string, result *Result) {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	if r.Details == nil {
		r.Details = make(map[string]*Result)
	}

	r.Details[name] = result
}
