package health

import (
	"sync"
	"time"
)

// Result represents the outcome of a health check.
//
// This struct encapsulates the status, timestamp, details, and any error
// information related to a health check. It is designed to be thread-safe
// for concurrent access and modification.
type Result struct {
	// mtx ensures thread-safe access to the Result fields.
	mtx *sync.RWMutex

	// Status indicates the current health status of the check.
	Status Status `json:"status"`

	// Timestamp records the time when the health check was performed.
	Timestamp *time.Time `json:"timestamp,omitempty"`

	// Details contains additional information or nested results of the health check.
	Details map[string]*Result `json:"details,omitempty"`

	// Error holds any error message returned by the health check.
	Error string `json:"error,omitempty"`
}

// NewResult creates and initializes a new Result instance.
//
// This function returns a pointer to a Result struct with the following default values:
// - A new read-write mutex for thread-safe operations.
// - Status set to StatusUnknown.
// - An empty map for storing nested health check results.
func NewResult() *Result {
	return &Result{
		mtx:     new(sync.RWMutex),
		Status:  StatusUnknown,
		Details: make(map[string]*Result),
	}
}

// SetTimestamp updates the timestamp of the Result instance in a thread-safe manner.
//
// This method locks the mutex to ensure safe concurrent access, initializes the
// Timestamp field if it is nil, and then sets it to the provided time value.
func (r *Result) SetTimestamp(t time.Time) {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	if r.Timestamp == nil {
		r.Timestamp = new(time.Time)
	}

	*r.Timestamp = t
}

// SetStatus updates the health status of the Result instance in a thread-safe manner.
//
// This method locks the mutex to ensure safe concurrent access and updates the
// Status field only if the provided status is worse (lower) than the current status.
//
// Parameters:
//   - status: The new health status to be set.
func (r *Result) SetStatus(status Status) {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	// Only set the status if it is worse than the current status.
	if status < r.Status {
		r.Status = status
	}
}

// addDetail adds a nested health check result to the Details map in a thread-safe manner.
//
// This method locks the mutex to ensure safe concurrent access, initializes the
// Details map if it is nil, and then adds the provided Result instance under the
// specified name.
//
// Parameters:
//   - name: The key under which the nested result will be stored.
//   - result: The Result instance to be added as a detail.
func (r *Result) addDetail(name string, result *Result) {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	if r.Details == nil {
		r.Details = make(map[string]*Result)
	}

	r.Details[name] = result
}
