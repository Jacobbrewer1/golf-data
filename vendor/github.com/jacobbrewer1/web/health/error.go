package health

// StatusError represents an error with an associated health status.
//
// This struct is used to encapsulate an error along with a specific health status,
// providing additional context for error handling in the health checking system.
// The embedded error interface allows it to be used as a standard error.
type StatusError struct {
	// error is the embedded standard error interface.
	error

	// Status provides additional information about the health status
	// associated with the error, such as whether the system is down or degraded.
	Status Status
}

// NewStatusError creates and returns a new StatusError.
//
// Parameters:
//   - err: The underlying error to associate with the status.
//   - status: The status to associate with the error.
//
// Returns:
//   - *StatusError: A pointer to the newly created StatusError instance.
func NewStatusError(err error, status Status) *StatusError {
	return &StatusError{
		error:  err,
		Status: status,
	}
}
