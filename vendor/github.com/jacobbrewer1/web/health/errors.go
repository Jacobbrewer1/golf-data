package health

// StatusError represents an error with an associated health status.
type StatusError struct {
	// error is the embedded standard error interface.
	error

	// Status provides additional information about the health status
	// associated with the error, such as whether the system is down or degraded.
	Status Status
}

// NewStatusError creates and returns a new StatusError.
func NewStatusError(err error, status Status) *StatusError {
	return &StatusError{
		error:  err,
		Status: status,
	}
}
