package health

// StatusError is an error that contains a status with the error.
type StatusError struct {
	error
	Status Status
}

// NewStatusError creates a new StatusError with the given error and status.
func NewStatusError(err error, status Status) *StatusError {
	return &StatusError{
		error:  err,
		Status: status,
	}
}
